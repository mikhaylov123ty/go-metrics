package storage

import (
	"fmt"
	"strconv"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

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

// Конструктор записи метрики
func NewData(metricType string, name string, value string) (*Data, error) {
	var err error

	data := &Data{
		Type: metricType,
		Name: name,
	}

	// В зависимости от типа конвертируем значение в требуемый формат
	switch metricType {
	case gauge:
		if data.Value, err = (strconv.ParseFloat(value, 64)); err != nil {
			return nil, fmt.Errorf("invalid new data: %w", err)
		}
	case counter:
		if data.Value, err = (strconv.ParseInt(value, 10, 64)); err != nil {
			return nil, fmt.Errorf("invalid new data: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid data type: %s", metricType)
	}

	return data, nil
}
