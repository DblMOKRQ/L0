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
	shutdownCh   chan struct{} // Добавляем канал для остановки Kafka
}

// TODO: сделать redis в http сервер и проверить его
// TODO: сделать заполнение кэша при запуске
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

func (a *App) Run(ttl time.Duration) error {
	// Канал для сигналов ОС
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем Kafka consumer
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.startKafka(ctx, ttl)
	}()

	// Запускаем HTTP сервер в отдельной горутине
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

	// Начинаем graceful shutdown
	if shutdownErr := a.shutdown(ctx); shutdownErr != nil {
		a.log.Error("Shutdown error", zap.Error(shutdownErr))
		if runErr == nil {
			runErr = shutdownErr
		}
	}

	return runErr
}

func (a *App) startKafka(ctx context.Context, ttl time.Duration) {
	defer a.log.Info("Kafka consumer stopped")

	// Создаем отдельный контекст для Kafka, который не зависит от основного

	for {
		select {
		case <-ctx.Done(): // Основной контекст приложения
			a.log.Info("Main context cancelled, stopping Kafka consumer")
			return
		case <-a.shutdownCh: // Сигнал shutdown
			a.log.Info("Shutdown signal received, stopping Kafka consumer")
			return
		default:
			order, err := a.orderService.ReadMessage(ctx)

			if err != nil {

				if errors.Is(err, context.DeadlineExceeded) {
					a.log.Info("1231231231231")
					continue
				}
				if errors.Is(err, context.Canceled) {
					return
				}

				a.log.Error("Error reading message", zap.Error(err))
				time.Sleep(2 * time.Second)
				continue
			}

			// Сохраняем заказ используя основной контекст
			if err := a.orderService.SaveOrder(ctx, order); err != nil {
				a.log.Error("Error saving order to DB", zap.Error(err))
				continue
			}

			if err := a.orderService.SetOrder(ctx, order, ttl); err != nil {
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

	// Гарантируем однократный вызов shutdown
	a.shutdownOnce.Do(func() {
		a.log.Info("Initiating graceful shutdown...")

		// 1. Закрываем канал shutdown чтобы остановить Kafka consumer
		close(a.shutdownCh)

		// 2. Останавливаем HTTP сервер
		shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Error("HTTP server shutdown error", zap.Error(err))
			shutdownErr = fmt.Errorf("HTTP shutdown error: %w", err)
		} else {
			a.log.Info("HTTP server stopped gracefully")
		}

		// 3. Закрываем Kafka соединение
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

		// 4. Закрываем репозиторий

		a.orderService.CloseRepo()

		// 5. Ждем завершения всех горутин с таймаутом
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
