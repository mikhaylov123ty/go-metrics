package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"log"
	"math/rand/v2"
	"strconv"
	"sync"

	"runtime"
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
		statsBuf:       collectMetrics(&Stats{mu: sync.RWMutex{}, Data: make(map[string]interface{})}),
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
		wg.Add(2)
		//here?
		jobs := make(chan *[]byte)
		res := make(chan *restyResponse)

		//TODO here create jobs channel?
		for range a.rateLimit {
			go a.postWorker(singleHandlerPath, jobs, res)
		}

		go func() {
			fmt.Println("STARTING SEND METRICS")
			if err := a.sendMetrics(jobs); err != nil {
				log.Println("send metric err:", err)
			}
			wg.Done()
			fmt.Println("ENDING SEND METRICS")
		}()

		go func() {
			fmt.Println("STARTING SEND METRICS BATCH")
			if err := a.sendMetricsBatch(jobs); err != nil {
				log.Println("Send metrics batch err:", err)
			}
			wg.Done()
			fmt.Println("Ending SEND METRICS BATCH")
		}()

		go func() {
			wg.Wait()
			fmt.Println("WAIT DONE, CLOSE JOBS")
			close(jobs)
		}()

		for range 40 {
			fmt.Println("reading from res")
			fmt.Println(<-res)

		}

		fmt.Println("END SEND CYCLE")

	}
}

// Метод сбора метрик с счетчиком
func collectMetrics(statsBuf *Stats) statsBuf {
	counter := 1
	return func() *Stats {
		wg := &sync.WaitGroup{}
		wg.Add(2)

		// Cборка основных метрик
		go func() {
			fmt.Println("STARTING GOROUTINE BASE METRIC")
			defer wg.Done()
			// Чтение метрик
			rt := &runtime.MemStats{}
			runtime.ReadMemStats(rt)

			statsBuf.mu.Lock()

			// Присвоение полей для каждой метрики
			(statsBuf.Data)["Alloc"] = float64(rt.Alloc)
			(statsBuf.Data)["BuckHashSys"] = float64(rt.BuckHashSys)
			(statsBuf.Data)["Frees"] = float64(rt.Frees)
			(statsBuf.Data)["GCCPUFraction"] = float64(rt.GCCPUFraction)
			(statsBuf.Data)["GCSys"] = float64(rt.GCSys)
			(statsBuf.Data)["HeapAlloc"] = float64(rt.HeapAlloc)
			(statsBuf.Data)["HeapIdle"] = float64(rt.HeapIdle)
			(statsBuf.Data)["HeapInuse"] = float64(rt.HeapInuse)
			(statsBuf.Data)["HeapObjects"] = float64(rt.HeapObjects)
			(statsBuf.Data)["HeapReleased"] = float64(rt.HeapReleased)
			(statsBuf.Data)["HeapSys"] = float64(rt.HeapSys)
			(statsBuf.Data)["LastGC"] = float64(rt.LastGC)
			(statsBuf.Data)["Lookups"] = float64(rt.Lookups)
			(statsBuf.Data)["MCacheInuse"] = float64(rt.MCacheInuse)
			(statsBuf.Data)["MCacheSys"] = float64(rt.MCacheSys)
			(statsBuf.Data)["MSpanInuse"] = float64(rt.MSpanInuse)
			(statsBuf.Data)["MSpanSys"] = float64(rt.MSpanSys)
			(statsBuf.Data)["Mallocs"] = float64(rt.Mallocs)
			(statsBuf.Data)["NextGC"] = float64(rt.NextGC)
			(statsBuf.Data)["NumForcedGC"] = float64(rt.NumForcedGC)
			(statsBuf.Data)["NumGC"] = float64(rt.NumGC)
			(statsBuf.Data)["OtherSys"] = float64(rt.OtherSys)
			(statsBuf.Data)["PauseTotalNs"] = float64(rt.PauseTotalNs)
			(statsBuf.Data)["StackInuse"] = float64(rt.StackInuse)
			(statsBuf.Data)["StackSys"] = float64(rt.StackSys)
			(statsBuf.Data)["Sys"] = float64(rt.Sys)
			(statsBuf.Data)["TotalAlloc"] = float64(rt.TotalAlloc)

			// Генерация произвольного значения
			(statsBuf.Data)["RandomValue"] = rand.Float64()

			// Увеличение счетчика
			(statsBuf.Data)["PollCount"] = int64(counter)

			statsBuf.mu.Unlock()

			counter++
			fmt.Println("ENDING GOROUTINE BASE METRIC")

		}()

		// Сборка дополнительных метрик
		go func() {
			defer wg.Done()
			fmt.Println("STARTING GOROUTINE ADVANCED METRIC")

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
			(statsBuf.Data)["TotalMemory"] = float64(vmStats.Total)
			(statsBuf.Data)["FreeMemory"] = float64(vmStats.Free)

			// Генератор статистики по каждому ЦПУ
			for i, cpuStat := range cpuStats {
				recordString := "CPUutilization" + strconv.Itoa(i+1)
				(statsBuf.Data)[recordString] = cpuStat
			}

			statsBuf.mu.Unlock()
			fmt.Println("END ADVANCED METRIC")
		}()

		wg.Wait()
		fmt.Println("RETURN GOROUTINE ALL  METRIC")
		return statsBuf
	}
}

type restyResponse struct {
	response *resty.Response
	err      error
}

// TODO количество иходящих запросов на сервер ограничить лимитом
// res <- chan?
// Метод отправки запроса
func (a *Agent) postWorker(handler string, postChan <-chan *[]byte, res chan<- *restyResponse) {
	for data := range postChan {
		fmt.Println("START SENDING DATA FROM WORKER")
		URL := a.baseURL + handler

		// Формирование и выполнение запроса
		resp, err := a.withSign(a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(*data)).Post(URL)
		if err != nil {
			fmt.Println("ERR SEND DATA FROM WORKER", err.Error())
		}
		result := &restyResponse{
			response: resp,
			err:      err,
		}

		res <- result
	}
	fmt.Println("ENDING SENDING DATA FROM WORKER")
}

// todo buf chan <- send metrics
// workers rate limit for range send metrics will send to post updates
func (a *Agent) worker(jobs <-chan *[]byte, results chan<- *storage.Data, sf sendFunc, path string) {
	for j := range jobs {
		// Передача метрики в функцию отправки с опцией повторения
		// при ошибках с подключением
		resp, err := sf.withRetry(path, j)
		if err != nil {
			log.Printf("%s, metric: %v\n", err.Error(), j)
			return
		}
		log.Printf("post update: metric: %v, URI: %s, Status Code: %d\n", j, resp.Request.URL, resp.StatusCode())
	}
}

// Метод отправки метрик
func (a *Agent) sendMetrics(jobs chan<- *[]byte) error {
	if len(a.metrics) == 0 {
		return nil
	}
	fmt.Println("STARTING SEND METRICS in func")
	//todo make jobs channel

	// Запуск параллельной отправки метрик горутинами
	for _, metric := range a.metrics {
		// Сериализация метрики
		data, err := json.Marshal(metric)
		if err != nil {
			fmt.Println("marshal metrics error: %w", err)
			return fmt.Errorf("marshal metrics error: %w", err)
		}

		fmt.Println("SENDING DATA TO JOB", string(data))
		jobs <- &data
	}

	fmt.Println("ENDING SEND METRICS in func")
	return nil
}

// Метод отправки метрик батчами
func (a *Agent) sendMetricsBatch(jobs chan<- *[]byte) error {
	if len(a.metrics) == 0 {
		return nil
	}

	// Сериализация метрик
	data, err := json.Marshal(a.metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics error: %w", err)
	}

	fmt.Println("SENDING BATCH DATA TO JOBS", string(data))
	jobs <- &data
	fmt.Println("ENDING SEND METRICS BATCH in func")
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
