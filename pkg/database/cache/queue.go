package cache

import (
	"context"
	"sync"

	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/storage"
)

// Event is an event that is emitted when a document is written to storage.
type Event struct {
	// Operation is the operation that was performed on the document.
	Operation Operation
	// Document is the document that was written to storage.
	Document *document.Document
}

// Queue is a queue of events to be processed by the cache
// so they can eventually be written to the storage.
//
// If the maximum number of attempts is reached, the event is discarded.
//
// The Queue will finish processing all events when receiving
// cancel signal and stop safely.
type Queue struct {
	Storage storage.Storage

	// queue is the queue of events to be processed.
	queue chan Event
	sync.RWMutex
	// draining is a flag that indicates whether the queue is draining.
	draining bool
	// drained is a channel that is closed when the queue is drained.
	drained chan struct{}
}

// Push pushes a write event to the queue.
func (q *Queue) Push(event Event) {
	q.RLock()
	defer q.RUnlock()
	if q.draining {
		return
	}

	q.queue <- event
}

// Start starts the queue.
func (q *Queue) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			q.Lock()
			q.draining = true
			q.drained = make(chan struct{})
			q.Unlock()

			// drain the queue
			close(q.queue)
			for event := range q.queue {
				q.process(event)
			}

			// close the drained channel
			q.drained <- struct{}{}
			close(q.drained)
			return
		case event := <-q.queue:
			q.process(event)
		}
	}
}

// process processes an event.
func (q *Queue) process(event Event) {
	switch event.Operation {
	case OperationCreate:
		q.processCreate(event.Document)
	case OperationUpdate:
		q.processCreate(event.Document)
	case OperationDelete:
		q.processDelete(event.Document)
	}
}

// processCreate processes a create event.
func (q *Queue) processCreate(doc *document.Document) {
	_ = q.Storage.Write(doc)
}

// processDelete processes a delete event.
func (q *Queue) processDelete(doc *document.Document) {
	_ = q.Storage.Delete(doc)
}

// WaitForShutdown waits for the queue to finish processing all events.
func (q *Queue) WaitForShutdown() {
	for {
		q.RLock()
		if q.drained != nil {
			q.RUnlock()
			break
		}
		q.RUnlock()
	}

	<-q.drained
}

// NewQueue returns a new queue.
func NewQueue(storage storage.Storage) *Queue {
	return &Queue{
		Storage: storage,
		queue:   make(chan Event),
	}
}
