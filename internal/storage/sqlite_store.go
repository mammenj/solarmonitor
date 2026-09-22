package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"solarmonitor/internal/domain"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	if dbPath == "" {
		dbPath = "solar_readings.db"
	}

	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize sqlite schema: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) init() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS meter_records (
            date TEXT NOT NULL PRIMARY KEY,
            import REAL NOT NULL,
            export REAL NOT NULL,
            solar_gen REAL NOT NULL DEFAULT 0
        );
    `)
	return err
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	log.Println("...closing DB")
	return s.db.Close()
}

func (s *SQLiteStore) GetAll(ctx context.Context) ([]domain.MeterRecord, error) {
	log.Println("...GetAll DB")
	rows, err := s.db.QueryContext(ctx, `
        SELECT date, import, export, solar_gen
        FROM meter_records
        ORDER BY date ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.MeterRecord, 0)
	for rows.Next() {
		var dateStr string
		var importValue, exportValue, solarGen float64

		if err := rows.Scan(&dateStr, &importValue, &exportValue, &solarGen); err != nil {
			return nil, err
		}

		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse meter date %q: %w", dateStr, err)
		}

		records = append(records, domain.MeterRecord{
			Date:     parsedDate,
			Import:   importValue,
			Export:   exportValue,
			SolarGen: solarGen,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *SQLiteStore) Save(ctx context.Context, record domain.MeterRecord) error {
	log.Println("...Save DB")

	if record.Date.IsZero() {
		return fmt.Errorf("meter record date cannot be zero")
	}

	_, err := s.db.ExecContext(
		ctx,
		`
            INSERT INTO meter_records (date, import, export, solar_gen)
            VALUES (?, ?, ?, ?)
            ON CONFLICT(date) DO UPDATE SET
                import = excluded.import,
                export = excluded.export,
                solar_gen = excluded.solar_gen
        `,
		record.Date.Format("2006-01-02"),
		record.Import,
		record.Export,
		record.SolarGen,
	)
	return err
}
