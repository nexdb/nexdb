package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/nexdb/nexdb/pkg/api/rest"
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/errors"
	"github.com/nexdb/nexdb/pkg/services/reader"
	"github.com/nexdb/nexdb/pkg/services/writer"

	"github.com/gorilla/mux"
)

// WriteDocument is a handler that writes a document to the database.
func WriteDocument(w *writer.Writer) rest.JsonHandler {
	return func(r *http.Request) *rest.Response {
		vars := mux.Vars(r)

		// get the collection name
		collection := vars["collection"]

		defer r.Body.Close()

		// get the data
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			return rest.JsonResponse(
				rest.WithError(err),
			)
		}

		doc, err := w.WriteDocument(r.Context(), collection, data)
		if err != nil {
			code := http.StatusBadRequest
			if internalErr, ok := err.(*errors.Error); ok {
				if internalErr.ErrorCode == errors.ErrDocumentNotFound {
					code = http.StatusNotFound
				}
			}

			return rest.JsonResponse(
				rest.WithError(err),
				rest.SetStatus(code),
			)
		}

		return rest.JsonResponse(
			rest.SetStatus(http.StatusCreated),
			rest.SetBody(doc),
			rest.SetWrap("data"),
		)
	}
}

// GetDocument is a handler that gets a document from the database.
func GetDocument(readerSvc *reader.Reader) rest.JsonHandler {
	return func(r *http.Request) *rest.Response {
		vars := mux.Vars(r)

		// get the document id
		id := vars["id"]

		defer r.Body.Close()

		doc, err := readerSvc.GetDocument(r.Context(), id)
		if err != nil {
			code := http.StatusBadRequest
			if internalErr, ok := err.(*errors.Error); ok {
				if internalErr.ErrorCode == errors.ErrDocumentNotFound {
					code = http.StatusNotFound
				}
			}

			return rest.JsonResponse(
				rest.WithError(err),
				rest.SetStatus(code),
			)
		}

		return rest.JsonResponse(
			rest.SetStatus(http.StatusOK),
			rest.SetBody(doc),
			rest.SetWrap("data"),
		)
	}
}

// DeleteDocument is a handler that deletes a document from the database.
func DeleteDocument(w *writer.Writer) rest.JsonHandler {
	return func(r *http.Request) *rest.Response {
		vars := mux.Vars(r)

		// get the document id
		id := vars["id"]

		defer r.Body.Close()

		err := w.DeleteDocument(r.Context(), id)
		if err != nil {
			code := http.StatusBadRequest
			if internalErr, ok := err.(*errors.Error); ok {
				if internalErr.ErrorCode == errors.ErrDocumentNotFound {
					code = http.StatusNotFound
				}
			}

			return rest.JsonResponse(
				rest.WithError(err),
				rest.SetStatus(code),
			)
		}

		return rest.JsonResponse(
			rest.SetStatus(http.StatusOK),
			rest.SetBody(map[string]interface{}{}),
			rest.SetWrap("data"),
		)
	}
}

// SearchDocuments is a handler that searches documents in the database.
func SearchDocuments(readerSvc *reader.Reader) rest.JsonHandler {
	return func(r *http.Request) *rest.Response {
		vars := mux.Vars(r)

		// get the collection name
		collection := vars["collection"]

		defer r.Body.Close()

		// get the data
		var q cache.Query
		err := json.NewDecoder(r.Body).Decode(&q)
		if err != nil {
			return rest.JsonResponse(
				rest.WithError(err),
			)
		}

		docs, err := readerSvc.SearchDocuments(r.Context(), collection, q)
		if err != nil {
			return rest.JsonResponse(
				rest.WithError(err),
				rest.SetStatus(http.StatusBadRequest),
			)
		}

		return rest.JsonResponse(
			rest.SetStatus(http.StatusOK),
			rest.SetBody(docs),
			rest.SetWrap("data"),
		)
	}
}
