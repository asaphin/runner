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

var timeoutSet bool
var cancel context.CancelFunc
var shutdownCalled bool

// Service defines the interface for services that can be gracefully started and stopped.
type Service interface {
	// Run starts the service and blocks until an error occurs or Shutdown is called.
	Run(ctx context.Context) error

	// Shutdown stops the service gracefully. It should wait for any ongoing operations
	// to finish before returning.
	Shutdown() error
}

type Option func(ctx context.Context) context.Context

func WithValue(key string, value interface{}) Option {
	return func(ctx context.Context) context.Context { return context.WithValue(ctx, key, value) }
}

func WithValues(key string, values map[string]interface{}) Option {
	return func(ctx context.Context) context.Context {
		for k, v := range values {
			ctx = context.WithValue(ctx, k, v)
		}

		return ctx
	}
}

func WithTimeout(timeout time.Duration) Option {
	timeoutSet = true

	return func(ctx context.Context) context.Context {
		ctx, cancel = context.WithTimeout(ctx, timeout)

		return ctx
	}
}

func Run(svc Service, opts ...Option) {
	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)

	for _, opt := range opts {
		ctx = opt(ctx)
	}

	shutdownSync := sync.Once{}
	shutdownComplete := make(chan struct{})

	shutdown := func() {
		shutdownSync.Do(func() {
			shutdownCalled = true
			shutdownErr := svc.Shutdown()
			if shutdownErr != nil {
				defer os.Exit(1)
				log.Printf("service shutdown error: %s\n", shutdownErr.Error())
			}

			cancel()
			shutdownComplete <- struct{}{}
		})
	}

	if timeoutSet {
		go func() {
			<-ctx.Done()
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

	err := svc.Run(ctx)
	if err != nil && !shutdownCalled {
		log.Printf("service run error: %s\n", err.Error())
		shutdown()
		log.Println("service shutdown")

		os.Exit(1)
	}

	<-shutdownComplete
}
