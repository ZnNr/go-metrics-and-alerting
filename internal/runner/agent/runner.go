package agent

import (
	"context"
	"fmt"
	metricagent "github.com/ZnNr/go-musthave-metrics.git/internal/agent"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Runner struct {
	params  *flags.Params
	logger  *zap.SugaredLogger
	signals chan os.Signal
}

func newLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}

func New(params *flags.Params) *Runner {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	logger, err := newLogger()
	if err != nil {
		fmt.Println("error while creating logger, exit")
		return nil
	}

	return &Runner{
		params:  params,
		signals: sigs,
		logger:  logger,
	}
}

func (r *Runner) Run(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup

	// Создание экземпляра metricagent.
	agent, err := metricagent.New(r.params, storage.New(collector.Collector()), r.logger)
	if err != nil {
		r.logger.Fatalw(err.Error(), "error", "creating agent")
	}

	// Запуск сбора и отправки метрик параллельно.
	wg.Add(1)
	go func() {
		agent.CollectMetrics(runCtx)
		wg.Done()
	}()

	// send metrics on server by timer internally
	wg.Add(1)
	go func() {
		if err = agent.SendMetrics(runCtx); err != nil {
			r.logger.Errorf("send metrics loop exited with error: %s", err.Error())
			wg.Done()
			cancel()
		}
		wg.Done()
	}()

	// catch signals
	wg.Add(1)
	go func() {
		sig := <-r.signals
		r.logger.Info(fmt.Sprintf("got signal: %s", sig.String()))
		if err = agent.SendMetrics(runCtx); err != nil {
			r.logger.Errorf("send metrics after signal %q exited with error: %s", sig.String(), err.Error())
		} else {
			r.logger.Infof("metrics successfully sent after signal %q", sig.String())
		}
		cancel()
		wg.Done()
	}()

	// wait for all goroutines to complete
	wg.Wait()
}
