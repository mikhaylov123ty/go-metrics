package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"metrics/internal/storage"
)

// Структура запроса
type updatePostRequest struct {
	id   string
	data *storage.Data
}

type getValueRequest struct {
	id string
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
func (h *Handler) UpdatePost(w http.ResponseWriter, req *http.Request) {
	var err error
	var query = &updatePostRequest{}

	//// Проверка хедера
	//if req.Header.Get("Content-Type") != "text/plain" {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	//TODO change to chi after test is solved
	// Парсинг даты и констурктор записи
	query.data, err = storage.NewData(
		strings.ToLower(req.PathValue("type")),
		strings.ToLower(req.PathValue("name")),
		req.PathValue("value"),
	)
	if err != nil {
		log.Println("update handler error:", err)
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
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ValueGet(w http.ResponseWriter, req *http.Request) {
	var err error
	var query = &getValueRequest{}

	query.id = strings.ToLower(req.PathValue("type")) + "_" + strings.ToLower(req.PathValue("name"))

	data, err := h.repo.Read(query.id)
	if err != nil {
		log.Println("get handler error:", err)
	}

	if data == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response, err := json.Marshal(data.Value)

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		log.Println("get handler error:", err)
	}
}

func (h *Handler) IndexGet(w http.ResponseWriter, req *http.Request) {
	data, err := h.repo.ReadAll()
	if err != nil {
		log.Println("get handler error:", err)
	}

	if data == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(data)
	if err != nil {
		log.Println("get handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(resp); err != nil {
		log.Println("get handler error:", err)
	}
}
