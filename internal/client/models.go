package client

import (
	"metrics/internal/storage"
)

// Алиасы для типов
type (
	Stats    map[string]interface{}
	statsBuf func() *Stats
)

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
