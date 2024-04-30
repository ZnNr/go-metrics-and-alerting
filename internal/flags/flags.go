// Package flags предоставляет функционал для обработки флагов командной строки и настройки параметров приложения.
// Используется пакет flag стандартной библиотеки для работы с флагами и настройками.
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
	defaultReportInterval int = 5
	// Интервал опроса по умолчанию (в секундах)
	defaultPollInterval int = 1
	// Интервал сохранения по умолчанию (в секундах)
	defaultStoreInterval int = 15
	// Путь к файлу хранения по умолчанию
	defaultFileStoragePath string = "/tmp/metrics-db.json"
	// Восстанавливать состояние по умолчанию или нет
	defaultRestore bool = true
)

// Option - функция, которая изменяет поля структуры параметров
type Option func(params *Params)

// WithRateLimit создает опцию для установки ограничения запросов.
func WithRateLimit() Option {
	return func(p *Params) {
		flag.IntVar(&p.RateLimit, "l", 1, "max requests to send on server")
		if envKey := os.Getenv("RATE_LIMIT"); envKey != "" {
			p.Key = envKey
		}
	}
}

// WithKey создает опцию для установки ключа подписки.
func WithKey() Option {
	return func(p *Params) {
		flag.StringVar(&p.Key, "k", "", "key for using hash subscription")
		if envKey := os.Getenv("KEY"); envKey != "" {
			p.Key = envKey
		}
	}
}

// WithDatabase - Опция для указания подключения к базе данных
func WithDatabase() Option {
	return func(p *Params) {
		result := ""
		flag.StringVar(&result, "d", "", "connection string for db")
		if envDBAddr := os.Getenv("DATABASE_DSN"); envDBAddr != "" {
			result = envDBAddr
		}
		p.DatabaseAddress = result
	}
}

// WithAddr Опция для указания адреса сервера
func WithAddr() Option {
	return func(p *Params) {
		flag.StringVar(&p.FlagRunAddr, "a", defaultAddr, "address and port to run server")
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.FlagRunAddr = envRunAddr
		}
	}
}

// WithReportInterval Опция для указания интервала отчетов
func WithReportInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.ReportInterval, "r", defaultReportInterval, "report interval")
		if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
			reportIntervalEnv, err := strconv.Atoi(envReportInterval)
			if err == nil {
				p.ReportInterval = reportIntervalEnv
			}
		}
	}
}

// WithPollInterval Опция для указания интервала опроса
func WithPollInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.PollInterval, "p", defaultPollInterval, "poll interval")
		if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
			pollIntervalEnv, err := strconv.Atoi(envPollInterval)
			if err == nil {
				p.PollInterval = pollIntervalEnv
			}
		}
	}
}

// WithStoreInterval Опция, которая позволяет установить интервал сохранения данных
func WithStoreInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.StoreInterval, "i", defaultStoreInterval, "store interval in seconds")
		if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
			storeIntervalEnv, err := strconv.Atoi(envStoreInterval)
			if err == nil {
				p.StoreInterval = storeIntervalEnv
			}
		}
	}
}

// WithFileStoragePath Опция для указания путя хранения файла
func WithFileStoragePath() Option {
	return func(p *Params) {
		flag.StringVar(&p.FileStoragePath, "f", defaultFileStoragePath, "file name for metrics collection")
		if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
			fileStoragePath, err := strconv.Atoi(envFileStoragePath)
			if err == nil {
				p.StoreInterval = fileStoragePath
			}
		}
	}
}

// WithRestore Опция позволяет включить функциональность восстановления данных
func WithRestore() Option {
	return func(p *Params) {
		flag.BoolVar(&p.Restore, "r", defaultRestore, "restore data from file")
		if envRestore := os.Getenv("RESTORE"); envRestore != "" {
			restore, err := strconv.Atoi(envRestore)
			if err == nil {
				p.StoreInterval = restore
			}
		}
	}
}

// WithTLSKeyPath Опция устанавливает путь к криптографическому
func WithTLSKeyPath() Option {
	return func(p *Params) {
		flag.StringVar(&p.CryptoKeyPath, "crypto-key", "", "crypto key path")
		if envCryptoKeyPath := os.Getenv("CRYPTO_KEY"); envCryptoKeyPath != "" {
			p.CryptoKeyPath = envCryptoKeyPath
		}
	}
}

// Init Инициализация параметров с помощью опций
func Init(opts ...Option) *Params {
	p := &Params{}
	for _, opt := range opts {
		opt(p)
	}
	flag.Parse()
	return p
}

// Params содержит параметры приложения.
type Params struct {
	FlagRunAddr     string // Адрес и порт сервера
	DatabaseAddress string // Адрес базы данных
	ReportInterval  int    // Интервал отчетов
	PollInterval    int    // Интервал опроса
	StoreInterval   int    // Интервал сохранения
	FileStoragePath string // Путь к хранилищу файлов
	Restore         bool   // Флаг восстановления данных
	Key             string // Ключ подписки
	RateLimit       int    // Ограничение запросов
	CryptoKeyPath   string // Путь к криптографическому ключу
}
