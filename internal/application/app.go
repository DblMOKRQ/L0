package application

import (
	"L0/internal/router"
	"L0/internal/service"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	orderService *service.OrderService
	router       *router.Router
	httpServer   *http.Server
	log          *zap.Logger
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	shutdownCh   chan struct{}
}

func NewApp(service *service.OrderService, router *router.Router, addr string, log *zap.Logger) *App {
	return &App{
		orderService: service,
		router:       router,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: router.GetHTTPHandler(),
		},
		log:        log.Named("application"),
		shutdownCh: make(chan struct{}),
	}
}

func (a *App) Run(limit int) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.orderService.PreloadRecentOrder(ctx, limit); err != nil {
			a.log.Error("Error preloading order", zap.Error(err))
		}
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.startKafka(ctx)
	}()

	serverErr := make(chan error, 1)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.log.Info("Starting HTTP server", zap.String("address", a.httpServer.Addr))
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	a.log.Info("Application started successfully")

	// Ожидаем либо сигнал завершения, либо ошибку сервера
	var runErr error
	select {
	case sig := <-sigCh:
		a.log.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	case err := <-serverErr:
		a.log.Error("HTTP server error", zap.Error(err))
		runErr = err
		cancel()
	}

	if shutdownErr := a.shutdown(ctx); shutdownErr != nil {
		a.log.Error("Shutdown error", zap.Error(shutdownErr))
		if runErr == nil {
			runErr = shutdownErr
		}
	}

	return runErr
}

func (a *App) startKafka(ctx context.Context) {
	defer a.log.Info("Kafka consumer stopped")

	for {
		select {
		case <-ctx.Done():
			a.log.Info("Main context cancelled, stopping Kafka consumer")
			return
		case <-a.shutdownCh:
			a.log.Info("Shutdown signal received, stopping Kafka consumer")
			return
		default:
			order, err := a.orderService.ReadMessage(ctx)

			if err != nil {

				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				if errors.Is(err, context.Canceled) {
					return
				}

				a.log.Error("Error reading message", zap.Error(err))
				time.Sleep(2 * time.Second)
				continue
			}

			if err := a.orderService.SaveOrder(ctx, order); err != nil {
				a.log.Error("Error saving order to DB", zap.Error(err))
				continue
			}

			if err := a.orderService.SetOrder(ctx, order); err != nil {
				a.log.Error("Error caching order", zap.Error(err))
				continue
			}

			a.log.Info("Successfully processed order",
				zap.String("order_uid", order.OrderUID))
		}
	}
}

func (a *App) shutdown(ctx context.Context) error {
	var shutdownErr error

	a.shutdownOnce.Do(func() {
		a.log.Info("Initiating graceful shutdown...")

		close(a.shutdownCh)

		shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Error("HTTP server shutdown error", zap.Error(err))
			shutdownErr = fmt.Errorf("HTTP shutdown error: %w", err)
		} else {
			a.log.Info("HTTP server stopped gracefully")
		}

		if err := a.orderService.CloseConsumer(); err != nil {
			a.log.Error("Failed to close Kafka connection", zap.Error(err))
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%v, Kafka close error: %w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("Kafka close error: %w", err)
			}
		} else {
			a.log.Info("Kafka connection closed")
		}

		a.orderService.CloseRepo()

		waitDone := make(chan struct{})
		go func() {
			a.wg.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			a.log.Info("All goroutines finished")
		case <-shutdownCtx.Done():
			a.log.Warn("Shutdown timed out, some goroutines may still be running")
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%v, shutdown timeout", shutdownErr)
			} else {
				shutdownErr = fmt.Errorf("shutdown timeout")
			}
		}

		a.log.Info("Shutdown completed")
	})

	return shutdownErr
}
