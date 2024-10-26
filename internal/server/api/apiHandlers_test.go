package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics/internal/storage"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Update(t *testing.T) {
	handler := &Handler{repo: storage.NewMemoryStorage()}
	handlerFunc := http.HandlerFunc(handler.Update)

	srv := httptest.NewServer(handlerFunc)
	defer srv.Close()

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
			name: "wrong content type, error code 400",
			args: args{
				url:         "http://localhost:8080/update",
				method:      "POST",
				contentType: "application/json",
				metricType:  "gauge",
				metricName:  "alloc",
				metricValue: "32.4123",
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
				code:        200,
				contentType: "text/plain; charset=utf-8",
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
				code:        200,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := resty.New().R().
				SetHeader("Content-Type", "text/plain").
				SetPathParams(map[string]string{
					"type":  tt.args.metricType,
					"name":  tt.args.metricName,
					"value": tt.args.metricName,
				})

			resp, err := request.Post(srv.URL)

			require.NoError(t, err)

			assert.Equal(t, tt.want.code, resp.StatusCode(), "Codes are not equal")
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"), "Content types are note equal")
		})
	}
}
