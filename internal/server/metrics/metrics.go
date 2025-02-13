package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"metrics/internal/storage"
)

type MetricsFileStorage struct {
	metricsReadUpdate
	fileStorage string
}

type metricsReadUpdate interface {
	ReadAll() ([]*storage.Data, error)
	Update(*storage.Data) error
}

func NewMetricsFileStorage(m metricsReadUpdate, fileStorage string) *MetricsFileStorage {
	return &MetricsFileStorage{
		metricsReadUpdate: m,
		fileStorage:       fileStorage,
	}
}

// Метод записи метрик в файл
func (m *MetricsFileStorage) StoreMetrics() error {
	// Чтение всех метрик из хранилища
	metrics, err := m.ReadAll()
	if err != nil {
		return fmt.Errorf("read metrics: %w", err)
	}

	// Проверка наличия метрик
	if len(metrics) == 0 {
		log.Println("no metrics found")
		return nil
	}

	// Сериализация метрик
	data := make([]string, len(metrics))
	for i, v := range metrics {
		record, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal metrics: %w", err)
		}

		data[i] = string(record)
	}

	// Создание/обновление файла
	storageFile, err := os.Create(m.fileStorage)
	if err != nil {
		return fmt.Errorf("error create file: %w", err)
	}

	// Запись даты в файл
	if _, err = storageFile.Write([]byte(strings.Join(data, "\n"))); err != nil {
		return fmt.Errorf("write metrics: %w", err)
	}
	return nil
}

// Метод восстановления данных метрик из файла
func (m *MetricsFileStorage) InitMetricsFromFile() error {
	fileData, err := os.ReadFile(m.fileStorage)
	if err != nil {
		log.Println("no metrics file found, skipping restore")
		return nil
	}

	// Проверка, если файл пустой
	if len(fileData) < 1 {
		log.Println("no metrics found in file, skipping restore")
		return nil
	}

	// Разбивка по линиям файла
	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		storageData := &storage.Data{}
		if err = json.Unmarshal([]byte(line), storageData); err != nil {
			return fmt.Errorf("unmarshal metrics: %w", err)
		}

		// Запись в хранилище
		if err = m.Update(storageData); err != nil {
			return fmt.Errorf("update metrics: %w", err)
		}
	}

	return nil
}
