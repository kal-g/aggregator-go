package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	writer "github.com/kal-g/aggregator-go/writer"
	"github.com/rs/zerolog/log"
)

const (
	port       = 50051
	consumeURL = "/consume"
)

func main() {
	svc := writer.MakeNewService(os.Args[1])

	r := mux.NewRouter()
	r.HandleFunc(consumeURL, svc.Consume).Methods("GET", "POST")

	// Create listener for signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Configure server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// Start server
	go func() {
		log.Info().Msgf("Starting up server on port %d ...", port)
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
