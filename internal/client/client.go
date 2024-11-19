package client

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// Структура агента
type Agent struct {
	baseURL        string
	client         *resty.Client
	pollInterval   int
	reportInterval int
	metrics        []*Metrics
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

// Метод отправки запроса "POST /update"
func (a *Agent) postUpdate(metric *Metrics) (*resty.Response, error) {
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
		go func(metric *Metrics) {
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
