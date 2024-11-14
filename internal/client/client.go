package client

import (
	"encoding/json"
	"log"
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
	statsBuf       func() *Stats
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
func (a *Agent) postUpdate(metric []byte) *resty.Response {
	URL := a.baseURL + "/update"

	// Формирования и выполнение запроса
	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(metric).
		Post(URL)
	if err != nil {
		log.Println("post update error:", err)
	}

	log.Println("Request Post Update", resp.Request.URL, string(resp.Request.Body.([]byte)))

	return resp
}

// Метод отправки метрик
func (a *Agent) sendMetrics() {
	for _, metric := range a.metrics {
		resp, err := json.Marshal(metric)
		if err != nil {
			log.Println("json marshal error:", err)
		}
		a.postUpdate(resp)
	}
}
