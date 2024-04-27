package runner

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	runningTimeoutSet  bool
	shutdownCalled     bool
	shutdownTimeoutSet bool
	shutdownTimeout    time.Duration
)

var (
	cancelRun      context.CancelFunc
	cancelShutdown context.CancelFunc
)

// Service defines the interface for services that can be gracefully started and stopped.
type Service interface {
	// Run starts the service and blocks until an error occurs or Shutdown is called.
	Run(ctx context.Context) error

	// Shutdown stops the service gracefully. It should wait for any ongoing operations
	// to finish before returning.
	Shutdown(ctx context.Context) error
}

type Option func(ctx context.Context) context.Context

func WithValue(key string, value interface{}) Option {
	return func(ctx context.Context) context.Context { return context.WithValue(ctx, key, value) }
}

func WithValues(values map[string]interface{}) Option {
	return func(ctx context.Context) context.Context {
		for k, v := range values {
			ctx = context.WithValue(ctx, k, v)
		}

		return ctx
	}
}

func GetValueFromContext[T any](ctx context.Context, key string) (T, bool) {
	value, ok := ctx.Value(key).(T)

	return value, ok
}

func WithRunTimeout(timeout time.Duration) Option {
	runningTimeoutSet = true

	return func(ctx context.Context) context.Context {
		ctx, cancelRun = context.WithTimeout(ctx, timeout)

		return ctx
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(ctx context.Context) context.Context {
		shutdownTimeoutSet = true
		shutdownTimeout = timeout

		return ctx
	}
}

func Run(svc Service, opts ...Option) {
	runCtx := context.Background()
	runCtx, cancelRun = context.WithCancel(runCtx)

	for _, opt := range opts {
		runCtx = opt(runCtx)
	}

	shutdownCtx := context.Background()
	if shutdownTimeoutSet {
		shutdownCtx, cancelShutdown = context.WithTimeout(shutdownCtx, shutdownTimeout)
	}

	shutdownSync := sync.Once{}
	shutdownComplete := make(chan struct{})

	shutdown := func() {
		shutdownSync.Do(func() {
			shutdownCalled = true
			shutdownErr := svc.Shutdown(shutdownCtx)
			if shutdownErr != nil {
				defer os.Exit(1)
				log.Printf("service shutdown error: %s\n", shutdownErr.Error())
			}

			cancelRun()
			shutdownComplete <- struct{}{}
		})
	}

	if runningTimeoutSet {
		go func() {
			<-runCtx.Done()
			log.Printf("timeout exceeded")

			shutdown()
			log.Println("service shutdown")

			os.Exit(0)
		}()
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		sig := <-sigCh
		log.Printf("received signal: %s\n", sig.String())

		shutdown()
		log.Println("service shutdown")

		os.Exit(0)
	}()

	err := svc.Run(runCtx)
	if err != nil && !shutdownCalled {
		log.Printf("service run error: %s\n", err.Error())
		shutdown()
		log.Println("service shutdown")

		os.Exit(1)
	}

	<-shutdownComplete
}
