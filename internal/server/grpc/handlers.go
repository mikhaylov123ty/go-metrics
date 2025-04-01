package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"metrics/internal/models"
	pb "metrics/internal/server/proto"
)

// Handler - структура gRPC хендлера
type Handler struct {
	pb.UnimplementedHandlersServer
	storageCommands *StorageCommands
}

// StorageCommands - команды для взаимодействия с хранилищем
type StorageCommands struct {
	dataReader
	dataUpdater
}

// dataReader - интерфейс хендлера для чтения из базы
type dataReader interface {
	Read(string) (*models.Data, error)
}

// dataUpdater - интерфейс хендлера для записи в базу
type dataUpdater interface {
	UpdateBatch([]*models.Data) error
}

// NewHandler - конструктор хендлера
func NewHandler(gRPCStorageCommands *StorageCommands) *Handler {
	return &Handler{
		storageCommands: gRPCStorageCommands,
	}
}

// NewStorageService - конструктор  сервиса, т.к. размещение инетрфейсов по месту использования
// предполгает, что они неэкспортируемые
func NewStorageService(dataReader dataReader, dataUpdater dataUpdater) *StorageCommands {
	return &StorageCommands{
		dataReader:  dataReader,
		dataUpdater: dataUpdater,
	}
}

// PostUpdates - хендлер запросав передачи метрик батчами
func (h *Handler) PostUpdates(ctx context.Context, request *pb.PostUpdatesRequest) (*pb.PostUpdatesResponse, error) {
	var err error
	var response pb.PostUpdatesResponse

	storageData := []*models.Data{}
	if err = json.Unmarshal(request.Metrics, &storageData); err != nil {
		return nil, fmt.Errorf("UpdatesPostGRPC failed unmarshall request body: %w", err)
	}

	// Проверка невалидных значений
	for _, data := range storageData {
		// Проверка невалидных значений
		if err = data.CheckData(); err != nil {
			return nil, status.Errorf(codes.Internal, "data values error: %s", err.Error())
		}
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.UpdateBatch(storageData); err != nil {
		return nil, status.Errorf(codes.Internal, "updates handler error: %s", err.Error())
	}

	return &response, nil
}
