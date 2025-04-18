// Модуль metrics делает бэкап метрик в файл и восстанавливает при старте сервера
package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"metrics/internal/models"
)

// MetricsFileStorage - структура для взаимодействия метрик с файлом и хранилищем
type MetricsFileStorage struct {
	storageCommands metricsReaderUpdater
	fileStorage     string
}

// metricsReaderUpdater - интерфейс для взаимодействия с БД
type metricsReaderUpdater interface {
	ReadAll() ([]*models.Data, error)
	Update(*models.Data) error
}

// NewMetricsFileStorage - конструктор структуры метрик
func NewMetricsFileStorage(m metricsReaderUpdater, fileStorage string) *MetricsFileStorage {
	return &MetricsFileStorage{
		storageCommands: m,
		fileStorage:     fileStorage,
	}
}

// StoreMetrics сохраняет метрики в файл
func (m *MetricsFileStorage) StoreMetrics() error {
	// Чтение всех метрик из хранилища
	metrics, err := m.storageCommands.ReadAll()
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
		record, marshalErr := json.Marshal(v)
		if marshalErr != nil {
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

// InitMetricsFromFile восстанавливает метрики из файла
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
		storageData := &models.Data{}
		if err = json.Unmarshal([]byte(line), storageData); err != nil {
			return fmt.Errorf("unmarshal metrics: %w", err)
		}

		// Запись в хранилище
		if err = m.storageCommands.Update(storageData); err != nil {
			return fmt.Errorf("update metrics: %w", err)
		}
	}

	return nil
}
