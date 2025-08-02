package main

import (
	"fmt"
	"net/http"
	"log/slog"
	"os"
	"os/signal"
	"context"
	"sync"

	"gateway-purch/gateway"
)


func configureMux() *http.ServeMux {
	mux := http.NewServeMux()

	// root handleFunc
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("endpoint '/' hit, responding...")
		_, err := fmt.Fprintln(w, "Gateway-Purch:v0.1.0")
		// target, _ := url.Parse("http://localhost:8080/budgets/sync-transactions")
		// _, err := fmt.Fprintf(w, "url.Parse('http://localhost:8080/budgets/sync-transactions'): %s", target)
		if err != nil {
			slog.Error("error generating response.", "endpoint", r.URL.Path, "error", err.Error())
		}
	})

	// test handleFunc
	mux.HandleFunc("/foo/bar/test", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("endpoint '/foo/bar/test hit, responding...")
		_, err := fmt.Fprintf(w, "r.URL.Path is: %s\n", r.URL.Path)
		if err != nil {
			slog.Error("error generating response.", "endpoint", r.URL.Path, "error", err.Error())
		}
	})

	return mux
}

func main() {
	slog.Info("starting gateway server...")
	ctx := context.Background()
	config := gateway.ReadConfig()

	mux := configureMux()
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