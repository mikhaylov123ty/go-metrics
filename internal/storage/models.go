package storage

type Data struct {
	Type  string
	Name  string
	Value any
}

func (d *Data) UniqueId() string {
	return d.Type + "_" + d.Name
}
