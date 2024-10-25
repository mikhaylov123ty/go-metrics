package storage

import "log"

// Структура хранилища
type MemoryStorage struct {
	metrics map[string]*Data
}

// Реализация интерфеса
func NewMemoryStorage() Storage {
	return &MemoryStorage{metrics: make(map[string]*Data)}
}

// Метод получения записи из хранилища
func (m *MemoryStorage) Read(id string) (*Data, error) {
	res, ok := m.metrics[id]
	if !ok {
		return nil, nil
	}

	return res, nil
}

// Метод создания или обновления существующей записи из хранилища
func (m *MemoryStorage) Update(id string, query *Data) error {
	m.metrics[id] = query

	for k, v := range m.metrics {
		log.Println("MEMORY", "key:", k, "value:", *v)
	}

	return nil
}

// Метод удаления записи из хранилища
func (m *MemoryStorage) Delete(id string) error {
	delete(m.metrics, id)
	return nil
}
