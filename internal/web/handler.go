package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"solarmonitor/internal/domain"
)

type Handler struct {
	service *domain.SolarService
	tmpl    *template.Template
}
type EntryPageData struct {
	Overview domain.SystemOverview
	Error    string
}

func NewHandler(service *domain.SolarService, tmpl *template.Template) *Handler {
	return &Handler{service: service, tmpl: tmpl}
}

// Standalone package helper in internal/web/handler.go
func (h *Handler) renderEntryWithError(w http.ResponseWriter, r *http.Request, errMsg string) {
	overview, _ := h.service.GetOverview(r.Context())
	data := struct {
		Overview domain.SystemOverview
		Error    string
	}{
		Overview: overview,
		Error:    errMsg,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "entry.html", data)
}

func (h *Handler) HandleCreateReading(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	solarGenStr := r.FormValue("solar_gen")
	exportStr := r.FormValue("export")
	importStr := r.FormValue("import")

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.renderEntryWithError(w, r, "Invalid date format")
		return
	}

	solarGenVal, err := strconv.ParseFloat(solarGenStr, 64)
	if err != nil {
		h.renderEntryWithError(w, r, "Invalid solar generation value")
		return
	}

	exportVal, err := strconv.ParseFloat(exportStr, 64)
	if err != nil {
		h.renderEntryWithError(w, r, "Invalid export reading value")
		return
	}

	importVal, err := strconv.ParseFloat(importStr, 64)
	if err != nil {
		h.renderEntryWithError(w, r, "Invalid import reading value")
		return
	}

	record := domain.MeterRecord{
		Date:     parsedDate,
		SolarGen: solarGenVal, // Map parsed form value here
		Export:   exportVal,
		Import:   importVal,
	}

	if err := h.service.AddReading(r.Context(), record); err != nil {
		log.Printf("Error saving record: %v", err)
		h.renderEntryWithError(w, r, err.Error())
		return
	}

	overview, err := h.service.GetOverview(r.Context())
	if err != nil {
		http.Error(w, "Failed to update overview", http.StatusInternalServerError)
		return
	}

	data := struct {
		Overview domain.SystemOverview
		Error    string
	}{
		Overview: overview,
		Error:    "",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "entry.html", data)
}

func (h *Handler) HandleCreateReading1(w http.ResponseWriter, r *http.Request) {
	errMsg := ""

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		errMsg = "Method Not Allowed"
	}

	// 1. Parse form values
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		errMsg = "Invalid data"
	}

	dateStr := r.FormValue("date")
	exportStr := r.FormValue("export")
	importStr := r.FormValue("import")

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		errMsg = "Invalid Date format"
	}

	exportVal, err := strconv.ParseFloat(exportStr, 64)
	if err != nil {
		http.Error(w, "Invalid export reading", http.StatusBadRequest)
		errMsg = "Invaild Export Reading"
	}

	importVal, err := strconv.ParseFloat(importStr, 64)
	if err != nil {
		http.Error(w, "Invalid import reading", http.StatusBadRequest)
		errMsg = "Invalid Import Reading"
	}

	// 2. Build record and save via service
	record := domain.MeterRecord{
		Date:   parsedDate,
		Export: exportVal,
		Import: importVal,
	}

	if err := h.service.AddReading(r.Context(), record); err != nil {
		log.Printf("Error saving record: %v", err)
		http.Error(w, "Failed to save reading", http.StatusInternalServerError)
		errMsg = "Failed to save reading"
	}

	// 3. Fetch updated overview and re-render entry.html
	overview, err := h.service.GetOverview(r.Context())
	if err != nil {
		log.Printf("Error fetching overview: %v", err)
		http.Error(w, "Failed to load updated data", http.StatusInternalServerError)
		errMsg = "Failed to update data"
	}

	//		data := struct {
	//		Overview domain.SystemOverview
	//	}{
	//
	//		Overview: overview,
	//	}
	data := EntryPageData{
		Overview: overview,
		Error:    errMsg,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "entry.html", data); err != nil {
		log.Printf("Error rendering entry.html template: %v", err)
	}
}

func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.GetOverview(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Overview domain.SystemOverview
	}{
		Overview: overview,
	}

	if err := h.tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Template render error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) HandleDashboardTab(w http.ResponseWriter, r *http.Request) {
	overview, _ := h.service.GetOverview(r.Context())
	data := struct {
		Overview domain.SystemOverview
	}{
		Overview: overview,
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := h.tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
			log.Printf("Dashboard tab render error: %v", err)
		}
		return
	}

	h.tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *Handler) HandleEntryTab(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	overview, err := h.service.GetOverview(r.Context())
	if err != nil {
		log.Printf("Error fetching overview for entry tab: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Include the Error field so entry.html can evaluate .Error safely
	data := struct {
		Overview domain.SystemOverview
		Error    string
	}{
		Overview: overview,
		Error:    "",
	}

	if err := h.tmpl.ExecuteTemplate(w, "entry.html", data); err != nil {
		log.Printf("ERROR rendering entry.html: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
