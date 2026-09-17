package web

import (
	"html/template"
	"net/http"
	"solarmonitor/internal/domain"
)

type ChartHandler struct {
	tmpl        *template.Template
	getOverview func() domain.SystemOverview
}

func NewChartHandler(tmpl *template.Template, getOverview func() domain.SystemOverview) *ChartHandler {
	return &ChartHandler{tmpl: tmpl, getOverview: getOverview}
}

func (h *ChartHandler) RenderEnergyFlowSVG(w http.ResponseWriter, r *http.Request) {
	overview := h.getOverview()

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")

	// Render pure Go SVG template
	err := h.tmpl.ExecuteTemplate(w, "energy_flow.svg", map[string]interface{}{
		"ChartData": overview.ChartData(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
