package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"metrics/internal/storage"

	"github.com/go-resty/resty/v2"
)

// Алиасы ручек
const (
	singleHandlerPath = "/update"
	batchHandlerPath  = "/updates"
)

// Структура агента
type Agent struct {
	baseURL        string
	client         *resty.Client
	pollInterval   int
	reportInterval int
	metrics        []*storage.Data
	statsBuf       statsBuf
	key            string
}

// Конструктор агента
func NewAgent(baseURL string, pollInterval int, reportInterval int, key string) *Agent {
	return &Agent{
		baseURL:        "http://" + baseURL,
		client:         resty.New(),
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		statsBuf:       collectMetrics(&Stats{}),
		key:            key,
	}
}

// Запуск агента
func (a *Agent) Run() {
	wg := &sync.WaitGroup{}

	// Запуск горутины по сбору метрик с интервалом pollInterval
	go func() {
		for {
			time.Sleep(time.Duration(a.pollInterval) * time.Second)

			a.metrics = a.statsBuf().buildMetrics()
		}
	}()

	// Запуск бесконечного цикла параллельной отправки метрики с интервалом reportInterval
	for {
		time.Sleep(time.Duration(a.reportInterval) * time.Second)

		wg.Add(2)

		go func() {
			a.sendMetrics()
			wg.Done()
		}()

		go func() {
			if err := a.sendMetricsBatch(); err != nil {
				log.Println("Send metrics batch err:", err)
			}
			wg.Done()
		}()

		wg.Wait()
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

// Обертка для запросов с подписью
func (a *Agent) withSign(request *resty.Request) *resty.Request {
	if a.key != "" {
		h := hmac.New(sha256.New, []byte(a.key))
		h.Write([]byte(fmt.Sprintf("%s", request.Body)))
		hash := hex.EncodeToString(h.Sum(nil))

		request.SetHeader("HashSHA256", hash)
	}

	return request
}

// Метод отправки запроса
func (a *Agent) postUpdates(handler string, data *[]byte) (*resty.Response, error) {
	URL := a.baseURL + handler

	// Формирование и выполнение запроса
	resp, err := a.withSign(a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(*data)).Post(URL)
	if err != nil {
		return nil, fmt.Errorf("post updates error: %w", err)
	}

	return resp, nil
}

// Метод отправки метрик
func (a *Agent) sendMetrics() {
	if len(a.metrics) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(a.metrics))

	// Запуск параллельной отправки метрик горутинами
	for _, metric := range a.metrics {
		go func(metric *storage.Data) {
			defer wg.Done()

			// Сериализация метрики
			data, err := json.Marshal(metric)
			if err != nil {
				return
			}

			// Передача метрики в функцию отправки с опцией повторения
			// при ошибках с подключением
			resp, err := sendFunc(a.postUpdates).withRetry(singleHandlerPath, &data)
			if err != nil {
				log.Printf("%s, metric: %v\n", err.Error(), metric)
				return
			}
			log.Printf("post update: metric: %v, URI: %s, Status Code: %d\n", metric, resp.Request.URL, resp.StatusCode())
		}(metric)
	}

	wg.Wait()
}

// Метод отправки метрик батчами
func (a *Agent) sendMetricsBatch() error {
	if len(a.metrics) == 0 {
		return nil
	}

	// Сериализация метрик
	data, err := json.Marshal(a.metrics)
	if err != nil {
		return err
	}

	// Передача метрик в функцию отправки с опцией повторения
	// при ошибках с подключением
	resp, err := sendFunc(a.postUpdates).withRetry(batchHandlerPath, &data)
	if err != nil {
		return err
	}

	log.Printf("post batch updates: metrics:  URI: %s, Status Code: %d", resp.Request.URL, resp.StatusCode())
	return nil
}
