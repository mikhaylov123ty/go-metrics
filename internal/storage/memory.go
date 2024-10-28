package storage

import (
	"log"
	"sync"
)

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

// Метод создания или обновления существующей записи из хранилища
func (m *MemoryStorage) Update(id string, query *Data) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[id] = query

	log.Println("Update record", query)

	return nil
}

// Метод удаления записи из хранилища
func (m *MemoryStorage) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.metrics, id)
	return nil
}
