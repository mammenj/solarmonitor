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
	if cache == nil {
		return nil, fmt.Errorf("cache is nil, initialize..")
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
						id INTEGER PRIMARY KEY,
						date TEXT NOT NULL UNIQUE,
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
	////
	records := make([]domain.MeterRecord, 0)

	itmes := s.cache.Items()
	if len(itmes) > 0 {
		log.Println("found cache.........# cache", len(itmes))
		for _, value := range itmes {
			records = append(records, value)
		}
		return records, nil
	}

	/// missed cache

	log.Println("Missed cached or 0 items, Going to DB now")
	rows, err := s.db.QueryContext(ctx, `
        SELECT date, import, export, solar_gen,addedon
        FROM meter_records
        ORDER BY date ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dateStr, addedDateStr string
		var importValue, exportValue, solarGen float64
		if err := rows.Scan(&dateStr, &importValue, &exportValue, &solarGen, &addedDateStr); err != nil {
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

	return records, nil
}

func (s *SQLiteStore) getLastRecordV2(ctx context.Context) (domain.MeterRecord, error) {
	var last_record domain.MeterRecord
	var dateStr string
	var addedDateStr string
	//// geting from cache

	var latestDate string
	for dateKey := range s.cache.Items() {
		// Standard alphanumeric string comparison works perfectly for YYYY-MM-DD
		if latestDate == "" || dateKey > latestDate {
			latestDate = dateKey
		}
	}

	last_record = s.cache.Get(latestDate)
	log.Println("Got last record from cache for ::", latestDate)
	if latestDate != "" {
		return last_record, nil
	}
	//
	log.Println("Going to DB for last_record")
	err := s.db.QueryRowContext(
		ctx,
		"SELECT date, import, export, solar_gen,addedon FROM meter_records ORDER BY id DESC LIMIT 1",
	).Scan(
		&dateStr, &last_record.Import, &last_record.Export, &last_record.SolarGen, &addedDateStr,
	)
	if err != nil {
		return domain.MeterRecord{}, err
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return domain.MeterRecord{}, fmt.Errorf("parse meter date %q: %w", dateStr, err)
	}

	parsedAddedDate, err := time.Parse("2006-01-02 15:04", addedDateStr)
	if err != nil {
		return domain.MeterRecord{}, fmt.Errorf("parse added date %q: %w", addedDateStr, err)
	}

	last_record.Date = parsedDate
	last_record.AddedOn = parsedAddedDate
	return last_record, nil
}

func (s *SQLiteStore) Save(ctx context.Context, record domain.MeterRecord) error {
	log.Println("...Save DB")
	last_record, errLast := s.getLastRecordV2(ctx)
	if errLast != nil {
		log.Println("No last record found, so initializing the db... ", errLast)
		//return fmt.Errorf("No data found, initialize DB...", errLast)
	}

	log.Printf("Last record is %v\n", last_record)

	now := time.Now()
	location, _ := time.LoadLocation("Asia/Kolkata")
	record.AddedOn = now.In(location)

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
	if err == nil {
		log.Println("Setting cache for ::", record.Date.Format("2006-01-02"))
		s.cache.Set(record.Date.Format("2006-01-02"), record)
	}
	return err
}
