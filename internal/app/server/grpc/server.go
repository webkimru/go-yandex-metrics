package grpc

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/server/models"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories"
	pb "github.com/webkimru/go-yandex-metrics/internal/proto"
	"golang.org/x/net/context"
)

var Repo *MetricsServer

// MetricsServer поддерживает все необходимые методы сервера.
type MetricsServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedMetricsServer
	// добавляем хранилище
	Store repositories.StoreRepository
}

func (s *MetricsServer) UpdateBatchMetrics(ctx context.Context, in *pb.RequestMetricBatch) (*pb.ResponseMetric, error) {
	var response pb.ResponseMetric
	var metrics []models.Metrics

	for _, request := range in.RequestMetrics {
		metrics = append(metrics, models.Metrics{
			Delta: &request.Delta,
			Value: &request.Value,
			ID:    request.Id,
			MType: request.Type,
		})
	}

	err := s.Store.UpdateBatchMetrics(ctx, metrics)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func NewRepo(repository repositories.StoreRepository) *MetricsServer {
	return &MetricsServer{
		Store: repository,
	}
}

func NewMetricHandlers(r *MetricsServer) {
	Repo = r
}
