package server

import (
	"log"
	"metrics/internal/api"
	"metrics/internal/storage"
	"net/http"
)

type Server struct {
	port    string
	storage *storage.Storage
}

func New(port string, storage *storage.Storage) *Server {
	return &Server{port: port, storage: storage}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	handlers := api.NewHandlers(*s.storage)

	s.addHandlers(mux, handlers)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) addHandlers(mux *http.ServeMux, handlers *api.Handlers) {
	mux.HandleFunc("POST /update/{type}/{name}/{value}", handlers.Update)

}
