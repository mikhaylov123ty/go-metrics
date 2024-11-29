package client

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"metrics/internal/storage"

	"github.com/go-resty/resty/v2"
)

// Структура агента
type Agent struct {
	baseURL        string
	client         *resty.Client
	pollInterval   int
	reportInterval int
	metrics        []*storage.Data
	statsBuf       statsBuf
}

// Конструктор агента
func NewAgent(baseURL string, pollInterval int, reportInterval int) *Agent {
	return &Agent{
		baseURL:        "http://" + baseURL,
		client:         resty.New(),
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		statsBuf:       collectMetrics(&Stats{}),
	}
}

// Запуск агента
func (a *Agent) Run() {
	// Запуск горутины по сбору метрик с интервалом pollInterval
	go func() {
		for {
			a.metrics = a.statsBuf().buildMetrics()

			time.Sleep(time.Duration(a.pollInterval) * time.Second)
		}
	}()

	// Запуск бесконечного цикла отправки метрики с интервалом reportInterval
	for {
		a.sendMetrics()

		time.Sleep(time.Duration(a.reportInterval) * time.Second)
	}
}

// Метод сбора метрик с счетчиком
func collectMetrics(statsBuf *Stats) statsBuf {
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

// Метод отправки запроса "POST /update"
func (a *Agent) postUpdate(metric *storage.Data) (*resty.Response, error) {
	URL := a.baseURL + "/update"

	// Формирования и выполнение запроса
	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(metric).
		Post(URL)
	if err != nil {
		return nil, fmt.Errorf("post update error: %w", err)
	}

	return resp, nil
}

// Метод отправки метрик
func (a *Agent) sendMetrics() {
	wg := sync.WaitGroup{}
	wg.Add(len(a.metrics))

	// Запуск параллельной отправки метрик горутинами
	for _, metric := range a.metrics {
		go func(metric *storage.Data) {
			defer wg.Done()
			resp, err := a.postUpdate(metric)
			if err != nil {
				log.Printf("%s, metric: %v", err.Error(), metric)
				return
			}
			log.Printf("post update: metric: %v, URI: %s, Status Code: %d", metric, resp.Request.URL, resp.StatusCode())
		}(metric)
	}
	wg.Wait()
}
