package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"metrics/internal/server/api"
	"metrics/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// Структура сервера
type Server struct {
	storage         storage.Storage
	logger          *logrus.Logger
	storeInterval   int
	fileStoragePath string
	restore         bool
}

// Конструктор инстанса сервера
func New(storage storage.Storage, logger *logrus.Logger, storeInterval int, fileStoragePath string, restore bool) *Server {
	return &Server{
		storage:         storage,
		logger:          logger,
		storeInterval:   storeInterval,
		fileStoragePath: fileStoragePath,
		restore:         restore,
	}
}

// Метод запуска сервера
func (s *Server) Start(address string) {

	if s.restore {
		if err := s.initMetricsFromFile(); err != nil {
			s.logger.Fatal("error restore metrics from file: ", err)
		}
	}

	go func() {
		for {
			time.Sleep(time.Duration(s.storeInterval) * time.Second)

			if err := s.storeMetrics(); err != nil {
				s.logger.Errorf("store metrics: failed read metrics: %s", err.Error())
			}
		}
	}()

	// Создание роутера
	router := chi.NewRouter()

	// Назначение соответствий хендлеров
	s.addHandlers(router, api.NewHandler(s.storage))

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
		r.Post("/", s.withGZipEncode(s.withLogger(handler.UpdatePostJSON)))
		r.Post("/{type}/{name}/{value}", s.withGZipEncode(s.withLogger(handler.UpdatePost)))
	})

	// /value
	router.Route("/value", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(handler.ValueGetJSON)))
		r.Get("/{type}/{name}", s.withGZipEncode(s.withLogger(handler.ValueGet)))
	})

	// index
	router.Get("/", s.withGZipEncode(s.withLogger(handler.IndexGet)))
}

// middleware для эндпоинтов для логирования
func (s *Server) withLogger(next http.HandlerFunc) http.HandlerFunc {
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
func (s *Server) withGZipEncode(next http.HandlerFunc) http.HandlerFunc {
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

		// TODO find a way to reuse writer instead creating new one
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

func (s *Server) storeMetrics() error {
	metrics, err := s.storage.ReadAll()
	if err != nil {
		return fmt.Errorf("read metrics: %w", err)
	}

	if len(metrics) == 0 {
		s.logger.Infof("no metrics found")
		return nil
	}

	data := []byte{}
	metricsLength := len(metrics) - 1
	for i, v := range metrics {
		record, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal metrics: %w", err)
		}
		data = append(data, record...)
		if i < metricsLength {
			data = append(data, '\n')
			i++
		}
	}
	s.logger.Infof("data: %s:endData", string(data))

	storageFile, err := os.Create(s.fileStoragePath)
	if err != nil {
		return fmt.Errorf("error create file: %w", err)
	}

	if _, err = storageFile.Write(data); err != nil {
		return fmt.Errorf("write metrics: %w", err)
	}
	return nil
}

// TODO make custom marshaller
func (s *Server) initMetricsFromFile() error {
	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		s.logger.Infof("no metrics file found, skipping restore")
		return nil
	}
	fmt.Println("DATA", string(data), "END DATA")
	arrData := strings.Split(string(data), "\n")
	fmt.Println("ARRDATA", arrData)
	for _, v := range arrData {
		storageData := &storage.Data{}
		if err = json.Unmarshal([]byte(v), storageData); err != nil {
			return fmt.Errorf("unmarshal metrics: %w", err)
		}
		dataID := storageData.UniqueID()

		newStorageData := &storage.Data{}
		newStorageData.Type = storageData.Type
		newStorageData.Name = storageData.Name

		if storageData.Type == "counter" {
			value := int64(storageData.Value.(float64))
			newStorageData.Value = &value

		} else {
			value := storageData.Value.(float64)
			newStorageData.Value = &value
		}
		if err = s.storage.Update(dataID, newStorageData); err != nil {
			return fmt.Errorf("update metrics: %w", err)
		}
	}

	return nil
}
