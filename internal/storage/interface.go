package storage

// Интерфейс хранилища
type Storage interface {
	Read(id string) (*Data, error)
	ReadAll() ([]*Data, error)
	Update(query *Data) error
	UpdateBatch(queries []*Data) error
	Delete(id string) error
	Ping() error
}
