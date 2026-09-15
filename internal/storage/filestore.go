package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"solarmonitor/internal/domain"
)

type FileStore struct {
	filepath string
	mu       sync.RWMutex
}

func NewFileStore(filepath string) *FileStore {
	return &FileStore{filepath: filepath}
}

func (f *FileStore) GetAll(ctx context.Context) ([]domain.MeterRecord, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, err := os.Open(f.filepath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []domain.MeterRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		d, err1 := time.Parse("2006-01-02", parts[0])
		imp, err2 := strconv.ParseFloat(parts[1], 64)
		exp, err3 := strconv.ParseFloat(parts[2], 64)

		if err1 == nil && err2 == nil && err3 == nil {
			var gen float64
			if len(parts) >= 4 {
				if parsedGen, err4 := strconv.ParseFloat(parts[3], 64); err4 == nil {
					gen = parsedGen
				}
			}
			records = append(records, domain.MeterRecord{
				Date:     d,
				Import:   imp,
				Export:   exp,
				SolarGen: gen,
			})
		}
	}
	return records, scanner.Err()
}

func (f *FileStore) Save(ctx context.Context, r domain.MeterRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	line := fmt.Sprintf("%s|%.2f|%.2f|%.2f\n", r.Date.Format("2006-01-02"), r.Import, r.Export, r.SolarGen)
	if _, err := file.WriteString(line); err != nil {
		file.Close()
		return err
	}

	return file.Close()
}
