package database

import (
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/storage"
)

// Database is a database of documents, it acts as a proxy to the cache.
type Database struct {
	*cache.Cache
}

// Load loads the database from storage.
func (d *Database) Load(store storage.Storage) error {
	str, err := store.Stream()
	if str == nil || err != nil {
		return nil
	}

	for doc := range str {
		if err := d.Put(doc, true); err != nil {
			return err
		}
	}

	return nil
}
