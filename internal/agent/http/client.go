package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/net/bpf"
	"log"
	"syscall"
	"time"

	"net"

	"github.com/go-resty/resty/v2"
)

const (
	singleHandlerPath = "/update"
	batchHandlerPath  = "/updates"
)

type HTTPClient struct {
	client  *resty.Client
	baseURL string
	key     string
}

type httpRequest resty.Request

type httpResponse struct {
	response *resty.Response
	err      error
}

func NewHTTPClient(baseURL string, key string, client *resty.Client) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		key:     key,
		client:  client,
	}
}

func (h *HTTPClient) PostUpdates(ctx context.Context, body []byte) error {
	request := httpRequest(*h.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(body)).
		withRealIP().
		withSign(h.key)

	// Обработка тела запроса

	// Формирование и выполнение запроса

	// Запись в результирующий канал
	res <- result
}

// Middleware для запросов с подписью
func (req httpRequest) withSign(key string) *httpRequest {
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(fmt.Sprintf("%s", req.Body)))
		hash := hex.EncodeToString(h.Sum(nil))

		req.Header.Add("HashSHA256", hash)
	}

	return &req
}

func (req httpRequest) withRealIP() *httpRequest {
	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("failed to get interface addresses: %s", err.Error())
	}

	for _, v := range interfaces {
		if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				req.Header.Add("X-Real-IP", ipnet.IP.String())
				break
			}
		}
	}
	return &req
}

// Middleware повтора функции отправки метрик на сервер
func (h httpRequest) withRetry(ctx context.Context) error {
	var err error
	wait := 1 * time.Second

	// Попытки выполнения запроса и возврат при успешном выполнении
	for range attempts {
		err = poster.PostUpdates(ctx, body)
		if err == nil {
			return nil
		}
		// Проверка ошибки для сценария недоступности сервера
		switch {
		case errors.Is(err, syscall.ECONNREFUSED):
			log.Printf("Worker: %d, retrying after error: %s\n", w, err.Error())
			time.Sleep(wait)
			wait += interval

		// Возврат ошибки по умолчанию
		default:
			return err
		}
	}

	return err
}
