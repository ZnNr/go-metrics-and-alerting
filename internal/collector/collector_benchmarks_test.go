package collector

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

var testBenchCollector = createTestBenchCollector()

func createTestBenchCollector() collector {
	return collector{
		[]StoredMetric{
			{
				ID:         "Alloc",
				MType:      "gauge",
				GaugeValue: PtrFloat64(10),
				TextValue:  PtrString("10"),
			},
			{
				ID:         "GCCPUFraction",
				MType:      "gauge",
				GaugeValue: PtrFloat64(5.543),
				TextValue:  PtrString("5.543"),
			},
			{
				ID:           "IO",
				MType:        "counter",
				CounterValue: PtrInt64(5),
				TextValue:    PtrString("5"),
			},
			{
				ID:         "Mem",
				MType:      "gauge",
				GaugeValue: PtrFloat64(500.1992),
				TextValue:  PtrString("500.1992"),
			},
			{
				ID:           "Requests",
				MType:        "counter",
				CounterValue: PtrInt64(100500),
				TextValue:    PtrString("100500"),
			},
		},
	}
}

func BenchmarkCollector_Collect(b *testing.B) {
	log.Println("collect benchmark")
	metric := MetricRequest{
		ID:    "new",
		MType: "gauge",
		Value: PtrFloat64(50.1001),
	}
	var err error
	for i := 0; i < b.N; i++ {
		err = testBenchCollector.Collect(metric, "50.1001")
	}
	assert.NoError(b, err)

}
func BenchmarkCollector_GetAvailableMetrics(b *testing.B) {
	log.Println("get available metrics benchmark")
	for i := 0; i < b.N; i++ {
		testBenchCollector.GetAvailableMetrics()
	}

}

func BenchmarkCollector_GetMetric(b *testing.B) {
	log.Println("get metric benchmark")
	metricName := "Requests"
	var err error // Глобальная переменная для ошибок
	for i := 0; i < b.N; i++ {
		_, err = testBenchCollector.GetMetric(metricName)
	}
	assert.NoError(b, err)

}

func BenchmarkCollector_GetMetricJSON(b *testing.B) {
	log.Println("get metric json benchmark")
	metricName := "Requests"
	var err error // Глобальная переменная для ошибок
	for i := 0; i < b.N; i++ {
		_, err = testBenchCollector.GetMetricJSON(metricName)
	}
	assert.NoError(b, err)

}

func BenchmarkCollector_UpsertMetric(b *testing.B) {
	log.Println("upsert metric benchmark")
	metric := StoredMetric{
		ID:         "Alloc",
		MType:      "gauge",
		GaugeValue: PtrFloat64(3),
		TextValue:  PtrString("3"),
	}
	for i := 0; i < b.N; i++ {
		testBenchCollector.UpsertMetric(metric)
	}

}

// тестирование бенчмарка в комплексном (грубо упрощенном) сценарии из двух условных методов
type testCase struct {
	name    string
	metrics MetricRequest
}

func (tc *testCase) method1(b *testing.B) {
	log.Println("get available metrics benchmark")
	for i := 0; i < b.N; i++ {
		testBenchCollector.GetAvailableMetrics()
	}

}

func (tc *testCase) method2(b *testing.B) {
	log.Println("upsert metric benchmark")
	metric := StoredMetric{
		ID:         "Alloc",
		MType:      "gauge",
		GaugeValue: PtrFloat64(3),
		TextValue:  PtrString("3"),
	}
	for i := 0; i < b.N; i++ {
		testBenchCollector.UpsertMetric(metric)
	}

}

func BenchmarkCollector_ComplexScenario(b *testing.B) {
	testCases := []testCase{
		{
			name: "Scenario 1",
			metrics: MetricRequest{
				ID:    "IO",
				MType: "counter",
				Value: PtrFloat64(50.1001),
			}, // Инициализация метрик для сценария 1
		},
		{
			name: "Scenario 2",
			metrics: MetricRequest{

				ID:    "Alloc",
				MType: "gauge",
				Value: PtrFloat64(50.1001),
			}, // Инициализация метрик для сценария 2
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			tc.method1(b) // Запуск бенчмарка для метода 1
		})
		b.Run(tc.name, func(b *testing.B) {
			tc.method2(b) // Запуск бенчмарка для метода 2
		})
	}
}
