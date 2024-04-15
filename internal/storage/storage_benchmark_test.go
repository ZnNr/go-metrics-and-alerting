package storage

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"log"
	"testing"
)

func BenchmarkStore_GopsutilMetricStore(b *testing.B) {
	log.Println("gopsutil metrics store benchmark")
	for i := 0; i < b.N; i++ {
		metricsCollector := &collector.Collector
		metricsStore := New(metricsCollector)
		metricsStore.GopsutilMetricStore()
	}

}

func BenchmarkStorage_RuntimeMetricStore(b *testing.B) {
	log.Println("runtime metrics store benchmark")
	for i := 0; i < b.N; i++ {
		metricsCollector := &collector.Collector
		metricsStore := New(metricsCollector)
		metricsStore.RuntimeMetricStore()
	}

}
