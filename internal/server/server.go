package server

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"encoding/hex"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"metrics/internal/server/api"
	"metrics/internal/server/config"
	"metrics/internal/server/metrics"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// Структура сервера
type Server struct {
	services        services
	logger          *logrus.Logger
	storeInterval   int
	fileStoragePath string
	restore         bool
	key             string
}

type services struct {
	storageCommands    *api.StorageCommands
	metricsFileStorage *metrics.MetricsFileStorage
}

// Конструктор инстанса сервера
func New(storageCommands *api.StorageCommands, metricsFileStorage *metrics.MetricsFileStorage, logger *logrus.Logger, cfg *config.ServerConfig) *Server {
	return &Server{
		services: services{
			storageCommands:    storageCommands,
			metricsFileStorage: metricsFileStorage,
		},
		logger:          logger,
		storeInterval:   cfg.FileStorage.StoreInterval,
		fileStoragePath: cfg.FileStorage.FileStoragePath,
		restore:         cfg.FileStorage.Restore,
		key:             cfg.Key,
	}
}

// Метод запуска сервера
func (s *Server) Start(address string) {
	// Инициализация даты из файла
	if s.restore {
		if err := s.services.metricsFileStorage.InitMetricsFromFile(); err != nil {
			s.logger.Fatal("error restore metrics from file: ", err)
		}
		s.logger.Infof("metrics file storage restored")
	}

	// Запуск горутины сохранения метрик с интервалом
	go func() {
		for {
			time.Sleep(time.Duration(s.storeInterval) * time.Second)

			if err := s.services.metricsFileStorage.StoreMetrics(); err != nil {
				s.logger.Errorf("store metrics: failed read metrics: %s", err.Error())
			}
		}
	}()

	// Создание роутера
	router := chi.NewRouter()

	// Назначение соответствий хендлеров
	s.addHandlers(router, api.NewHandler(s.services.storageCommands))

	// Старт сервера
	s.logger.Infof("Starting server on %v", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatal(err)
	}
}

// Наполнение сервера методами хендлера
func (s *Server) addHandlers(router *chi.Mux, handler *api.Handler) {
	// /debug profiler
	router.Mount("/debug", middleware.Profiler())

	// /update
	router.Route("/update", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(s.withHash(handler.UpdatePostJSON))))
		r.Post("/{type}/{name}/{value}", s.withGZipEncode(s.withLogger(s.withHash(handler.UpdatePost))))
	})

	// /updates
	router.Route("/updates", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(s.withHash(handler.UpdatesPostJSON))))
	})

	// /value
	router.Route("/value", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(s.withHash(handler.ValueGetJSON))))
		r.Get("/{type}/{name}", s.withGZipEncode(s.withLogger(s.withHash(handler.ValueGet))))
	})

	// index
	router.Get("/", s.withGZipEncode(s.withLogger(s.withHash(handler.IndexGet))))

	// /ping
	router.Get("/ping", s.withLogger(handler.PingGet))
}

// middleware эндпоинтов для логирования
func (s *Server) withLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создание обертки для ResponseWriter
		lw := &LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData: &ResponseData{
				Status: 0,
				Size:   0,
			},
		}

		// Переход к следующему хендлеру
		next(lw, r)

		s.logger.Infof("Incoming HTTP Request: URI: %s, Method: %v, Time Duration: %v", r.RequestURI, r.Method, time.Since(start))
		s.logger.Infof("Outgoing HTTP Response: Status Code: %v, Content Length:%v", lw.ResponseData.Status, lw.ResponseData.Size)
	}
}

// middleware эндпоинтов для компрессии
func (s *Server) withGZipEncode(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" && r.Header.Get("Accept") != "text/html" {
			s.logger.Infof("client accepts content is not json or html: %s", r.Header.Get("Accept"))
			next(w, r)
			return
		}

		// Проверка хедеров
		headers := strings.Split(r.Header.Get("Accept-Encoding"), ",")
		if !ArrayContains(headers, "gzip") {
			next(w, r)
			return
		}

		// TODO find a way to reuse writer instead creating new one
		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			s.logger.Error("gZip encode error:", err)
		}
		defer gz.Close()

		s.logger.Debugln("compressing request with gzip")

		w.Header().Set("Content-Encoding", "gzip")

		next(GzipWriter{ResponseWriter: w, Writer: gz}, r)
	}
}

// middleware для эндпоинтов для хеширования и подписи
func (s *Server) withHash(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//Декодирование хедера
		requestHeader, err := hex.DecodeString(r.Header.Get("HashSHA256"))
		if err != nil {
			s.logger.Error("error decoding hash header:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Создание обертки для ResponseWriter
		hashWriter := &HashResponseWriter{
			ResponseWriter: w,
		}

		// Проверка наличия ключа из флага и в запросе
		if len(s.key) > 0 && len(requestHeader) > 0 {
			// Чтение тела запроса, закрытие и копирование
			// для передачи далее по пайплайну
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Вычисление и валидация хэша
			hash := getHash(s.key, body)
			if !hmac.Equal(hash, requestHeader) {
				s.logger.Error("invalid hash")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hashWriter.key = s.key
		}

		next(hashWriter, r)
	}
}
