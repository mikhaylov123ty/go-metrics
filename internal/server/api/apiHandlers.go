package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"metrics/internal/storage"
)

// Структура хендлера
type Handler struct {
	repo storage.Storage
}

// Конструктор обработчика
func NewHandler(repo storage.Storage) *Handler {
	return &Handler{repo: repo}
}

// Метод ручки "POST /update с телом JSON"
func (h *Handler) UpdatePostJSON(w http.ResponseWriter, req *http.Request) {
	var err error

	if req.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics := Metrics{}
	fmt.Println("BODY", string(body), "URL", req.URL)
	if err = json.Unmarshal(body, &metrics); err != nil {
		log.Println("failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storageData := &storage.Data{
		Type: metrics.MType,
		Name: metrics.ID,
	}
	// Формирование уникального идентификатора
	dataID := storageData.UniqueID()

	// Проверка предыдущего значения, если тип "counter"
	if storageData.Type == "counter" {
		if metrics.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storageData.Value = metrics.Delta
		prevData, err := h.repo.Read(dataID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Сложение значений, если найдено в хранилище
		if prevData != nil {
			value := *storageData.Value.(*int64) + *prevData.Value.(*int64)
			storageData.Value = &value
		}
	} else {

		storageData.Value = metrics.Value
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.repo.Update(dataID, storageData); err != nil {
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// Метод ручки "POST /update/{type}/{name}/{value}"
func (h *Handler) UpdatePost(w http.ResponseWriter, req *http.Request) {
	var err error
	var query = &updatePost{}

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
	} else {
		query.data.Value = query.data.Value.(float64)
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

// Метод ручки "POST /value"
func (h *Handler) ValueGetJSON(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics := Metrics{}
	fmt.Println("BODY", string(body), "URL", req.URL)
	if err = json.Unmarshal(body, &metrics); err != nil {
		log.Println("failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if metrics.MType == "" || metrics.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storageData := &storage.Data{
		Type: metrics.MType,
		Name: metrics.ID,
	}
	// Формирование уникального идентификатора
	dataID := storageData.UniqueID()

	// Получение данных записи
	data, err := h.repo.Read(dataID)
	if err != nil {
		log.Println("get handler: read repo:", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	if data == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if metrics.MType == "counter" {
		metrics.Delta = data.Value.(*int64)
	} else {
		metrics.Value = data.Value.(*float64)
	}

	// Сериализация данных
	response, err := json.Marshal(metrics)
	if err != nil {
		log.Println("get handler: marshal data:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Передача данных в ответ
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(response); err != nil {
		log.Println("get handler error:", err)
	}
}

// Метод ручки "GET /value/{type}/{name}"
func (h *Handler) ValueGet(w http.ResponseWriter, req *http.Request) {
	var err error
	var query = &valueGet{}

	// Формирование ключа записи
	query.id = req.PathValue("type") + "_" + req.PathValue("name")

	// Получение данных записи
	data, err := h.repo.Read(query.id)
	if err != nil {
		log.Println("get handler: read repo:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if data == nil {
		log.Println("get handler: read repo: not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Сериализация данных
	response, err := json.Marshal(data.Value)
	if err != nil {
		log.Println("get handler: marshal data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Передача данных в ответ

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(response); err != nil {
		log.Println("get handler error:", err)
	}
}

// Метод ручки "GET /"
func (h *Handler) IndexGet(w http.ResponseWriter, req *http.Request) {
	// Получение всех записей
	data, err := h.repo.ReadAll()
	if err != nil {
		log.Println("get handler error:", err)
	}

	if len(data) == 0 {
		log.Println("get handler error: not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Сериализация данных
	resp, err := json.Marshal(data)
	if err != nil {
		log.Println("get handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Передача данных в ответ
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(resp); err != nil {
		log.Println("get handler error:", err)
	}
}
