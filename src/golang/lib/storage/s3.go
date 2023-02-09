package storage

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"path"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Storage struct {
	s3Config *shared.S3Config
}

func newS3Storage(s3Config *shared.S3Config) *s3Storage {
	return &s3Storage{
		s3Config: s3Config,
	}
}

// parseBucketAndKey takes the bucket in the form of s3://bucket/path
// and a key and parses the bucket name and the key.
func (s *s3Storage) parseBucketAndKey(key string) (string, string, error) {
	u, err := url.Parse(s.s3Config.Bucket)
	if err != nil {
		return "", "", err
	}

	bucket := u.Host

	u.Path = strings.TrimLeft(u.Path, "/")
	key = path.Join(u.Path, key)

	return bucket, key, nil
}

func (s *s3Storage) Get(ctx context.Context, key string) ([]byte, error) {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return nil, err
	}

	bucket, key, err := s.parseBucketAndKey(key)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	// Get the object
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return content, err
}

func (s *s3Storage) Put(ctx context.Context, key string, value []byte) error {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return err
	}

	bucket, key, err := s.parseBucketAndKey(key)
	if err != nil {
		return err
	}

	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(value),
	})
	return err
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return err
	}

	s3Client := s3.New(sess)

	bucket, key, err := s.parseBucketAndKey(key)
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteObjectWithContext(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	return err
}

func (s *s3Storage) Exists(ctx context.Context, key string) bool {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		return false
	}

	s3Client := s3.New(sess)

	bucket, key, err := s.parseBucketAndKey(key)
	if err != nil {
		return false
	}

	_, err = s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return false
			default:
				return false
			}
		}
		return false
	}
	return true
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
