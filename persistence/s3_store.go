package persistence

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Store struct {
	sess       *session.Session
	bucketName string
	key        string
}

func NewS3Store(region string, bucketName string, key string) (*s3Store, error) {
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		return nil, fmt.Errorf("initialize new session: %w", err)
	}
	return &s3Store{
		sess:       sess,
		bucketName: bucketName,
		key:        key,
	}, nil
}

func (s *s3Store) Load(ctx context.Context) ([]byte, error) {
	d := s3manager.NewDownloader(s.sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	input := &s3.GetObjectInput{Bucket: &s.bucketName, Key: &s.key}
	if _, err := d.DownloadWithContext(ctx, buf, input); err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *s3Store) Save(ctx context.Context, body []byte) error {
	u := s3manager.NewUploader(s.sess)
	reader := bytes.NewReader(body)
	input := &s3manager.UploadInput{Bucket: &s.bucketName, Key: &s.key, Body: reader}
	if _, err := u.UploadWithContext(ctx, input); err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}
