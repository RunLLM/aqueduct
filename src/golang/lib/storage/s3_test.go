package storage

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/stretchr/testify/require"
)

func TestParseBucketAndKey(t *testing.T) {
	type test struct {
		inputBucket    string
		inputKey       string
		expectedBucket string
		expectedKey    string
	}

	tests := []test{
		{
			inputBucket:    "s3://aqueduct/test/folder",
			inputKey:       "key",
			expectedBucket: "aqueduct",
			expectedKey:    "test/folder/key",
		},
		{
			inputBucket:    "s3://aqueduct",
			inputKey:       "key",
			expectedBucket: "aqueduct",
			expectedKey:    "key",
		},
	}

	for _, tc := range tests {
		storage := s3Storage{
			s3Config: &shared.S3Config{
				Bucket: tc.inputBucket,
			},
		}

		bucket, key, err := storage.parseBucketAndKey(tc.inputKey)
		require.Nil(t, err)
		require.Equal(t, tc.expectedBucket, bucket)
		require.Equal(t, tc.expectedKey, key)
	}
}
