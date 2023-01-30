package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"google.golang.org/api/option"
)

type gcsStorage struct {
	gcsConfig *shared.GCSConfig
}

func newGCSStorage(gcsConfig *shared.GCSConfig) *gcsStorage {
	return &gcsStorage{
		gcsConfig: gcsConfig,
	}
}

// parseBucketAndKey uses a bucket in the form of bucket/path and a key and
// returns the bucket name and the full key.
func (g *gcsStorage) parseBucketAndKey(key string) (string, string) {
	parts := strings.Split(g.gcsConfig.Bucket, "/")
	if len(parts) == 1 {
		// There is no subpath for this bucket
		return parts[0], key
	}

	bucket := parts[0]
	keyPath := parts[1:]

	// The subpath should be prefixed to the key to get the full key path
	keyPath = append(keyPath, key)
	key = path.Join(keyPath...)

	return bucket, key
}

func (g *gcsStorage) Get(ctx context.Context, key string) ([]byte, error) {
	client, err := g.newClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	bucket, key := g.parseBucketAndKey(key)

	// Check if object exists
	_, err = client.Bucket(bucket).Object(key).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, errors.New("Object does not exist in storage.")
		}
		return nil, err
	}

	rc, err := client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *gcsStorage) Put(ctx context.Context, key string, value []byte) error {
	client, err := g.newClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	bucket, key := g.parseBucketAndKey(key)

	buf := bytes.NewBuffer(value)
	wc := client.Bucket(bucket).Object(key).NewWriter(ctx)

	if _, err = io.Copy(wc, buf); err != nil {
		return err
	}

	return wc.Close()
}

func (g *gcsStorage) Delete(ctx context.Context, key string) error {
	client, err := g.newClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	bucket, key := g.parseBucketAndKey(key)

	return client.Bucket(bucket).Object(key).Delete(ctx)
}

// newClient returns a GCS client for this storage object.
// The caller must call `defer client.Close()` on the returned storage client.
func (g *gcsStorage) newClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithCredentialsJSON([]byte(g.gcsConfig.ServiceAccountCredentials)))
}
