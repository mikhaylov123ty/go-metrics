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

	// /ping
	router.Get("/ping", s.withLogger(handler.PingGet))
}

// middleware для эндпоинтов для логирования
func (s *Server) withLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создание обертки для ResponseWriter
		lw := &api.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData: &api.ResponseData{
				Status: 0,
				Size:   0,
			},
		}

		// Переход к следующему хендлеру
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

		// Проверка хедеров
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

		s.logger.Debugln("compressing request with gzip")

		w.Header().Set("Content-Encoding", "gzip")

		next(api.GzipWriter{ResponseWriter: w, Writer: gz}, r)
	}
}

// Метод записи метрик в файл
func (s *Server) storeMetrics() error {
	// Чтение всех метрик из хранилища
	metrics, err := s.storage.ReadAll()
	if err != nil {
		return fmt.Errorf("read metrics: %w", err)
	}

	// Проверка наличия метрик
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

	s.logger.Debugf("data:%s:endData", string(data))

	storageFile, err := os.Create(s.fileStoragePath)
	if err != nil {
		return fmt.Errorf("error create file: %w", err)
	}

	if _, err = storageFile.Write(data); err != nil {
		return fmt.Errorf("write metrics: %w", err)
	}
	return nil
}

// Метод восстановления данных метрик из файла
func (s *Server) initMetricsFromFile() error {
	fileData, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		s.logger.Infof("no metrics file found, skipping restore")
		return nil
	}

	// Проверка, если файл пустой
	if len(fileData) < 1 {
		s.logger.Infof("no metrics found in file, skipping restore")
		return nil
	}

	// Разбивка по линиям файла
	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		//Десериализация в буфер
		bufData := map[string]any{}
		if err = json.Unmarshal([]byte(line), &bufData); err != nil {
			return fmt.Errorf("unmarshal metrics: %w", err)
		}

		// Конструктор хранилища даты
		storageData := &storage.Data{}
		storageData.Type = bufData["type"].(string)
		storageData.Name = bufData["name"].(string)
		dataID := storageData.UniqueID()

		// Форматирование типа значения
		if storageData.Type == "counter" {
			value := int64(bufData["value"].(float64))
			storageData.Value = &value
		} else {
			value := bufData["value"].(float64)
			storageData.Value = &value
		}

		// Забись в хранилище
		if err = s.storage.Update(dataID, storageData); err != nil {
			return fmt.Errorf("update metrics: %w", err)
		}
	}

	return nil
}
