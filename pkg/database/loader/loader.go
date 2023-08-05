package loader

import (
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/storage"
)

// Loader is a loader of documents, used only for when the cache is empty,
// usually on startup.
type Loader struct {
	Cache   *cache.Cache
	Storage storage.Storage
}
