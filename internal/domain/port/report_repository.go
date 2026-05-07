package port

import (
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
)

// ReportRepository defines the interface for execution reports
type ReportRepository interface {
	Save(report *entity.ExecutionReport) error
	Load(txHash string) (*entity.ExecutionReport, error)
	ListByDate(date time.Time) ([]*entity.ExecutionReport, error)
	ListAll() ([]*entity.ExecutionReport, error)
}
