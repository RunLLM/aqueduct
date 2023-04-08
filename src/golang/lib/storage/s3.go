package storage

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

type s3Storage struct {
	s3Config *shared.S3Config
}

func newS3Storage(s3Config *shared.S3Config) *s3Storage {
	return &s3Storage{
		s3Config: s3Config,
	}
}

// parseBucketAndKey takes a key and resolves to into the form s3://bucket/[root dirpath/]path,
// returning the bucket name and the full, usable key.
func (s *s3Storage) parseBucketAndKey(key string) (string, string, error) {
	log.Errorf("Parse bucket and key: %s", key)
	log.Errorf("S3Config bucket %s", s.s3Config.Bucket)

	u, err := url.Parse(s.s3Config.Bucket)
	if err != nil {
		return "", "", err
	}

	bucket := u.Host

	log.Errorf("U.Path %v", u.Path)

	dirPath := strings.TrimLeft(u.Path, "/")
	if s.s3Config.RootDir != "" {
		dirPath = path.Join(dirPath, s.s3Config.RootDir)
	}
	key = path.Join(dirPath, key)

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
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, errors.Wrapf(ErrObjectDoesNotExist(), "Unable to fetch key `%s` from bucket `%s`.", key, bucket)
			default:
				return nil, err
			}
		}
		return nil, err
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return content, err
}

func (s *s3Storage) Put(ctx context.Context, key string, value []byte) error {
	sess, err := CreateS3Session(s.s3Config)
	if err != nil {
		log.Errorf("HELLO")
		return err
	}

	bucket, key, err := s.parseBucketAndKey(key)
	if err != nil {
		return err
	}

	log.Errorf("HELLO: s3storage.put() bucket %s: %s", bucket, key)

	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(value),
	})
	log.Errorf("HELLO: put object failure: %v", err)
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
	// TODO: ENG-2428 we should explicitly surface other error types to the caller
	// instead of just returning `false` for non s3.ErrCodeNoSuchKey errors.
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
