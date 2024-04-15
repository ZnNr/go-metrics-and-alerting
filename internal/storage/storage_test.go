package storage

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestStorage_GopsutilMetricStore(t *testing.T) {
	log.Println("gopsutil metrics store test")
	c := collector.Collector
	metricsCollector := &c
	metricsStore := New(metricsCollector)
	metricsStore.GopsutilMetricStore()
	assert.Equal(t, metricsCollector.GetAvailableMetrics(), []string{"FreeMemory", "TotalMemory", "CPUutilization1"})

}

func TestStorage_RuntimeMetricStore(t *testing.T) {
	log.Println("runtime metrics store test")
	c := collector.Collector
	metricsCollector := &c
	metricsStore := New(metricsCollector)
	metricsStore.RuntimeMetricStore()
	assert.Equal(t, metricsCollector.GetAvailableMetrics(), []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue", "LastGC", "PollCount"})

}
