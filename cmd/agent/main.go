package main

import (
	"context"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/runner/agent"
)

func main() {
	//Инициализируются параметры программы, используя пакет flags.
	params := flags.Init(
		flags.WithConfig(),
		flags.WithPollInterval(),
		flags.WithReportInterval(),
		flags.WithAddr(),
		flags.WithKey(),
		flags.WithRateLimit(),
		flags.WithTLSKeyPath(),
	)

	// Создание контекста для возможности отмены операций.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Создаем экземпляр Runner на основе параметров.
	runner := agent.New(params)
	// Запускаем Runner, передавая контекст выполнения.
	runner.Run(ctx)
}
