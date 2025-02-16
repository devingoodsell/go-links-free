package models

import (
	"context"
	"fmt"
	"time"

	"github.com/devingoodsell/go-links-free/internal/db"
)

type LogRetentionPolicy struct {
	DetailedRetentionDays  int // Keep detailed logs for this many days
	AggregateRetentionDays int // Keep aggregated statistics for this many days
	BatchSize              int // Number of records to delete in each batch
	MaxDeletionsPerRun     int // Maximum number of records to delete per cleanup run
}

type LogManager struct {
	db     *db.DB
	policy LogRetentionPolicy
}

func NewLogManager(db *db.DB, policy LogRetentionPolicy) *LogManager {
	if policy.DetailedRetentionDays == 0 {
		policy.DetailedRetentionDays = 30 // Default to 30 days
	}
	if policy.AggregateRetentionDays == 0 {
		policy.AggregateRetentionDays = 90 // Default to 90 days
	}
	if policy.BatchSize == 0 {
		policy.BatchSize = 1000 // Default to 1000 records per batch
	}
	if policy.MaxDeletionsPerRun == 0 {
		policy.MaxDeletionsPerRun = 10000 // Default to 10000 records per run
	}

	return &LogManager{
		db:     db,
		policy: policy,
	}
}

// Aggregate logs before deletion
func (m *LogManager) AggregateLogs(ctx context.Context, date time.Time) error {
	query := `
		INSERT INTO request_log_aggregates (
			date,
			total_requests,
			avg_response_time,
			error_count,
			status_2xx,
			status_3xx,
			status_4xx,
			status_5xx
		)
		SELECT
			DATE(timestamp),
			COUNT(*),
			AVG(response_time),
			COUNT(*) FILTER (WHERE status_code >= 500),
			COUNT(*) FILTER (WHERE status_code >= 200 AND status_code < 300),
			COUNT(*) FILTER (WHERE status_code >= 300 AND status_code < 400),
			COUNT(*) FILTER (WHERE status_code >= 400 AND status_code < 500),
			COUNT(*) FILTER (WHERE status_code >= 500)
		FROM request_logs
		WHERE DATE(timestamp) = DATE($1)
		GROUP BY DATE(timestamp)
		ON CONFLICT (date) DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			avg_response_time = EXCLUDED.avg_response_time,
			error_count = EXCLUDED.error_count,
			status_2xx = EXCLUDED.status_2xx,
			status_3xx = EXCLUDED.status_3xx,
			status_4xx = EXCLUDED.status_4xx,
			status_5xx = EXCLUDED.status_5xx`

	_, err := m.db.ExecContext(ctx, query, date)
	return err
}

// Clean up old logs
func (m *LogManager) CleanupOldLogs(ctx context.Context) error {
	// Start a transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Get the cutoff dates
	detailedCutoff := time.Now().AddDate(0, 0, -m.policy.DetailedRetentionDays)
	aggregateCutoff := time.Now().AddDate(0, 0, -m.policy.AggregateRetentionDays)

	// Aggregate logs before deletion
	dates, err := m.getDistinctDates(ctx, detailedCutoff)
	if err != nil {
		return fmt.Errorf("failed to get distinct dates: %v", err)
	}

	for _, date := range dates {
		if err := m.AggregateLogs(ctx, date); err != nil {
			return fmt.Errorf("failed to aggregate logs for %v: %v", date, err)
		}
	}

	// Delete old detailed logs in batches
	totalDeleted := 0
	for totalDeleted < m.policy.MaxDeletionsPerRun {
		result, err := tx.ExecContext(ctx, `
			DELETE FROM request_logs
			WHERE id IN (
				SELECT id FROM request_logs
				WHERE timestamp < $1
				LIMIT $2
			)`,
			detailedCutoff,
			m.policy.BatchSize,
		)
		if err != nil {
			return fmt.Errorf("failed to delete detailed logs: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %v", err)
		}

		if rowsAffected == 0 {
			break // No more records to delete
		}

		totalDeleted += int(rowsAffected)
	}

	// Delete old aggregated logs
	_, err = tx.ExecContext(ctx, `
		DELETE FROM request_log_aggregates
		WHERE date < $1`,
		aggregateCutoff,
	)
	if err != nil {
		return fmt.Errorf("failed to delete aggregated logs: %v", err)
	}

	return tx.Commit()
}

func (m *LogManager) getDistinctDates(ctx context.Context, cutoff time.Time) ([]time.Time, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT DISTINCT DATE(timestamp)
		FROM request_logs
		WHERE timestamp < $1
		ORDER BY DATE(timestamp)`,
		cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}

	return dates, nil
}
