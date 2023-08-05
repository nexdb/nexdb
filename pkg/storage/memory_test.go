package storage_test

import (
	"testing"

	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_Memory(t *testing.T) {
	for _, tc := range []struct {
		name   string
		doc    *document.Document
		before func(s storage.Storage)
		verify func(t *testing.T, s storage.Storage, doc *document.Document, err error)
	}{
		{
			name:   "write nil document, expect ErrDocumentInvalid to be returned",
			before: func(s storage.Storage) {},
			verify: func(t *testing.T, s storage.Storage, doc *document.Document, err error) {
				require.ErrorIs(t, err, storage.ErrDocumentInvalid)
			},
		},
		{
			name:   "write document, expect document to exist in storage",
			doc:    document.New().SetCollection("test"),
			before: func(s storage.Storage) {},
			verify: func(t *testing.T, s storage.Storage, doc *document.Document, err error) {
				require.NoError(t, err)

				c, err := s.Stream()
				require.NoError(t, err)
				require.NotNil(t, c)
				found := false
				for d := range c {
					if d.ID == doc.ID {
						found = true
						break
					}
				}
				assert.True(t, found)
			},
		},
		{
			name: "write document, expect document to be updated in storage",
			doc:  document.New().SetID("00000000000000000000000000").SetCollection("test"),
			before: func(s storage.Storage) {
				s.Write(document.New().
					SetID("00000000000000000000000000").
					SetCollection("test").
					SetData(map[string]interface{}{
						"foo": "bar",
					}))
			},
			verify: func(t *testing.T, s storage.Storage, doc *document.Document, err error) {
				require.NoError(t, err)

				c, err := s.Stream()
				require.NoError(t, err)
				require.NotNil(t, c)

				var got *document.Document
				for d := range c {
					if d.ID == doc.ID {
						got = d
						break
					}
				}
				assert.NotNil(t, got)

				// assert the document has been updated
				want := map[string]interface{}{}
				assert.Equal(t, want, got.Data)
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.New(storage.MemoryDriver)
			require.NoError(t, err)

			tc.before(s)

			err = s.Write(tc.doc)

			tc.verify(t, s, tc.doc, err)
		})
	}
}

func TestStorage_Memory_Delete(t *testing.T) {
	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	d := document.New().
		SetID("00000000000000000000000000").
		SetCollection("test").
		SetData(map[string]interface{}{
			"foo": "",
		})
	err = s.Write(d)
	require.NoError(t, err)

	// assert the document exists
	c, err := s.Stream()
	require.NoError(t, err)
	require.NotNil(t, c)

	found := getDocumentFromStream(c, d.ID.String())
	require.NotNil(t, found)

	// delete the document
	err = s.Delete(d)
	require.NoError(t, err)

	// assert the document has been deleted
	c, err = s.Stream()
	require.NoError(t, err)
	require.NotNil(t, c)

	found = getDocumentFromStream(c, d.ID.String())
	require.Nil(t, found)
}

func getDocumentFromStream(c <-chan *document.Document, id string) *document.Document {
	for d := range c {
		if d.ID.String() == id {
			return d
		}
	}
	return nil
}
