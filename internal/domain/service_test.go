package domain_test

import (
	"context"
	"testing"
	"time"

	"solarmonitor/internal/domain"
)

type mockRepo struct {
	records []domain.MeterRecord
}

func (m *mockRepo) GetAll(ctx context.Context) ([]domain.MeterRecord, error) {
	return m.records, nil
}

func (m *mockRepo) Save(ctx context.Context, r domain.MeterRecord) error {
	m.records = append(m.records, r)
	return nil
}

func TestCalculateSummaries(t *testing.T) {
	svc := domain.NewSolarService(&mockRepo{})

	t1, _ := time.Parse("2006-01-02", "2026-03-01")
	t2, _ := time.Parse("2006-01-02", "2026-03-11") // 10 days

	recs := []domain.MeterRecord{
		{Date: t1, Import: 100, Export: 50, SolarGen: 0},
		{Date: t2, Import: 150, Export: 80, SolarGen: 100}, // impDiff=50, expDiff=30
	}

	summaries := svc.CalculateSummaries(recs)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	if s.Days != 10 {
		t.Errorf("expected 10 days, got %f", s.Days)
	}
	if s.ImportDiff != 50 {
		t.Errorf("expected 50 import diff, got %f", s.ImportDiff)
	}
	if s.ExportDiff != 30 {
		t.Errorf("expected 30 export diff, got %f", s.ExportDiff)
	}
	// Consumption = SolarGen(100) + ImpDiff(50) - ExpDiff(30) = 120
	if s.Consumption != 120 {
		t.Errorf("expected 120 consumption, got %f", s.Consumption)
	}
	// Daily Avg = 120 / 10 = 12
	if s.DailyAvgConsom != 12 {
		t.Errorf("expected 12 daily avg, got %f", s.DailyAvgConsom)
	}
}
