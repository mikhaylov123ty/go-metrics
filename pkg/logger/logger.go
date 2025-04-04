// Модуль logger позволяет реализовать продвинутое логирование
package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// New - констуктор логгера
func New(level string) (*logrus.Logger, error) {
	newLogger := logrus.New()
	newLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}

	newLogger.SetLevel(newLevel)

	return newLogger, nil
}
