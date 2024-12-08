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
	services        services
	logger          *logrus.Logger
	storeInterval   int
	fileStoragePath string
	restore         bool
}

type services struct {
	storageCommands *api.StorageCommands
}

// Конструктор инстанса сервера
func New(storageCommands *api.StorageCommands, logger *logrus.Logger, storeInterval int, fileStoragePath string, restore bool) *Server {
	return &Server{
		services: services{
			storageCommands: storageCommands},
		logger:          logger,
		storeInterval:   storeInterval,
		fileStoragePath: fileStoragePath,
		restore:         restore,
	}
}

// Метод запуска сервера
func (s *Server) Start(address string) {

	// Инициализация даты из файла
	if s.restore {
		if err := s.initMetricsFromFile(); err != nil {
			s.logger.Fatal("error restore metrics from file: ", err)
		}
	}

	// Запуск горутины сохранения метрик с интервалом
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
	s.addHandlers(router, api.NewHandler(s.services.storageCommands))

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

	// /updates
	router.Route("/updates", func(r chi.Router) {
		r.Post("/", s.withGZipEncode(s.withLogger(handler.UpdatesPostJSON)))
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

// Метод записи метрик в файл
func (s *Server) storeMetrics() error {
	// Чтение всех метрик из хранилища
	metrics, err := s.services.storageCommands.ReadAll()
	if err != nil {
		return fmt.Errorf("read metrics: %w", err)
	}

	// Проверка наличия метрик
	if len(metrics) == 0 {
		s.logger.Infof("no metrics found")
		return nil
	}

	// Сериализация метрик
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

	// Создание/обновление файла
	storageFile, err := os.Create(s.fileStoragePath)
	if err != nil {
		return fmt.Errorf("error create file: %w", err)
	}

	// Запись даты в файл
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
		storageData := &storage.Data{}
		if err = json.Unmarshal([]byte(line), storageData); err != nil {
			return fmt.Errorf("unmarshal metrics: %w", err)
		}

		// Забись в хранилище
		if err = s.services.storageCommands.Update(storageData); err != nil {
			return fmt.Errorf("update metrics: %w", err)
		}
	}

	return nil
}
