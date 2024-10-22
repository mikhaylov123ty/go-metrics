package storage

type Storage interface {
	Create(query *Data) error
	Read(id string) (*Data, error)
	Update(query *Data) error
	Delete(id string) error
}
