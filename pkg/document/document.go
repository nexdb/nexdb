package document

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"

	"github.com/oklog/ulid/v2"
)

// Document is a document that can be stored in a database.
type Document struct {
	ID         ulid.ULID              `json:"_id"`
	Collection string                 `json:"collection"`
	Data       map[string]interface{} `json:"data"`
}

// SetID sets the ID of the document. Should only be used for testing.
func (d *Document) SetID(id string) *Document {
	d.ID = ulid.MustParse(id)

	return d
}

// SetCollection sets the collection of the document.
func (d *Document) SetCollection(name string) *Document {
	d.Collection = name

	return d
}

// SetData sets the fields of the document.
func (d *Document) SetData(data map[string]interface{}) *Document {
	if data == nil {
		data = make(map[string]interface{})
	}

	d.Data = data

	return d
}

// ToStorage returns the document as a byte slice that can be stored in a database,
// if an encryption key is provided, the document will be encrypted.
func (d *Document) ToStorage(encryptionKey []byte) ([]byte, error) {
	b, err := json.Marshal(&d)
	if err != nil {
		return nil, err
	}

	if len(encryptionKey) > 0 {
		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		return gcm.Seal(nonce, nonce, b, nil), nil
	}

	return b, nil
}

// FromStorage returns a document from a byte slice that was stored in a database,
// if an encryption key is provided, the document will be decrypted.
func FromStorage(b []byte, encryptionKey []byte) (*Document, error) {
	if len(encryptionKey) > 0 {
		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		nonceSize := gcm.NonceSize()
		if len(b) < nonceSize {
			return nil, err
		}

		nonce, ciphertext := b[:nonceSize], b[nonceSize:]
		b, err = gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, err
		}
	}

	d := &Document{}
	if err := json.Unmarshal(b, d); err != nil {
		return nil, err
	}

	return d, nil
}

// New returns a new document.
func New() *Document {
	return &Document{
		ID:         ulid.Make(),
		Collection: "",
		Data:       make(map[string]interface{}),
	}
}
