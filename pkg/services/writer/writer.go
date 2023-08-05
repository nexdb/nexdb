package writer

import (
	"context"

	"github.com/nexdb/nexdb/pkg/database"
	"github.com/nexdb/nexdb/pkg/database/validation"
	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/errors"
)

// Writer is a service that handles requests from handlers to write to the database.
type Writer struct {
	database *database.Database
}

// WriteDocument writes a document to the database.
func (w *Writer) WriteDocument(ctx context.Context, collection string, data map[string]interface{}) (*document.Document, error) {
	if err := validation.ValidateCollectionName(collection); err != nil {
		return nil, err
	}

	// if the document has an id, then it already exists in the database
	// and we need to update it.
	if data["_id"] != nil {
		existing := w.database.GetByID(data["_id"].(string))
		if existing == nil {
			return nil, errors.New(errors.ErrDocumentNotFound)
		}

		// remove _id from the data
		delete(data, "_id")

		existing.SetData(data)

		err := w.database.Put(existing, false)
		return existing, err
	}

	// if the document does not have an id, then we need to create a new one.
	doc := document.New().SetCollection(collection).SetData(data)
	err := w.database.Put(doc, false)

	return doc, err
}

// DeleteDocument deletes a document from the database.
func (w *Writer) DeleteDocument(ctx context.Context, id string) error {
	if doc := w.database.GetByID(id); doc == nil {
		return errors.New(errors.ErrDocumentNotFound)
	}

	return w.database.Delete(id)
}

// New returns a new instance of Writer.
func New(d *database.Database) *Writer {
	return &Writer{
		database: d,
	}
}
