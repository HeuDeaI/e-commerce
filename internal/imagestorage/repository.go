package imagestorage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type ImageStorageRepository interface {
	Upload(ctx context.Context, file io.Reader, id int) (string, error)
	Delete(ctx context.Context, id int) error
}

type imageStorageRepository struct {
	client *minio.Client
	bucket string
}

func NewImageStorageRepository(client *minio.Client, bucket string) ImageStorageRepository {
	return &imageStorageRepository{
		client: client,
		bucket: bucket,
	}
}

func (r *imageStorageRepository) Upload(ctx context.Context, file io.Reader, id int) (string, error) {
	objectName := fmt.Sprintf("%d.jpg", id)
	_, err := r.client.PutObject(ctx, r.bucket, objectName, file, -1, minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s/%s/%s", r.client.EndpointURL().Host, r.bucket, objectName), nil
}

func (r *imageStorageRepository) Delete(ctx context.Context, id int) error {
	objectName := fmt.Sprintf("%d.jpg", id)
	return r.client.RemoveObject(ctx, r.bucket, objectName, minio.RemoveObjectOptions{})
}
