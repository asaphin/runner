package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/asaphin/runner"
	"log"
	"net/http"
)

func main() {
	svc := NewService()

	runner.Run(svc, runner.WithValue("port", 8080))
}

type Service struct {
	api *API
}

func NewService() *Service {
	return &Service{
		api: NewAPI(),
	}
}

func (s *Service) Run(ctx context.Context) error {
	port, ok := ctx.Value("port").(int)
	if !ok {
		return errors.New("no port value found in context")
	}

	return s.api.Run(fmt.Sprintf(":%d", port))
}

func (s *Service) Shutdown() error {
	log.Println("server shutdown requested")
	err := s.api.Shutdown()
	log.Println("server shutdown")

	return err
}

type API struct {
	server *http.Server
}

func NewAPI() *API {
	api := new(API)

	mux := http.NewServeMux()

	mux.HandleFunc("/", api.handler)

	api.server = &http.Server{
		Handler: mux,
	}

	return api
}

func (a *API) Run(addr string) error {
	a.server.Addr = addr

	return a.server.ListenAndServe()
}

func (a *API) Shutdown() error {
	log.Println("API shutdown requested")
	err := a.server.Shutdown(context.Background())
	log.Println("API shutdown")

	return err
}

func (a *API) handler(w http.ResponseWriter, _ *http.Request) {
	n, err := fmt.Fprint(w, "Hello, this is a simple web server!")
	if err != nil {
		log.Printf("response write error: %s\n", err.Error())
	}

	log.Printf("%d bytes written\n", n)
}
