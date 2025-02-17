package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// LoggingResponseWriter - обертка логирования для интерфейса writer
type LoggingResponseWriter struct {
	http.ResponseWriter
	ResponseData *ResponseData
}

// ResponseData - структура даты ответа
type ResponseData struct {
	Status int
	Size   int
}

// HashResponseWriter - обертка для хэширования
type HashResponseWriter struct {
	http.ResponseWriter
	key string
}

// GzipWriter - обертка компрессии gzip для интерфейса writer
type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Обертка метода Write для хеширования ответа и записи в хедер
func (w *HashResponseWriter) Write(b []byte) (int, error) {
	if w.key != "" {
		hash := getHash(w.key, b)
		w.ResponseWriter.Header().Set("HashSHA256", hex.EncodeToString(hash))
	}

	size, err := w.ResponseWriter.Write(b)

	return size, err
}

// Обертка метода WriteHeader для дублирования данных в структуру ответа
func (w *LoggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseData.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Обертка метода Write для дублирования данных в структуру ответа
func (w *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.ResponseData.Size += size

	return size, err
}

// Обертка метода Write для записи компрессированных сообщений
func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// ArrayContains проверяет наличие строки в массиве строк
func ArrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// Метод создания хэша из сообщения и подписи ключом
func getHash(key string, msg []byte) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(msg)

	return h.Sum(nil)
}
