package client

import (
	"github.com/go-resty/resty/v2"
)

// Структура для канала заданий метрик
type metricJob struct {
	Data    *[]byte
	URLPath string
}

// Структура для канала ответов заданий метрик
type restyResponse struct {
	Response *resty.Response
	Err      error
	Worker   int
}
