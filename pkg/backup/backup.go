package backup

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/nnqq/kube-mongodump/pkg/minio"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"os"
	"path"
	"time"
)

type backup struct {
	mongodbURL             string
	numParallelCollections int
	backupBucket           *minio.Bucket
	logger                 zerolog.Logger
}

func NewBackup(logger zerolog.Logger, mongodbURL string, numParallelCollections int, backupBucket *minio.Bucket) *backup {
	return &backup{
		mongodbURL:             mongodbURL,
		numParallelCollections: numParallelCollections,
		backupBucket:           backupBucket,
		logger:                 logger,
	}
}

func (b *backup) Do(ctx context.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.GetWd: %w", err)
	}

	archiveName := fmt.Sprintf("%d.gz", time.Now().Unix())
	archivePath := path.Join(wd, archiveName)
	defer func() {
		e := os.Remove(archivePath)
		if e != nil {
			b.logger.Error().Err(e).Send()
		}
	}()

	uri, err := connstring.Parse(b.mongodbURL)
	if err != nil {
		return fmt.Errorf("connstring.Parse: %w", err)
	}

	md := &mongodump.MongoDump{
		ToolOptions: &options.ToolOptions{
			URI: &options.URI{
				ConnectionString: b.mongodbURL,
				ConnString:       uri,
			},
			Namespace: &options.Namespace{},
			Auth:      &options.Auth{},
			Connection: &options.Connection{
				Timeout: 5,
			},
		},
		OutputOptions: &mongodump.OutputOptions{
			Archive:                archivePath,
			Gzip:                   true,
			Oplog:                  true,
			NumParallelCollections: b.numParallelCollections,
		},
		InputOptions: &mongodump.InputOptions{},
	}

	err = md.Init()
	if err != nil {
		return fmt.Errorf("md.Init: %w", err)
	}

	err = md.Dump()
	if err != nil {
		return fmt.Errorf("md.Dump: %w", err)
	}
	b.logger.Info().Msg("mongodump OK")

	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("f.Stat: %w", err)
	}

	err = b.backupBucket.Put(ctx, archiveName, f, stat.Size())
	if err != nil {
		return fmt.Errorf("b.backupBucket.Put: %w", err)
	}
	b.logger.Info().Msg("S3 put OK")

	return nil
}
