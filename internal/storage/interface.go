package storage

type Storage interface {
	Create(query *Data) error
	Read(id string) (*Data, error)
	Update(id string, query *Data) error
	Delete(id string) error
}
