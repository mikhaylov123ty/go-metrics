package client

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
)

type Agent struct {
	BaseUrl string
	Client  *http.Client
	Stats   Stats
}

func NewAgent(baseUrl string) *Agent {
	return &Agent{
		BaseUrl: baseUrl,
		Client:  http.DefaultClient,
	}
}

func (a *Agent) PostUpdate(metricType string, metricName string, metricValue string) *http.Response {
	request, err := http.NewRequest("POST", a.BaseUrl+"update/"+metricType+"/"+metricName+"/"+metricValue, nil)
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Set("Content-Type", "text/plain")

	resp, err := a.Client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(resp)

	return resp
}

func (a *Agent) CollectData() {
	rt := &runtime.MemStats{}
	runtime.ReadMemStats(rt)

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
	a.Stats.Gauge.RandomValue = rand.Float64()

	a.Stats.Counter.PollCount++
}
