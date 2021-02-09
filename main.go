package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"syscall"
	"fmt"
	"github.com/sentenceshen/codingchallenge/handler"

	"github.com/gorilla/mux"
)

func main() {

	listenAddr := ":443"

	var logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)
	logger.Printf("Server is starting...")

	router := mux.NewRouter()

	router.HandleFunc("/jobs/start", handler.LogApi(handler.Start, logger))
	router.HandleFunc("/jobs/query", handler.LogApi(handler.Query, logger))
	router.HandleFunc("/jobs/stop", handler.LogApi(handler.Stop, logger))

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
	}

	done := make(chan bool)

	go func() {
		logger.Println("Server is ready to handle requests at", listenAddr)
		if err := server.ListenAndServeTLS("ssl/ca.pem", "ssl/ca.key"); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
		}
		close(done)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

	for s := range quit{
		// wait for all resources are released
		fmt.Printf("catch signal %v, now exit...\n", s)
		break
	}

	logger.Printf("Server is shutting down...")

	// shut down the server first
	ctxServer, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctxServer); err != nil {
		logger.Printf("Failed shutdown the server: %v\n", err)
	}
	<-done
	logger.Printf("Server stopped")
}
