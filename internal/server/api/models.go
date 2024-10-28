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
