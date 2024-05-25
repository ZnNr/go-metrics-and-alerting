package grpc

import (
	"context"
	"github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	pb "github.com/ZnNr/go-musthave-metrics.git/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

// MetricsServer определяет структуру сервера метрик.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
}

// SaveMetricFromJSON сохраняет метрику из JSON и возвращает ответ.
func (s *MetricsServer) SaveMetricFromJSON(ctx context.Context, in *pb.MetricRequest) (*pb.SaveMetricResponse, error) {
	c := collector.Collector()
	metric := collector.MetricRequest{
		ID:    in.ID,
		MType: in.MType,
	}

	// Получение значения метрики.
	var metricValue string
	switch in.MType {
	case collector.Counter:
		metricValue = strconv.Itoa(int(in.Delta))
		metric.Delta = &in.Delta
	case collector.Gauge:
		metricValue = strconv.FormatFloat(in.Value, 'f', -1, 64)
		metric.Value = &in.Value
	default:
		// Возвращаем ошибку, если тип метрики не поддерживается.
		return &pb.SaveMetricResponse{
			ResultJSON: nil,
		}, status.Error(codes.Unimplemented, collector.ErrNotImplemented.Error())
	}

	if err := c.Collect(metric, metricValue); err != nil {
		// Возвращаем ошибку, если коллектор не смог собрать метрику.
		return &pb.SaveMetricResponse{
			ResultJSON: nil,
		}, status.Error(codes.Internal, err.Error())
	}

	// Получаем сохраненную метрику в формате JSON для ответа.
	resultJSON, err := c.GetMetricJSON(metric.ID)
	if err != nil {
		// Возвращаем ошибку, если не удалось получить метрику в формате JSON.
		return &pb.SaveMetricResponse{
			ResultJSON: nil,
		}, status.Error(codes.Internal, err.Error())
	}
	log.Println(string(resultJSON))
	return &pb.SaveMetricResponse{
		ResultJSON: resultJSON,
	}, nil
}
