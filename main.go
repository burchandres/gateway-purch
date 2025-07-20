package main

import (
	"fmt"
	"net/http"
	"log/slog"
	"os"
	"os/signal"
	"context"
)


func configureMux() *http.ServeMux {
	mux := http.NewServeMux()

	// root handleFunc
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("endpoint '/' hit, responding...")
		_, err := fmt.Fprintln(w, "Gateway-Purch:v0.1.0")
		if err != nil {
			slog.Error("error with root handle response.", "error", err.Error())
		}
	})

	return mux
}

func main() {
	slog.Info("starting gateway server...")
	ctx := context.Background()
	config := ReadConfig()

	mux := configureMux()
	server := http.Server{
		Addr: config.ServerAddress,
		Handler: mux,
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
		slog.Error("error shutting down gateway server.", "error", err.Error())
	}

	slog.Info("gateway server shutdown.")
}