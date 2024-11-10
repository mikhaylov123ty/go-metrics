package server

import (
	"log"
	"net/http"
	"time"

	"metrics/internal/server/api"
	"metrics/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// Структура сервера
type Server struct {
	storage *storage.Storage
	logger  *logrus.Logger
}

// Конструктор инстанса сервера
func New(storage *storage.Storage, logger *logrus.Logger) *Server {
	return &Server{storage: storage, logger: logger}
}

// Метод запуска сервера
func (s *Server) Start(address string) {
	// Создание роутера
	router := chi.NewRouter()

	// Назначение соответствий хендлеров
	s.addHandlers(router, api.NewHandler(*s.storage))

	// Старт сервера
	s.logger.Infof("Starting server on %v", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatal(err)
	}
}

// Наполнение сервера методами хендлера
func (s *Server) addHandlers(router *chi.Mux, handler *api.Handler) {
	// /update
	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", s.WithLogger(handler.UpdatePost))
	})

	// /value
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", s.WithLogger(handler.ValueGet))
	})

	// index
	router.Get("/", s.WithLogger(handler.IndexGet))
}

// middleware для эндпоинтов для логирования
func (s *Server) WithLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &api.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData: &api.ResponseData{
				Status: 0,
				Size:   0,
			},
		}

		next(lw, r)

		s.logger.Infof("Incoming HTTP Request: URI: %s, Method: %v, Time Duration: %v", r.RequestURI, r.Method, time.Since(start))
		s.logger.Infof("Outgoing HTTP Response: Status Code: %v, Content Length:%v", lw.ResponseData.Status, lw.ResponseData.Size)
	}
}
