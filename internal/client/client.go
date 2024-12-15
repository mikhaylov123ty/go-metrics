package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"metrics/internal/client/config"
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
	rateLimit      int
}

// Конструктор агента
func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{
		baseURL:        "http://" + cfg.String(),
		client:         resty.New(),
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		statsBuf:       collectMetrics(&stats{mu: sync.RWMutex{}, data: make(map[string]interface{})}),
		key:            cfg.Key,
		rateLimit:      cfg.RateLimit,
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

		// Создание каналов для связи горутин
		jobs := make(chan *metricJob)
		res := make(chan *restyResponse)

		// Ограничение рабочих, которые выполняют одновременные запросы к серверу
		for range a.rateLimit {
			go a.postWorker(jobs, res)
		}

		wg.Add(2)

		// Запуск горутины для отправки запросов по каждой метрике
		go func() {
			if err := a.sendMetrics(jobs); err != nil {
				log.Println("send metric err:", err)
			}
			wg.Done()
		}()

		// Запуск горутины для отправки метрик батчем
		go func() {
			if err := a.sendMetricsBatch(jobs); err != nil {
				log.Println("Send metrics batch err:", err)
			}
			wg.Done()
		}()

		// Запуск горутины ожидания окончания всех отправок и закрытие канала
		go func() {
			wg.Wait()
			close(jobs)
		}()

		// Чтение результатов из результирующего канала по количеству заданий
		// количество всех метрик + 1 батч запрос
		for range len(a.metrics) + 1 {
			r := <-res
			if r.err != nil {
				log.Printf("Error sending metric: %s", r.err.Error())
				continue
			}
			log.Printf("Sent metric. Code: %d, URL: %s, Body: %s", r.response.StatusCode(), r.response.Request.URL, r.response.Request.Body)
		}
	}
}

// Метод отправки запроса
func (a *Agent) postWorker(jobs <-chan *metricJob, res chan<- *restyResponse) {
	// Чтение из канала с заданиями
	for data := range jobs {
		URL := a.baseURL + data.urlPath

		// Формирование и выполнение запроса
		resp, err := withRetry(a.withSign(a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(*data.data)), URL)

		// Создание ответа для передачи в результирующий канал
		result := &restyResponse{
			response: resp,
			err:      err,
		}

		// Запись в результирующий канал
		res <- result
	}
}

// Метод отправки метрик
func (a *Agent) sendMetrics(jobs chan<- *metricJob) error {
	if len(a.metrics) == 0 {
		return nil
	}

	// Запись каждой метрики в канал с заданиями
	for _, metric := range a.metrics {
		channelJob := &metricJob{urlPath: singleHandlerPath}

		// Сериализация метрики
		data, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("marshal metrics error: %w", err)
		}
		channelJob.data = &data

		jobs <- channelJob
	}

	return nil
}

// Метод отправки метрик батчами
func (a *Agent) sendMetricsBatch(jobs chan<- *metricJob) error {
	if len(a.metrics) == 0 {
		return nil
	}

	channelJob := &metricJob{urlPath: batchHandlerPath}

	// Сериализация метрик
	data, err := json.Marshal(a.metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics error: %w", err)
	}

	// Запись метрик в канал с заданиями
	channelJob.data = &data

	jobs <- channelJob
	return nil
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
