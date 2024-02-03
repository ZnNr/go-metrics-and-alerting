package main

import (
	"context"
	metricagent "github.com/ZnNr/go-musthave-metrics.git/internal/agent"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	//Инициализируются параметры программы, используя пакет flags.
	params := flags.Init(
		flags.WithPollInterval(),
		flags.WithReportInterval(),
		flags.WithAddr(),
		flags.WithKey(),
	)

	errs, ctx := errgroup.WithContext(context.Background())
	agent := metricagent.New(params, storage.New(&collector.Collector))
	errs.Go(func() error {
		return agent.CollectMetrics(ctx)
	})

	errs.Go(func() error {

		return agent.SendMetrics(ctx)
	})

	_ = errs.Wait()
}
