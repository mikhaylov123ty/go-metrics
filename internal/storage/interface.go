package storage

// Интерфейс хранилища
type Storage interface {
	Read(id string) (*Data, error)
	ReadAll() ([]*Data, error)
	Update(id string, query *Data) error
	Delete(id string) error
}
