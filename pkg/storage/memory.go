package storage

import (
	"sync"

	"github.com/nexdb/nexdb/pkg/document"
)

// Memory is a storage implementation that stores all data in memory.
type Memory struct {
	encryptionKey []byte
	mx            sync.RWMutex
	data          []*document.Document
}

// Write writes a document to the storage, if the document already exists
// with the same ID it will be overwritten.
func (m *Memory) Write(doc *document.Document) error {
	if doc == nil {
		return ErrDocumentInvalid
	}

	m.mx.Lock()
	defer m.mx.Unlock()

	// check if document already exists
	// and overwrite if it does.
	for i, d := range m.data {
		if d.ID.String() == doc.ID.String() {
			m.data[i] = doc
			return nil
		}
	}

	m.data = append(m.data, doc)

	return nil
}

// Stream streams documents from the storage.
func (m *Memory) Stream() (<-chan *document.Document, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	c := make(chan *document.Document)
	go func() {
		defer close(c)
		for _, d := range m.data {
			c <- d
		}
	}()

	return c, nil
}

// Delete deletes a document from the storage.
func (m *Memory) Delete(doc *document.Document) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	for i, d := range m.data {
		if d.ID.Compare(doc.ID) == 0 {
			m.data = append(m.data[:i], m.data[i+1:]...)
			return nil
		}
	}

	return nil
}

// WithEncryptionKey sets the encryption key.
func (m *Memory) WithEncryptionKey(key []byte) (Storage, error) {
	m.encryptionKey = key
	return m, nil
}

// NewMemory returns a new Memory storage implementation.
func NewMemory() (*Memory, error) {
	return &Memory{}, nil
}
