package domain

import (
	"testing"
	"time"
)

func TestSystemOverview_Totals(t *testing.T) {
	overview := SystemOverview{
		Summaries: []PeriodSummary{
			{
				FromDate:    time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
				ToDate:      time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC),
				SolarGen:    25.0,
				ExportDiff:  8.0,
				ImportDiff:  6.0,
				Consumption: 23.0,
			},
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
	summary := PeriodSummary{
		FromDate: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC),
	}

	expected := "01-Mar-2026 to 15-Mar-2026"
	if got := summary.PeriodName(); got != expected {
		t.Errorf("PeriodName() = %q, want %q", got, expected)
	}
}

func TestSystemOverview_ChartData(t *testing.T) {
	from := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)

	overview := SystemOverview{
		Summaries: []PeriodSummary{
			{
				FromDate:    from,
				ToDate:      to,
				SolarGen:    150.0,
				ExportDiff:  80.0,
				ImportDiff:  30.0,
				Consumption: 100.0,
			},
		},
	}

	chartPoints := overview.ChartData()

	if len(chartPoints) != 1 {
		t.Fatalf("ChartData() returned %d points, want 1", len(chartPoints))
	}

	pt := chartPoints[0]
	expectedLabel := "01-Mar-2026 to 15-Mar-2026"
	if pt.Label != expectedLabel {
		t.Errorf("Label = %q, want %q", pt.Label, expectedLabel)
	}
	if pt.SolarGen != 150.0 {
		t.Errorf("SolarGen = %v, want 150.0", pt.SolarGen)
	}
	if pt.Export != 80.0 {
		t.Errorf("Export = %v, want 80.0", pt.Export)
	}
	if pt.Import != 30.0 {
		t.Errorf("Import = %v, want 30.0", pt.Import)
	}
	if pt.Consumption != 100.0 {
		t.Errorf("Consumption = %v, want 100.0", pt.Consumption)
	}
}
