package imagestorage

import (
	"context"
	"e-commerce/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ImageStorage struct {
	Client *minio.Client
	config *config.MinioConfig
}

func New(ctx context.Context, cfg *config.MinioConfig) (*ImageStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &ImageStorage{
		Client: client,
		config: cfg,
	}, nil
}
