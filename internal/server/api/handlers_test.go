package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	counterVal := int64(100)
	gaugeVal := float64(200)
	newMemStorage.UpdateBatch([]*models.Data{
		{
			Type:  "counter",
			Name:  "testCounter",
			Delta: &counterVal,
		},
		{
			Type:  "gauge",
			Name:  "testGauge",
			Value: &gaugeVal,
		},
	})
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
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode, "Codes are not equal")
		})
	}
}

func ExampleHandler_UpdatePostJSON() {
	// Конструктор структуры тела запроса
	val := int64(200)
	data := &models.Data{
		Type:  "counter",
		Name:  "testCounter",
		Delta: &val,
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
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_UpdatesPostJSON() {
	// Конструктор структуры тела запроса
	gaugeVal := float64(200)
	counterVal := int64(200)
	data := []*models.Data{
		{
			Type:  "gauge",
			Name:  "testGauge",
			Value: &gaugeVal,
		},
		{
			Type:  "counter",
			Name:  "testCounter",
			Delta: &counterVal,
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
	defer res.Body.Close()

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
	request.SetPathValue("type", "counter")
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
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

func ExampleHandler_ValueGetJSON() {
	// Конструктор структуры тела запроса
	data := &models.Data{
		Type: "counter",
		Name: "testCounter",
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
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// {"type":"counter","id":"testCounter","delta":500}
}

func ExampleHandler_ValueGet() {
	// Составление запроса с параметрами URL
	request := httptest.NewRequest(
		"GET",
		"http://localhost:8080/value",
		nil)
	request.Header.Add("Content-Type", "text/plain")
	request.SetPathValue("type", "counter")
	request.SetPathValue("name", "testCounter")

	// Создание интерфейса записи
	w := httptest.NewRecorder()

	// Регистрация нового обработчика
	handler := NewHandler(commands)

	// Выполнение запроса
	handler.ValueGet(w, request)

	// Получение ответа
	res := w.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// 500
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
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// [{"type":"counter","id":"testCounter","delta":500},{"type":"gauge","id":"testGauge","value":200},{"type":"counter","id":"alloc","delta":6502650468}]
}
