package metrics

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"runtime"
	"strconv"
)

// RuntimeMetricStore метод используется для сбора метрик и сохранения их в хранилище.
// a method for capturing and upserting runtime metrics.
func (st *Storage) RuntimeMetricStore() {
	metrics := runtime.MemStats{}  //создается переменная metrics типа runtime.MemStats, которая представляет собой статистику памяти
	runtime.ReadMemStats(&metrics) //вызывается функциякоторая заполняет структуру metrics актуальными данными о памяти.
	//Происходит вызов метода Collect() объекта st.metricsCollector для каждого из собранных показателей памяти.
	//Каждый вызов передает имя метрики, тип и значение метрики, преобразованное в строку.
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "Alloc", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.Alloc)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.Alloc), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "BuckHashSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.BuckHashSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.BuckHashSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "Frees", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.Frees)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.Frees), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "GCCPUFraction", MType: "gauge", GaugeValue: &metrics.GCCPUFraction, TextValue: collector.PtrString(strconv.FormatFloat(metrics.GCCPUFraction, 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "GCSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.GCSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.GCSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "HeapAlloc", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.HeapAlloc)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.HeapAlloc), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "HeapIdle", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.HeapIdle)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.HeapIdle), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "HeapInuse", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.HeapInuse)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.HeapInuse), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "HeapObjects", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.HeapObjects)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.HeapObjects), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "HeapReleased", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.HeapReleased)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.HeapReleased), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "HeapSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.HeapSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.HeapSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "Lookups", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.Lookups)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.Lookups), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "MCacheInuse", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.MCacheInuse)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.MCacheInuse), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "MCacheSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.MCacheSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.MCacheSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "MSpanInuse", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.MSpanInuse)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.MSpanInuse), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "MSpanSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.MSpanSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.MSpanSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "Mallocs", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.Mallocs)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.Mallocs), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "NextGC", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.NextGC)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.NextGC), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "NumForcedGC", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.NumForcedGC)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.NumForcedGC), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "NumGC", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.NumGC)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.NumGC), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "OtherSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.OtherSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.OtherSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "PauseTotalNs", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.PauseTotalNs)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.PauseTotalNs), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "StackInuse", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.StackInuse)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.StackInuse), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "StackSys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.StackSys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.StackSys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "Sys", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.Sys)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.Sys), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "TotalAlloc", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.TotalAlloc)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.TotalAlloc), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "RandomValue", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(rand.Int())), TextValue: collector.PtrString(strconv.FormatFloat(float64(rand.Int()), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "LastGC", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(metrics.LastGC)), TextValue: collector.PtrString(strconv.FormatFloat(float64(metrics.LastGC), 'f', 11, 64))})

	cnt, _ := st.metricsCollector.GetMetric("PollCount")
	counter := int64(0)
	if cnt.CounterValue != nil {
		counter = *cnt.CounterValue + 1
	}
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "PollCount", MType: "counter", CounterValue: collector.PtrInt64(counter), TextValue: collector.PtrString(strconv.Itoa(int(counter)))})
}

// GopsutilMetricStore метод для сбора и сохранения метрик gopsutil
func (st *Storage) GopsutilMetricStore() {
	v, _ := mem.VirtualMemory()
	cp, _ := cpu.Percent(0, false)

	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "FreeMemory", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(v.Free)), TextValue: collector.PtrString(strconv.FormatFloat(float64(v.Free), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "TotalMemory", MType: "gauge", GaugeValue: collector.PtrFloat64(float64(v.Total)), TextValue: collector.PtrString(strconv.FormatFloat(float64(v.Total), 'f', 11, 64))})
	st.metricsCollector.UpsertMetric(collector.StoredMetric{ID: "CPUutilization1", MType: "gauge", GaugeValue: collector.PtrFloat64(cp[0]), TextValue: collector.PtrString(strconv.FormatFloat(cp[0], 'f', 11, 64))})

}

// New - это конструктор, который создает и возвращает новый экземпляр структуры metrics.
// Он принимает аргумент metricsCollector, который должен быть реализацией интерфейса collectorImpl
func New(metricsCollector collectorImpl) *Storage {
	return &Storage{
		metricsCollector: metricsCollector,
	}
}

// Storage определены два поля:
// metricsCollector - тип этого поля задан как collectorImpl, это поле будет использоваться для сбора и хранения метрик.
// полю metricsCollector можно присвоить любое значение, которое соответствует интерфейсу collectorImpl.
type Storage struct {
	metricsCollector collectorImpl
}

// Интерфейс collectorImpl определяет только один метод Collect, который принимает три аргумента: metricName (имя метрики), metricType (тип метрики) и metricValue (значение метрики), и возвращает ошибку
type collectorImpl interface {
	UpsertMetric(metric collector.StoredMetric)
	GetMetric(metricName string) (collector.StoredMetric, error)
}
