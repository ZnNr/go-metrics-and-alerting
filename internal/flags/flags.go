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
	// Интервал сохранения по умолчанию (в секундах)
	defaultStoreInterval int = 300
	// Путь к файлу хранения по умолчанию
	defaultFileStoragePath string = "/tmp/metrics-db.json"
	// Восстанавливать состояние по умолчанию или нет
	defaultRestore bool = true
)

// Option - функция, которая изменяет поля структуры параметров
type Option func(params *Params)

// WithAddr Опция для указания адреса сервера
func WithAddr() Option {
	return func(p *Params) {
		flag.StringVar(&p.FlagRunAddr, "a", defaultAddr, "address and port to run server") // Установка флага командной строки для адреса и порта сервера
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.FlagRunAddr = envRunAddr // Если переменная среды ADDRESS задана, используется ее значение
		}
	}
}

// WithReportInterval Опция для указания интервала отчетов
func WithReportInterval() Option {
	return func(p *Params) {
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
	return func(p *Params) {
		flag.IntVar(&p.PollInterval, "p", defaultPollInterval, "poll interval") // Установка флага командной строки для интервала опроса
		if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
			pollIntervalEnv, err := strconv.Atoi(envPollInterval)
			if err == nil {
				p.PollInterval = pollIntervalEnv // Если переменная среды POLL_INTERVAL задана, проверяется и используется ее значение
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

// Init Инициализация параметров с помощью опций
func Init(opts ...Option) *Params {
	p := &Params{}
	for _, opt := range opts {
		opt(p)
	}
	flag.Parse()
	return p
}

type Params struct {
	FlagRunAddr     string // Адрес и порт сервера
	ReportInterval  int    // Интервал отчетов
	PollInterval    int    // Интервал опроса
	StoreInterval   int    // Интервал сохранения
	FileStoragePath string // Путь к хранилищу файлов
	Restore         bool   // Флаг восстановления данных
}
