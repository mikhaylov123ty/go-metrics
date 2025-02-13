package client

import (
	"github.com/go-resty/resty/v2"
)

// Структура для канала заданий метрик
type MetricJob struct {
	Data    *[]byte
	URLPath string
}

// Структура для канала ответов заданий метрик
type RestyResponse struct {
	Response *resty.Response
	Err      error
	Worker   int
}
