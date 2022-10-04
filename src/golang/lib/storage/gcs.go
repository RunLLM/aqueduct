package storage

import (
	"bytes"
	"context"
	"io"

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

func (g *gcsStorage) Get(ctx context.Context, key string) ([]byte, error) {
	client, err := g.newClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Check if object exists
	_, err = client.Bucket(g.gcsConfig.Bucket).Object(key).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, ErrObjectDoesNotExist
		}
		return nil, err
	}

	rc, err := client.Bucket(g.gcsConfig.Bucket).Object(key).NewReader(ctx)
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

	buf := bytes.NewBuffer(value)
	wc := client.Bucket(g.gcsConfig.Bucket).Object(key).NewWriter(ctx)

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

	return client.Bucket(g.gcsConfig.Bucket).Object(key).Delete(ctx)
}

// newClient returns a GCS client for this storage object.
// The caller must call `defer client.Close()` on the returned storage client.
func (g *gcsStorage) newClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, option.WithCredentialsJSON([]byte(g.gcsConfig.ServiceAccountCredentials)))
}
