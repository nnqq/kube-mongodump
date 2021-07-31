package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"io"
)

type Bucket struct {
	client     *minio.Client
	bucketName string
}

func NewBucket(
	ctx context.Context,
	client *minio.Client,
	bucketName string,
	region string,
	ttlDays int,
) (
	*Bucket,
	error,
) {
	err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: region,
	})
	if err != nil &&
		err.Error() != "Your previous request to create the named bucket succeeded and you already own it." &&
		err.Error() != "Bucket already exists" {
		return nil, fmt.Errorf("client.MakeBucket: %w", err)
	}

	if ttlDays != 0 {
		e := client.SetBucketLifecycle(ctx, bucketName, &lifecycle.Configuration{
			Rules: []lifecycle.Rule{{
				ID:     "Remove expired files",
				Status: "Enabled",
				Expiration: lifecycle.Expiration{
					Days: lifecycle.ExpirationDays(ttlDays),
				},
			}, {
				ID:     "Remove expired multipart upload",
				Status: "Enabled",
				AbortIncompleteMultipartUpload: lifecycle.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: 1,
				},
			}},
		})
		if e != nil {
			return nil, fmt.Errorf("client.SetBucketLifecycle: %w", e)
		}
	}

	return &Bucket{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (b *Bucket) Put(ctx context.Context, name string, payload io.Reader, size int64) error {
	_, err := b.client.PutObject(ctx, b.bucketName, name, payload, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("b.client.PutObject: %w", err)
	}

	return nil
}
