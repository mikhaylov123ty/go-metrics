package storage

import (
	"fmt"
	"strconv"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

type Data struct {
	Type  string
	Name  string
	Value any
}

func (d *Data) UniqueId() string {
	return d.Type + "_" + d.Name
}

func (d *Data) NewData(metricType string, name string, value string) error {
	var err error
	d.Type = metricType
	d.Name = name
	switch metricType {
	case gauge:
		if d.Value, err = (strconv.ParseFloat(value, 64)); err != nil {
			return fmt.Errorf("invalid new data: %w", err)
		}
	case counter:
		if d.Value, err = (strconv.ParseInt(value, 10, 64)); err != nil {
			return fmt.Errorf("invalid new data: %w", err)
		}
	default:
		return fmt.Errorf("invalid data type: %s", metricType)
	}

	return nil
}
