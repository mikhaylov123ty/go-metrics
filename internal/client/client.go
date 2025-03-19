// Модуль client реализует бизнес логику агента сбора метрик и передачу на сервер
package client

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
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
	pollInterval   float64
	reportInterval float64
	metrics        []*models.Data
	statsBuf       collector.StatsBuf
	key            string
	rateLimit      int
	certFile       string
}

// NewAgent - конструктор агента
func NewAgent(cfg *config.AgentConfig) *Agent {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(cfg.CryptoKey))
	return &Agent{
		baseURL:        "http://" + cfg.String(),
		client:         resty.New(),
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		statsBuf: collector.CollectMetrics(&collector.Stats{
			Data: make(map[string]interface{})}),
		key:       cfg.Key,
		rateLimit: cfg.RateLimit,
		certFile:  cfg.CryptoKey,
	}
}

// Run запускает агента
func (a *Agent) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Создание каналов для связи горутин отправки метрик
	jobs := make(chan *metricJob)
	res := make(chan *restyResponse)

	// Запуск горутины по сбору метрик с интервалом pollInterval
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Build Metrics Done")
				return
			default:
				time.Sleep(time.Duration(a.pollInterval) * time.Second)
				a.metrics = a.statsBuf().BuildMetrics()
			}
		}
	}()

	// Ограничение рабочих, которые выполняют одновременные запросы к серверу
	for i := range a.rateLimit {
		go a.postWorker(i, jobs, res)
	}

	// Запуск бесконечного цикла отправки метрики с интервалом reportInterval
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Send Metrics Done")
			close(jobs)
			return
		default:
			time.Sleep(time.Duration(a.reportInterval) * time.Second)

			// Проверка пустых батчей
			if len(a.metrics) == 0 {
				continue
			}

			// Запуск задания отправки метрик батчем
			if err := a.sendMetricsBatch(jobs); err != nil {
				log.Println("Send metrics batch err:", err)
			}

			// Запуск горутины чтения результирующего канала
			// В интерпретации задания, предполагается, что следующая отрпавка может быть выполнена,
			// не дожидаясь окончания предыдущей итерации.
			// Для ожидания достаточно запустить обычное чтение вне горутины.
			go func(res chan *restyResponse) {
				// Чтение результатов из результирующего канала по количеству заданий
				// В сценарии с отправкой батчем, количество = 1
				// В сценарии с отправкой каждой метрики по отдельности = кол-во отправок
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

}

// Метод отправки запроса
func (a *Agent) postWorker(i int, jobs <-chan *metricJob, res chan<- *restyResponse) {
	var err error
	for {
		data, ok := <-jobs
		if !ok {
			return
		}

		// Чтение из канала с заданиями
		URL := a.baseURL + data.urlPath

		// Создание ответа для передачи в результирующий канал
		result := &restyResponse{
			worker: i,
		}

		// Обработка тела запроса
		var body []byte
		body, err = a.encryptRequest(*data.data)
		if err != nil {
			result.err = fmt.Errorf("encrypt failed: %w", err)
			res <- result
			continue
		}

		// Формирование и выполнение запроса
		result.response, result.err = withRetry(a.withSign(a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(body)), URL, i)

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

// Middleware для запросов с подписью
func (a *Agent) withSign(request *resty.Request) *resty.Request {
	if a.key != "" {
		h := hmac.New(sha256.New, []byte(a.key))
		h.Write([]byte(fmt.Sprintf("%s", request.Body)))
		hash := hex.EncodeToString(h.Sum(nil))

		request.SetHeader("HashSHA256", hash)
	}

	return request
}

// Middleware повтора функции отправки метрик на сервер
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

// Шифрует тело запроса при наличии флага сертификата
func (a *Agent) encryptRequest(body []byte) ([]byte, error) {
	// Пропуск обработки, если флаг не задан
	if a.certFile == "" {
		return body, nil
	}

	// Чтение pem файла
	certPEM, err := os.ReadFile(a.certFile)
	if err != nil {
		return nil, fmt.Errorf("error reading tls public key: %w", err)
	}
	// Поиск блока сертификата
	pubKeyBlock, _ := pem.Decode(certPEM)
	// Парсинг сертификата
	parsedCert, err := x509.ParseCertificate(pubKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing tls public key: %w", err)
	}
	// Присвоение публичного ключа
	pubKey := parsedCert.PublicKey

	// Проверка срока истечения сертификата
	if parsedCert.NotAfter.Before(time.Now()) {
		return nil, fmt.Errorf("tls public key expired")
	}

	// Установка длины частей шифрования
	blockLen := pubKey.(*rsa.PublicKey).Size() - 11

	// Шифрование тела запроса частями
	var encryptedBytes []byte
	for start := 0; start < len(body); start += blockLen {
		end := start + blockLen
		if start+blockLen > len(body) {
			end = len(body)
		}

		var encryptedChunk []byte
		encryptedChunk, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), body[start:end])
		if err != nil {
			return nil, fmt.Errorf("error encrypting random text: %w", err)
		}

		encryptedBytes = append(encryptedBytes, encryptedChunk...)
	}

	return encryptedBytes, nil
}
