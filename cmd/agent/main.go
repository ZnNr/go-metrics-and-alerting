package main

import (
	"context"
	agentRunner "github.com/ZnNr/go-musthave-metrics.git/internal/agent/runner"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
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
		flags.WithGrpcAddr(),
	)

	// Создание контекста для возможности отмены операций.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Создаем экземпляр Runner на основе параметров.
	runner := agentRunner.New(params)
	// Запускаем Runner, передавая контекст выполнения.
	runner.Run(ctx)
}
