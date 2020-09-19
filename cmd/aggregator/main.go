package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
	"github.com/kal-g/aggregator-go/internal/service"
	"github.com/rs/zerolog/log"
)

const (
	consumeURL          = "/consume"
	countURL            = "/count"
	namespaceGetInfoURL = "/namespace/get_info"
)

type configEnv struct {
	Port     string `env:"PORT_NUMBER" envDefault:"50051"`
	RedisURL string `env:"REDIS_URL,required"`
	ZkURL    string `env:"ZOOKEEPER_URL"`
}

func main() {
	var cfg configEnv
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	svc := service.MakeNewService(cfg.RedisURL)

	r := mux.NewRouter()
	r.HandleFunc(consumeURL, svc.Consume).Methods("GET", "POST")
	r.HandleFunc(countURL, svc.Count).Methods("GET", "POST")
	r.HandleFunc(namespaceGetInfoURL, svc.NamespaceGetInfo).Methods("GET", "POST")

	// Create listener for signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Configure server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	// Start server
	go func() {
		log.Info().Msgf("Starting up server on port %d ...", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("The server was unable to start up or continue to run.")
		}
	}()

	// Wait for signals
	sig := <-sigs
	log.Info().Msgf("Shutting down server due to receiving a signal of %s ...", sig)
	if err := server.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("There was a problem trying to shutdown the server.")
	}
	log.Info().Msg("The server has been shutdown.")
}
