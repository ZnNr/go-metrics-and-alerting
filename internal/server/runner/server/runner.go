package server

import (
	"context"
	"database/sql"
	"fmt"
	collector "github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	serverGRPC "github.com/ZnNr/go-musthave-metrics.git/internal/server/grpc"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/server/middlewares/logger"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/router"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/saver/database"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/saver/file"
	pb "github.com/ZnNr/go-musthave-metrics.git/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Константы, связанные с адресом pprof и информацией о сборке приложения.
const (
	pprofAddr    string = ":8090"
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// Runner Структура, представляющая собой главный компонент приложения сервера.
type Runner struct {
	saver           saver
	metricsInterval time.Duration
	isRestore       bool
	storeInterval   int
	tlsKey          string
	appSrv          httpServer
	pprofSrv        httpServer
	grpcServer      grpcServer
	listener        listener
	logger          *zap.SugaredLogger
	signals         chan os.Signal
}

// New создает экземпляр Runner с использованием параметров flags.Params.
func New(params *flags.Params) *Runner {
	// Инициализация логгера.
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit") //сообщаем об ошибки пакетом fmt потому что логгер не создался
		return nil
	}
	defer logger.Sync()
	log.SugarLogger = *logger.Sugar()

	// Инициализация saver (файл или база данных).
	saver, err := initSaver(params)
	if err != nil {
		log.SugarLogger.Fatalw(err.Error(), "error", "init metrics saver")
	}
	// Инициализация роутера.
	r, err := router.New(*params)
	if err != nil {
		log.SugarLogger.Fatalw(err.Error(), "error", "creating router")
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	runner := &Runner{
		saver:           saver,
		metricsInterval: time.Duration(params.StoreInterval),
		isRestore:       params.Restore,
		storeInterval:   params.StoreInterval,
		tlsKey:          params.CryptoKeyPath,
		appSrv: &http.Server{
			Addr:    params.FlagRunAddr,
			Handler: r,
		},
		pprofSrv: &http.Server{
			Addr:    pprofAddr,
			Handler: nil,
		},
		signals: sigs,
		logger:  &log.SugarLogger,
	}
	if !params.DisableGrpc {
		// Создание gRPC сервера.
		s := grpc.NewServer()
		// Регистрация gRPC сервера.
		pb.RegisterMetricsServer(s, &serverGRPC.MetricsServer{})

		listen, err := net.Listen("tcp", params.GrpcRunAddr)
		if err != nil {
			log.SugarLogger.Fatalw(err.Error(), "event", "start listen tcp")
		}

		runner.listener = listen
		runner.grpcServer = s
	}

	return runner
}

// Run запускает основной цикл приложения, обрабатывая сохранение метрик, pprof и сигналы.
func (r *Runner) Run(ctx context.Context) {
	// Логгирование информации о сборке.
	r.logger.Info("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	// Восстановление предыдущих метрик, если необходимо.
	if r.isRestore {
		metrics, err := r.saver.Restore(ctx)
		if err != nil {
			r.logger.Error(err.Error(), "restore error")
		}
		collector.Collector().Metrics = metrics
		r.logger.Info("metrics restored")
	}

	// Регулярное сохранение метрик.
	go r.saveMetrics(ctx, r.storeInterval)

	// Запуск pprof.
	go func() {
		if err := r.pprofSrv.ListenAndServe(); err != nil {
			r.logger.Fatalw(err.Error(), "pprof", "start pprof httpServer")
		}
	}()

	// Обработка сигналов.
	go func() {
		sig := <-r.signals
		r.logger.Info(fmt.Sprintf("got signal: %s", sig.String()))
		// save metrics
		if err := r.saver.Save(ctx, collector.Collector().Metrics); err != nil {
			r.logger.Error(err.Error(), "save error")
		} else {
			r.logger.Info("metrics was successfully saved")
		}
		// gracefully shutdown
		if err := r.appSrv.Shutdown(ctx); err != nil {
			r.logger.Error(fmt.Sprintf("error while httpServer shutdown: %s", err.Error()), "httpServer shutdown error")
			return
		}
	}()

	// Запуск gRPC сервера.
	if r.grpcServer != nil {
		go func() {
			r.logger.Info("Starting gRPC httpServer")

			if err := r.grpcServer.Serve(r.listener); err != nil {
				r.logger.Errorf("error while serving grpc server: %s", err.Error())
			}
			defer r.grpcServer.GracefulStop()
			defer r.listener.Close()
		}()
	}

	// Запуск http httpServer.
	r.logger.Info("Starting http httpServer")
	if err := r.appSrv.ListenAndServe(); err != nil {
		r.logger.Fatalw(err.Error(), "event", "start http httpServer")
	}
}

// saveMetrics сохраняет метрики с указанным интервалом.
func (r *Runner) saveMetrics(ctx context.Context, interval int) {
	ticker := time.NewTicker(time.Duration(interval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.saver.Save(ctx, collector.Collector().Metrics); err != nil {
				r.logger.Error(err.Error(), "save error")
			}
		}
	}
}

// initSaver инициализирует saver (файл или базу данных) в зависимости от параметров.
func initSaver(params *flags.Params) (saver, error) {
	if params.DatabaseAddress != "" {
		return initDatabaseSaver(params.DatabaseAddress)
	} else if params.FileStoragePath != "" {
		return initFileSaver(params.FileStoragePath), nil
	}
	return nil, fmt.Errorf("neither file path nor database address was specified")
}

// initDatabaseSaver инициализирует saver для базы данных с указанным адресом.
func initDatabaseSaver(databaseAddress string) (saver, error) {
	db, err := sql.Open("pgx", databaseAddress)
	if err != nil {
		return nil, err
	}
	return database.New(db)
}

// initFileSaver инициализирует saver для работы с файлом по указанному пути.
func initFileSaver(fileStoragePath string) saver {
	return file.New(fileStoragePath)
}

//go:generate mockery --inpackage --disable-version-string --filename saver_mock.go --name saver
type saver interface {
	Restore(ctx context.Context) ([]collector.StoredMetric, error)
	Save(ctx context.Context, metrics []collector.StoredMetric) error
}

//go:generate mockery --inpackage --disable-version-string --filename http_server_mock.go --name httpServer
type httpServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

//go:generate mockery --inpackage --disable-version-string --filename grpc_server_mock.go --name grpcServer
type grpcServer interface {
	Serve(lis net.Listener) error
	GracefulStop()
}

//go:generate mockery --inpackage --disable-version-string --filename listener_mock.go --name listener
type listener interface {
	Close() error
	Accept() (net.Conn, error)
	Addr() net.Addr
}
