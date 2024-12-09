package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// Структура обертки логирования для интерфейса writer
type LoggingResponseWriter struct {
	http.ResponseWriter
	ResponseData *ResponseData
}

// Структура даты ответа
type ResponseData struct {
	Status int
	Size   int
}

type HashResponseWriter struct {
	http.ResponseWriter
	key string
}

// Структура обертки компрессии gzip для интерфейса writer
type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

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

// Метод проверки строки в массиве строк
func ArrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getHash(key string, msg []byte) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(msg)

	return h.Sum(nil)
}
