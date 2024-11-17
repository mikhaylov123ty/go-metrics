package server

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"
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
		r.Post("/", s.WithGZipEncode(s.WithLogger(handler.UpdatePostJSON)))
		r.Post("/{type}/{name}/{value}", s.WithGZipEncode(s.WithLogger(handler.UpdatePost)))
	})

	// /value
	router.Route("/value", func(r chi.Router) {
		r.Post("/", s.WithGZipEncode(s.WithLogger(handler.ValueGetJSON)))
		r.Get("/{type}/{name}", s.WithGZipEncode(s.WithLogger(handler.ValueGet)))
	})

	// index
	router.Get("/", s.WithGZipEncode(s.WithLogger(handler.IndexGet)))
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

		s.logger.Infof("Incoming HTTP Request: URI: %s, Method: %v, Headers: %v, Time Duration: %v", r.RequestURI, r.Method, r.Header, time.Since(start))
		s.logger.Infof("Outgoing HTTP Response: Status Code: %v, Content Length:%v", lw.ResponseData.Status, lw.ResponseData.Size)
	}
}

// middleware для компрессии
func (s *Server) WithGZipEncode(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" && r.Header.Get("Accept") != "text/html" {
			s.logger.Infof("client accepts content is not json or html: %s", r.Header.Get("Accept"))
			next(w, r)
			return
		}

		headers := strings.Split(r.Header.Get("Accept-Encoding"), ",")
		if !api.ArrayContains(headers, "gzip") {
			next(w, r)
			return
		}

		// TODO find a way to reuse writer insead creating new one
		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			s.logger.Error("gZip encode error:", err)
		}

		defer gz.Close()

		s.logger.Infof("compressing request with gzip")

		w.Header().Set("Content-Encoding", "gzip")
		next(api.GzipWriter{ResponseWriter: w, Writer: gz}, r)
	}
}
