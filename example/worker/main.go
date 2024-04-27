package main

import (
	"context"
	"github.com/asaphin/runner"
	"log"
	"time"
)

func main() {
	svc := NewEchoService("I'm running", time.Second)

	runner.Run(svc, runner.WithRunTimeout(5*time.Second), runner.WithShutdownTimeout(500*time.Millisecond))
}

type EchoService struct {
	message      string
	period       time.Duration
	shutdownChan chan struct{}
}

func (s *EchoService) Run(_ context.Context) error { // Use context if necessary
	defer log.Println("service stopped")

	for {
		// Put service operations here
		log.Println(s.message)
		time.Sleep(s.period)

		select {
		case <-s.shutdownChan:
			log.Println("service shutdown signal received")
			return nil
		default:
		}
	}
}

func (s *EchoService) Shutdown(ctx context.Context) error {
	log.Println("service Shutdown() method called")

	var errChan = make(chan error)

	go func() {
		<-ctx.Done()
		log.Println("shutdown context done")
		s.shutdownChan <- struct{}{}
		log.Println("service shutdown signal sent")
		errChan <- ctx.Err()
	}()

	go func() {
		// Put shutdown actions here
		time.Sleep(1000 * time.Millisecond)
		log.Println("service shutdown actions completed")

		//if s.shutdownChan != nil {
		//	log.Println("shutdown channel is open")
		s.shutdownChan <- struct{}{}
		log.Println("service shutdown signal sent")
		errChan <- nil
	}()

	return <-errChan
}

func NewEchoService(message string, period time.Duration) *EchoService {
	return &EchoService{
		message:      message,
		period:       period,
		shutdownChan: make(chan struct{}),
	}
}
