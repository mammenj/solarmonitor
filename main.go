package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"solarmonitor/internal/domain"
	"solarmonitor/internal/storage"
	"solarmonitor/internal/web"
)

func main() {
	// repo1 := storage.NewFileStore("solar_readings.txt")
	repo, err := storage.NewSQLiteStore("meter_logs.db")
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
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

	fmt.Println("⚡ Solar Monitor running at http://0.0.0.0:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
