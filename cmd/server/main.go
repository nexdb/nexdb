package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/nexdb/nexdb/pkg/database"
	"github.com/nexdb/nexdb/pkg/database/cache"
	"github.com/nexdb/nexdb/pkg/document"
	"github.com/nexdb/nexdb/pkg/handlers"
	"github.com/nexdb/nexdb/pkg/services/auth"
	"github.com/nexdb/nexdb/pkg/services/reader"
	"github.com/nexdb/nexdb/pkg/services/writer"
	"github.com/nexdb/nexdb/pkg/storage"

	"github.com/gorilla/mux"
)

var (
	db        *database.Database
	dbCache   *cache.Cache
	queue     *cache.Queue
	store     storage.Storage
	wr        *writer.Writer
	readerSvc *reader.Reader
	authSvc   *auth.AuthService
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get storage driver from env
	storageDriver := storage.MemoryDriver
	driver := os.Getenv("NEXDB_STORAGE_DRIVER")
	if driver != "" {
		switch driver {
		case "aws-s3":
			storageDriver = storage.AWS3Driver
		}
	}

	// create storage driver
	var err error
	store, err = storage.New(storageDriver)
	if err != nil {
		log.Fatal(err)
	}

	// queue
	queue = cache.NewQueue(store)
	go queue.Start(ctx)

	// cache
	dbCache = cache.NewCache(ctx, queue)

	// database
	db = &database.Database{Cache: dbCache}

	// writer
	wr = writer.New(db)

	// reader service
	readerSvc = reader.New(db)

	// auth service
	authSvc = auth.New(db)
	// << end services setup >>

	// << start database setup >>
	// load the database from storage
	if err := db.Load(store); err != nil {
		log.Fatal(err)
	}

	// check if we have any api keys in the database
	if err := initilaiseAuthentication(); err != nil {
		log.Fatal(err)
	}
	// << end database setup >>

	// << start router setup >>
	r := mux.NewRouter()
	r.HandleFunc("/v1/collections/{collection}", handlers.WriteDocument(wr).ServeHTTP).Methods("PUT")
	r.HandleFunc("/v1/collections/{collection}/{id}", handlers.GetDocument(readerSvc).ServeHTTP).Methods("GET")
	r.HandleFunc("/v1/collections/{collection}/{id}", handlers.DeleteDocument(wr).ServeHTTP).Methods("DELETE")
	r.HandleFunc("/v1/collections/{collection}", handlers.SearchDocuments(readerSvc).ServeHTTP).Methods("POST")

	// << start middleware setup >>
	authMiddleware := &handlers.AuthMiddleware{AuthService: authSvc}
	r.Use(authMiddleware.IsAuthenticated)
	// << end middleware setup >>

	// << end router setup >>

	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatal(err)
	}

	// wait for the queue to drain
	queue.WaitForShutdown()
}

func initilaiseAuthentication() error {
	apiKeys := db.Filter("_api_keys", cache.Query{})
	if len(apiKeys) == 0 {
		if apiKey := os.Getenv("NEXDB_API_KEY"); apiKey != "" {
			doc := document.New().SetCollection("_api_keys").SetData(map[string]interface{}{
				"key": apiKey,
			})

			err := db.Put(doc, false)
			if err != nil {
				return errors.New("failed to add api key to database")
			}

			return nil
		}
		return errors.New("no api keys found in database, please add one")
	}

	return nil
}
