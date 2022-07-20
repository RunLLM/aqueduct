package storage

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/stretchr/testify/require"
)

func TestParseBucketAndKey(t *testing.T) {
	config := &shared.S3Config{
		Region:             "us-east-2",
		Bucket:             "s3://aqueduct/test/folder",
		CredentialsPath:    "/home/users/aqueduct/.aws",
		CredentialsProfile: "default",
	}
	storage := s3Storage{
		s3Config: config,
	}

	bucket, key, err := storage.parseBucketAndKey("key")
	require.Nil(t, err)
	require.Equal(t, "aqueduct", bucket)
	require.Equal(t, "/test/folder/key", key)
}
