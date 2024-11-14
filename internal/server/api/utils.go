package api

import "net/http"

// Структура обертки для интерфейса writer
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
