package pkg

import (
	"errors"
	"fmt"
	"log"
	"syscall"
	"time"
)

const (
	attempts = 4
	interval = 2 * time.Second
)

// AnyFunc - вспомогательный тип для методов примитивных функций
type AnyFunc func() error

// WithRetry повторяет примитивую функцию
func (af AnyFunc) WithRetry() error {
	var err error
	wait := 1 * time.Second

	// Попытки выполнения запроса и возврат при успешном выполнении
	for range attempts {
		if err = af(); err == nil {
			return nil
		}

		// Проверка ошибки для сценария недоступности соединения
		switch {
		case errors.Is(err, syscall.ECONNREFUSED):
			time.Sleep(wait)
			log.Printf("trying to reconnect, error: %s", err.Error())
			wait += interval

		default:
			return err
		}
	}

	return fmt.Errorf("failed after %d attempts", attempts)
}

type ContextKey struct{}
