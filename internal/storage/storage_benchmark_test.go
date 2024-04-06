package storage

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"testing"
)

func BenchmarkStore_GopsutilMetricStore(b *testing.B) {
	b.Run("gopsutil metrics store benchmark", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metricsCollector := &collector.Collector
			metricsStore := New(metricsCollector)
			metricsStore.GopsutilMetricStore()
		}
	})
}

func BenchmarkStorage_RuntimeMetricStore(b *testing.B) {
	b.Run("runtime metrics store benchmark", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metricsCollector := &collector.Collector
			metricsStore := New(metricsCollector)
			metricsStore.RuntimeMetricStore()
		}
	})
}
