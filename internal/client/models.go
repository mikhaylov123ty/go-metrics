package client

import (
	"math/rand/v2"
	"runtime"
)

type Stats map[string]interface{}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Метод сбора метрик
func collectMetrics(statsBuf *Stats) func() *Stats {
	counter := 1
	return func() *Stats {
		// Чтение метрик
		rt := &runtime.MemStats{}
		runtime.ReadMemStats(rt)

		// Присвоение полей для каждой метрики
		(*statsBuf)["Alloc"] = float64(rt.Alloc)
		(*statsBuf)["BuckHashSys"] = float64(rt.BuckHashSys)
		(*statsBuf)["Frees"] = float64(rt.Frees)
		(*statsBuf)["GCCPUFraction"] = float64(rt.GCCPUFraction)
		(*statsBuf)["GCSys"] = float64(rt.GCSys)
		(*statsBuf)["HeapAlloc"] = float64(rt.HeapAlloc)
		(*statsBuf)["HeapIdle"] = float64(rt.HeapIdle)
		(*statsBuf)["HeapInuse"] = float64(rt.HeapInuse)
		(*statsBuf)["HeapObjects"] = float64(rt.HeapObjects)
		(*statsBuf)["HeapReleased"] = float64(rt.HeapReleased)
		(*statsBuf)["HeapSys"] = float64(rt.HeapSys)
		(*statsBuf)["LastGC"] = float64(rt.LastGC)
		(*statsBuf)["Lookups"] = float64(rt.Lookups)
		(*statsBuf)["MCacheInuse"] = float64(rt.MCacheInuse)
		(*statsBuf)["MCacheSys"] = float64(rt.MCacheSys)
		(*statsBuf)["MSpanInuse"] = float64(rt.MSpanInuse)
		(*statsBuf)["MSpanSys"] = float64(rt.MSpanSys)
		(*statsBuf)["Mallocs"] = float64(rt.Mallocs)
		(*statsBuf)["NextGC"] = float64(rt.NextGC)
		(*statsBuf)["NumForcedGC"] = float64(rt.NumForcedGC)
		(*statsBuf)["NumGC"] = float64(rt.NumGC)
		(*statsBuf)["OtherSys"] = float64(rt.OtherSys)
		(*statsBuf)["PauseTotalNs"] = float64(rt.PauseTotalNs)
		(*statsBuf)["StackInuse"] = float64(rt.StackInuse)
		(*statsBuf)["StackSys"] = float64(rt.StackSys)
		(*statsBuf)["Sys"] = float64(rt.Sys)
		(*statsBuf)["TotalAlloc"] = float64(rt.TotalAlloc)

		// Генерация произвольного значения
		(*statsBuf)["RandomValue"] = rand.Float64()

		// Увеличение счетчика
		(*statsBuf)["PollCount"] = int64(counter)

		counter++

		return statsBuf
	}

}

func (s *Stats) buildMetrics() []*Metrics {
	res := []*Metrics{}
	for k, v := range *s {
		metric := Metrics{ID: k}
		switch t := v.(type) {
		case float64:
			metric.MType = "gauge"
			metric.Value = &t
		case int64:
			metric.MType = "counter"
			metric.Delta = &t
		}
		res = append(res, &metric)
	}
	return res
}
