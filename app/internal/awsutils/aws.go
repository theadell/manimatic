package awsutils

import (
	"manimatic/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	localStackEndpoint = "http://localhost:4566"
	localStackRegion   = "eu-central-1"
)

func NewS3Client(cfg config.Config, awsCfg aws.Config) *s3.Client {
	var opts func(*s3.Options)
	if cfg.EnvLocal {
		opts = func(o *s3.Options) {
			o.BaseEndpoint = aws.String(localStackEndpoint)
			o.Region = *aws.String(localStackRegion)
			o.UsePathStyle = true
		}
	} else {
		opts = func(o *s3.Options) {}
	}
	return s3.NewFromConfig(awsCfg, opts)
}

func NewSQSClient(cfg config.Config, awsCfg aws.Config) *sqs.Client {
	var opts func(*sqs.Options)
	if cfg.EnvLocal {
		opts = func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(localStackEndpoint)
			o.Region = *aws.String(localStackRegion)
		}
	} else {
		opts = func(o *sqs.Options) {}
	}
	return sqs.NewFromConfig(awsCfg, opts)
}
