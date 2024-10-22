package server

import (
	"log"
	"net/http"

	"metrics/internal/api"
	"metrics/internal/storage"
)

// Структура сервера
type Server struct {
	storage *storage.Storage
}

// Конструктор инстанса сервера
func New(storage *storage.Storage) *Server {
	return &Server{storage: storage}
}

// Метод запуска сервера
func (s *Server) Start(port string) {
	mux := http.NewServeMux()

	handler := api.NewHandler(*s.storage)

	s.addHandlers(mux, handler)

	// Старт сервера
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

// Наполнение сервера методами хендлера
func (s *Server) addHandlers(mux *http.ServeMux, handler *api.Handler) {
	mux.HandleFunc("POST /update/{type}/{name}/{value}", handler.Update)

}
