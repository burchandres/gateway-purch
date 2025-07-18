package main

import (
	"fmt"
	"net/http"
	"log/slog"
	"os"
	"os/signal"
	"context"
)

func configureServerHandler() *http.ServeMux {
	s := http.NewServeMux()

	// root handleFunc
	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Gateway-Purch v0.1.0")
		if err != nil {
			slog.Error("error with root handle response", "error", err.Error())
		}
	})

	return s
}

func main() {
	slog.Info("starting gateway server...")
	ctx := context.Background()

	serverMux := configureServerHandler()
	server := http.Server{
		Addr: "localhost:8080",
		Handler: serverMux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	slog.Info("shutting down gateway server...")

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("error shutting down gateway server", "error", err.Error())
	}
}