package collector

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/dataStorage"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

// определена функция TestCollector_Collect, которая тестирует метод CollectMetrics объекта Collector
func TestCollector_Collect(t *testing.T) {
	testCases := []struct {
		name     string                 // Название тестового случая
		storage  dataStorage.MemStorage // Инициализация хранилища данных типа MemStorage
		metric   runtime.MemStats       // Входные данные - метрики
		expected dataStorage.MemStorage // Ожидаемые данные - хранилище данных типа MemStorage
	}{
		{
			name:    "case0",
			storage: dataStorage.MemStorage{Metrics: map[string]dataStorage.Metric{}},
			metric:  runtime.MemStats{Alloc: 1, Sys: 1, GCCPUFraction: 5.543},
			expected: dataStorage.MemStorage{Metrics: map[string]dataStorage.Metric{
				"Alloc":         {MetricType: "gauge", Value: uint64(1)},
				"Sys":           {MetricType: "gauge", Value: uint64(1)},
				"GCCPUFraction": {MetricType: "gauge", Value: 5.543},
			}},
		},
	}
	// Итерация по тестовым случаям
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metric := runtime.MemStats{Alloc: 1, Sys: 1, GCCPUFraction: 5.543}
			// Создание экземпляра Collector с передачей хранилища данных
			metricsCollector := New(&tt.storage)
			// Сбор метрик для заданных входных данных
			metricsCollector.CollectMetrics(&metric)
			// Проверка совпадения ожидаемых данных с фактическими данными в хранилище
			assert.Equal(t, tt.expected.Metrics["Alloc"], tt.storage.Metrics["Alloc"])
			assert.Equal(t, tt.expected.Metrics["Sys"], tt.storage.Metrics["Sys"])
			assert.Equal(t, tt.expected.Metrics["GCCPUFraction"], tt.storage.Metrics["GCCPUFraction"])
		})
	}
}
