package client

import (
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// Временное решение до реализации конфига
const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

// Структура агента
type Agent struct {
	BaseURL string
	Client  *resty.Client
	Stats   Stats
}

// Конструктор агента
func NewAgent(baseURL string) *Agent {
	return &Agent{
		BaseURL: baseURL,
		Client:  resty.New(),
	}
}

// Запуск агента
func (a *Agent) Run() {
	// Запуск горутины по сбору метрик с интервалом pollInterval
	go func() {
		for {
			a.collectMetrics()

			time.Sleep(pollInterval)
		}
	}()

	// Запуск бесконечного цикла отправки метрики с интервалом reportInterval
	for {
		a.sendMetrics()

		time.Sleep(reportInterval)
	}
}

// Метод отправки запроса "POST /update/{type}/{name}/{value}"
func (a *Agent) postUpdate(metricType string, metricName string, metricValue string) *resty.Response {
	URL := a.BaseURL + "update/" + metricType + "/" + metricName + "/" + metricValue

	// Формирования и выполнение запроса
	resp, err := a.Client.R().
		SetHeader("Content-Type", "text/plain").
		Post(URL)
	if err != nil {
		log.Println("post update error:", err)
	}

	log.Println("Request Post Update", URL)

	return resp
}

// Метод сбора метрик
func (a *Agent) collectMetrics() {
	// Чтение метрик
	rt := &runtime.MemStats{}
	runtime.ReadMemStats(rt)

	// Присвоение полей для каждой метрики
	a.Stats.Gauge.Alloc = float64(rt.Alloc)
	a.Stats.Gauge.BuckHashSys = float64(rt.BuckHashSys)
	a.Stats.Gauge.Frees = float64(rt.Frees)
	a.Stats.Gauge.GCCPUFraction = float64(rt.GCCPUFraction)
	a.Stats.Gauge.GCSys = float64(rt.GCSys)
	a.Stats.Gauge.HeapAlloc = float64(rt.HeapAlloc)
	a.Stats.Gauge.HeapIdle = float64(rt.HeapIdle)
	a.Stats.Gauge.HeapInuse = float64(rt.HeapInuse)
	a.Stats.Gauge.HeapObjects = float64(rt.HeapObjects)
	a.Stats.Gauge.HeapReleased = float64(rt.HeapReleased)
	a.Stats.Gauge.HeapSys = float64(rt.HeapSys)
	a.Stats.Gauge.LastGC = float64(rt.LastGC)
	a.Stats.Gauge.Lookups = float64(rt.Lookups)
	a.Stats.Gauge.MCacheInuse = float64(rt.MCacheInuse)
	a.Stats.Gauge.MCacheSys = float64(rt.MCacheSys)
	a.Stats.Gauge.MSpanInuse = float64(rt.MSpanInuse)
	a.Stats.Gauge.MSpanSys = float64(rt.MSpanSys)
	a.Stats.Gauge.Mallocs = float64(rt.Mallocs)
	a.Stats.Gauge.NextGC = float64(rt.NextGC)
	a.Stats.Gauge.NumForcedGC = float64(rt.NumForcedGC)
	a.Stats.Gauge.NumGC = float64(rt.NumGC)
	a.Stats.Gauge.OtherSys = float64(rt.OtherSys)
	a.Stats.Gauge.PauseTotalNs = float64(rt.PauseTotalNs)
	a.Stats.Gauge.StackInuse = float64(rt.StackInuse)
	a.Stats.Gauge.StackSys = float64(rt.StackSys)
	a.Stats.Gauge.Sys = float64(rt.Sys)
	a.Stats.Gauge.TotalAlloc = float64(rt.TotalAlloc)

	// Генерация произвольного значения
	a.Stats.Gauge.RandomValue = rand.Float64()

	// Увеличение счетчика
	a.Stats.Counter.PollCount++
}

// Метод отправки метрик
func (a *Agent) sendMetrics() {
	stats, err := a.Stats.Map()
	if err != nil {
		log.Println("error convert stats to map: ", err)
	}

	for types, typesData := range stats {
		for k, v := range typesData.(map[string]interface{}) {
			vStr := strconv.FormatFloat(v.(float64), 'f', -1, 64)
			a.postUpdate(types, k, vStr)
		}
	}
}
