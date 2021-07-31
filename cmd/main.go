package main

import (
	"context"
	"github.com/nnqq/healthz"
	"github.com/nnqq/kube-mongodump/pkg/backup"
	"github.com/nnqq/kube-mongodump/pkg/config"
	"github.com/nnqq/kube-mongodump/pkg/logger"
	"github.com/nnqq/kube-mongodump/pkg/minio"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	logg, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	minioClient, err := minio.NewClient(
		ctx,
		cfg.S3.AccessKeyID,
		cfg.S3.SecretAccessKey,
		cfg.S3.Endpoint,
		cfg.S3.Secure,
	)
	if err != nil {
		panic(err)
	}

	backupBucket, err := minio.NewBucket(
		ctx,
		minioClient,
		cfg.S3.BucketBackup,
		cfg.S3.Region,
		cfg.S3.BucketBackupTTLDays,
	)
	if err != nil {
		panic(err)
	}

	go healthz.NewHealthz(healthz.Logger(&logg), healthz.Addr("0.0.0.0:"+cfg.HealthzPort))
	err = backup.NewBackup(logg, cfg.MongoDB.URL, backupBucket).Do(ctx)
	if err != nil {
		panic(err)
	}
}
