package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/logger"
	"github.com/ZnNr/go-musthave-metrics.git/internal/router"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/database"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/file"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const pprofAddr string = ":6060"

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit")
		return

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
		saver = file.New(params.FileStoragePath)
	} else if params.DatabaseAddress != "" {
		db, err := sql.Open("pgx", params.DatabaseAddress)
		if err != nil {
			log.SugarLogger.Fatal(err.Error(), "open db error")
			return
		}
		saver, err = database.New(db)
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
		go saveMetrics(ctx, saver, params.StoreInterval)
	}

	if err := http.ListenAndServe(pprofAddr, nil); err != nil {
		log.SugarLogger.Fatalw(err.Error(), "pprof", "start pprof server")
	}

	// запуск сервера
	if err := http.ListenAndServe(params.FlagRunAddr, r); err != nil {
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}

}

// saveMetrics — горутина, которая периодически сохраняет метрики
func saveMetrics(ctx context.Context, saver saver, interval int) {
	for {
		if err := saver.Save(ctx, collector.Collector.Metrics); err != nil {
			log.SugarLogger.Error(err.Error(), "save error")
		}
		time.Sleep(time.Duration(interval) * time.Second) // добавляем небольшую задержку перед очередным сохранением
	}
}

// saver — интерфейс, определяющий методы восстановления и сохранения метрик.
type saver interface {
	Restore(ctx context.Context) ([]collector.StoredMetric, error)
	Save(ctx context.Context, metrics []collector.StoredMetric) error
}
