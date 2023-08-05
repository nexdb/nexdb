package cache

import (
	"context"
	"sync"

	"github.com/nexdb/nexdb/pkg/document"
)

// Operation is the type of operation to perform on a document.
type Operation int

const (
	// OperationCreate creates a document.
	OperationCreate Operation = iota
	// OperationUpdate updates a document.
	OperationUpdate
	// OperationDelete deletes a document.
	OperationDelete
)

// Cache is a cache of documents, used for primary access to documents.
//
// Writes/Deletes to the cache are queued and eventually written/removed to/from the storage
// via the txQueue.
//
// Reads from the cache are performed directly.
//
// To initialise a new Cache use the NewCache function which accepts
// a type of Queue.
type Cache struct {
	txQueue *Queue
	sync.RWMutex
	// documents is a slice of documents in the cache.
	documents []*document.Document
	// idx is a map of document ID to documents slice index
	idx map[string]int
}

// Put puts a document into the cache.
func (c *Cache) Put(d *document.Document, blackhole bool) error {
	// determine the operation
	op := OperationCreate
	if d := c.GetByID(d.ID.String()); d != nil {
		op = OperationUpdate
	}

	// push the event to the queue
	if !blackhole {
		c.txQueue.Push(Event{
			Operation: op,
			Document:  d,
		})
	}

	// update the cache
	if op == OperationCreate {
		return c.createDocument(d)
	}

	// update the document
	return c.updateDocument(d)
}

// createDocument creates a document in the cache. It will lock the cache
// and release it when the function returns.
func (c *Cache) createDocument(d *document.Document) error {
	c.Lock()
	defer c.Unlock()

	c.documents = append(c.documents, d)
	c.idx[d.ID.String()] = len(c.documents) - 1

	return nil
}

// updateDocument updates a document in the cache. It will lock the cache
// and release it when the function returns.
func (c *Cache) updateDocument(d *document.Document) error {
	c.Lock()
	defer c.Unlock()

	// get the document index
	idx, ok := c.idx[d.ID.String()]

	// if the document doesn't exist, return
	if !ok {
		return nil
	}

	// update the document
	c.documents[idx] = d

	return nil
}

// GetByID gets a document from the cache by ID. It will lock the cache
// and release it when the function returns.
func (c *Cache) GetByID(id string) *document.Document {
	c.RLock()
	defer c.RUnlock()

	idx, ok := c.idx[id]
	if !ok {
		return nil
	}

	return c.documents[idx]
}

// Delete deletes a document from the cache. It will lock the cache
// and release it when the function returns.
func (c *Cache) Delete(id string) error {
	c.Lock()
	defer c.Unlock()

	// grab the document index
	idx, ok := c.idx[id]
	if !ok {
		return nil
	}

	// get the document
	d := c.documents[idx]

	// create a copy to pass to the queue
	dCopy := document.New().SetCollection(d.Collection).SetID(d.ID.String()).SetData(d.Data)

	// push the event to the queue
	c.txQueue.Push(Event{
		Operation: OperationDelete,
		Document:  dCopy,
	})

	// delete the document from the slice
	c.documents = append(c.documents[:idx], c.documents[idx+1:]...)

	// update the index
	delete(c.idx, id)

	return nil
}

// NewCache returns a new cache.
func NewCache(ctx context.Context, txQueue *Queue) *Cache {
	c := &Cache{
		txQueue:   txQueue,
		documents: make([]*document.Document, 0),
		idx:       make(map[string]int),
	}

	go c.txQueue.Start(ctx)

	return c
}
