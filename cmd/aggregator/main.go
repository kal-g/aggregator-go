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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	consumeURL          = "/consume"
	countURL            = "/count"
	namespaceGetInfoURL = "/namespace/get_info"
)

type configEnv struct {
	Port     int    `env:"PORT_NUMBER" envDefault:"50051"`
	NodeName string `env:"NODE_NAME,required"`
	RedisURL string `env:"REDIS_URL,required"`
	ZkURL    string `env:"ZOOKEEPER_URL"`
	CfgFile  string `env:"CONFIG_FILE"`
}

var logger zerolog.Logger = zerolog.New(os.Stderr).With().Str("source", "SVC").Logger()

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	var cfg configEnv
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	svc := service.MakeNewService(cfg.RedisURL, cfg.ZkURL, cfg.NodeName, cfg.CfgFile)

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
		logger.Info().Msgf("Starting up server on port %d ...", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("The server was unable to start up or continue to run.")
		}
	}()

	// Wait for signals
	sig := <-sigs
	logger.Info().Msgf("Shutting down server due to receiving a signal of %s ...", sig)
	if err := server.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("There was a problem trying to shutdown the server.")
	}
	logger.Info().Msg("The server has been shutdown.")
}
