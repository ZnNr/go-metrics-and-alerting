package server

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/middlewares/logger"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/database"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/file"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/router"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Константы, связанные с адресом pprof и информацией о сборке приложения.
const (
	pprofAddr    string = ":8090"
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// Runner Структура, представляющая собой главный компонент приложения сервера.
type Runner struct {
	saver           saver
	metricsInterval time.Duration
	isRestore       bool
	storeInterval   int
	tlsKey          string
	appSrv          server
	pprofSrv        server
	logger          *zap.SugaredLogger
	signals         chan os.Signal
}

// New создает экземпляр Runner с использованием параметров flags.Params.
func New(params *flags.Params) *Runner {
	// init logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit")
		return nil
	}
	defer logger.Sync()
	log.SugarLogger = *logger.Sugar()

	// init saver (file or db)
	saver, err := initSaver(params)
	if err != nil {
		log.SugarLogger.Fatalw(err.Error(), "error", "init metrics saver")
	}

	// init router
	r, err := router.New(*params)
	if err != nil {
		log.SugarLogger.Fatalw(err.Error(), "error", "creating router")
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	return &Runner{
		saver:           saver,
		metricsInterval: time.Duration(params.StoreInterval),
		isRestore:       params.Restore,
		storeInterval:   params.StoreInterval,
		tlsKey:          params.CryptoKeyPath,
		appSrv: &http.Server{
			Addr:    params.FlagRunAddr,
			Handler: r,
		},
		pprofSrv: &http.Server{
			Addr:    pprofAddr,
			Handler: nil,
		},
		signals: sigs,
		logger:  &log.SugarLogger,
	}
}

// Run запускает основной цикл приложения, обрабатывая сохранение метрик, pprof и сигналы.
func (r *Runner) Run(ctx context.Context) {
	// Логгирование информации о сборке.
	r.logger.Info("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	// Восстановление предыдущих метрик, если необходимо.
	if r.isRestore {
		metrics, err := r.saver.Restore(ctx)
		if err != nil {
			r.logger.Error(err.Error(), "restore error")
		}
		collector.Collector().Metrics = metrics
		r.logger.Info("metrics restored")
	}

	// Регулярное сохранение метрик.
	go r.saveMetrics(ctx, r.storeInterval)

	// Запуск pprof.
	go func() {
		if err := r.pprofSrv.ListenAndServe(); err != nil {
			r.logger.Fatalw(err.Error(), "pprof", "start pprof server")
		}
	}()

	// Обработка сигналов.
	go func() {
		sig := <-r.signals
		r.logger.Info(fmt.Sprintf("got signal: %s", sig.String()))
		// Сохранение метрик.
		if err := r.saver.Save(ctx, collector.Collector().Metrics); err != nil {
			r.logger.Error(err.Error(), "save error")
		} else {
			r.logger.Info("metrics was successfully saved")
		}
		// Graceful shutdown.
		if err := r.appSrv.Shutdown(ctx); err != nil {
			r.logger.Error(fmt.Sprintf("error while server shutdown: %s", err.Error()), "server shutdown error")
			return
		}
	}()

	// Запуск сервера.
	r.logger.Info("Starting server")
	if err := r.appSrv.ListenAndServe(); err != nil {
		r.logger.Fatalw(err.Error(), "event", "start server")
	}
}

// saveMetrics сохраняет метрики с указанным интервалом.
func (r *Runner) saveMetrics(ctx context.Context, interval int) {
	ticker := time.NewTicker(time.Duration(interval))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.saver.Save(ctx, collector.Collector().Metrics); err != nil {
				r.logger.Error(err.Error(), "save error")
			}
		}
	}
}

// initSaver инициализирует saver (файл или базу данных) в зависимости от параметров.
func initSaver(params *flags.Params) (saver, error) {
	if params.DatabaseAddress != "" {
		return initDatabaseSaver(params.DatabaseAddress)
	} else if params.FileStoragePath != "" {
		return initFileSaver(params.FileStoragePath), nil
	}
	return nil, fmt.Errorf("neither file path nor database address was specified")
}

// initDatabaseSaver инициализирует saver для базы данных с указанным адресом.
func initDatabaseSaver(databaseAddress string) (saver, error) {
	db, err := sql.Open("pgx", databaseAddress)
	if err != nil {
		return nil, err
	}
	return database.New(db)
}

// initFileSaver инициализирует saver для работы с файлом по указанному пути.
func initFileSaver(fileStoragePath string) saver {
	return file.New(fileStoragePath)
}

// saver - интерфейс для работы с метриками.
//
//go:generate mockery --inpackage --disable-version-string --filename saver_mock.go --name saver
type saver interface {
	Restore(ctx context.Context) ([]collector.StoredMetric, error)
	Save(ctx context.Context, metrics []collector.StoredMetric) error
}

// server - интерфейс для работы с HTTP-сервером.
//
//go:generate mockery --inpackage --disable-version-string --filename server_mock.go --name server
type server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
