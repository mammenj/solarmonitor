package domain

import (
	"fmt"
	"time"
)

type MeterRecord struct {
	Date     time.Time `json:"date"`
	Import   float64   `json:"import"`
	Export   float64   `json:"export"`
	SolarGen float64   `json:"solar_gen"`
	AddedOn  time.Time `json:"added_on"`
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

func (ps PeriodSummary) PeriodLabel() string {
	return ps.ToDate.Format("02-Jan")
}

func (ps PeriodSummary) PeriodName() string {
	// WRONG: ps.FromDate.Format("02-Jan-2016") or similar
	// CORRECT: Must use "02-Jan-2006" as the standard reference layout
	return fmt.Sprintf(
		"%s to %s",
		ps.FromDate.AddDate(0, 0, 1).Format("02-Jan-2006"),
		ps.ToDate.Format("02-Jan-2006"),
	)
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

func (so SystemOverview) TotalHomeConsumption() float64 {
	var totalGen, totalExport, totalImport float64

	for _, s := range so.Summaries {
		totalGen += s.SolarGen
		totalExport += s.ExportDiff
		totalImport += s.ImportDiff
	}

	return totalGen + totalImport - totalExport
}

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

func (so SystemOverview) ChartData() []ChartPoint {
	points := make([]ChartPoint, 0, len(so.Summaries))
	for _, s := range so.Summaries {
		points = append(points, ChartPoint{
			Label:           s.PeriodLabel(),
			SolarGen:        s.SolarGen,
			Export:          s.ExportDiff,
			Import:          s.ImportDiff,
			Consumption:     s.Consumption,
			SelfConsumption: s.SelfConsumptionRatio(),
			SelfSufficiency: s.SelfSufficiencyRatio(),
		})
	}
	return points
}

func (ps PeriodSummary) SelfConsumptionRatio() float64 {
	if ps.SolarGen <= 0 {
		return 0
	}
	selfConsumed := ps.SolarGen - ps.Export()
	if selfConsumed < 0 {
		selfConsumed = 0
	}
	return (selfConsumed / ps.SolarGen) * 100
}

// SelfSufficiencyRatio returns the % of home consumption covered by solar (0 - 100%).
func (ps PeriodSummary) SelfSufficiencyRatio() float64 {
	if ps.Consumption <= 0 {
		return 0
	}
	solarCovered := ps.Consumption - ps.Import()
	if solarCovered < 0 {
		solarCovered = 0
	}
	return (solarCovered / ps.Consumption) * 100
}

// ChartPoint represents a single data point formatted for Chart.js / Alpine datasets

type ChartPoint struct {
	Label           string  `json:"label"`
	SolarGen        float64 `json:"solar_gen"`
	Export          float64 `json:"export"`
	Import          float64 `json:"import"`
	Consumption     float64 `json:"consumption"`
	SelfConsumption float64 `json:"self_consumption_pct"`
	SelfSufficiency float64 `json:"self_sufficiency_pct"`
}
