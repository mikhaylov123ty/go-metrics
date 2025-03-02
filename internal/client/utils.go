package client

import (
	"github.com/go-resty/resty/v2"
)

// Структура для канала заданий метрик
type metricJob struct {
	data    *[]byte
	urlPath string
}

// Структура для канала ответов заданий метрик
type restyResponse struct {
	response *resty.Response
	err      error
	worker   int
}
