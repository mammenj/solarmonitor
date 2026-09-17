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

// Helper methods on PeriodSummary for easy template display without structural changes

func (ps PeriodSummary) PeriodName() string {
	return ps.FromDate.Format("02-Jan-2026") + " to " + ps.ToDate.Format("02-Jan-2006")
}

func (ps PeriodSummary) Import() float64 {
	return ps.ImportDiff
}

func (ps PeriodSummary) Export() float64 {
	return ps.ExportDiff
}

func (ps PeriodSummary) NetEnergy() float64 {
	return ps.NetBalance
}

//
//

// TotalExport calculates total kWh exported across all period summaries.
func (so SystemOverview) TotalExport() float64 {
	var total float64
	for _, s := range so.Summaries {
		total += s.ExportDiff
	}
	return total
}

// TotalImport calculates total kWh imported across all period summaries.
func (so SystemOverview) TotalImport() float64 {
	var total float64
	for _, s := range so.Summaries {
		total += s.ImportDiff
	}
	return total
}

// TotalSolarGen calculates total kWh generated across all period summaries.
func (so SystemOverview) TotalSolarGen() float64 {
	var total float64
	for _, s := range so.Summaries {
		total += s.SolarGen
	}
	return total
}

// NetConsumption calculates total Net Balance (Total Export - Total Import).
func (so SystemOverview) NetConsumption() float64 {
	return so.TotalExport() - so.TotalImport()
}
