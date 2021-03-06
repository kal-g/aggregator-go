package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
	"github.com/kal-g/aggregator-go/internal/service"
	"github.com/rs/zerolog"
)

const (
	consumeURL            = "/consume"
	countURL              = "/count"
	debugSetLogLevelURL   = "/debug/set_log_level"
	namespaceGetInfoURL   = "/namespace/get_info"
	namespaceSetConfigURL = "/namespace/config/set"
	namespaceGetConfigURL = "/namespace/config/get"
	namespaceDeleteURL    = "/namespace/delete"
)

type configEnv struct {
	Port     int    `env:"PORT_NUMBER" envDefault:"50051"`
	NodeName string `env:"NODE_NAME,required"`
	RedisURL string `env:"REDIS_URL,required"`
	ZkURL    string `env:"ZOOKEEPER_URL,required"`
}

var logger zerolog.Logger = zerolog.New(os.Stderr).With().
	Str("source", "MAIN").
	Timestamp().
	Logger()

type configFlags []string

func (i *configFlags) String() string {
	return strings.Join(*i, "-")
}

func (i *configFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var configFiles configFlags

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	var cfg configEnv
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	flag.Var(&configFiles, "config", "Config files")
	flag.Parse()

	logger.Info().Msgf("Loading config: %v\n", configFiles)

	svc := service.MakeNewService(cfg.RedisURL, cfg.ZkURL, cfg.NodeName, configFiles)

	r := getRouter(&svc)

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
			logger.Fatal().Err(err).Msg("The server was unable to start up or continue to run.")
		}
	}()

	// Wait for signals
	sig := <-sigs
	logger.Info().Msgf("Shutting down server due to receiving a signal of %s ...", sig)
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Err(err).Msg("There was a problem trying to shutdown the server.")
	}
	logger.Info().Msg("The server has been shutdown.")
}

func getRouter(svc *service.Service) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(consumeURL, svc.Consume).Methods("GET", "POST")
	r.HandleFunc(countURL, svc.Count).Methods("GET", "POST")
	r.HandleFunc(debugSetLogLevelURL, svc.DebugSetLogLevel).Methods("GET", "POST")
	r.HandleFunc(namespaceGetInfoURL, svc.NamespaceGetInfo).Methods("GET", "POST")
	r.HandleFunc(namespaceSetConfigURL, svc.NamespaceSetConfig).Methods("GET", "POST")
	r.HandleFunc(namespaceGetConfigURL, svc.NamespaceGetConfig).Methods("GET", "POST")
	r.HandleFunc(namespaceDeleteURL, svc.NamespaceDelete).Methods("GET", "POST")
	return r
}
