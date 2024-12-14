
import * as cdk from "aws-cdk-lib"
import * as ec2 from "aws-cdk-lib/aws-ec2"
import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as route53Targets from 'aws-cdk-lib/aws-route53-targets';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as elbv2_targets from 'aws-cdk-lib/aws-elasticloadbalancingv2-targets';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';

import { Construct } from "constructs";

interface AppStackProps extends cdk.StackProps {
    vpc: ec2.Vpc
}


export class AppStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props: AppStackProps) {
        super(scope, id, props)

        const securityGroup = new ec2.SecurityGroup(this, 'apiEC2InstanceSG', {
            vpc: props.vpc,
            allowAllOutbound: true,
        })

        securityGroup.addIngressRule(
            ec2.Peer.anyIpv4(),
            ec2.Port.tcp(22),
            'Allow SSH Access',
        )

        securityGroup.addIngressRule(
            ec2.Peer.anyIpv4(),
            ec2.Port.tcp(80),
            'Allow HTTP web traffic'
        );

        const keyPair = ec2.KeyPair.fromKeyPairName(this, 'ec2-keyPair', 'dev-mini-01')

        const apiInstance = new ec2.Instance(this, 'apiEC2Instance', {
            vpc: props.vpc,
            vpcSubnets: {
                subnetType: ec2.SubnetType.PUBLIC
            },
            instanceType: ec2.InstanceType.of(
                ec2.InstanceClass.T2,
                ec2.InstanceSize.MICRO,
            ),
            machineImage: ec2.MachineImage.latestAmazonLinux2023(),
            securityGroup: securityGroup,
            keyPair: keyPair
        })

        const workerSecurityGroup = new ec2.SecurityGroup(this, 'workerEC2InstanceSG', {
            vpc: props.vpc,
            allowAllOutbound: true,
        })

        workerSecurityGroup.addIngressRule(
            ec2.Peer.anyIpv4(),
            ec2.Port.tcp(22),
            'Allow SSH Access',
        )


        const workerInstance = new ec2.Instance(this, 'workerEC2Instance', {
            vpc: props.vpc,
            vpcSubnets: {
                subnetType: ec2.SubnetType.PRIVATE_ISOLATED
            },
            instanceType: ec2.InstanceType.of(
                ec2.InstanceClass.T2,
                ec2.InstanceSize.MICRO,
            ),
            machineImage: ec2.MachineImage.latestAmazonLinux2023(),
            securityGroup: workerSecurityGroup,
            keyPair: keyPair
        })

        const manimaticDLQ = new sqs.Queue(this, 'Queue-DLQ', {
            queueName: 'manimatic-DLQ',
        })

        const taskQueue = new sqs.Queue(this, 'Queue-Tasks', {
            queueName: 'manimatic-tasks',
            deadLetterQueue: {
                maxReceiveCount: 2,
                queue: manimaticDLQ
            },
        })

        const animationQueue = new sqs.Queue(this, 'Queue-Animations', {
            queueName: 'manimatic-animations',
            deadLetterQueue: {
                maxReceiveCount: 2,
                queue: manimaticDLQ
            },
        })

        taskQueue.grantSendMessages(apiInstance)
        taskQueue.grantConsumeMessages(workerInstance)

        animationQueue.grantSendMessages(workerInstance)
        animationQueue.grantConsumeMessages(apiInstance)


        const zone = route53.HostedZone.fromLookup(this, 'devHostedZone', {
            domainName: 'dev.pluseinslab.com'
        })

        const certificate = new acm.Certificate(this, 'MyCertificate', {
            domainName: 'api.manimatic.dev.pluseinslab.com',
            validation: acm.CertificateValidation.fromDns(zone),
        });

        // ALB 
        const albSecurityGroup = new ec2.SecurityGroup(this, 'ALBSecurityGroup', {
            vpc: props.vpc,
            allowAllOutbound: true,
            description: 'Security group for Application Load Balancer'
        });

        // Allow HTTP and HTTPS inbound traffic from anywhere
        albSecurityGroup.addIngressRule(
            ec2.Peer.anyIpv4(),
            ec2.Port.tcp(80),
            'Allow HTTP traffic'
        );
        albSecurityGroup.addIngressRule(
            ec2.Peer.anyIpv4(),
            ec2.Port.tcp(443),
            'Allow HTTPS traffic'
        )

        const alb = new elbv2.ApplicationLoadBalancer(this, 'APILoadBalancer', {
            vpc: props.vpc,
            internetFacing: true,
            securityGroup: albSecurityGroup
        })


        const targetGroup = new elbv2.ApplicationTargetGroup(this, 'APITargetGroup', {
            vpc: props.vpc,
            port: 80,
            targetType: elbv2.TargetType.INSTANCE,
            healthCheck: {
                path: '/healthz',
                healthyHttpCodes: '200',
            },
        })

        targetGroup.addTarget(
            new elbv2_targets.InstanceTarget(apiInstance, 80)
        )

        const httpsListener = alb.addListener('HTTPSListener', {
            port: 443,
            certificates: [certificate],
            open: true
        })

        httpsListener.addTargetGroups('DefaultTargetGroup', {
            targetGroups: [targetGroup]
        })

        alb.addListener('HTTPRedirectListener', {
            port: 80,
            open: true,
        }).addAction('RedirectToHTTPS', {
            action: elbv2.ListenerAction.redirect({
                port: '443',
                protocol: elbv2.ApplicationProtocol.HTTPS
            })
        })

        const albR53Target = new route53Targets.LoadBalancerTarget(alb)

        new route53.ARecord(this, 'APILoadBalancerDNSRecord', {
            zone: zone,
            recordName: 'api.manimatic.dev.pluseinslab.com',
            target: route53.RecordTarget.fromAlias(albR53Target)
        });


        new cdk.CfnOutput(this, 'instancePublicIp', {
            value: apiInstance.instancePublicIp
        })

        new cdk.CfnOutput(this, 'Tasks Queue URL', {
            value: taskQueue.queueUrl,
            description: 'The URL of the Tasks SQS queue'
        })
        new cdk.CfnOutput(this, 'Animations Queue URL', {
            value: animationQueue.queueUrl,
            description: 'The URL of the Animations SQS queue'
        })

    }

}