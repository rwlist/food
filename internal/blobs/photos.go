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

type Photos struct {
	s3Client *s3.Client
	bucket   string
}

func NewPhotos(s3Client *s3.Client, bucket string) *Photos {
	return &Photos{s3Client: s3Client, bucket: bucket}
}

func (p *Photos) UploadFile(ctx context.Context, file string) (string, error) {
	slog.Info("uploading file to photos bucket", "file", file)

	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(data)

	key := "photos/" + filepath.Base(file)
	_, err = p.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &p.bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

func (p *Photos) GetObject(ctx context.Context, file string) ([]byte, error) {
	resp, err := p.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &file,
	})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
