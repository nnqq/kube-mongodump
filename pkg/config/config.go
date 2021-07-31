package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HealthzPort string
	MongoDB     mongodb
	S3          s3
	LogLevel    string
}

type s3 struct {
	Endpoint            string
	Region              string
	Secure              bool
	AccessKeyID         string
	SecretAccessKey     string
	BucketBackup        string
	BucketBackupTTLDays int
}

type mongodb struct {
	URL                    string
	NumParallelCollections int
}

func NewConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("envconfig.Process: %w", err)
	}
	return cfg, nil
}
