package main

import (
	"context"
	"github.com/asaphin/runner"
	"log"
	"time"
)

func main() {
	svc := NewEchoService("I'm running", time.Second)

	runner.Run(svc, runner.WithTimeout(5*time.Second))
}

type EchoService struct {
	message      string
	period       time.Duration
	shutdownChan chan struct{}
}

func (s *EchoService) Run(_ context.Context) error { // Use context if necessary
	//defer func() {
	//	close(s.shutdownChan)
	//	s.shutdownChan = nil
	//}()

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

func (s *EchoService) Shutdown() error {
	log.Println("service Shutdown() method called")

	// Put shutdown actions here
	time.Sleep(500 * time.Millisecond)
	log.Println("service shutdown actions completed")

	//if s.shutdownChan != nil {
	//	log.Println("shutdown channel is open")
	s.shutdownChan <- struct{}{}
	log.Println("service shutdown signal sent")
	//}

	return nil
}

func NewEchoService(message string, period time.Duration) *EchoService {
	return &EchoService{
		message:      message,
		period:       period,
		shutdownChan: make(chan struct{}),
	}
}
