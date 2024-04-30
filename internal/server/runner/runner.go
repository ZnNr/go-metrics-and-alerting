package runner

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/logger"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/database"
	"github.com/ZnNr/go-musthave-metrics.git/internal/saver/file"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/router"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	pprofAddr    string = ":6060"
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

type Runner struct {
	saver           saver
	metricsInterval time.Duration
	router          *chi.Mux
	isRestore       bool
	storeInterval   int
	runAddress      string
	tlsKey          string
}

func New(params *flags.Params) *Runner {
	// init restorer
	saver, err := initSaver(params)
	if err != nil {
		log.SugarLogger.Fatalw(err.Error(), "error", "init metrics saver")
	}
	r, err := router.New(*params)
	if err != nil {
		log.SugarLogger.Fatalw(err.Error(), "error", "creating router")
	}
	return &Runner{
		saver:           saver,
		metricsInterval: time.Duration(params.StoreInterval),
		router:          r,
		isRestore:       params.Restore,
		storeInterval:   params.StoreInterval,
		runAddress:      params.FlagRunAddr,
		tlsKey:          params.CryptoKeyPath,
	}
}

func (r *Runner) Run(ctx context.Context) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit")
		return
	}
	defer logger.Sync()
	log.SugarLogger = *logger.Sugar()

	log.SugarLogger.Infow("Build information", "version", buildVersion, "date", buildDate, "commit", buildCommit)

	if r.isRestore {
		if metrics, err := r.saver.Restore(ctx); err != nil {
			log.SugarLogger.Errorw("Restore error", "error", err.Error())
		} else {
			collector.Collector.Metrics = metrics
			log.SugarLogger.Info("Metrics restored")
		}
	}

	// Regularly save metrics if needed
	go r.saveMetrics(ctx, r.storeInterval)

	// Start pprof server
	go func() {
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			log.SugarLogger.Fatalw("Failed to start pprof server", "error", err.Error())
		}
	}()

	// Run server
	log.SugarLogger.Infow("Starting server", "addr", r.runAddress)
	if err := http.ListenAndServe(r.runAddress, r.router); err != nil {
		log.SugarLogger.Fatalw("Failed to start server", "error", err.Error())
	}
}

func (r *Runner) runServer() {
	if err := http.ListenAndServe(r.runAddress, r.router); err != nil {
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}

func (r *Runner) saveMetrics(ctx context.Context, interval int) {
	ticker := time.NewTicker(time.Duration(interval))
	defer ticker.Stop()
	if err := r.saver.Save(ctx, collector.Collector.Metrics); err != nil {
		log.SugarLogger.Error(err.Error(), "save error")
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.saver.Save(ctx, collector.Collector.Metrics); err != nil {
				log.SugarLogger.Error(err.Error(), "save error")
			}
		}
	}
}

func initSaver(params *flags.Params) (saver, error) {
	if params.DatabaseAddress != "" {
		return initDatabaseSaver(params.DatabaseAddress)
	} else if params.FileStoragePath != "" {
		return initFileSaver(params.FileStoragePath), nil
	}
	return nil, fmt.Errorf("neither file path nor database address was specified")
}

func initDatabaseSaver(databaseAddress string) (saver, error) {
	db, err := sql.Open("pgx", databaseAddress)
	if err != nil {
		return nil, err
	}
	return database.New(db)
}

func initFileSaver(fileStoragePath string) saver {
	return file.New(fileStoragePath)
}

type saver interface {
	Restore(ctx context.Context) ([]collector.StoredMetric, error)
	Save(ctx context.Context, metrics []collector.StoredMetric) error
}
