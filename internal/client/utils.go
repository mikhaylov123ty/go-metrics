package client

import (
	"errors"
	"fmt"
	"log"
	"syscall"
	"time"

	"metrics/internal/storage"

	"github.com/go-resty/resty/v2"
)

const (
	attempts = 4
	interval = 2 * time.Second
)

type (
	// Вспомогательные типы для методов функций
	metricFunc func(*storage.Data) (*resty.Response, error)
	batchFunc  func() (*resty.Response, error)
	Stats      map[string]interface{}
	statsBuf   func() *Stats
)

// Метод повтора функции отправки метрик на сервер
func (mf metricFunc) withRetry(metric *storage.Data) (*resty.Response, error) {
	wait := 1 * time.Second

	// Попытки выполнения запроса и возврат при успешном выполнении
	for range attempts {
		resp, err := mf(metric)
		if err == nil {
			return resp, nil
		}

		// Проверка ошибки для сценария недоступности сервера
		switch {
		case errors.Is(err, syscall.ECONNREFUSED):
			log.Println("retrying after error:", err, "metric name", metric.Name)
			time.Sleep(wait)
			wait += interval

			// Возврат ошибки по умолчанию
		default:
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed after %d attempts", attempts)
}

// Метод повтора функции отправки батча метрик на сервер
func (bf batchFunc) withRetry() (*resty.Response, error) {
	wait := 1 * time.Second

	// Попытки выполнения запроса и возврат при успешном выполнении
	for range attempts {
		resp, err := bf()
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
	return nil, fmt.Errorf("failed after %d attempts", attempts)
}

// Метод конструктора метрик в структры
func (s *Stats) buildMetrics() []*storage.Data {
	res := []*storage.Data{}
	for k, v := range *s {
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
