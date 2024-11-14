package api

import "metrics/internal/storage"

// Структуры запросов

// Структура для запроса "POST /update/{type}/{name}/{value}"
type updatePost struct {
	id   string
	data *storage.Data
}

// Структура для запроса "GET /update/{type}/{name}"
type valueGet struct {
	id string
}

// Структура для запроса "POST /update" с телом JSON
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
