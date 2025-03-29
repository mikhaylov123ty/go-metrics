package gRPC

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"metrics/internal/models"
	pb "metrics/internal/server/proto"
)

type Handler struct {
	pb.UnimplementedHandlersServer
	storageCommands *StorageCommands
}

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
	Update(*models.Data) error
	UpdateBatch([]*models.Data) error
}

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

func (h *Handler) PostUpdate(ctx context.Context, request *pb.PostUpdateRequest) (*pb.PostUpdateResponse, error) {
	var err error
	var response pb.PostUpdateResponse
	// Десериализация тела запроса
	storageData := models.Data{
		Type:  request.Metric.Type,
		Name:  request.Metric.Id,
		Value: &request.Metric.Value,
		Delta: &request.Metric.Delta,
	}
	// Проверка невалидных значений
	if err = storageData.CheckData(); err != nil {
		return nil, status.Errorf(codes.Internal, "data values error: %s", err.Error())
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.Update(&storageData); err != nil {
		log.Println("UpdatePostJSON: update handler error:", err)
		return nil, status.Errorf(codes.Internal, "update handler error: %s", err.Error())
	}

	return &response, nil
}

func (h *Handler) PostUpdates(ctx context.Context, request *pb.PostUpdatesRequest) (*pb.PostUpdatesResponse, error) {
	var err error
	var response pb.PostUpdatesResponse

	storageData := make([]*models.Data, len(request.Metric))
	for i, v := range request.Metric {
		// Десериализация тела запроса
		storageData[i].Type = v.Type
		storageData[i].Name = v.Id
		storageData[i].Value = &v.Value
		storageData[i].Delta = &v.Delta

		// Проверка невалидных значений
		if err = storageData[i].CheckData(); err != nil {
			return nil, status.Errorf(codes.Internal, "data values error: %s", err.Error())
		}
	}

	// Обновление или сохранение новой записи в хранилище
	if err = h.storageCommands.UpdateBatch(storageData); err != nil {
		return nil, status.Errorf(codes.Internal, "updates handler error: %s", err.Error())
	}

	return &response, nil
}
func (h *Handler) GetValue(ctx context.Context, request *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	var err error

	metric, err := h.storageCommands.Read(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get handler error: %s", err.Error())
	}

	response := &pb.GetValueResponse{
		Metric: &pb.Metric{
			Type:  metric.Type,
			Id:    metric.Name,
			Value: *metric.Value,
			Delta: *metric.Delta,
		},
	}

	return response, nil
}
