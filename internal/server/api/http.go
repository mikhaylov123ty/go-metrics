package api

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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"io"
	"log"
	"metrics/internal/server/utils"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type HTTPServer struct {
	auth   *auth
	Server *http.Server
	logger *logrus.Logger
}

type auth struct {
	cryptoKey     string
	hashKey       string
	trustedSubnet *net.IPNet
}

func NewServer(address string, cryptoKey string, hashKey string, trustedSubnet *net.IPNet, storageCommands *StorageCommands, logger *logrus.Logger) *HTTPServer {
	router := chi.NewRouter()

	instance := &HTTPServer{
		auth: &auth{
			cryptoKey:     cryptoKey,
			hashKey:       hashKey,
			trustedSubnet: trustedSubnet,
		},
		Server: &http.Server{
			Addr:    address,
			Handler: router,
		},
		logger: logger,
	}

	fmt.Println("AUTH", *instance.auth)

	// Назначение соответствий хендлеров
	instance.addHandlers(router, NewHandler(storageCommands))

	return instance
}

// Наполнение сервера методами хендлера
func (s *HTTPServer) addHandlers(router *chi.Mux, handler *Handler) {
	router.Use(
		middleware.RequestID,
		s.withLogger,
		s.withTrustedSubnet,
		s.withHash,
		s.withGZipEncode,
	)

	// /debug profiler
	router.Mount("/debug", middleware.Profiler())

	// /update
	router.Route("/update", func(r chi.Router) {
		r.Post("/", s.withDecrypt(handler.UpdatePostJSON))
		r.Post("/{type}/{name}/{value}", handler.UpdatePost)
	})

	// /updates
	router.Route("/updates", func(r chi.Router) {
		r.Post("/", s.withDecrypt(handler.UpdatesPostJSON))
	})

	// /value
	router.Route("/value", func(r chi.Router) {
		r.Post("/", s.withDecrypt(handler.ValueGetJSON))
		r.Get("/{type}/{name}", handler.ValueGet)
	})

	// index
	router.Get("/", handler.IndexGet)

	// /ping
	router.Get("/ping", handler.PingGet)
}

// middleware эндпоинтов для логирования
func (s *HTTPServer) withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создание обертки для ResponseWriter
		lw := &utils.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData: &utils.ResponseData{
				Status: 0,
				Size:   0,
			},
		}

		// Переход к следующему хендлеру
		next.ServeHTTP(lw, r)

		requestID, ok := r.Context().Value(middleware.RequestIDKey).(string)
		if !ok {
			requestID = "unknown"
		}

		s.logger.Infof("Incoming HTTP Request: URI: %s, Method: %v, Time Duration: %v, Request ID: %v", r.RequestURI, r.Method, time.Since(start), requestID)
		s.logger.Infof("Outgoing HTTP Response: Status Code: %v, Content Length:%v, Request ID: %v\"", lw.ResponseData.Status, lw.ResponseData.Size, requestID)
	})
}

// middleware эндпоинтов для компрессии
func (s *HTTPServer) withGZipEncode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" && r.Header.Get("Accept") != "text/html" {
			s.logger.Infof("client accepts content is not json or html: %s", r.Header.Get("Accept"))
			next.ServeHTTP(w, r)
			return
		}

		// Проверка хедеров
		headers := strings.Split(r.Header.Get("Accept-Encoding"), ",")
		if !utils.ArrayContains(headers, "gzip") {
			next.ServeHTTP(w, r)
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

		next.ServeHTTP(utils.GzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// middleware для эндпоинтов для хеширования
func (s *HTTPServer) withHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Декодирование хедера
		requestHeader, err := hex.DecodeString(r.Header.Get("HashSHA256"))
		if err != nil {
			s.logger.Error("error decoding hash header:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Создание обертки для ResponseWriter
		hashWriter := &utils.HashResponseWriter{
			ResponseWriter: w,
		}

		// Проверка наличия ключа из флага и в запросе
		if len(s.auth.hashKey) > 0 && len(requestHeader) > 0 {
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
			hash := utils.GetHash(s.auth.hashKey, body)
			if !hmac.Equal(hash, requestHeader) {
				s.logger.Error("invalid hash")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hashWriter.Key = s.auth.hashKey
		}

		next.ServeHTTP(hashWriter, r)
	})
}

// Middleware для дешифровки тела запроса
func (s *HTTPServer) withDecrypt(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO вынести чтение файла в инит конфига
		// Проверка флага приватнрго ключа
		if s.auth.cryptoKey != "" {
			// Чтение pem файла
			privatePEM, err := os.ReadFile(s.auth.cryptoKey)
			if err != nil {
				s.logger.Error("error reading private key", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// Поиск блока приватного ключа
			privateKeyBlock, _ := pem.Decode(privatePEM)
			// Парсинг приватного ключа
			privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
			if err != nil {
				s.logger.Error("error parsing private key", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err = privateKey.Validate(); err != nil {
				s.logger.Error("error validate private key", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Чтение тела запроса
			var body []byte
			body, err = io.ReadAll(r.Body)
			if err != nil {
				s.logger.Error("error reading body", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Отложенное закрытие тела
			defer func() {
				if err = r.Body.Close(); err != nil {
					log.Println("Decode middleware: failed close request body", err)
				}
			}()

			// Установка длины частей публичного ключа
			blockLen := privateKey.PublicKey.Size()

			// Дешифровка тела запроса частями
			var decryptedBytes []byte
			for start := 0; start < len(body); start += blockLen {
				end := start + blockLen
				if start+blockLen > len(body) {
					end = len(body)
				}

				var decryptedChunk []byte
				decryptedChunk, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, body[start:end])
				if err != nil {
					s.logger.Errorf("error decrypting random text: %s", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				decryptedBytes = append(decryptedBytes, decryptedChunk...)
			}

			// Подмена тела запроса
			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBytes))
		}

		next(w, r)
	}
}

func (s *HTTPServer) withTrustedSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.auth.trustedSubnet != nil {
			requestIP := net.ParseIP(r.Header.Get("X-Real-IP"))
			if requestIP == nil {
				s.logger.Errorln("error parsing X-Real-IP header")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if !s.auth.trustedSubnet.Contains(requestIP) {
				s.logger.Errorln("IP address is not trusted")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
