package reader

import (
	"context"

	"github.com/nexdb/nexdb/pkg/database"
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/errors"
)

// Reader is a service that handles requests from handlers to read from the database.
type Reader struct {
	database *database.Database
}

// WriteDocument writes a document to the database.
func (r *Reader) GetDocument(ctx context.Context, id string) (*document.Document, error) {
	doc := r.database.GetByID(id)
	if doc == nil {
		return nil, errors.New(errors.ErrDocumentNotFound)
	}

	return doc, nil
}

// SearchDocuments searches the database for documents that match the query.
func (r *Reader) SearchDocuments(ctx context.Context, collection string, query cache.Query) ([]*document.Document, error) {
	docs := r.database.Filter(collection, query)
	return docs, nil
}

// New returns a new instance of Reader
func New(d *database.Database) *Reader {
	return &Reader{
		database: d,
	}
}
