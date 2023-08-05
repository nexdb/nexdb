package writer_test

import (
	"context"
	"testing"

	"github.com/nexdb/nexdb/pkg/database"
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/errors"
	"github.com/nexdb/nexdb/pkg/services/writer"
	"github.com/nexdb/nexdb/pkg/storage"

	"github.com/stretchr/testify/require"
)

func TestWriter_WriteDocument(t *testing.T) {
	for _, tc := range []struct {
		name       string
		collection string
		data       map[string]interface{}
		verify     func(t *testing.T, doc *document.Document, err error)
	}{
		{
			name:       "valid, no error should be returned",
			collection: "users",
			data:       map[string]interface{}{"name": "John"},
			verify: func(t *testing.T, doc *document.Document, err error) {
				require.NoError(t, err)
				require.NotNil(t, doc)
			},
		},
		{
			name:       "collection name is empty, expect ErrCollectionInvalid to be returned",
			collection: "",
			data:       map[string]interface{}{"name": "John"},
			verify: func(t *testing.T, doc *document.Document, err error) {
				require.Equal(t, err.Error(), errors.New(errors.ErrCollectionNameIsEmpty).Error())
				require.Nil(t, doc)
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s, err := storage.New(storage.MemoryDriver)
			require.NoError(t, err)

			q := cache.NewQueue(s)

			d := &database.Database{
				Cache: cache.NewCache(ctx, q),
			}
			wr := writer.New(d)

			got, err := wr.WriteDocument(ctx, tc.collection, tc.data)
			tc.verify(t, got, err)
		})
	}
}
