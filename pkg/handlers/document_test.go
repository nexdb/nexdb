package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexdb/nexdb/pkg/database"
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/errors"
	"github.com/nexdb/nexdb/pkg/handlers"
	"github.com/nexdb/nexdb/pkg/services/writer"
	"github.com/nexdb/nexdb/pkg/storage"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocument_WriteDocument(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	q := cache.NewQueue(s)

	d := &database.Database{
		Cache: cache.NewCache(ctx, q),
	}
	wr := writer.New(d)

	rr := httptest.NewRecorder()

	// set the request body
	data := []byte(`{"name": "John"}`)

	req := httptest.NewRequest("POST", "/collection/users", bytes.NewReader(data))
	req = mux.SetURLVars(req, map[string]string{
		"collection": "users",
	})

	// call the handler
	handlers.WriteDocument(wr).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestDocument_WriteDocumentWithInvalidCollectionName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	q := cache.NewQueue(s)

	d := &database.Database{
		Cache: cache.NewCache(ctx, q),
	}
	wr := writer.New(d)

	rr := httptest.NewRecorder()

	// set the request body
	data := []byte(`{"name": "John"}`)

	req := httptest.NewRequest("POST", "/collection/users-invalid", bytes.NewReader(data))
	req = mux.SetURLVars(req, map[string]string{
		"collection": "users-invalid",
	})

	// call the handler
	handlers.WriteDocument(wr).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	expectedErr := errors.New(errors.ErrCollectionNameIsInvalid)
	expectedCode := expectedErr.Code().ToString()
	assert.Equal(t, `{"error":"`+expectedErr.Error()+`","code":`+expectedCode+`}`, rr.Body.String())
}
