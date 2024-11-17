package api

import (
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

// Структура обертки компрессии gzip для интерфейса writer
type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Обертка метода Write для записи компрессированных сообщений
func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func ArrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
