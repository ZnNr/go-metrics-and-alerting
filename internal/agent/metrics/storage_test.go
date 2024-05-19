package metrics

import (
	collector2 "github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestStorage_GopsutilMetricStore(t *testing.T) {
	log.Println("gopsutil metrics store test")
	metricsCollector := collector2.Collector()
	metricsCollector.Metrics = []collector2.StoredMetric{}
	metricsStore := New(metricsCollector)
	metricsStore.GopsutilMetricStore()
	assert.Equal(t, metricsCollector.GetAvailableMetrics(), []string{"FreeMemory", "TotalMemory", "CPUutilization1"})

}

func TestStorage_RuntimeMetricStore(t *testing.T) {
	log.Println("runtime metrics store test")
	metricsCollector := collector2.Collector()
	metricsCollector.Metrics = []collector2.StoredMetric{}
	metricsStore := New(metricsCollector)
	metricsStore.RuntimeMetricStore()
	assert.Equal(t, metricsCollector.GetAvailableMetrics(), []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue", "LastGC", "PollCount"})

}
