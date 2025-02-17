// Модуль client реализует бизнес логику агента сбора метрик и передачу на сервер
package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

	"metrics/internal/client/collector"
	"metrics/internal/client/config"
	"metrics/internal/models"
)

const (
	singleHandlerPath = "/update"
	batchHandlerPath  = "/updates"

	attempts = 3
	interval = 2 * time.Second
)

// Agent - структура агента
type Agent struct {
	baseURL        string
	client         *resty.Client
	pollInterval   int
	reportInterval int
	metrics        []*models.Data
	statsBuf       collector.StatsBuf
	key            string
	rateLimit      int
}

// NewAgent - конструктор агента
func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{
		baseURL:        "http://" + cfg.String(),
		client:         resty.New(),
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		statsBuf: collector.CollectMetrics(&collector.Stats{
			Data: make(map[string]interface{})}),
		key:       cfg.Key,
		rateLimit: cfg.RateLimit,
	}
}

// Run запускает агента
func (a *Agent) Run() {
	// Запуск горутины по сбору метрик с интервалом pollInterval
	go func() {
		for {
			time.Sleep(time.Duration(a.pollInterval) * time.Second)
			a.metrics = a.statsBuf().BuildMetrics()
		}
	}()

	// Создание каналов для связи горутин отправки метрик
	jobs := make(chan *metricJob)
	res := make(chan *restyResponse)

	// Ограничение рабочих, которые выполняют одновременные запросы к серверу
	for i := range a.rateLimit {
		go a.postWorker(i, jobs, res)
	}

	// Запуск бесконечного цикла отправки метрики с интервалом reportInterval
	for {
		time.Sleep(time.Duration(a.reportInterval) * time.Second)

		// Проверка пустых батчей
		if len(a.metrics) == 0 {
			continue
		}

		// Запуск отправки метрик батчем
		if err := a.sendMetricsBatch(jobs); err != nil {
			log.Println("Send metrics batch err:", err)
		}

		// Запуск горутины чтения результирующего канала
		// В интерпретации задания, предполагается, что следующая отрпавка может быть выполнена,
		// не дожидаясь окончания предыдущей итерации.
		// Для ожидания достаточно запустить обычное чтение вне горутины.
		go func(res chan *restyResponse) {
			// Чтение результатов из результирующего канала по количеству заданий(горутин в цикле)
			for range 1 {
				r := <-res
				if r.err != nil {
					log.Printf("Worker: %d, Failed sending metric: %s", r.worker, r.err.Error())
					continue
				}
				log.Printf(" Worker: %d Metric sent, Code: %d, URL: %s, Body: %s\n", r.worker, r.response.StatusCode(), r.response.Request.URL, r.response.Request.Body)
			}
		}(res)
	}
}

// Метод отправки запроса
func (a *Agent) postWorker(i int, jobs <-chan *metricJob, res chan<- *restyResponse) {
	// Чтение из канала с заданиями
	for data := range jobs {
		URL := a.baseURL + data.urlPath

		// Создание ответа для передачи в результирующий канал
		result := &restyResponse{
			worker: i,
		}

		// Формирование и выполнение запроса
		result.response, result.err = withRetry(a.withSign(a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(*data.data)), URL, i)

		// Запись в результирующий канал
		res <- result
	}
}

// Метод отправки метрик батчами
func (a *Agent) sendMetricsBatch(jobs chan<- *metricJob) error {
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

// Метод повтора функции отправки метрик на сервер
func withRetry(request *resty.Request, URL string, w int) (*resty.Response, error) {
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
			log.Printf("Worker: %d, retrying after error: %s\n", w, err.Error())
			time.Sleep(wait)
			wait += interval

		// Возврат ошибки по умолчанию
		default:
			return nil, err
		}
	}

	return nil, err
}
