package domain

import (
	"testing"
	"time"
)

func TestSystemOverview_Totals(t *testing.T) {
	overview := SystemOverview{
		Records: []MeterRecord{
			{Date: time.Now(), SolarGen: 15.0, Export: 5.0, Import: 2.0},
			{Date: time.Now(), SolarGen: 10.0, Export: 3.0, Import: 4.0},
		},
	}

	if got := overview.TotalSolarGen(); got != 25.0 {
		t.Errorf("TotalSolarGen() = %v, want 25.0", got)
	}
	if got := overview.TotalExport(); got != 8.0 {
		t.Errorf("TotalExport() = %v, want 8.0", got)
	}
	if got := overview.TotalImport(); got != 6.0 {
		t.Errorf("TotalImport() = %v, want 6.0", got)
	}
	if got := overview.NetConsumption(); got != 2.0 {
		t.Errorf("NetConsumption() = %v, want 2.0", got)
	}
}

func TestPeriodSummary_TemplateGetters(t *testing.T) {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	ps := PeriodSummary{
		FromDate:   from,
		ToDate:     to,
		ImportDiff: 12.5,
		ExportDiff: 30.0,
		NetBalance: 17.5,
	}

	if got := ps.PeriodName(); got != "Mar 01 - Mar 15, 2026" {
		t.Errorf("PeriodName() = %v, want 'Mar 01 - Mar 15, 2026'", got)
	}
	if got := ps.Import(); got != 12.5 {
		t.Errorf("Import() = %v, want 12.5", got)
	}
	if got := ps.Export(); got != 30.0 {
		t.Errorf("Export() = %v, want 30.0", got)
	}
	if got := ps.NetEnergy(); got != 17.5 {
		t.Errorf("NetEnergy() = %v, want 17.5", got)
	}
}

func TestSystemOverview_TotalsFromSummaries(t *testing.T) {
	overview := SystemOverview{
		Summaries: []PeriodSummary{
			{
				FromDate:    time.Now().AddDate(0, 0, -14),
				ToDate:      time.Now().AddDate(0, 0, -7),
				ImportDiff:  200.0,
				ExportDiff:  800.0,
				SolarGen:    1000.0,
				Consumption: 400.0,
			},
			{
				FromDate:    time.Now().AddDate(0, 0, -7),
				ToDate:      time.Now(),
				ImportDiff:  227.0,
				ExportDiff:  564.0,
				SolarGen:    800.0,
				Consumption: 463.0,
			},
		},
	}

	if got := overview.TotalImport(); got != 427.0 {
		t.Errorf("TotalImport() = %v, want 427.0", got)
	}
	if got := overview.TotalExport(); got != 1364.0 {
		t.Errorf("TotalExport() = %v, want 1364.0", got)
	}
	if got := overview.NetConsumption(); got != 937.0 { // 1364 - 427
		t.Errorf("NetConsumption() = %v, want 937.0", got)
	}
}
