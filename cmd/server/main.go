package main

import (
	"context"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/runner/server"
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
	)
	fmt.Println(params)

	// Создание контекста для возможности отмены операций.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Восстановление предыдущих метрик.
	serverRunner := server.New(params)

	serverRunner.Run(ctx)
}
