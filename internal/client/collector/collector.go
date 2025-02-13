package collector

import (
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"

	"metrics/internal/models"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Вспомогательные типы для методов функций
type (
	StatsBuf func() *Stats
)

// Структура статистики
type Stats struct {
	Data map[string]interface{}
}

// Метод сбора метрик с счетчиком
func CollectMetrics(statsBuf *Stats) StatsBuf {
	counter := 1
	return func() *Stats {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		// Сбор основных метрик
		// Чтение метрик
		rt := &runtime.MemStats{}
		runtime.ReadMemStats(rt)

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

		(statsBuf.Data)["PollCount"] = int64(counter)

		// Генерация произвольного значения
		(statsBuf.Data)["RandomValue"] = rand.Float64()

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
			(statsBuf.Data)["TotalMemory"] = float64(vmStats.Total)
			(statsBuf.Data)["FreeMemory"] = float64(vmStats.Free)

			// Генератор статистики по каждому ЦПУ
			for i, cpuStat := range cpuStats {
				recordString := "CPUutilization" + strconv.Itoa(i+1)
				(statsBuf.Data)[recordString] = cpuStat
			}
		}()

		wg.Wait()

		// Увеличение счетчика
		counter++

		return statsBuf
	}
}

// Метод конструктора метрик в структры
func (s *Stats) BuildMetrics() []*models.Data {
	var res []*models.Data
	for k, v := range s.Data {
		metric := &models.Data{Name: k}
		switch t := v.(type) {
		case float64:
			metric.Type = "gauge"
			metric.Value = &t
		case int64:
			metric.Type = "counter"
			metric.Delta = &t
		}
		res = append(res, metric)
	}

	return res
}
