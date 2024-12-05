package client

import (
	"log"
	"net/http"
)

type RecordStorage struct {
}

func NewRecordStorage() RecordStorage {
	return RecordStorage{}
}

// ServeHTTP implements http.Handler.
func (r RecordStorage) ServeHTTP(http.ResponseWriter, *http.Request) {
	log.Println("asked to store record")
}
