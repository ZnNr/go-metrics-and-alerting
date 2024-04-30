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

func main() {
	//Инициализируются параметры программы, используя пакет flags.
	params := flags.Init(
		flags.WithConfig(),
		flags.WithPollInterval(),
		flags.WithReportInterval(),
		flags.WithAddr(), flags.WithKey(),
		flags.WithRateLimit(),
		flags.WithTLSKeyPath(),
	)
	// Создание контекста и группы ошибок.
	errGroup, ctx := errgroup.WithContext(context.Background())
	// Создание логгера.
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit")
		return
	}
	defer logger.Sync()
	log.SugarLogger = *logger.Sugar()
	// Создание экземпляра metricagent.
	agent, err := metricagent.New(params, storage.New(&collector.Collector), log.SugarLogger)
	if err != nil {
		log.SugarLogger.Fatalf("Error creating agent: %v", err)
	}
	// Запуск сбора и отправки метрик параллельно.
	errGroup.Go(func() error {
		if err := agent.CollectMetrics(ctx); err != nil {
			return err
		}
		return nil
	})
	errGroup.Go(func() error {
		if err := agent.SendMetrics(ctx); err != nil {
			return err
		}
		return nil
	})
	// Ожидание завершения всех операций и обработка ошибок.
	if err = errGroup.Wait(); err != nil {
		log.SugarLogger.Errorf("Error while running agent: %s", err.Error())
	}
}
