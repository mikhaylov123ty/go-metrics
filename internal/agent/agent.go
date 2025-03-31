// Модуль client реализует бизнес логику агента сбора метрик и передачу на сервер
package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"metrics/pkg"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"metrics/internal/agent/collector"
	"metrics/internal/agent/config"
	grpcClient "metrics/internal/agent/gRPC"
	httpClient "metrics/internal/agent/http"
	"metrics/internal/models"
	pb "metrics/internal/server/proto"
)

const (
	protocol = "http://"
	attempts = 3
	interval = 2 * time.Second
)

type UpdatesPoster interface {
	PostUpdates(context.Context, []byte) error
}

// Agent - структура агента
type Agent struct {
	client         UpdatesPoster
	pollInterval   float64
	reportInterval float64
	metrics        []*models.Data
	statsBuf       collector.StatsBuf
	rateLimit      int
	certFile       string
}

// NewAgent - конструктор агента
func NewAgent(cfg *config.AgentConfig) *Agent {
	var client UpdatesPoster

	if cfg.Host.GRPCPort != "" {
		interceptors := grpcClient.NewInterceptors(cfg.Key)
		conn, err := grpc.NewClient(":"+cfg.Host.GRPCPort, interceptors, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		client = grpcClient.New(pb.NewHandlersClient(conn), attempts, interval)
	} else {
		baseURL := protocol + cfg.Host.String()
		client = httpClient.New(resty.New(), baseURL, cfg.Key, attempts, interval)
	}

	return &Agent{
		client:         client,
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		statsBuf: collector.CollectMetrics(
			&collector.Stats{
				Data: make(map[string]interface{}),
			},
		),
		rateLimit: cfg.RateLimit,
		certFile:  cfg.CryptoKey,
	}
}

// Run запускает агента
func (a *Agent) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Создание каналов для связи горутин отправки метрик
	jobs := make(chan *metricJob)
	res := make(chan *jobResponse)

	// Запуск горутины по сбору метрик с интервалом pollInterval
	go func() {
		defer wg.Done()
		for {
			// Завершение горутины, если получен сигнал
			select {
			case <-ctx.Done():
				log.Println("Build Metrics Done")
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
	go func() {
		defer wg.Done()
		for {
			select {
			//Заверешине цикла и закрытые канала jobs,
			//если в очередной итерации был получен сигнал
			case <-ctx.Done():
				log.Println("Send Metrics Done")
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
				go func(res chan *jobResponse) {
					// Чтение результатов из результирующего канала по количеству заданий
					// В сценарии с отправкой батчем, количество = 1
					// В сценарии с отправкой каждой метрики по отдельности = кол-во отправок
					for range 1 {
						r := <-res
						if r.err != nil {
							log.Printf("Worker: %d, Failed sending metric: %s", r.worker, r.err.Error())
							continue
						}
						log.Printf("Worker: %d Metric sent", r.worker)
					}
				}(res)
			}
		}
	}()
	wg.Wait()
}

// Метод отправки запроса
func (a *Agent) postWorker(i int, jobs <-chan *metricJob, res chan<- *jobResponse) {
	var err error
	for {
		// Чтение заданий и проверка их наличия в канале
		data, ok := <-jobs
		if !ok {
			log.Printf("Worker %d finished", i)
			return
		}

		// Создание ответа для передачи в результирующий канал
		result := &jobResponse{
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

		// Передача id раннера в запрос
		ctx := context.WithValue(context.Background(), pkg.ContextKey{}, i)
		if err = a.client.PostUpdates(ctx, body); err != nil {
			log.Printf("Worker %d: PostUpdates failed: %s", i, err)
			result.err = fmt.Errorf("post updates failed: %w", err)
		}

		// Запись в результирующий канал
		res <- result
	}
}

// TODO depreciate data to struct
// Метод отправки метрик батчами
func (a *Agent) sendMetricsBatch(jobs chan<- *metricJob) error {
	// Сериализация метрик
	data, err := json.Marshal(a.metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics error: %w", err)
	}

	// Запись метрик в канал с заданиями

	jobs <- &metricJob{data: &data}

	return nil
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
