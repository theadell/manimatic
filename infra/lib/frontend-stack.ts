import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins';
import * as s3deployment from 'aws-cdk-lib/aws-s3-deployment';
import * as cm from 'aws-cdk-lib/aws-certificatemanager';
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as route53targets from 'aws-cdk-lib/aws-route53-targets';

import * as path from 'path';

import { Construct } from 'constructs';

interface FrontendStackProps extends cdk.StackProps {
  certificate: cm.Certificate;
}


export class FrontendStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: FrontendStackProps) {
    super(scope, id, props);

    const s3Bucket = new s3.Bucket(this, 'manimatic-website-bucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
      versioned: true,
    })

    const oac = new cloudfront.S3OriginAccessControl(this, 'oac', {
      originAccessControlName: 'manimatic-website-oac',
      signing: {
        protocol: cloudfront.SigningProtocol.SIGV4,
        behavior: cloudfront.SigningBehavior.ALWAYS,
      }
    })
    const s3Origin = origins.S3BucketOrigin.withOriginAccessControl(s3Bucket, {
      originAccessControl: oac
    }
    )

    const siteDomainName = "manimatic.aws.adelh.dev"

    const distribution = new cloudfront.Distribution(this, 'WebsiteDistribution', {
      defaultRootObject: 'index.html',
      domainNames: [siteDomainName],
      certificate: props.certificate,
      defaultBehavior: {
        origin: s3Origin,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
      },
      
    })

    const hostedZone = route53.HostedZone.fromLookup(this, 'HostedZone', {
      domainName: 'aws.adelh.dev',
    });

    new route53.ARecord(this, 'SiteAliasRecord', {
      zone: hostedZone,
      recordName: 'manimatic',
      target: route53.RecordTarget.fromAlias(new route53targets.CloudFrontTarget(distribution)),
    });

    new s3deployment.BucketDeployment(this, 'DeployStaticSite', {
      sources: [s3deployment.Source.asset(path.join(__dirname, '../../frontend/dist'))],
      destinationBucket: s3Bucket,
      distribution, // Invalidate cache in CloudFront
      distributionPaths: ['/*'], // Cache invalidation paths
    });


    new cdk.CfnOutput(this, 'BucketName', {
      value: s3Bucket.bucketName,
      description: 'The name of the S3 bucket hosting the static site',
    });
    
    new cdk.CfnOutput(this, 'WebsiteURL', {
      value: `https://${distribution.domainName}`,
      description: 'The URL of the deployed website',
    });    
    new cdk.CfnOutput(this, 'Custom-WebsiteURL', {
      value: `https://${siteDomainName}`,
      description: 'The Custom URL of the deployed website',
    });

  }
}
