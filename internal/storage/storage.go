package storage

//данный код представляет основу для работы с метриками и их хранения в памяти.

// Metric metric  структура, представляющая метрику.
type Metric struct {
	Value      interface{} // Значение метрики
	MetricType string      // Тип метрики
}

// MemStorage - структура, представляющая хранилище памяти для метрик.
type MemStorage struct {
	Metrics map[string]Metric //// Мапа метрик, где ключ - строковый идентификатор, значение - метрика
}

// MetricsStorage - переменная, представляющая хранилище метрик в памяти.
var MetricsStorage = MemStorage{Metrics: make(map[string]Metric)}
