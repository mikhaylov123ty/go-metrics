package storage

import (
	"fmt"
)

type MemoryStorage struct {
	metrics map[string]*Data
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{metrics: make(map[string]*Data)}
}

func (m *MemoryStorage) Create(query *Data) error {
	id := query.UniqueId()

	m.metrics[id] = query

	fmt.Println(m.metrics[id])

	return nil
}

func (m *MemoryStorage) Read(id string) (*Data, error) {
	res, ok := m.metrics[id]
	if !ok {
		return nil, nil
	}

	return res, nil
}

func (m *MemoryStorage) Update(id string, query *Data) error {

	m.metrics[id] = query

	fmt.Println(m.metrics)
	fmt.Println(m.metrics[id])
	return nil
}

func (m *MemoryStorage) Delete(id string) error {
	delete(m.metrics, id)
	return nil
}
