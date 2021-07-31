package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"time"
)

func NewClient(
	ctx context.Context,
	accessKeyID,
	secretAccessKey,
	endpoint string,
	secure bool,
) (
	*minio.Client,
	error,
) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			accessKeyID,
			secretAccessKey,
			"",
		),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}

	// ping
	_, err = client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("client.ListBuckets: %w", err)
	}

	return client, nil
}
