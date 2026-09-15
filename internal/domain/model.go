package domain

import "time"

type MeterRecord struct {
	Date     time.Time `json:"date"`
	Import   float64   `json:"import"`
	Export   float64   `json:"export"`
	SolarGen float64   `json:"solar_gen"`
}

type PeriodSummary struct {
	FromDate       time.Time `json:"from_date"`
	ToDate         time.Time `json:"to_date"`
	Days           float64   `json:"days"`
	ImportDiff     float64   `json:"import_diff"`
	ExportDiff     float64   `json:"export_diff"`
	NetBalance     float64   `json:"net_balance"` // Export - Import
	SolarGen       float64   `json:"solar_gen"`
	Consumption    float64   `json:"consumption"`
	DailyAvgConsom float64   `json:"daily_avg_consumption"`
	HasSolar       bool      `json:"has_solar"`
}

type SystemOverview struct {
	Records      []MeterRecord   `json:"records"`
	Summaries    []PeriodSummary `json:"summaries"`
	LastBaseline *MeterRecord    `json:"last_baseline"`
}
