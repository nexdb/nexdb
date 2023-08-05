package cache_test

import (
	"context"
	"testing"

	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Put(t *testing.T) {
	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	// create a new cache
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := cache.NewCache(ctx, cache.NewQueue(s))

	// create a new document
	d := document.New().SetCollection("test")

	// put the document
	err = c.Put(d, false)
	require.NoError(t, err)

	// get the document
	got := c.GetByID(d.ID.String())
	assert.Equal(t, got, d)
}

func TestCache_PutWithExistingDocument(t *testing.T) {
	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	// create a new cache
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := cache.NewCache(ctx, cache.NewQueue(s))

	// create a new document
	d := document.New().SetCollection("test")

	// put the document
	err = c.Put(d, false)
	require.NoError(t, err)

	// get the document
	got := c.GetByID(d.ID.String())
	assert.Equal(t, "test", got.Collection)

	// update the document
	d.SetCollection("test2")

	// put the document
	err = c.Put(d, false)
	require.NoError(t, err)

	// get the document
	got = c.GetByID(d.ID.String())
	assert.Equal(t, "test2", got.Collection)
}

func TestCache_Delete(t *testing.T) {
	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	// create a new cache
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := cache.NewCache(ctx, cache.NewQueue(s))

	// create a new document
	d := document.New().SetCollection("test")

	// put the document
	err = c.Put(d, false)
	require.NoError(t, err)

	// get the document
	got := c.GetByID(d.ID.String())
	assert.Equal(t, got, d)

	// delete the document
	err = c.Delete(d.ID.String())
	require.NoError(t, err)

	// get the document
	got = c.GetByID(d.ID.String())
	assert.Nil(t, got)
}
