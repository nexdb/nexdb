package cache_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue_SafeShutdown(t *testing.T) {
	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	// create our context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create a new queue
	q := cache.NewQueue(s)
	go q.Start(ctx)

	// push a create event to the queue
	go func() {
		for i := 0; i < 10; i++ {
			i := i
			q.Push(cache.Event{
				Operation: cache.OperationCreate,
				Document:  document.New().SetCollection("test" + strconv.Itoa(i)),
			})
		}
	}()

	// cancel the context
	cancel()

	// wait for the queue to drain
	q.WaitForShutdown()
}

func TestQueue_Push(t *testing.T) {
	s, err := storage.New(storage.MemoryDriver)
	require.NoError(t, err)

	// create our context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create a new queue
	q := cache.NewQueue(s)
	go q.Start(ctx)

	q.Push(cache.Event{
		Operation: cache.OperationCreate,
		Document:  document.New().SetCollection("test"),
	})
	cancel()
	q.WaitForShutdown()

	// check storage
	st, err := s.Stream()
	require.NoError(t, err)
	require.NotNil(t, st)

	var found bool
	for d := range st {
		if d.Collection == "test" {
			found = true
		}
	}

	assert.True(t, found)
}
