package awsutils

import (
	"context"
	"manimatic/internal/config"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	awsDefaultRegion = "eu-central-1"
)

func NewS3Client(cfg config.Config, awsCfg aws.Config) *s3.Client {
	var opts func(*s3.Options)
	if cfg.AWS.EndpointURL != "" {
		opts = func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.AWS.EndpointURL)
			o.Region = *aws.String(awsDefaultRegion)
			o.UsePathStyle = true
		}
	} else {
		opts = func(o *s3.Options) {}
	}
	return s3.NewFromConfig(awsCfg, opts)
}

type S3Presigner struct {
	client *s3.PresignClient
	bucket string
}

func NewS3PreSigner(client *s3.Client, bucketName string) *S3Presigner {
	return &S3Presigner{
		client: s3.NewPresignClient(client),
		bucket: bucketName,
	}
}

func (p *S3Presigner) PreSignGet(key string, expires int64) (*v4.PresignedHTTPRequest, error) {
	return p.client.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(expires) * time.Second
	})

}

func NewSQSClient(cfg config.Config, awsCfg aws.Config) *sqs.Client {
	var opts func(*sqs.Options)
	if cfg.AWS.EndpointURL != "" {
		opts = func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(cfg.AWS.EndpointURL)
			o.Region = *aws.String(awsDefaultRegion)
		}
	} else {
		opts = func(o *sqs.Options) {}
	}
	return sqs.NewFromConfig(awsCfg, opts)
}
