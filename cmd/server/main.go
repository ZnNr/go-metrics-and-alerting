package main

import (
	"context"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/logger"
	"github.com/ZnNr/go-musthave-metrics.git/internal/router"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/database"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/file"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	log.SugarLogger = *logger.Sugar()

	params := flags.Init(flags.WithAddr(), flags.WithStoreInterval(), flags.WithFileStoragePath(), flags.WithRestore(), flags.WithDatabase(), flags.WithKey())

	r := router.New(*params)

	log.SugarLogger.Infow("Starting server", "addr", params.FlagRunAddr)
	// Инициализация ресторера
	// инициализация переменной saver типа saver, которая будет использоваться для восстановления и сохранения метрик.
	var saver saver
	if params.FileStoragePath != "" && params.DatabaseAddress == "" {
		saver = file.New(params)
	} else if params.DatabaseAddress != "" {
		saver, err = database.New(params)
		if err != nil {
			log.SugarLogger.Errorf(err.Error())
		}
	}

	// востановление предыдущих метрик
	ctx := context.Background()
	if params.Restore && (params.FileStoragePath != "" || params.DatabaseAddress != "") {
		metrics, err := saver.Restore(ctx)
		if err != nil {
			log.SugarLogger.Error(err.Error(), "restore error")
		}
		collector.Collector.Metrics = metrics
		log.SugarLogger.Info("metrics restored")
	}

	// востановление метрик
	if params.DatabaseAddress != "" || params.FileStoragePath != "" {
		go saveMetrics(ctx, saver)
	}

	// запуск сервера
	if err := http.ListenAndServe(params.FlagRunAddr, r); err != nil {
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}

// saveMetrics — горутина, которая периодически сохраняет метрики
func saveMetrics(ctx context.Context, saver saver) {
	for {
		if err := saver.Save(ctx, collector.Collector.Metrics); err != nil {
			log.SugarLogger.Error(err.Error(), "save error")
		}
	}
}

// saver — интерфейс, определяющий методы восстановления и сохранения метрик.
type saver interface {
	Restore(ctx context.Context) ([]collector.StoredMetric, error)
	Save(ctx context.Context, metrics []collector.StoredMetric) error
}
