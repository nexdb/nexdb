package storage

import (
	"bytes"
	"io"

	"github.com/nexdb/nexdb/pkg/document"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var _ Storage = (*AWSS3)(nil)

// AWS S3 is a storage implementation that stores all data in AWS S3.
type AWSS3 struct {
	encryptionKey []byte
	bucket        string
	client        *s3.S3
	uploader      *s3manager.Uploader
}

// Delete implements Storage.
func (a *AWSS3) Delete(doc *document.Document) error {
	_, err := a.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(makeKey(doc.Collection, doc.ID.String())),
	})
	return err
}

// Stream implements Storage.
func (a *AWSS3) Stream() (<-chan *document.Document, error) {
	c := make(chan *document.Document)

	go func() {
		defer close(c)
		_ = a.client.ListObjectsV2Pages(
			&s3.ListObjectsV2Input{
				Bucket: aws.String(a.bucket),
			},
			func(page *s3.ListObjectsV2Output, lastPage bool) bool {
				for _, obj := range page.Contents {
					o, err := a.client.GetObject(&s3.GetObjectInput{
						Bucket: aws.String(a.bucket),
						Key:    obj.Key,
					})
					if err != nil {
						return false
					}

					// get bytes
					b, err := io.ReadAll(o.Body)
					if err != nil {
						return false
					}

					// decode the document
					doc, err := document.FromStorage(b, a.encryptionKey)
					if err != nil {
						return false
					}
					c <- doc
				}
				return true
			})
	}()

	return c, nil
}

// Write implements Storage.
func (a *AWSS3) Write(doc *document.Document) error {
	b, err := doc.ToStorage(a.encryptionKey)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)

	_, err = a.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(makeKey(doc.Collection, doc.ID.String())),
		Body:   r,
	})
	return err
}

// WithEncryptionKey sets the encryption key.
func (a *AWSS3) WithEncryptionKey(key []byte) (Storage, error) {
	a.encryptionKey = key
	return a, nil
}

// makeKey returns a key for the document.
func makeKey(collection, id string) string {
	return collection + "/" + id
}

// NewAWSS3 returns a new AWS S3 storage implementation.
func NewAWSS3(region, bucket string) (*AWSS3, error) {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(
		aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()).WithRegion(region),
	))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// create a client with the session
	client := s3.New(sess)

	return &AWSS3{
		bucket:   bucket,
		client:   client,
		uploader: uploader,
	}, nil
}
