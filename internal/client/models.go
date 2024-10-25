package client

import (
	"encoding/json"
	"fmt"
)

type Stats struct {
	Gauge   Gauge   `json:"gauge"`
	Counter Counter `json:"counter"`
}

type Gauge struct {
	Alloc         float64 `json:"alloc"`
	BuckHashSys   float64 `json:"buckHashSys"`
	Frees         float64 `json:"frees"`
	GCCPUFraction float64 `json:"gcCpuFraction"`
	GCSys         float64 `json:"gcSys"`
	HeapAlloc     float64 `json:"heapAlloc"`
	HeapIdle      float64 `json:"heapIdle"`
	HeapInuse     float64 `json:"heapInuse"`
	HeapObjects   float64 `json:"heapObjects"`
	HeapReleased  float64 `json:"heapReleased"`
	HeapSys       float64 `json:"heapSys"`
	LastGC        float64 `json:"lastGC"`
	Lookups       float64 `json:"lookups"`
	MCacheInuse   float64 `json:"mCacheInuse"`
	MCacheSys     float64 `json:"mCacheSys"`
	MSpanInuse    float64 `json:"mSpanInuse"`
	MSpanSys      float64 `json:"mSpanSys"`
	Mallocs       float64 `json:"mallocs"`
	NextGC        float64 `json:"nextGC"`
	NumForcedGC   float64 `json:"numForcedGC"`
	NumGC         float64 `json:"numGC"`
	OtherSys      float64 `json:"otherSys"`
	PauseTotalNs  float64 `json:"pauseTotalNs"`
	StackInuse    float64 `json:"stackInuse"`
	StackSys      float64 `json:"stackSys"`
	Sys           float64 `json:"sys"`
	TotalAlloc    float64 `json:"totalAlloc"`
	RandomValue   float64 `json:"randomValue"`
}

type Counter struct {
	PollCount int64 `json:"pollCount"`
}

func (s *Stats) Map() (map[string]any, error) {

	res, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("error marshalling %w", err)
	}

	postData := make(map[string]interface{})

	if err = json.Unmarshal(res, &postData); err != nil {
		return nil, fmt.Errorf("error unmarshalling %w", err)
	}

	return postData, nil
}
