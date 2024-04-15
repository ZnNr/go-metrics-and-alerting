package collector

const (
	Counter = "counter" // тип метрики для счетчика
	Gauge   = "gauge"   // тип метрики для датчика
)

type (
	// MetricRequest - a struct of metric request for upserting from the http request.
	MetricRequest struct {
		ID    string   `json:"id"`              // имя метрики
		MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	}

	// StoredMetric - a struct for storing metrics on the server.
	StoredMetric struct {
		ID           string   `json:"id"`                      // имя метрики
		MType        string   `json:"type"`                    // параметр, принимающий значение gauge или counter
		CounterValue *int64   `json:"counter_value,omitempty"` // значение метрики в случае передачи counter
		GaugeValue   *float64 `json:"gauge_value,omitempty"`   // значение метрики в случае передачи gauge
		TextValue    *string  `json:"text_value,omitempty"`    // значение метрики в случае передачи текста
	}

	collector struct {
		Metrics []StoredMetric
	}
)
