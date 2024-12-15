#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { VpcStack } from '../lib/vpc-stack';
import { AppStack } from '../lib/app-stack';

const app = new cdk.App();

const vpcStack = new VpcStack(app, 'VpcStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

new AppStack(app, 'AppStack', {
  vpc: vpcStack.vpc,
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

// const certStack = new CertStack(app, 'certStack', {
//   crossRegionReferences: true, 
//   env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' },
// })

/*
new InfraStack(app, 'InfraStack', {
  certificate: certStack.certificate,
  crossRegionReferences: true,
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },

 // For more information, see https://docs.aws.amazon.com/cdk/latest/guide/environments.html 
});
*/