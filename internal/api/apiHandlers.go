package api

import (
	"log"
	"net/http"

	"metrics/internal/storage"
)

// Структура запроса
type updateRequest struct {
	id   string
	data *storage.Data
}

// Структура хендлера
type Handler struct {
	repo storage.Storage
}

// Конструктор обработчика
func NewHandler(repo storage.Storage) *Handler {
	return &Handler{repo: repo}
}

// Метод ручки "POST /update/{type}/{name}/{value}"
func (h Handler) Update(w http.ResponseWriter, req *http.Request) {
	var err error
	var query = &updateRequest{}

	// Проверка хедера
	if req.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Парсинг даты и констурктор записи
	query.data, err = storage.NewData(
		req.PathValue("type"),
		req.PathValue("name"),
		req.PathValue("value"),
	)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Формирование уникального идентификатора
	query.id = query.data.UniqueID()

	// Проверка предыдущего значения, если тип "counter"
	if query.data.Type == "counter" {
		prevData, err := h.repo.Read(query.id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Сложение значений, если найдено в хранилище
		if prevData != nil {
			query.data.Value = prevData.Value.(int64) + query.data.Value.(int64)
		}
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.repo.Update(query.id, query.data); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
