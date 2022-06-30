package storage

import (
	"bytes"
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Storage struct {
	s3Config *shared.S3Config
}

func newS3Storage(s3Config *shared.S3Config) *s3Storage {
	return &s3Storage{
		s3Config: s3Config,
	}
}

func (s *s3Storage) Get(ctx context.Context, key string) ([]byte, error) {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return nil, err
	}

	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.DownloadWithContext(
		ctx,
		buff,
		&s3.GetObjectInput{
			Bucket: aws.String(s.s3Config.Bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (s *s3Storage) Put(ctx context.Context, key string, value []byte) error {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return err
	}

	file := bytes.NewReader(value)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.UploadWithContext(
		ctx,
		&s3manager.UploadInput{
			Bucket: aws.String(s.s3Config.Bucket),
			Key:    aws.String(key),
			Body:   file,
		})
	if err != nil {
		return err
	}
	return nil
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return err
	}

	s3Client := s3.New(sess)
	_, err = s3Client.DeleteObjectWithContext(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(s.s3Config.Bucket),
			Key:    aws.String(key),
		},
	)
	return err
}

func CreateS3Session(s3Config *shared.S3Config) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s3Config.Region),
		Credentials: credentials.NewSharedCredentials(
			s3Config.CredentialsPath,
			s3Config.CredentialsProfile,
		),
	})
	if err != nil {
		return nil, err
	}
	return sess, nil
}
