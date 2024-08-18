package blobs

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const PrefixPhotos = "photos/"

type Photos struct {
	s3Client *s3.Client
	bucket   string
}

func NewPhotos(s3Client *s3.Client, bucket string) *Photos {
	return &Photos{s3Client: s3Client, bucket: bucket}
}

// UploadFile uploads a file to the photos bucket.
func (p *Photos) UploadFile(ctx context.Context, filePath string) (string, error) {
	slog.Info("uploading file to photos bucket", "file", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(data)

	filebase := filepath.Base(filePath)
	key := PrefixPhotos + filebase
	_, err = p.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &p.bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}

	return filebase, nil
}

// GetPhoto returns an object from the photos bucket.
func (p *Photos) GetPhoto(ctx context.Context, name string) (*s3.GetObjectOutput, error) {
	key := PrefixPhotos + name
	return p.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &key,
	})
}
