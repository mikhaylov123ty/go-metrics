// Модуль api описывает эндпоинты
package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"metrics/internal/models"
)

//TODO разбить по файлам

// Handler - структура HTTP хендлера
type Handler struct {
	storageCommands *StorageCommands
}

// StorageCommands - команды для взаимодействия с хранилищем
type StorageCommands struct {
	dataReader
	dataUpdater
	pinger
}

// dataReader - интерфейс хендлера для чтения из базы
type dataReader interface {
	Read(string) (*models.Data, error)
	ReadAll() ([]*models.Data, error)
}

// dataUpdater - интерфейс хендлера для записи в базу
type dataUpdater interface {
	Update(*models.Data) error
	UpdateBatch([]*models.Data) error
}

// pinger - интерфейс хендлера для проверки базы
type pinger interface {
	Ping() error
}

// NewHandler - конструктор хендлера
func NewHandler(apiStorageCommands *StorageCommands) *Handler {
	return &Handler{
		storageCommands: apiStorageCommands,
	}
}

// NewStorageService - конструктор  сервиса, т.к. размещение инетрфейсов по месту использования
// предполгает, что они неэкспортируемые
func NewStorageService(dataReader dataReader, dataUpdater dataUpdater, ping pinger) *StorageCommands {
	return &StorageCommands{
		dataReader:  dataReader,
		dataUpdater: dataUpdater,
		pinger:      ping,
	}
}

// UpdatePostJSON - метод ручки "POST /update с телом JSON"
func (h *Handler) UpdatePostJSON(w http.ResponseWriter, req *http.Request) {
	var err error

	// Проверка хедера
	if req.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Чтение тела запроса
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("UpdatePostJSON: failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer func() {
		if err = req.Body.Close(); err != nil {
			log.Println("UpdatePostJSON: failed close request body", err)
		}
	}()

	// Десериализация тела запроса
	storageData := models.Data{}
	if err = json.Unmarshal(body, &storageData); err != nil {
		log.Println("UpdatePostJSON: failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка невалидных значений
	if err = storageData.CheckData(); err != nil {
		log.Println("UpdatePostJSON: failed check request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.Update(&storageData); err != nil {
		log.Println("UpdatePostJSON: update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.WriteHeader(http.StatusOK)
}

// UpdatesPostJSON - метод ручки "POST /updates с телом JSON" (Batches)
func (h *Handler) UpdatesPostJSON(w http.ResponseWriter, req *http.Request) {
	var err error

	// Проверка хедера
	if req.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Чтение тела запроса
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("UpdatesPostJSON: failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer func() {
		if err = req.Body.Close(); err != nil {
			log.Println("UpdatesPostJSON: failed close request body", err)
		}
	}()

	// Десериализация тела запроса
	storageData := []*models.Data{}
	if err = json.Unmarshal(body, &storageData); err != nil {
		log.Println("UpdatesPostJSON: failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустых батчей
	if len(storageData) == 0 {
		log.Println("UpdatesPostJSON: empty batch data")
		w.WriteHeader(http.StatusBadRequest)
	}

	// Проход по метрикам
	for _, data := range storageData {
		// Проверка невалидных значений
		if err = data.CheckData(); err != nil {
			log.Println("UpdatesPostJSON: failed check request body", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.UpdateBatch(storageData); err != nil {
		log.Println("UpdatesPostJSON: update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.WriteHeader(http.StatusOK)
}

// UpdatePost - метод ручки "POST /update/{type}/{name}/{value}"
func (h *Handler) UpdatePost(w http.ResponseWriter, req *http.Request) {
	// Конструктор даты хранилища
	storageData := &models.Data{
		Type: req.PathValue("type"),
		Name: req.PathValue("name"),
	}

	// Проверка типа метрики
	switch storageData.Type {
	case "counter":
		// Форматирование и присвоение значения counter
		delta, err := strconv.ParseInt(req.PathValue("value"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("UpdatePost: invalid new data:", err)
			return
		}
		storageData.Delta = &delta

	case "gauge":
		// Форматирование и присвоение значения gauge
		value, err := strconv.ParseFloat(req.PathValue("value"), 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("UpdatePost: invalid new data:", err)
			return
		}
		storageData.Value = &value

	default:
		log.Println("UpdatePost: invalid data type:", storageData.Type)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустых значений
	if (storageData.Value == nil && storageData.Delta == nil) || storageData.Type == "" || storageData.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Обновление или сохранение новой записи в хранилище
	if err := h.storageCommands.Update(storageData); err != nil {
		log.Println("UpdatePost: update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.WriteHeader(http.StatusOK)
}

// ValueGetJSON - метод ручки "POST /value"
func (h *Handler) ValueGetJSON(w http.ResponseWriter, req *http.Request) {
	// Проверка хедера
	if req.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Чтение тела
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("ValueGetJSON: failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer func() {
		if err = req.Body.Close(); err != nil {
			log.Println("ValueGetJSON: failed close request body", err)
		}
	}()

	// Десериализация тела
	storageData := models.Data{}
	if err = json.Unmarshal(body, &storageData); err != nil {
		log.Println("ValueGetJSON: failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получение данных записи
	metric, err := h.storageCommands.Read(storageData.Name)
	if err != nil {
		log.Println("ValueGetJSON: get handler: read repo:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустой даты
	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Сериализация данных
	response, err := json.Marshal(metric)
	if err != nil {
		log.Println("ValueGetJSON: get handler: marshal data:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Передача данных в ответ
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(response); err != nil {
		log.Println("get handler error:", err)
	}
}

// ValueGet - метод ручки "GET /value/{type}/{name}"
func (h *Handler) ValueGet(w http.ResponseWriter, req *http.Request) {
	var err error

	if req.PathValue("name") == "" {
		log.Println("ValueGet: empty path value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получение данных записи
	data, err := h.storageCommands.Read(req.PathValue("name"))
	if err != nil {
		log.Println("ValueGet: read repo:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустой даты
	if data == nil {
		log.Println("ValueGet: read repo: not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var response []byte
	// Сериализация данных
	if data.Type == "counter" {
		response, err = json.Marshal(data.Delta)
		if err != nil {
			log.Println("ValueGet:marshal data:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		response, err = json.Marshal(data.Value)
		if err != nil {
			log.Println("ValueGet: marshal data:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Передача данных в ответ
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(response); err != nil {
		log.Println("get handler error:", err)
	}
}

// IndexGet - метод ручки "GET /"
func (h *Handler) IndexGet(w http.ResponseWriter, req *http.Request) {
	// Получение всех записей
	data, err := h.storageCommands.ReadAll()
	if err != nil {
		log.Println("IndexGet: get handler error:", err)
	}

	// Проверка пустой даты
	if len(data) == 0 {
		log.Println("IndexGet: get handler error: not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Сериализация данных
	resp, err := json.Marshal(data)
	if err != nil {
		log.Println("IndexGet: get handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Передача данных в ответ
	w.Header().Set("Content-Type", "text/html")
	if _, err = w.Write(resp); err != nil {
		log.Println("get handler error:", err)
	}
}

// PingGet - метод ручки "GET /ping"
func (h *Handler) PingGet(w http.ResponseWriter, req *http.Request) {
	if h.storageCommands.pinger == nil {
		log.Println("working from memory")
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := h.storageCommands.Ping(); err != nil {
		log.Println("ping handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
