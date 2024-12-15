import * as cdk from "aws-cdk-lib"
import { Construct } from "constructs";
import * as ec2 from 'aws-cdk-lib/aws-ec2'

export class VpcStack extends cdk.Stack {

    public readonly vpc: ec2.Vpc

    constructor(scope: Construct, id: string, props: cdk.StackProps) {
        super(scope, id, props)
        this.vpc = new ec2.Vpc(this, 'm_vpc', {
            ipAddresses: ec2.IpAddresses.cidr('10.16.0.0/16'),
            maxAzs: 3,
            reservedAzs: 1,
            subnetConfiguration: [
                {
                    subnetType: ec2.SubnetType.PUBLIC,
                    name: 'Public',
                },
                {
                    // TODO: Use fck-nat 
                    subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
                    name: 'Worker',
                },
                {
                    subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
                    name: 'Private',
                },
                {
                    subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
                    name: 'Reserved',
                },
            ],

        })

        const sqsEndpointsSG = new ec2.SecurityGroup(this, 'sqsEndpointSG', {
            vpc: this.vpc,
            description: 'Security group for SQS VPC Endpoint',
            allowAllOutbound: true,
        })

        // Workaround to avoid cyclic depdency with the AppStack
        sqsEndpointsSG.addIngressRule(ec2.Peer.ipv4(this.vpc.vpcCidrBlock), ec2.Port.tcp(443), 'Allow VPC CIDR to access SQS endpoint');


        const sqsEndpoint = this.vpc.addInterfaceEndpoint('SQSVpcEndpoint', {
            service: ec2.InterfaceVpcEndpointAwsService.SQS,
            securityGroups: [sqsEndpointsSG]
        })


        this.vpc.publicSubnets.forEach((subnet, index) => {
            new cdk.CfnOutput(this, `PublicSubnet${index}`, {
                value: `Subnet ID: ${subnet.subnetId}, CIDR Block: ${subnet.ipv4CidrBlock || 'No CIDR Block'}`,
                description: `Public Subnet ${index}`,
            })
        })

        this.vpc.isolatedSubnets.forEach((subnet, index) => {
            new cdk.CfnOutput(this, `IsolatedSubnet${index}`, {
                value: `Subnet ID: ${subnet.subnetId}, CIDR Block: ${subnet.ipv4CidrBlock || 'No CIDR Block'}`,
                description: `Isolated Subnet ${index}`,
            })
        })

        this.vpc.privateSubnets.forEach((subnet, index) => {
            new cdk.CfnOutput(this, `PrivateSubnet${index}`, {
                value: `Subnet ID: ${subnet.subnetId}, CIDR Block: ${subnet.ipv4CidrBlock || 'No CIDR Block'}`,
                description: `Private Subnet ${index}`,
            })
        })

        new cdk.CfnOutput(this, 'SqsVpcEndpointId', {
            value: sqsEndpoint.vpcEndpointId,
          })

    }
}