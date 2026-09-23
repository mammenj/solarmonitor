package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"solarmonitor/internal/domain"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	log.Println("...new DB")
	if dbPath == "" {
		return nil, fmt.Errorf("No database file found...")
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
	log.Println("...init DB")
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS meter_records (
						id INTEGER PRIMARY KEY,
						date TEXT NOT NULL UNIQUE,
            import REAL NOT NULL,
            export REAL NOT NULL,
            solar_gen REAL NOT NULL,
						addedon TEXT DEFAULT CURRENT_TIMESTAMP 
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
        SELECT date, import, export, solar_gen,addedon
        FROM meter_records
        ORDER BY date ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.MeterRecord, 0)
	for rows.Next() {
		var dateStr, addedDateStr string
		var importValue, exportValue, solarGen float64
		if err := rows.Scan(&dateStr, &importValue, &exportValue, &solarGen, &addedDateStr); err != nil {
			return nil, err
		}

		parsedDate, err := time.Parse("2006-01-02", dateStr)
		parsedAddedDate, err := time.Parse("2006-01-02 15:04", addedDateStr)
		if err != nil {
			return nil, fmt.Errorf("parse meter date %q: %w", dateStr, err)
		}

		records = append(records, domain.MeterRecord{
			Date:     parsedDate,
			Import:   importValue,
			Export:   exportValue,
			SolarGen: solarGen,
			AddedOn:  parsedAddedDate,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *SQLiteStore) getLastRecord(ctx context.Context) (domain.MeterRecord, error) {
	log.Println("...getLastRecord DB")
	emptyRecord := domain.MeterRecord{
		Date:     time.Now(),
		Import:   0.0,
		Export:   0.0,
		SolarGen: 0.0,
		AddedOn:  time.Now(),
	}
	rows, err := s.db.QueryContext(ctx, `
        SELECT date, import, export, solar_gen,addedon
        FROM meter_records
        ORDER BY id DESC LIMIT 1
		`)
	log.Println("After QueryContext")
	if err != nil {
		return emptyRecord, err
	}
	defer rows.Close()

	records := make([]domain.MeterRecord, 0)
	count := 0
	for rows.Next() {
		count++
		var dateStr, addedDateStr string
		var importValue, exportValue, solarGen float64

		if err := rows.Scan(&dateStr, &importValue, &exportValue, &solarGen, &addedDateStr); err != nil {
			return emptyRecord, err
		}
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		parsedAddedDate, err := time.Parse("2006-01-02 15:04", addedDateStr)
		if err != nil {
			return emptyRecord, fmt.Errorf("parse meter date %q: %w", dateStr, err)
		}
		log.Println("importValue ", importValue)
		log.Println("exportValue", exportValue)
		log.Println("solarGen ", solarGen)
		log.Println("date ", parsedDate)

		records = append(records, domain.MeterRecord{
			Date:     parsedDate,
			Import:   importValue,
			Export:   exportValue,
			SolarGen: solarGen,
			AddedOn:  parsedAddedDate,
		})
	}

	log.Println("Count is ", count)
	if err := rows.Err(); err != nil {
		return emptyRecord, err
	}
	if count == 0 {
		return emptyRecord, nil
	}
	return records[0], nil
}

/*
*
rows, err := s.db.QueryContext(ctx, `
        SELECT date, import, export, solar_gen,addedon
        FROM meter_records
        ORDER BY id DESC LIMIT 1
    `)

*/

func (s *SQLiteStore) Save(ctx context.Context, record domain.MeterRecord) error {
	log.Println("...Save DB")
	last_record, errLast := s.getLastRecord(ctx)
	if errLast != nil {
		log.Println("No last record found %v", errLast)
		return fmt.Errorf("No data found, initialize DB...")
	}

	log.Println("Last record is %v", last_record)

	now := time.Now()
	record.AddedOn = now

	if record.Import < last_record.Import {
		return fmt.Errorf("Cannot be less than to the last IMPORT")
	}
	if record.Export < last_record.Export {
		return fmt.Errorf("Cannot be less than the last EXPORT")
	}
	if record.SolarGen < last_record.SolarGen {
		return fmt.Errorf("Cannot be less than the last Generated SOLAR")
	}

	_, err := s.db.ExecContext(
		ctx,
		`
            INSERT INTO meter_records (date, import, export, solar_gen,addedon)
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(date) DO UPDATE SET
                import = excluded.import,
                export = excluded.export,
                solar_gen = excluded.solar_gen,
								addedon = excluded.addedon 
        `,
		record.Date.Format("2006-01-02"),
		record.Import,
		record.Export,
		record.SolarGen,
		record.AddedOn.Format("2006-01-02 15:04"),
	)
	return err
}
