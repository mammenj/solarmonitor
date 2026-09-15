package domain

import "context"

type Repository interface {
	GetAll(ctx context.Context) ([]MeterRecord, error)
	Save(ctx context.Context, record MeterRecord) error
}
