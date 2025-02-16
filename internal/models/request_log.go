package models

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/devingoodsell/go-links-free/internal/db"
	"github.com/google/uuid"
)

type RequestLog struct {
	ID             int64           `json:"id"`
	Timestamp      time.Time       `json:"timestamp"`
	Path           string          `json:"path"`
	Method         string          `json:"method"`
	StatusCode     int             `json:"status_code"`
	ResponseTime   float64         `json:"response_time"` // in milliseconds
	UserID         *int64          `json:"user_id,omitempty"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	IPAddress      net.IP          `json:"ip_address"`
	UserAgent      string          `json:"user_agent"`
	Referer        string          `json:"referer"`
	RequestSize    int64           `json:"request_size"`
	ResponseSize   int64           `json:"response_size"`
	Host           string          `json:"host"`
	Protocol       string          `json:"protocol"`
	QueryParams    string          `json:"query_params"`
	RequestHeaders json.RawMessage `json:"request_headers"`
	TraceID        uuid.UUID       `json:"trace_id"`
}

type RequestLogRepository struct {
	db *db.DB
}

func NewRequestLogRepository(db *db.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

func (r *RequestLogRepository) Create(ctx context.Context, log *RequestLog) error {
	query := `
		INSERT INTO request_logs (
			timestamp, path, method, status_code, 
			response_time, user_id, error_message,
			ip_address, user_agent, referer,
			request_size, response_size, host,
			protocol, query_params, request_headers,
			trace_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		log.Timestamp,
		log.Path,
		log.Method,
		log.StatusCode,
		log.ResponseTime,
		log.UserID,
		log.ErrorMessage,
		log.IPAddress,
		log.UserAgent,
		log.Referer,
		log.RequestSize,
		log.ResponseSize,
		log.Host,
		log.Protocol,
		log.QueryParams,
		log.RequestHeaders,
		log.TraceID,
	).Scan(&log.ID)
}

// Batch insert for better performance
func (r *RequestLogRepository) CreateBatch(ctx context.Context, logs []*RequestLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO request_logs (
			timestamp, path, method, status_code, 
			response_time, user_id, error_message,
			ip_address, user_agent, referer,
			request_size, response_size, host,
			protocol, query_params, request_headers,
			trace_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, log := range logs {
		_, err = stmt.ExecContext(ctx,
			log.Timestamp,
			log.Path,
			log.Method,
			log.StatusCode,
			log.ResponseTime,
			log.UserID,
			log.ErrorMessage,
			log.IPAddress,
			log.UserAgent,
			log.Referer,
			log.RequestSize,
			log.ResponseSize,
			log.Host,
			log.Protocol,
			log.QueryParams,
			log.RequestHeaders,
			log.TraceID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
