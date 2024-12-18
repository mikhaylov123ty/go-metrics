package client

import (
	"errors"
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"metrics/internal/storage"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	attempts = 3
	interval = 2 * time.Second
)

// Вспомогательные типы для методов функций
type (
	statsBuf func() *stats
)

// Структура статистики
type stats struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// Структура для канала заданий метрик
type metricJob struct {
	data    *[]byte
	urlPath string
}

// Структура для канала ответов заданий метрик
type restyResponse struct {
	response *resty.Response
	err      error
	worker   int
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

			// Присвоение полей для каждой метрики
			statsBuf.mu.Lock()

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

			(statsBuf.data)["PollCount"] = int64(counter)

			// Генерация произвольного значения
			(statsBuf.data)["RandomValue"] = rand.Float64()

			statsBuf.mu.Unlock()

			// Увеличение счетчика
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

			// Присвоение полей статистики
			statsBuf.mu.Lock()

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

// Метод конструктора метрик в структры
func (s *stats) buildMetrics() []*storage.Data {
	res := []*storage.Data{}
	for k, v := range s.data {
		metric := storage.Data{Name: k}
		switch t := v.(type) {
		case float64:
			metric.Type = "gauge"
			metric.Value = &t
		case int64:
			metric.Type = "counter"
			metric.Delta = &t
		}
		res = append(res, &metric)
	}

	return res
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
