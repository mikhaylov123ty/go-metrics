// Модуль server реализует эндпоинты для взаимодействия и хранения метрик
package server

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"metrics/internal/server/api"
	"metrics/internal/server/config"
	"metrics/internal/server/metrics"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// Server - структура сервера
type Server struct {
	services       services
	logger         *logrus.Logger
	options        options
	privateKeyFile string
}

// services - структура команд БД и файла с бэкапом
type services struct {
	storageCommands    *api.StorageCommands
	metricsFileStorage *metrics.MetricsFileStorage
}

type options struct {
	storeInterval   int
	fileStoragePath string
	restore         bool
	key             string
}

// New - конструктор инстанса сервера
func New(storageCommands *api.StorageCommands, metricsFileStorage *metrics.MetricsFileStorage, logger *logrus.Logger, cfg *config.ServerConfig) *Server {
	return &Server{
		services: services{
			storageCommands:    storageCommands,
			metricsFileStorage: metricsFileStorage,
		},
		logger: logger,
		options: options{
			storeInterval:   cfg.FileStorage.StoreInterval,
			fileStoragePath: cfg.FileStorage.FileStoragePath,
			restore:         cfg.FileStorage.Restore,
			key:             cfg.Key,
		},
		privateKeyFile: cfg.PrivateKeyFile,
	}
}

// Start запускает сервера
func (s *Server) Start(address string) error {
	// Инициализация даты из файла
	if s.options.restore {
		if err := s.services.metricsFileStorage.InitMetricsFromFile(); err != nil {
			s.logger.Fatal("error restore metrics from file: ", err)
		}
		s.logger.Infof("metrics file storage restored")
	}

	// Запуск горутины сохранения метрик с интервалом
	go func() {
		for {
			time.Sleep(time.Duration(s.options.storeInterval) * time.Second)

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
	return http.ListenAndServe(address, router)
}

// Наполнение сервера методами хендлера
func (s *Server) addHandlers(router *chi.Mux, handler *api.Handler) {
	// /debug profiler
	router.Mount("/debug", middleware.Profiler())

	// /update
	router.Route("/update", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(s.withHash(s.withDecrypt(handler.UpdatePostJSON)))))
		r.Post("/{type}/{name}/{value}", s.withGZipEncode(s.withLogger(s.withHash(handler.UpdatePost))))
	})

	// /updates
	router.Route("/updates", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(s.withHash(s.withDecrypt(handler.UpdatesPostJSON)))))
	})

	// /value
	router.Route("/value", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(s.withHash(s.withDecrypt(handler.ValueGetJSON)))))
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

		defer func() {
			if err = gz.Close(); err != nil {
				log.Println("gZip middleware: failed close gZip writer", err)
			}
		}()

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
		if len(s.options.key) > 0 && len(requestHeader) > 0 {
			var body []byte
			// Чтение тела запроса, закрытие и копирование
			// для передачи далее по пайплайну
			body, err = io.ReadAll(r.Body)
			if err != nil {
				s.logger.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			defer func() {
				if err = r.Body.Close(); err != nil {
					log.Println("hash middleware: failed close request body", err)
				}
			}()

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Вычисление и валидация хэша
			hash := getHash(s.options.key, body)
			if !hmac.Equal(hash, requestHeader) {
				s.logger.Error("invalid hash")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hashWriter.key = s.options.key
		}

		next(hashWriter, r)
	}
}

func (s *Server) withDecrypt(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.privateKeyFile != "" {
			privatePEM, err := os.ReadFile(s.privateKeyFile)
			if err != nil {
				s.logger.Error("error reading tls private key", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			privateKeyBlock, _ := pem.Decode(privatePEM)
			privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
			if err != nil {
				s.logger.Error("error parsing tls private key", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var body []byte
			// Чтение тела запроса, закрытие и копирование
			// для передачи далее по пайплайну
			body, err = io.ReadAll(r.Body)
			if err != nil {
				s.logger.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			defer func() {
				if err = r.Body.Close(); err != nil {
					log.Println("Decode middleware: failed close request body", err)
				}
			}()

			var decryptedBytes []byte
			blockLen := privateKey.PublicKey.Size()

			for start := 0; start < len(body); start += blockLen {
				end := start + blockLen
				if start+blockLen > len(body) {
					end = len(body)
				}

				decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, body[start:end])
				fmt.Println("decrypted chunk:", string(decryptedChunk))
				if err != nil {
					s.logger.Errorf("error decrypting random text: %s", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				decryptedBytes = append(decryptedBytes, decryptedChunk...)
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBytes))
		}

		next(w, r)
	}
}
