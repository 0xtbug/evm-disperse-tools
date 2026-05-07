package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
)

// ReportsFileRepo implements ReportRepository using JSON files
type ReportsFileRepo struct {
	dirpath string
}

// NewReportsFileRepo creates a new ReportsFileRepo
func NewReportsFileRepo(dirpath string) *ReportsFileRepo {
	return &ReportsFileRepo{
		dirpath: dirpath,
	}
}

// Save saves an execution report to a JSON file
func (r *ReportsFileRepo) Save(report *entity.ExecutionReport) error {
	if err := os.MkdirAll(r.dirpath, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Use timestamp + txHash for unique filename
	timestamp := report.Timestamp.Format("2006-01-02")
	hashPrefix := report.TxHash
	if len(report.TxHash) >= 8 {
		hashPrefix = report.TxHash[:8]
	}
	filename := fmt.Sprintf("%s_%s.json", timestamp, hashPrefix)
	filepath := filepath.Join(r.dirpath, filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// Load loads a report by transaction hash
func (r *ReportsFileRepo) Load(txHash string) (*entity.ExecutionReport, error) {
	// Search for file matching the txHash
	entries, err := os.ReadDir(r.dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("report not found for txHash: %s", txHash)
		}
		return nil, fmt.Errorf("failed to read reports directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && len(name) >= 8 && len(txHash) >= 8 && name[:8] == txHash[:8] {
			filepath := filepath.Join(r.dirpath, name)
			data, err := os.ReadFile(filepath)
			if err != nil {
				return nil, fmt.Errorf("failed to read report file: %w", err)
			}

			var report entity.ExecutionReport
			if err := json.Unmarshal(data, &report); err != nil {
				return nil, fmt.Errorf("failed to unmarshal report: %w", err)
			}

			return &report, nil
		}
	}

	return nil, fmt.Errorf("report not found for txHash: %s", txHash)
}

// ListAll returns all reports across all dates, sorted by newest first.
func (r *ReportsFileRepo) ListAll() ([]*entity.ExecutionReport, error) {
	entries, err := os.ReadDir(r.dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entity.ExecutionReport{}, nil
		}
		return nil, fmt.Errorf("failed to read reports directory: %w", err)
	}

	var reports []*entity.ExecutionReport
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		fpath := filepath.Join(r.dirpath, name)
		data, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}
		var report entity.ExecutionReport
		if json.Unmarshal(data, &report) != nil {
			continue
		}
		reports = append(reports, &report)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp.After(reports[j].Timestamp)
	})

	return reports, nil
}

// ListByDate returns all reports for a specific date
func (r *ReportsFileRepo) ListByDate(date time.Time) ([]*entity.ExecutionReport, error) {
	entries, err := os.ReadDir(r.dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*entity.ExecutionReport{}, nil
		}
		return nil, fmt.Errorf("failed to read reports directory: %w", err)
	}

	dateStr := date.Format("2006-01-02")
	var reports []*entity.ExecutionReport

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && len(name) >= 10 && name[:10] == dateStr {
			filepath := filepath.Join(r.dirpath, name)
			data, err := os.ReadFile(filepath)
			if err != nil {
				return nil, fmt.Errorf("failed to read report file: %w", err)
			}

			var report entity.ExecutionReport
			if err := json.Unmarshal(data, &report); err != nil {
				return nil, fmt.Errorf("failed to unmarshal report: %w", err)
			}

			reports = append(reports, &report)
		}
	}

	return reports, nil
}
