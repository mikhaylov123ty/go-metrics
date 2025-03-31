// Модуль server реализует эндпоинты для взаимодействия и хранения метрик
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"metrics/internal/server/api"
	"metrics/internal/server/config"
	"metrics/internal/server/gRPC"
	"metrics/internal/server/metrics"
)

// Server - структура сервера
type Server struct {
	services services
	logger   *logrus.Logger
	options  *options
	auth     *auth
}

// services - структура команд БД и файла с бэкапом
type services struct {
	apiStorageCommands  *api.StorageCommands
	gRPCStorageCommands *gRPC.StorageCommands
	metricsFileStorage  *metrics.MetricsFileStorage
}

type options struct {
	storeInterval   float64
	fileStoragePath string
	restore         bool
}

type auth struct {
	cryptoKey     string
	hashKey       string
	trustedSubnet *net.IPNet
}

// New - конструктор инстанса сервера
func New(
	apiStorageCommands *api.StorageCommands,
	gRPCStorageCommands *gRPC.StorageCommands,
	metricsFileStorage *metrics.MetricsFileStorage,
	logger *logrus.Logger,
	cfg *config.ServerConfig) *Server {
	return &Server{
		services: services{
			apiStorageCommands:  apiStorageCommands,
			gRPCStorageCommands: gRPCStorageCommands,
			metricsFileStorage:  metricsFileStorage,
		},
		logger: logger,
		options: &options{
			storeInterval:   cfg.FileStorage.StoreInterval,
			fileStoragePath: cfg.FileStorage.FileStoragePath,
			restore:         cfg.FileStorage.Restore,
		},
		auth: &auth{
			cryptoKey:     cfg.CryptoKey,
			hashKey:       cfg.Key,
			trustedSubnet: cfg.Net.TrustedSubnet,
		},
	}
}

// Start запускает сервера
func (s *Server) Start(ctx context.Context, host *config.Host) error {
	// Инициализация даты из файла
	if s.options.restore {
		if err := s.services.metricsFileStorage.InitMetricsFromFile(); err != nil {
			s.logger.Fatal("error restore metrics from file: ", err)
		}
		s.logger.Infof("metrics file storage restored")
	}
	fmt.Printf("AUTH: %+v\n", s.auth)
	// Создание группы ожидания
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Запуск горутины сохранения метрик с интервалом
	go func() {
		s.logger.Infof("Starting store metrics worker. Interval: %f", s.options.storeInterval)
		defer wg.Done()
		for {
			//Останавливает горутину, если получен сигнал
			select {
			case <-ctx.Done():
				s.logger.Warn("shutting down file storage worker")
				return
			default:
				time.Sleep(time.Duration(s.options.storeInterval) * time.Second)

				if err := s.services.metricsFileStorage.StoreMetrics(); err != nil {
					s.logger.Errorf("store metrics: failed read metrics: %s", err.Error())
				}
			}
		}
	}()

	fmt.Println("AUTH", *s.auth)

	// HTTP Server
	httpSRV := api.NewServer(host.String(), s.auth.cryptoKey, s.auth.hashKey, s.auth.trustedSubnet, s.services.apiStorageCommands, s.logger)

	// Старт сервера
	go func() {
		s.logger.Infof("Starting server on %v", host.String())
		if err := httpSRV.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP Server Error:", err)
		}
	}()

	// gRPC Server
	listen, err := net.Listen("tcp", ":"+host.GRPCPort)
	if err != nil {
		return fmt.Errorf("gRPC could not listen on %v: %v", host.GRPCPort, err)
	}

	gRPCServer := gRPC.NewServer(s.auth.cryptoKey, s.auth.hashKey, s.auth.trustedSubnet, s.services.gRPCStorageCommands, s.logger)

	// Старт сервера
	go func() {
		s.logger.Infof("Starting gRPC server on %v", host.GRPCPort)
		if err = gRPCServer.Server.Serve(listen); err != nil {
			log.Fatal("gRPC Server Error:", err)
		}
	}()

	// Ожидание сигнала
	<-ctx.Done()

	// Остановка сервера
	if err := httpSRV.Server.Shutdown(ctx); err != nil && err != context.Canceled {
		log.Fatal("HTTP Server Shutdown Failed:", err)
	}

	gRPCServer.Server.GracefulStop()

	// Ожидание завершения горутин
	wg.Wait()

	return nil
}
