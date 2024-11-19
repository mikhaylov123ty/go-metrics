package storage

// Структура данных хранилища
type Data struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// Временный метод генерации id записи метрики
func (d *Data) UniqueID() string {
	return d.Type + "_" + d.Name
}
