package main

import (
	"context"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/runner"
)

func main() {
	// Инициализация параметров программы.
	params := flags.Init(
		flags.WithConfig(),
		flags.WithAddr(),
		flags.WithStoreInterval(),
		flags.WithFileStoragePath(),
		flags.WithRestore(),
		flags.WithDatabase(),
		flags.WithKey(),
		flags.WithTLSKeyPath(),
	)

	// Создание контекста для возможности отмены операций.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Восстановление предыдущих метрик.
	serverRunner := runner.New(params)
	serverRunner.Run(ctx)
}
