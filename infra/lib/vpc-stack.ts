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


        this.vpc.addGatewayEndpoint('S3Endpoint', {
            service: ec2.GatewayVpcEndpointAwsService.S3,
            subnets: [
                { subnetType: ec2.SubnetType.PRIVATE_ISOLATED }
            ]
        })

        new cdk.CfnOutput(this, 'VpcId', {
            value: this.vpc.vpcId,
            description: 'The ID of the VPC',
        });


    }
}