package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	singleHandlerPath = "/update"
	batchHandlerPath  = "/updates"
)

// TODO annotations
type HTTPClient struct {
	client   *resty.Client
	baseURL  string
	key      string
	attempts int
	interval time.Duration
}

type httpRequest struct {
	*resty.Request
}

func New(client *resty.Client, baseURL string, key string, attempts int, interval time.Duration) *HTTPClient {
	return &HTTPClient{
		client:   client,
		baseURL:  baseURL,
		key:      key,
		attempts: attempts,
		interval: interval,
	}
}

func (h *HTTPClient) PostUpdates(ctx context.Context, body []byte) error {
	request := httpRequest{h.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(body)}
	response, err := request.
		withRealIP().
		withSign(h.key).
		doWithRetry(h.attempts, h.baseURL+batchHandlerPath, h.interval)
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("http status code %d", response.StatusCode())
	}

	return nil
}

// Middleware для запросов с подписью
func (req *httpRequest) withSign(key string) *httpRequest {
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(fmt.Sprintf("%s", req.Body)))
		hash := hex.EncodeToString(h.Sum(nil))

		req.Header.Add("HashSHA256", hash)
	}

	return req
}

func (req *httpRequest) withRealIP() *httpRequest {
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

	return req
}

// Middleware повтора функции отправки метрик на сервер
func (req *httpRequest) doWithRetry(attempts int, url string, interval time.Duration) (*resty.Response, error) {
	var err error
	wait := 1 * time.Second

	// Попытки выполнения запроса и возврат при успешном выполнении
	for range attempts {
		var response *resty.Response
		response, err = req.Post(url)
		if err == nil {
			return response, nil
		}
		// Проверка ошибки для сценария недоступности сервера
		switch {
		case errors.Is(err, syscall.ECONNREFUSED):
			log.Printf("Worker: TODO HERE, retrying after error: %s\n", err.Error())
			time.Sleep(wait)
			wait += interval

		// Возврат ошибки по умолчанию
		default:
			return nil, err
		}
	}

	return nil, err
}
