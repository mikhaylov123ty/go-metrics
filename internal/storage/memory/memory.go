package memory

import (
	"sync"

	"metrics/internal/storage"
)

// Структура хранилища
type MemoryStorage struct {
	mu      sync.RWMutex
	metrics map[string]*storage.Data
}

// Реализация интерфеса
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{metrics: make(map[string]*storage.Data)}
}

// Метод получения записи из хранилища по id
func (m *MemoryStorage) Read(id string) (*storage.Data, error) {
	res, ok := m.metrics[id]
	if !ok {
		return nil, nil
	}

	return res, nil
}

// Метод получения записей из хранилища
func (m *MemoryStorage) ReadAll() ([]*storage.Data, error) {
	res := make([]*storage.Data, 0)
	for _, data := range m.metrics {
		res = append(res, data)
	}

	return res, nil
}

// Метод создания или обновления существующей записи в хранилище
func (m *MemoryStorage) Update(query *storage.Data) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, ok := m.metrics[query.Name]; ok && query.Type == "counter" && metric.Type == query.Type {
		*query.Delta += *metric.Delta
	}

	m.metrics[query.Name] = query

	return nil
}

// Метод создания или обновление существующих записей в хранилище
func (m *MemoryStorage) UpdateBatch(queries []*storage.Data) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, query := range queries {
		if metric, ok := m.metrics[query.Name]; ok && query.Type == "counter" && metric.Type == query.Type {
			*query.Delta += *metric.Delta
		}
		m.metrics[query.Name] = query
	}

	return nil
}
