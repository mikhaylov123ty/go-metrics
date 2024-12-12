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
	"strconv"
	"sync"
	"time"

	"metrics/internal/client/config"
	"metrics/internal/storage"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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
		jobs := make(chan *job)
		res := make(chan *restyResponse)

		// Ограничение рабочих, которые выполняют запрос к серверу
		for range a.rateLimit {
			go a.postWorker(jobs, res)
		}

		wg.Add(2)

		go func() {
			if err := a.sendMetrics(jobs); err != nil {
				log.Println("send metric err:", err)
			}
			wg.Done()
		}()

		go func() {
			if err := a.sendMetricsBatch(jobs); err != nil {
				log.Println("Send metrics batch err:", err)
			}
			wg.Done()
		}()

		go func() {
			wg.Wait()
			close(jobs)
		}()

		for range len(a.metrics) + 1 {
			if r := <-res; r.err != nil {
				log.Printf("Error sending metric: %s", r.err.Error())
			}
		}
		fmt.Println("END CYCLE")
	}
}

// Метод сбора метрик с счетчиком
func collectMetrics(statsBuf *stats) statsBuf {
	counter := 1
	return func() *stats {
		wg := &sync.WaitGroup{}
		wg.Add(2)

		// Сбор основных метрик
		go func() {
			defer wg.Done()
			// Чтение метрик
			rt := &runtime.MemStats{}
			runtime.ReadMemStats(rt)

			statsBuf.mu.Lock()

			// Присвоение полей для каждой метрики
			(statsBuf.data)["Alloc"] = float64(rt.Alloc)
			(statsBuf.data)["BuckHashSys"] = float64(rt.BuckHashSys)
			(statsBuf.data)["Frees"] = float64(rt.Frees)
			(statsBuf.data)["GCCPUFraction"] = float64(rt.GCCPUFraction)
			(statsBuf.data)["GCSys"] = float64(rt.GCSys)
			(statsBuf.data)["HeapAlloc"] = float64(rt.HeapAlloc)
			(statsBuf.data)["HeapIdle"] = float64(rt.HeapIdle)
			(statsBuf.data)["HeapInuse"] = float64(rt.HeapInuse)
			(statsBuf.data)["HeapObjects"] = float64(rt.HeapObjects)
			(statsBuf.data)["HeapReleased"] = float64(rt.HeapReleased)
			(statsBuf.data)["HeapSys"] = float64(rt.HeapSys)
			(statsBuf.data)["LastGC"] = float64(rt.LastGC)
			(statsBuf.data)["Lookups"] = float64(rt.Lookups)
			(statsBuf.data)["MCacheInuse"] = float64(rt.MCacheInuse)
			(statsBuf.data)["MCacheSys"] = float64(rt.MCacheSys)
			(statsBuf.data)["MSpanInuse"] = float64(rt.MSpanInuse)
			(statsBuf.data)["MSpanSys"] = float64(rt.MSpanSys)
			(statsBuf.data)["Mallocs"] = float64(rt.Mallocs)
			(statsBuf.data)["NextGC"] = float64(rt.NextGC)
			(statsBuf.data)["NumForcedGC"] = float64(rt.NumForcedGC)
			(statsBuf.data)["NumGC"] = float64(rt.NumGC)
			(statsBuf.data)["OtherSys"] = float64(rt.OtherSys)
			(statsBuf.data)["PauseTotalNs"] = float64(rt.PauseTotalNs)
			(statsBuf.data)["StackInuse"] = float64(rt.StackInuse)
			(statsBuf.data)["StackSys"] = float64(rt.StackSys)
			(statsBuf.data)["Sys"] = float64(rt.Sys)
			(statsBuf.data)["TotalAlloc"] = float64(rt.TotalAlloc)

			// Генерация произвольного значения
			(statsBuf.data)["RandomValue"] = rand.Float64()

			// Увеличение счетчика
			(statsBuf.data)["PollCount"] = int64(counter)

			statsBuf.mu.Unlock()

			counter++
		}()

		// Сбор дополнительных метрик
		go func() {
			defer wg.Done()

			// Сбор статистики памяти
			vmStats, err := mem.VirtualMemory()
			if err != nil {
				log.Println("collect metrics error: Mem.VirtualMemory err:", err)
			}

			// Сбор статистики ЦПУ
			cpuStats, err := cpu.Percent(0, true)
			if err != nil {
				log.Println("collect metrics error: cpu.Percent err:", err)
			}

			statsBuf.mu.Lock()
			// Присвоение полей статистики
			(statsBuf.data)["TotalMemory"] = float64(vmStats.Total)
			(statsBuf.data)["FreeMemory"] = float64(vmStats.Free)

			// Генератор статистики по каждому ЦПУ
			for i, cpuStat := range cpuStats {
				recordString := "CPUutilization" + strconv.Itoa(i+1)
				(statsBuf.data)[recordString] = cpuStat
			}

			statsBuf.mu.Unlock()
		}()

		wg.Wait()
		return statsBuf
	}
}

type restyResponse struct {
	response *resty.Response
	err      error
}

// Метод отправки запроса
func (a *Agent) postWorker(postChan <-chan *job, res chan<- *restyResponse) {
	for data := range postChan {
		URL := a.baseURL + data.urlPath

		// Формирование и выполнение запроса
		resp, err := withRetry(a.withSign(a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(*data.data)), URL)

		result := &restyResponse{
			response: resp,
			err:      err,
		}

		res <- result
	}
}

// Метод отправки метрик
func (a *Agent) sendMetrics(jobs chan<- *job) error {
	if len(a.metrics) == 0 {
		return nil
	}

	// Запуск параллельной отправки метрик горутинами
	for _, metric := range a.metrics {
		channelJob := &job{urlPath: singleHandlerPath}
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
func (a *Agent) sendMetricsBatch(jobs chan<- *job) error {
	if len(a.metrics) == 0 {
		return nil
	}
	channelJob := &job{urlPath: batchHandlerPath}

	// Сериализация метрик
	data, err := json.Marshal(a.metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics error: %w", err)
	}
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
