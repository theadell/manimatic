#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { VpcStack } from '../lib/vpc-stack';
import { AppStack } from '../lib/app-stack';
import { CloudfrontTlsCertStack } from '../lib/cloudfront-tls-cert-stack';
import { FrontendStack } from '../lib/frontend-stack';

const app = new cdk.App();

const vpcStack = new VpcStack(app, 'VpcStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

const appStack = new AppStack(app, 'AppStack', {
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

const cloudfrontTlsCertStack = new CloudfrontTlsCertStack(app, 'CloudfrontTlsCertStack', {
  crossRegionReferences: true,
  // To use the certificate in cloudfront it must be created in us-east-1
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: 'us-east-1' }
})


new FrontendStack(app, 'FrontendStack', {
  certificate: cloudfrontTlsCertStack.certificate,
  crossRegionReferences: true,
  env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },

});
