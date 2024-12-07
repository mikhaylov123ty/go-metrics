package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"metrics/internal/storage"
)

// Структура хендлера
type Handler struct {
	storageCommands *StorageCommands
}

// Комманды хендлера
type StorageCommands struct {
	read
	readAll
	update
	updateBatch
	ping
}

// Интерфейсы хендлера
type read interface {
	Read(string) (*storage.Data, error)
}
type readAll interface {
	ReadAll() ([]*storage.Data, error)
}
type update interface {
	Update(*storage.Data) error
}
type updateBatch interface {
	UpdateBatch([]*storage.Data) error
}
type ping interface {
	Ping() error
}

// Конструктор хендлера
func NewHandler(storageCommands *StorageCommands) *Handler {
	return &Handler{
		storageCommands: storageCommands,
	}
}

// Конструктор  сервиса, т.к. размещение инетрфейсов по месту использования
// предполгает, что они неимпортируемые
func NewStorageService(read read, readAll readAll, update update, updateBatch updateBatch, ping ping) *StorageCommands {
	return &StorageCommands{
		read:        read,
		readAll:     readAll,
		update:      update,
		updateBatch: updateBatch,
		ping:        ping,
	}
}

// Метод ручки "POST /update с телом JSON"
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
		log.Println("failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Десериализация тела запроса
	storageData := storage.Data{}
	if err = json.Unmarshal(body, &storageData); err != nil {
		log.Println("failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка невалидных значений
	if err = storageData.CheckData(); err != nil {
		log.Println("failed check request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.Update(&storageData); err != nil {
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.WriteHeader(http.StatusOK)
}

// Метод ручки "POST /updates с телом JSON" (Batches)
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
		log.Println("failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Десериализация тела запроса
	storageData := []*storage.Data{}
	if err = json.Unmarshal(body, &storageData); err != nil {
		log.Println("failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустых батчей
	if len(storageData) == 0 {
		log.Println("empty batch data")
		w.WriteHeader(http.StatusBadRequest)
	}

	// Проход по метрикам
	for _, data := range storageData {
		// Проверка невалидных значений
		if err = data.CheckData(); err != nil {
			log.Println("failed check request body", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.UpdateBatch(storageData); err != nil {
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.WriteHeader(http.StatusOK)
}

// Метод ручки "POST /update/{type}/{name}/{value}"
func (h *Handler) UpdatePost(w http.ResponseWriter, req *http.Request) {
	var err error

	// Конструктор даты хранилища
	storageData := &storage.Data{
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
			log.Println("invalid new data:", err)
			return
		}
		storageData.Delta = &delta

	case "gauge":
		// Форматирование и присвоение значения gauge
		value, err := strconv.ParseFloat(req.PathValue("value"), 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("invalid new data:", err)
			return
		}
		storageData.Value = &value

	default:
		log.Println("invalid data type:", storageData.Type)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустых значений
	if (storageData.Value == nil && storageData.Delta == nil) || storageData.Type == "" || storageData.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.Update(storageData); err != nil {
		log.Println("update handler error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Назначение хедера и статуса
	w.WriteHeader(http.StatusOK)
}

// Метод ручки "POST /value"
func (h *Handler) ValueGetJSON(w http.ResponseWriter, req *http.Request) {
	// Проверка хедера
	if req.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Чтение тела
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("failed read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Десериализация тела
	storageData := storage.Data{}
	if err = json.Unmarshal(body, &storageData); err != nil {
		log.Println("failed unmarshall request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получение данных записи
	metric, err := h.storageCommands.Read(storageData.Name)
	if err != nil {
		log.Println("get handler: read repo:", err)
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

// Метод ручки "GET /value/{type}/{name}"
func (h *Handler) ValueGet(w http.ResponseWriter, req *http.Request) {
	var err error

	if req.PathValue("name") == "" {
		log.Println("value get: empty path value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получение данных записи
	data, err := h.storageCommands.Read(req.PathValue("name"))
	if err != nil {
		log.Println("get handler: read repo:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка пустой даты
	if data == nil {
		log.Println("get handler: read repo: not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var response []byte
	// Сериализация данных
	if data.Type == "counter" {
		response, err = json.Marshal(data.Delta)
		if err != nil {
			log.Println("get handler: marshal data:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		response, err = json.Marshal(data.Value)
		if err != nil {
			log.Println("get handler: marshal data:", err)
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

// Метод ручки "GET /"
func (h *Handler) IndexGet(w http.ResponseWriter, req *http.Request) {
	// Получение всех записей
	data, err := h.storageCommands.ReadAll()
	if err != nil {
		log.Println("get handler error:", err)
	}

	// Проверка пустой даты
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
	w.Header().Set("Content-Type", "text/html")
	if _, err = w.Write(resp); err != nil {
		log.Println("get handler error:", err)
	}
}

// Метод ручки "GET /ping"
func (h *Handler) PingGet(w http.ResponseWriter, req *http.Request) {
	if h.storageCommands.ping == nil {
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
