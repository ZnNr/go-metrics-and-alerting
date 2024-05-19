// Package flags предоставляет функционал для обработки флагов командной строки и настройки параметров приложения.
// Используется пакет flag стандартной библиотеки для работы с флагами и настройками.
package flags

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

const (
	defaultRateLimit = 1
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

// WithTrustedSubnet возвращает опцию для установки доверенной подсети.
func WithTrustedSubnet() Option {
	return func(p *Params) {
		flag.StringVar(&p.TrustedSubnet, "t", p.TrustedSubnet, "trusted subnet")
		if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
			p.TrustedSubnet = envTrustedSubnet
		}
	}
}

// WithRateLimit создает опцию для установки ограничения запросов.
func WithRateLimit() Option {
	return func(p *Params) {
		flag.IntVar(&p.RateLimit, "l", p.RateLimit, "max requests to send on server")
		if envKey := os.Getenv("RATE_LIMIT"); envKey != "" {
			p.Key = envKey
		}
	}
}

// WithKey создает опцию для установки ключа подписки.
func WithKey() Option {
	return func(p *Params) {
		flag.StringVar(&p.Key, "k", p.Key, "key for using hash subscription")
		if envKey := os.Getenv("KEY"); envKey != "" {
			p.Key = envKey
		}
	}
}

// WithDatabase - Опция для указания подключения к базе данных
func WithDatabase() Option {
	return func(p *Params) {
		result := ""
		flag.StringVar(&result, "d", p.DatabaseAddress, "connection string for db")
		if envDBAddr := os.Getenv("DATABASE_DSN"); envDBAddr != "" {
			result = envDBAddr
		}
		p.DatabaseAddress = result
	}
}

// WithAddr Опция для указания адреса сервера
func WithAddr() Option {
	return func(p *Params) {
		flag.StringVar(&p.FlagRunAddr, "a", p.FlagRunAddr, "address and port to run server")
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.FlagRunAddr = envRunAddr
		}
	}
}

// WithReportInterval Опция для указания интервала отчетов
func WithReportInterval() Option {
	return func(p *Params) {
		flag.IntVar(&p.ReportInterval, "r", p.ReportInterval, "report interval")
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
		flag.IntVar(&p.PollInterval, "p", p.PollInterval, "poll interval")
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
		flag.IntVar(&p.StoreInterval, "i", p.StoreInterval, "store interval in seconds")
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
		flag.StringVar(&p.FileStoragePath, "f", p.FileStoragePath, "file name for metrics collection")
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
		flag.BoolVar(&p.Restore, "r", p.Restore, "restore data from file")
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
		flag.StringVar(&p.CryptoKeyPath, "crypto-key", p.CryptoKeyPath, "crypto key path")
		if envCryptoKeyPath := os.Getenv("CRYPTO_KEY"); envCryptoKeyPath != "" {
			p.CryptoKeyPath = envCryptoKeyPath
		}
	}
}

func WithConfig() Option {
	return func(p *Params) {
		var configPath string
		flag.StringVar(&configPath, "c", "", "config path")
		for i, arg := range os.Args {
			if arg == "-c" || arg == "-config" {
				configPath = os.Args[i+1]
			}
		}
		// priority for the env variables
		if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
			configPath = envConfigPath
		}
		if configPath != "" {
			config, err := os.ReadFile(configPath)
			if err != nil {
				log.Printf("config path was provided, but an error ocurred while opening: %s\n", err.Error())
				log.Println("using default values, values from command line and from env variables...")
				return
			}
			if err = json.Unmarshal(config, p); err != nil {
				log.Printf("error while parsing config: %s\n", err.Error())
			}
		}
	}
}

// Init Инициализация параметров с помощью опций
func Init(opts ...Option) *Params {
	p := &Params{
		RateLimit:       defaultRateLimit,
		FlagRunAddr:     defaultAddr,
		ReportInterval:  defaultReportInterval,
		PollInterval:    defaultPollInterval,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestore,
	}

	for _, opt := range opts {
		opt(p)
	}
	flag.Parse()
	return p
}

type Params struct {
	FlagRunAddr     string `json:"address"`         // Адрес и порт сервера
	DatabaseAddress string `json:"database_dsn"`    // Адрес базы данных
	ReportInterval  int    `json:"report_interval"` // Интервал отчетов
	PollInterval    int    `json:"poll_interval"`   // Интервал опроса
	StoreInterval   int    `json:"store_interval"`  // Интервал сохранения
	FileStoragePath string `json:"store_file"`      // Путь к хранилищу файлов
	Restore         bool   `json:"restore"`         // Флаг восстановления данных
	Key             string `json:"hash_key"`        // Ключ подписки
	RateLimit       int    `json:"rate_limit"`      // Ограничение запросов
	CryptoKeyPath   string `json:"crypto_key"`      // Путь к криптографическому ключу
	TrustedSubnet   string `json:"trusted_subnet"`  // доверенная подсеть
}
