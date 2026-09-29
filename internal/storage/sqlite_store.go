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
	db    *sql.DB
	cache *Cache[string, domain.MeterRecord]
}

func NewSQLiteStore(dbPath string, cache *Cache[string, domain.MeterRecord]) (*SQLiteStore, error) {
	log.Println("...new DB:: ", dbPath)
	if dbPath == "" {
		return nil, fmt.Errorf("no database file found")
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
	if cache == nil {
		return nil, fmt.Errorf("cache is nil, initialize")
	}

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("location of the timezone couldnt be found %v\n", location)
		return nil, fmt.Errorf("timezone location not found error: %w", err)
	}

	store := &SQLiteStore{db: db, cache: cache}
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
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						date TEXT UNIQUE NOT NULL,
            import REAL NOT NULL,
            export REAL NOT NULL,
            solar_gen REAL NOT NULL,
						addedon TEXT NOT NULL 
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
	log.Println("...GetAll from Cache")

	cachelen := s.cache.Len()
	if cachelen > 0 {
		log.Println("found cache, length:: ", cachelen)
		records := s.cache.AllValues()
		return records, nil
	}

	/// missed cache

	log.Println("Missed cached, Going to DB now")
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, date, import, export, solar_gen,addedon
        FROM meter_records
        ORDER BY date ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.MeterRecord
	for rows.Next() {
		var dateStr, addedDateStr string
		var id int64
		var importValue, exportValue, solarGen float64
		if err := rows.Scan(&id, &dateStr, &importValue, &exportValue, &solarGen, &addedDateStr); err != nil {
			return nil, err
		}

		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse meter date %q: %w", dateStr, err)
		}

		parsedAddedDate, err := time.Parse("2006-01-02 15:04", addedDateStr)
		if err != nil {
			return nil, fmt.Errorf("parse meter date %q: %w", addedDateStr, err)
		}
		record := domain.MeterRecord{
			Id:       id,
			Date:     parsedDate,
			Import:   importValue,
			Export:   exportValue,
			SolarGen: solarGen,
			AddedOn:  parsedAddedDate,
		}
		records = append(records, record)
		log.Printf("Set item in cache for :: %v\n", dateStr)
		s.cache.Set(dateStr, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	log.Printf("Returning from DB get all records: %v\n", len(records))
	return records, nil
}

func (s *SQLiteStore) Save(ctx context.Context, record domain.MeterRecord) error {
	log.Printf("...Save DB %v\n", record.Id)

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return err
	}
	record.AddedOn = time.Now().In(location)
	dateKey := record.Date.Format("2006-01-02")

	// RETURNING id fixes the fallback update tracking error cleanly
	query := `
		INSERT INTO meter_records (date, import, export, solar_gen, addedon) 
		VALUES (?, ?, ?, ?, ?) 
		ON CONFLICT(date) DO UPDATE SET 
			import = excluded.import, 
			export = excluded.export, 
			solar_gen = excluded.solar_gen, 
			addedon = excluded.addedon
		RETURNING id`

	var actualID int64
	err = s.db.QueryRowContext(
		ctx, query,
		dateKey,
		record.Import,
		record.Export,
		record.SolarGen,
		record.AddedOn.Format("2006-01-02 15:04"),
	).Scan(&actualID)
	if err != nil {
		return err
	}
	log.Printf("Actual ID is %v\n", actualID)
	record.Id = actualID
	s.cache.Set(dateKey, record)

	return nil
}
