package storage

import "fmt"

type MemoryStorage struct {
	metrics map[string]*Data
}

func NewMemoryStorage() Storage {
	return &MemoryStorage{metrics: make(map[string]*Data)}
}

func (m *MemoryStorage) Create(query *Data) error {
	id := query.UniqueId()

	m.metrics[id] = query

	fmt.Println(m.metrics)
	fmt.Println(m.metrics[id])

	return nil
}

func (m *MemoryStorage) Read(id string) (*Data, error) {
	return m.metrics[id], nil
}

func (m *MemoryStorage) Update(query *Data) error {
	id := query.UniqueId()

	m.metrics[id] = query

	fmt.Println(m.metrics)
	fmt.Println(m.metrics[id])
	return nil
}

func (m *MemoryStorage) Delete(id string) error {
	delete(m.metrics, id)
	return nil
}
