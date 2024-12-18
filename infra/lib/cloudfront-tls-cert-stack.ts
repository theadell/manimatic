import * as cdk from 'aws-cdk-lib';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import * as route53 from 'aws-cdk-lib/aws-route53';
import { Construct } from 'constructs';

export class CloudfrontTlsCertStack extends cdk.Stack {
  public readonly certificate: acm.Certificate;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const hostedZone = route53.HostedZone.fromLookup(this, 'HostedZone', {
      domainName: 'aws.adelh.dev',
    });

    this.certificate = new acm.Certificate(this, 'TLSCertificate', {
      domainName: 'manimatic.aws.adelh.dev',
      validation: acm.CertificateValidation.fromDns(hostedZone),
    });
  }
}
