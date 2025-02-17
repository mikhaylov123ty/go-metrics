package memory

import (
	"sync"

	"metrics/internal/models"
)

// MemoryStorage - структура хранилища памяти
type MemoryStorage struct {
	mu      sync.RWMutex
	metrics map[string]*models.Data
}

// NewMemoryStorage - конструктор хранилища
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{metrics: make(map[string]*models.Data)}
}

// Read получает метрику из хранилища по названию
func (m *MemoryStorage) Read(id string) (*models.Data, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res, ok := m.metrics[id]
	if !ok {
		return nil, nil
	}

	return res, nil
}

// ReadAll получает все метрики из хранилища
func (m *MemoryStorage) ReadAll() ([]*models.Data, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*models.Data, 0)
	for _, data := range m.metrics {
		res = append(res, data)
	}

	return res, nil
}

// Update создает новую или обновляет существующую запись метрики в хранилище
func (m *MemoryStorage) Update(query *models.Data) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, ok := m.metrics[query.Name]; ok && query.Type == "counter" && metric.Type == query.Type {
		*query.Delta += *metric.Delta
	}

	m.metrics[query.Name] = query

	return nil
}

// UpdateBatch создает новые или обновляет существующие записи метрики в хранилище
func (m *MemoryStorage) UpdateBatch(queries []*models.Data) error {
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
