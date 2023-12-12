package flags

import (
	"flag"
	"os"
	"strconv"
)

const (
	// Адрес и порт сервера по умолчанию
	defaultAddr string = "localhost:8080"
	// Интервал отчетов по умолчанию (в секундах)
	defaultReportInterval int = 10
	// Интервал опроса по умолчанию (в секундах)
	defaultPollInterval int = 2
)

// Option - функция, которая изменяет поля структуры параметров
type Option func(params2 *params)

// WithAddr Опция для указания адреса сервера
func WithAddr() Option {
	return func(p *params) {
		flag.StringVar(&p.FlagRunAddr, "a", defaultAddr, "address and port to run server") // Установка флага командной строки для адреса и порта сервера
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.FlagRunAddr = envRunAddr // Если переменная среды ADDRESS задана, используется ее значение
		}
	}
}

// WithReportInterval Опция для указания интервала отчетов
func WithReportInterval() Option {
	return func(p *params) {
		flag.IntVar(&p.ReportInterval, "r", defaultReportInterval, "report interval") // Установка флага командной строки для интервала отчетов
		if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
			reportIntervalEnv, err := strconv.Atoi(envReportInterval)
			if err == nil {
				p.ReportInterval = reportIntervalEnv // Если переменная среды REPORT_INTERVAL задана, проверяется и используется ее значение
			}
		}
	}
}

// WithPollInterval Опция для указания интервала опроса
func WithPollInterval() Option {
	return func(p *params) {
		flag.IntVar(&p.PollInterval, "p", defaultPollInterval, "poll interval") // Установка флага командной строки для интервала опроса
		if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
			pollIntervalEnv, err := strconv.Atoi(envPollInterval)
			if err == nil {
				p.PollInterval = pollIntervalEnv // Если переменная среды POLL_INTERVAL задана, проверяется и используется ее значение
			}
		}
	}
}

// Init Инициализация параметров с помощью опций
func Init(opts ...Option) *params {
	p := &params{}
	for _, opt := range opts {
		opt(p)
	}
	flag.Parse() // Обработка флагов командной строки
	return p
}

type params struct {
	FlagRunAddr    string // Адрес и порт сервера
	ReportInterval int    // Интервал отчетов
	PollInterval   int    // Интервал опроса
}
