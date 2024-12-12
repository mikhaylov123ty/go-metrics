package client

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"log"
	"sync"
	"syscall"
	"time"

	"metrics/internal/storage"
)

const (
	attempts = 4
	interval = 2 * time.Second
)

type (
	// Вспомогательные типы для методов функций
	statsBuf func() *stats
)

type stats struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

type job struct {
	data    *[]byte
	urlPath string
}

// Метод повтора функции отправки метрик на сервер
func withRetry(request *resty.Request, URL string) (*resty.Response, error) {
	var resp *resty.Response
	var err error
	wait := 1 * time.Second

	// Попытки выполнения запроса и возврат при успешном выполнении
	for range attempts {
		resp, err = request.Post(URL)
		if err == nil {
			return resp, nil
		}

		// Проверка ошибки для сценария недоступности сервера
		switch {
		case errors.Is(err, syscall.ECONNREFUSED):
			log.Println("retrying after error:", err)
			time.Sleep(wait)
			wait += interval

			// Возврат ошибки по умолчанию
		default:
			return nil, err
		}
	}

	return nil, err
}

// Метод конструктора метрик в структры
func (s *stats) buildMetrics() []*storage.Data {
	res := []*storage.Data{}
	for k, v := range s.data {
		metric := storage.Data{Name: k}
		switch t := v.(type) {
		case float64:
			metric.Type = "gauge"
			metric.Value = &t
		case int64:
			metric.Type = "counter"
			metric.Delta = &t
		}
		res = append(res, &metric)
	}
	return res
}
