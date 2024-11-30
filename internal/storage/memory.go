package storage

import "sync"

// Структура хранилища
type MemoryStorage struct {
	mu      sync.RWMutex
	metrics map[string]*Data
}

// Реализация интерфеса
func NewMemoryStorage() Storage {
	return &MemoryStorage{metrics: make(map[string]*Data)}
}

// Метод получения записи из хранилища по id
func (m *MemoryStorage) Read(id string) (*Data, error) {
	res, ok := m.metrics[id]
	if !ok {
		return nil, nil
	}

	return res, nil
}

// Метод получения записей из хранилища
func (m *MemoryStorage) ReadAll() ([]*Data, error) {
	res := make([]*Data, 0)
	for _, data := range m.metrics {
		res = append(res, data)
	}

	return res, nil
}

// Метод создания или обновления существующей записи в хранилище
func (m *MemoryStorage) Update(query *Data) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metric, ok := m.metrics[query.Name]; ok && query.Type == "counter" {
		*query.Delta += *metric.Delta
	}

	m.metrics[query.Name] = query

	return nil
}

// Метод создания или обновление существующих записей в хранилище
func (m *MemoryStorage) UpdateBatch(queries []*Data) error {
	for _, query := range queries {
		m.mu.Lock()

		if metric, ok := m.metrics[query.Name]; ok && query.Type == "counter" {
			*query.Delta += *metric.Delta
		}
		m.metrics[query.Name] = query

		m.mu.Unlock()
	}

	return nil
}

// Метод удаления записи из хранилища
func (m *MemoryStorage) Delete(id string) error {
	delete(m.metrics, id)

	return nil
}

// Метод проверки доступности БД
func (m *MemoryStorage) Ping() error {
	return nil
}
