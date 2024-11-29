package storage

// Структура данных хранилища
type Data struct {
	Type  string   `json:"type"`
	Name  string   `json:"id"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

// Временный метод генерации id записи метрики
func (d *Data) UniqueID() string {
	return d.Type + "_" + d.Name
}
