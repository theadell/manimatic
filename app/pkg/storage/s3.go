package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3 struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	presigner  *s3.PresignClient
	bucket     string
	log        *slog.Logger
}

func NewS3(client *s3.Client, bucket string, log *slog.Logger) *S3 {
	return &S3{
		client:     client,
		uploader:   manager.NewUploader(client),
		downloader: manager.NewDownloader(client),
		presigner:  s3.NewPresignClient(client),
		bucket:     bucket,
		log:        log,
	}
}

func (s *S3) UploadAndPresign(ctx context.Context, outputPath string, sessionId string) (string, error) {

	videoFile, err := os.Open(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open animation file %w", err)
	}
	ext := filepath.Ext(outputPath)
	if ext == "" {
		return "", fmt.Errorf("animation file has no extension %w", err)
	}

	key := fmt.Sprintf("manim_outputs/%s/%d%s",
		sessionId,
		time.Now().UnixNano(),
		ext,
	)

	err = s.Upload(ctx, key, videoFile)
	if err != nil {
		return "", err
	}

	return s.PresignGet(ctx, key, time.Minute*3)

}

func (s *S3) Upload(ctx context.Context, key string, file io.Reader) error {
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("S3 upload failed: %w", err)
	}

	s.log.Info("file uploaded to S3", "bucket", s.bucket, "key", key)
	return nil
}

func (s *S3) Download(ctx context.Context, key string, w io.WriterAt) error {
	_, err := s.downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

func (s *S3) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list failed: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}
	return keys, nil
}

func (s *S3) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, fmt.Errorf("head failed: %w", err)
	}
	return true, nil
}

func (s *S3) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {

	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign: %w", err)
	}

	return req.URL, nil
}
