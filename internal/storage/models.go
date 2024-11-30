package storage

import (
	"fmt"
)

// Структура данных хранилища
type Data struct {
	Type  string   `json:"type"`
	Name  string   `json:"id"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

// Метод проверки входящих данных
func (d *Data) CheckData() error {
	if d.Value == nil && d.Delta == nil {
		return fmt.Errorf("empty metrics values")
	}
	if d.Value == nil && d.Type == "gauge" {
		return fmt.Errorf("empty gauge value")
	}
	if d.Delta == nil && d.Type == "counter" {
		return fmt.Errorf("empty counter delta")
	}

	return nil
}
