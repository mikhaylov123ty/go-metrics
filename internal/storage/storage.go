package storage

import (
	"log"

	"metrics/internal/server/api"
	"metrics/internal/server/gRPC"
	"metrics/internal/server/metrics"
	"metrics/internal/storage/memory"
	"metrics/internal/storage/psql"
)

// Storage - структура интерфейсов для сервисов сервера
type Storage struct {
	ApiStorageCommands  *api.StorageCommands
	GRPCStorageCommands *gRPC.StorageCommands
	MetricsFileStorage  *metrics.MetricsFileStorage
	Closer              func()
}

// NewStorage собирает и инициализирует инстанс хранилища
func NewStorage(dsn string, fsPath string) *Storage {
	var err error
	s := &Storage{}

	switch {
	// Проверка dsn БД
	case dsn != "":
		var psqlStorage *psql.DataBase
		psqlStorage, err = psql.NewPSQLDataBase(
			dsn,
		)
		if err != nil {
			log.Fatal("Build Server Storage Connection Error:", err)
		}

		if err = psqlStorage.Ping(); err != nil {
			log.Fatal("Build Server Storage Ping Error:", err)
		}

		// Подготовка БД
		if err = psqlStorage.BootStrap(dsn); err != nil {
			log.Fatal("Build Server Storage Bootstrap Error:", err)
		}

		// Присвоение интерфейса для сервера HTTP
		s.ApiStorageCommands = api.NewStorageService(
			psqlStorage,
			psqlStorage,
			psqlStorage,
		)

		// Присвоение интерфейса для сервера gRPC
		s.GRPCStorageCommands = gRPC.NewStorageService(
			psqlStorage,
			psqlStorage,
		)

		//Присвоение интерфейса для файла с метриками
		s.MetricsFileStorage = metrics.NewMetricsFileStorage(psqlStorage, fsPath)

		// Передача функции закрытия подключения в инстанс
		s.Closer = func() {
			log.Printf("Closing Server Storage Postgres Instance")
			if err = psqlStorage.Instance.Close(); err != nil {
				log.Printf("Closing Server Storage Instance Error: %s", err.Error())
			}
		}

		log.Println("Storage: postgres")

	default:
		// Хранилище в памяти по умолчанию
		memStorage := memory.NewMemoryStorage()

		// Присвоение интерфейса для сервера HTTP
		s.ApiStorageCommands = api.NewStorageService(
			memStorage,
			memStorage,
			nil,
		)

		// Присвоение интерфейса для сервера gRPC
		s.GRPCStorageCommands = gRPC.NewStorageService(
			memStorage,
			memStorage,
		)

		// Присвоение интерфейса для файла с метриками
		s.MetricsFileStorage = metrics.NewMetricsFileStorage(memStorage, fsPath)

		// Передача функции закрытия подключения в инстанс
		s.Closer = func() {
			log.Printf("Closing Server Storage Memory Instance")
		}

		log.Println("Storage: memory")
	}

	return s
}
