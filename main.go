package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"gateway-purch/gateway"
)

// we will probably just spin up a db connection within the authService

func configureMux(
	gate *gateway.Gateway, 
	authService gateway.AuthService,
) *http.ServeMux {
	mux := http.NewServeMux()

	// root handleFunc
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("endpoint '/' hit, responding...")
		_, err := fmt.Fprintln(w, "Gateway-Purch:v0.1.0")
		if err != nil {
			slog.Error(
				"error generating response.", 
				"endpoint", r.URL.Path, 
				"error", err.Error(),
			)
		}
	})

	// add auth middleware
	mux.Handle("/api/", gateway.AuthMiddleware(gate, authService))

	return mux
}

func main() {
	slog.Info("starting gateway server...")
	ctx := context.Background()
	config := gateway.ReadConfig()
	gate := gateway.NewGateway()
	authService := gateway.NewAuthService()

	mux := configureMux(gate, *authService)
	server := http.Server{
		Addr: config.ServerAddress,
		Handler: mux,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	slog.Info("shutting down gateway server...")

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("error shutting down gateway server.", "error", err.Error())
	}

	wg.Wait()

	slog.Info("gateway server shutdown.")
}