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
	docs map[string]*document.Document
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

	c.docs[d.ID.String()] = d

	return nil
}

// updateDocument updates a document in the cache. It will lock the cache
// and release it when the function returns.
func (c *Cache) updateDocument(d *document.Document) error {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.docs[d.ID.String()]; !ok {
		return nil
	}

	// update the document
	c.docs[d.ID.String()] = d

	return nil
}

// GetByID gets a document from the cache by ID. It will lock the cache
// and release it when the function returns.
func (c *Cache) GetByID(id string) *document.Document {
	c.RLock()
	defer c.RUnlock()

	return c.docs[id]
}

// Delete deletes a document from the cache. It will lock the cache
// and release it when the function returns.
func (c *Cache) Delete(id string) error {
	c.Lock()
	defer c.Unlock()

	// get the document
	d := c.docs[id]

	// create a copy to pass to the queue
	dCopy := document.New().SetCollection(d.Collection).SetID(d.ID.String()).SetData(d.Data)

	// push the event to the queue
	c.txQueue.Push(Event{
		Operation: OperationDelete,
		Document:  dCopy,
	})

	// delete the document from the slice
	delete(c.docs, id)

	return nil
}

// NewCache returns a new cache.
func NewCache(ctx context.Context, txQueue *Queue) *Cache {
	c := &Cache{
		txQueue: txQueue,
		docs:    make(map[string]*document.Document),
	}

	go c.txQueue.Start(ctx)

	return c
}
