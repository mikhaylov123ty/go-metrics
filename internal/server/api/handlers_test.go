package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"testing"

	"metrics/internal/models"
	"metrics/internal/storage/memory"

	"github.com/stretchr/testify/assert"
)

var commands *StorageCommands

func init() {
	newMemStorage := memory.NewMemoryStorage()
	commands = &StorageCommands{
		dataReader:  newMemStorage,
		dataUpdater: newMemStorage,
	}

	gaugeVal := float64(3251325234)
	if err := newMemStorage.UpdateBatch([]*models.Data{
		{
			Type:  "gauge",
			Name:  "alloc",
			Value: &gaugeVal,
		},
	}); err != nil {
		panic(err)
	}
}

func TestHandler_Update(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}

	type args struct {
		url         string
		method      string
		contentType string
		metricType  string
		metricName  string
		metricValue string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "no url values, error code 400",
			args: args{
				url:    "http://localhost:8080/update",
				method: "POST",
			},
			want: want{
				code: 400,
			},
		},

		{
			name: "valid gauge",
			args: args{
				url:         "http://localhost:8080/update",
				method:      "POST",
				contentType: "text/plain",
				metricType:  "gauge",
				metricName:  "alloc",
				metricValue: "32.4123",
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "valid counter",
			args: args{
				url:         "http://localhost:8080/update",
				method:      "POST",
				contentType: "text/plain",
				metricType:  "counter",
				metricName:  "alloc",
				metricValue: "3251325234",
			},
			want: want{
				code: 200,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.args.method, tt.args.url, nil)
			request.Header.Add("Content-Type", tt.args.contentType)
			request.SetPathValue("type", tt.args.metricType)
			request.SetPathValue("name", tt.args.metricName)
			request.SetPathValue("value", tt.args.metricValue)

			w := httptest.NewRecorder()

			handler := NewHandler(commands)
			handler.UpdatePost(w, request)

			res := w.Result()

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Println("error closing response body", err)
				}
			}()

			assert.Equal(t, tt.want.code, res.StatusCode, "Codes are not equal")
		})
	}
}

func ExampleHandler_UpdatePostJSON() {
	// Конструктор структуры тела запроса
	val := float64(3251325234)
	data := &models.Data{
		Type:  "gauge",
		Name:  "alloc",
		Value: &val,
	}

	// Сериализация в JSON
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"POST",
		"http://localhost:8080/update",
		io.Reader(bytes.NewBuffer(body)))
	request.Header.Add("Content-Type", "application/json")

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.UpdatePostJSON(w, request)

	// Получение ответа
	res := w.Result()

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_UpdatesPostJSON() {
	// Конструктор структуры тела запроса
	gaugeVal := float64(3251325234)
	data := []*models.Data{
		{
			Type:  "gauge",
			Name:  "alloc",
			Value: &gaugeVal,
		},
	}

	// Сериализация в JSON
	body, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"POST",
		"http://localhost:8080/updates",
		io.Reader(bytes.NewBuffer(body)))
	request.Header.Add("Content-Type", "application/json")

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.UpdatesPostJSON(w, request)

	// Получение ответа
	res := w.Result()

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_UpdatePost() {
	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"POST",
		"http://localhost:8080/update",
		nil)
	request.Header.Add("Content-Type", "text/plain")
	request.SetPathValue("type", "gauge")
	request.SetPathValue("name", "alloc")
	request.SetPathValue("value", "3251325234")

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.UpdatePost(w, request)

	// Получение ответа
	res := w.Result()

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_ValueGetJSON() {
	// Конструктор структуры тела запроса
	data := &models.Data{
		Type: "gauge",
		Name: "alloc",
	}

	// Сериализация в JSON
	body, _ := json.Marshal(data)

	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"POST",
		"http://localhost:8080/value",
		io.Reader(bytes.NewBuffer(body)))
	request.Header.Add("Content-Type", "application/json")

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.ValueGetJSON(w, request)

	// Получение ответа
	res := w.Result()

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()

	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// {"type":"gauge","id":"alloc","value":3251325234}
}

func ExampleHandler_ValueGet() {
	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"GET",
		"http://localhost:8080/value",
		nil)
	request.Header.Add("Content-Type", "text/plain")
	request.SetPathValue("type", "gauge")
	request.SetPathValue("name", "alloc")

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.ValueGet(w, request)

	// Получение ответа
	res := w.Result()

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()

	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// 3251325234
}

func ExampleHandler_IndexGet() {
	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"GET",
		"http://localhost:8080/",
		nil)

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.IndexGet(w, request)

	// Получение ответа
	res := w.Result()

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()

	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// [{"type":"gauge","id":"alloc","value":3251325234}]
}
