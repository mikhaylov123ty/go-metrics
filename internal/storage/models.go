package storage

// Структура данных хранилища
type Data struct {
	Type  string
	Name  string
	Value any
}

// Временный метод генерации id записи метрики
func (d *Data) UniqueID() string {
	return d.Type + "_" + d.Name
}
