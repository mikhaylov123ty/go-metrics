package requestTemplates

import "metrics/internal/storage"

// Структуры запросов
type UpdatePost struct {
	Id   string
	Data *storage.Data
}

type GetValue struct {
	Id string
}
