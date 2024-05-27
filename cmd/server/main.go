package main

import (
	"context"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	serverRunner "github.com/ZnNr/go-musthave-metrics.git/internal/server/runner/server"
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
		flags.WithTrustedSubnet(),
		flags.WithGrpc(),
		flags.WithGrpcAddr(),
	)

	// Создание контекста для возможности отмены операций.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Восстановление предыдущих метрик.
	runner := serverRunner.New(params)

	runner.Run(ctx)
}
