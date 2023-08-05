package document_test

import (
	"encoding/json"
	"testing"

	"github.com/nexdb/nexdb/pkg/document"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocument_New(t *testing.T) {
	d := document.New()
	require.NotNil(t, d)

	assert.NotEmpty(t, d.ID)
	assert.Empty(t, d.Collection)
	require.NotNil(t, d.Data)
}

func TestDocument_SetCollection(t *testing.T) {
	for _, tc := range []struct {
		name       string
		collection string
	}{
		{
			name:       "set collection: collection is set",
			collection: "test",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := document.New()

			d.SetCollection(tc.collection)

			assert.Equal(t, tc.collection, d.Collection)
		})
	}
}

func TestDocument_SetData(t *testing.T) {
	for _, tc := range []struct {
		name   string
		data   map[string]interface{}
		verify func(t *testing.T, d *document.Document)
	}{
		{
			name: "nil data should set as map[string]interface{}{}",
			data: nil,
			verify: func(t *testing.T, d *document.Document) {
				assert.NotNil(t, d.Data)
			},
		},
		{
			name: "data should set",
			data: map[string]interface{}{
				"foo": "bar",
			},
			verify: func(t *testing.T, d *document.Document) {
				assert.NotNil(t, d.Data)
				assert.Equal(t, "bar", d.Data["foo"])
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := document.New()

			d.SetData(tc.data)

			tc.verify(t, d)
		})
	}
}

func TestDocument_ToStorage(t *testing.T) {
	for _, tc := range []struct {
		name          string
		doc           *document.Document
		encryptionKey []byte
		verify        func(t *testing.T, d *document.Document, b []byte, err error)
	}{
		{
			name: "without encryption key: should return bytes of document",
			doc: document.New().SetCollection("test").SetData(map[string]interface{}{
				"foo": "bar",
			}),
			encryptionKey: []byte{},
			verify: func(t *testing.T, d *document.Document, b []byte, err error) {
				require.NoError(t, err)
				require.NotNil(t, b)

				var got *document.Document
				err = json.Unmarshal(b, &got)
				require.NoError(t, err)
				assert.Equal(t, d, got)
			},
		},
		{
			name: "with encryption key: should return bytes encrypted of document",
			doc: document.New().SetCollection("test").SetData(map[string]interface{}{
				"foo": "bar",
			}),
			encryptionKey: []byte("key-that-is-thirty-2-bytes-long!"),
			verify: func(t *testing.T, d *document.Document, b []byte, err error) {
				require.NoError(t, err)
				require.NotNil(t, b)

				var got *document.Document
				err = json.Unmarshal(b, &got)
				require.Error(t, err)

				// decode
				got, err = document.FromStorage(b, []byte("key-that-is-thirty-2-bytes-long!"))
				require.NoError(t, err)
				assert.Equal(t, d, got)
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.doc.ToStorage(tc.encryptionKey)
			tc.verify(t, tc.doc, b, err)
		})
	}
}
