package domain

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

var (
	ErrOutofOrderDate = errors.New("new reading date must be after the last recorded date")
	ErrInvalidReading = errors.New("current reading cannot be less than previous meter reading")
)

type SolarService struct {
	repo Repository
}

func NewSolarService(repo Repository) *SolarService {
	return &SolarService{repo: repo}
}

func (s *SolarService) GetOverview(ctx context.Context) (SystemOverview, error) {
	records, err := s.repo.GetAll(ctx)
	if err != nil {
		return SystemOverview{}, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.Before(records[j].Date)
	})

	summaries := s.CalculateSummaries(records)

	var lastBaseline *MeterRecord
	if len(records) > 0 {
		lastBaseline = &records[len(records)-1]
	}

	return SystemOverview{
		Records:      records,
		Summaries:    summaries,
		LastBaseline: lastBaseline,
	}, nil
}

func (s *SolarService) AddReading(ctx context.Context, rec MeterRecord) error {
	existing, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		sort.Slice(existing, func(i, j int) bool {
			return existing[i].Date.Before(existing[j].Date)
		})
		last := existing[len(existing)-1]

		if !rec.Date.After(last.Date) {
			return fmt.Errorf("%w (%s <= %s)", ErrOutofOrderDate, rec.Date.Format("2006-01-02"), last.Date.Format("2006-01-02"))
		}
		if rec.Import < last.Import || rec.Export < last.Export {
			return ErrInvalidReading
		}
	}

	return s.repo.Save(ctx, rec)
}

func (s *SolarService) CalculateSummaries(records []MeterRecord) []PeriodSummary {
	if len(records) < 2 {
		return nil
	}

	summaries := make([]PeriodSummary, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		prev := records[i-1]
		curr := records[i]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		impDiff := curr.Import - prev.Import
		expDiff := curr.Export - prev.Export
		solarGenDiff := curr.SolarGen - prev.SolarGen
		net := expDiff - impDiff

		summary := PeriodSummary{
			FromDate:   prev.Date,
			ToDate:     curr.Date,
			Days:       days,
			ImportDiff: impDiff,
			ExportDiff: expDiff,
			NetBalance: net,
			SolarGen:   solarGenDiff,
		}

		if curr.SolarGen > 0 {
			summary.HasSolar = true
			summary.Consumption = (solarGenDiff + impDiff) - expDiff
			if days > 0 {
				summary.DailyAvgConsom = summary.Consumption / days
			}
		}

		summaries = append(summaries, summary)
	}

	return summaries
}
