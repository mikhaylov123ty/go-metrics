package api

import (
	"log"
	"net/http"
	"strings"

	"metrics/internal/storage"

	"github.com/go-chi/chi/v5"
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
		strings.ToLower(chi.URLParam(req, "type")),
		strings.ToLower(chi.URLParam(req, "name")),
		chi.URLParam(req, "value"),
	)
	if err != nil {
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Формирование уникального идентификатора
	query.id = query.data.UniqueID()

	// Обновление или сохранение новой записи в хранилище
	if err = h.repo.Update(query.id, query.data); err != nil {
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
