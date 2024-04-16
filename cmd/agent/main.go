package main

import (
	"context"
	"fmt"
	metricagent "github.com/ZnNr/go-musthave-metrics.git/internal/agent"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/logger"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", BuildVersion, BuildDate, BuildCommit)

	//Инициализируются параметры программы, используя пакет flags.
	params := flags.Init(flags.WithPollInterval(), flags.WithReportInterval(), flags.WithAddr(), flags.WithKey(), flags.WithRateLimit())

	errs, ctx := errgroup.WithContext(context.Background())

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit")

	}
	defer logger.Sync()
	log.SugarLogger = *logger.Sugar()

	agent := metricagent.New(params, storage.New(&collector.Collector), log.SugarLogger)
	errs.Go(func() error {
		return agent.CollectMetrics(ctx)
	})
	errs.Go(func() error {
		return agent.SendMetrics(ctx)
	})
	if err = errs.Wait(); err != nil {
		log.SugarLogger.Errorf(fmt.Sprintf("error while running agent: %s", err.Error()))
	}
}
