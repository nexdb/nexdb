package storage

import (
	"errors"
	"log"
	"os"

	"github.com/nexdb/nexdb/pkg/document"
)

// Driver is a storage driver.
type Driver string

const (
	MemoryDriver Driver = "memory"
	AWS3Driver   Driver = "aws-s3"
)

var (
	// ErrUnknownDriver is returned when a driver is not recognised.
	ErrUnknownDriver = errors.New("unknown driver")
	// ErrDocumentInvalid is returned when a document is invalid.
	ErrDocumentInvalid = errors.New("document is invalid")
)

// Storage is an interface for storage implementations.
type Storage interface {
	// Write writes a document to the storage, if the document already exists
	// with the same ID it will be overwritten.
	Write(doc *document.Document) error
	// Delete deletes a document from the storage.
	Delete(doc *document.Document) error
	// Stream streams documents from the storage.
	Stream() (<-chan *document.Document, error)
}

// New returns a new storage implementation.
func New(d Driver) (Storage, error) {
	encryptionKey := []byte(os.Getenv("NEXDB_ENCRYPTION_KEY"))
	if len(encryptionKey) > 0 {
		if len(encryptionKey) != 32 {
			log.Fatal("encryption key must be 32 bytes")
		}
	}

	switch d {
	case MemoryDriver:
		m, err := NewMemory()
		if err != nil {
			return nil, err
		}

		return m.WithEncryptionKey(encryptionKey)
	case AWS3Driver:
		a, err := NewAWSS3(os.Getenv("AWS_REGION"), os.Getenv("AWS_BUCKET"))
		if err != nil {
			return nil, err
		}

		return a.WithEncryptionKey(encryptionKey)
	default:
		return nil, ErrUnknownDriver
	}
}
