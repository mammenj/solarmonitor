package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"solarmonitor/internal/domain"
	"solarmonitor/internal/storage"
	"solarmonitor/internal/web"
)

func main() {
	dbname := "meter_logs.db"
	//dbname := "test.db"
	repo, err := storage.NewSQLiteStore(dbname + "?_timelayout=2006-01-02%2015:04")
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	// Defer close as a safety net in case the application panics or exits early
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error cleaning up repo on panic: %v", err)
		}
	}()
	service := domain.NewSolarService(repo)

	tmpl, err := web.InitTemplates()
	if err != nil {
		log.Fatalf("Template init error: %v", err)
	}

	handler := web.NewHandler(service, tmpl)
	mux := http.NewServeMux()

	// 1. Embedded Static Assets
	staticSub, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		log.Fatalf("Static FS error: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// 2. Application Routes (Plain Path Patterns)
	mux.HandleFunc("/", handler.HandleRoot)
	mux.HandleFunc("/app/dashboard", handler.HandleDashboardTab)
	mux.HandleFunc("/app/entry", handler.HandleEntryTab)
	// Form submission endpoint for HTMX
	mux.HandleFunc("/api/readings", handler.HandleCreateReading)

	// 2. Explicitly configure http.Server instead of using shorthand http.ListenAndServe
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	////
	// 3. Start the HTTP server in a separate background goroutine
	go func() {
		fmt.Println("⚡ Solar Monitor running at http://0.0.0.0:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// 4. Setup channels to catch termination signals (Ctrl+C, kill command)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// This blocks code execution until a signal is received
	<-quit
	log.Println("Shutting down Solar Monitor server gracefully...")

	// 5. Establish a 10-second bounded window to drain active web requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 6. Step One: Stop accepting new requests and wait for ongoing HTMX/Dashboard loads to finish
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped accepting connections.")
	}

	// 7. Step Two: Safely close the SQLite Database connections now that handlers are idle
	log.Println("Closing SQLite database connection pool...")
	if err := repo.Close(); err != nil {
		log.Printf("Error closing SQLite storage safely: %v", err)
	} else {
		log.Println("SQLite database closed cleanly.")
	}

	log.Println("Solar Monitor stopped successfully.")

	//
	//fmt.Println("⚡ Solar Monitor running at http://0.0.0.0:8080")
	//if err := http.ListenAndServe(":8080", mux); err != nil {
	//	log.Fatal(err)
	//}
}
