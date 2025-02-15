package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yourusername/go-links/internal/db"
	"github.com/lib/pq"
)

type LinkRepository struct {
	db *db.DB
}

func NewLinkRepository(db *db.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(ctx context.Context, link *Link) error {
	query := `
		INSERT INTO links (alias, destination_url, created_by, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		link.Alias,
		link.DestinationURL,
		link.CreatedBy,
		link.ExpiresAt,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		return err
	}

	// Initialize stats
	statsQuery := `
		INSERT INTO link_stats (link_id, daily_count, weekly_count, total_count)
		VALUES ($1, 0, 0, 0)`

	_, err = r.db.ExecContext(ctx, statsQuery, link.ID)
	return err
}

func (r *LinkRepository) GetByAlias(ctx context.Context, alias string) (*Link, error) {
	query := `
		SELECT l.id, l.alias, l.destination_url, l.created_by, l.expires_at,
			   l.created_at, l.updated_at,
			   s.daily_count, s.weekly_count, s.total_count, s.last_accessed_at
		FROM links l
		LEFT JOIN link_stats s ON l.id = s.link_id
		WHERE l.alias = $1`

	link := &Link{Stats: &LinkStats{}}
	err := r.db.QueryRowContext(ctx, query, alias).Scan(
		&link.ID, &link.Alias, &link.DestinationURL, &link.CreatedBy,
		&link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt,
		&link.Stats.DailyCount, &link.Stats.WeeklyCount, &link.Stats.TotalCount,
		&link.Stats.LastAccessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (r *LinkRepository) IncrementStats(ctx context.Context, linkID int64) error {
	query := `
		UPDATE link_stats
		SET daily_count = daily_count + 1,
			weekly_count = weekly_count + 1,
			total_count = total_count + 1,
			last_accessed_at = NOW()
		WHERE link_id = $1`

	_, err := r.db.ExecContext(ctx, query, linkID)
	return err
}

func (r *LinkRepository) ListByUser(ctx context.Context, userID int64) ([]*Link, error) {
	query := `
		SELECT l.id, l.alias, l.destination_url, l.created_by, l.expires_at,
			   l.created_at, l.updated_at,
			   s.daily_count, s.weekly_count, s.total_count, s.last_accessed_at
		FROM links l
		LEFT JOIN link_stats s ON l.id = s.link_id
		WHERE l.created_by = $1
		ORDER BY l.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*Link
	for rows.Next() {
		link := &Link{Stats: &LinkStats{}}
		err := rows.Scan(
			&link.ID, &link.Alias, &link.DestinationURL, &link.CreatedBy,
			&link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt,
			&link.Stats.DailyCount, &link.Stats.WeeklyCount, &link.Stats.TotalCount,
			&link.Stats.LastAccessedAt,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

func (r *LinkRepository) Update(ctx context.Context, link *Link) error {
	query := `
		UPDATE links
		SET destination_url = $1, expires_at = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		link.DestinationURL,
		link.ExpiresAt,
		link.ID,
	).Scan(&link.UpdatedAt)

	return err
}

func (r *LinkRepository) Delete(ctx context.Context, id int64) error {
	// Start a transaction to ensure both link and stats are deleted
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete stats first due to foreign key constraint
	_, err = tx.ExecContext(ctx, "DELETE FROM link_stats WHERE link_id = $1", id)
	if err != nil {
		return err
	}

	// Delete the link
	result, err := tx.ExecContext(ctx, "DELETE FROM links WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("link not found")
	}

	return tx.Commit()
}

type ListOptions struct {
	Limit  int
	Offset int
	Search string
	Status string
	SortBy string
}

type ListResult struct {
	Links      []*Link
	TotalCount int
	HasMore    bool
}

func (r *LinkRepository) ListByUserWithPagination(ctx context.Context, userID int64, opts ListOptions) (*ListResult, error) {
	// Get total count first
	var totalCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM links WHERE created_by = $1", userID).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	// If no limit is specified, use a reasonable default
	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	query := `
		SELECT l.id, l.alias, l.destination_url, l.created_by, l.expires_at,
			   l.created_at, l.updated_at,
			   s.daily_count, s.weekly_count, s.total_count, s.last_accessed_at
		FROM links l
		LEFT JOIN link_stats s ON l.id = s.link_id
		WHERE l.created_by = $1
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, opts.Limit+1, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*Link
	for rows.Next() {
		link := &Link{Stats: &LinkStats{}}
		err := rows.Scan(
			&link.ID, &link.Alias, &link.DestinationURL, &link.CreatedBy,
			&link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt,
			&link.Stats.DailyCount, &link.Stats.WeeklyCount, &link.Stats.TotalCount,
			&link.Stats.LastAccessedAt,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	// Check if there are more results
	hasMore := len(links) > opts.Limit
	if hasMore {
		links = links[:opts.Limit] // Remove the extra item we fetched
	}

	return &ListResult{
		Links:      links,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

func (r *LinkRepository) ListByUserWithFilters(ctx context.Context, userID int64, opts ListOptions) ([]*Link, error) {
	query := `
		SELECT l.id, l.alias, l.destination_url, l.created_by, l.expires_at,
			   l.created_at, l.updated_at,
			   s.daily_count, s.weekly_count, s.total_count, s.last_accessed_at
		FROM links l
		LEFT JOIN link_stats s ON l.id = s.link_id
		WHERE l.created_by = $1`

	args := []interface{}{userID}
	argCount := 1

	// Add search condition
	if opts.Search != "" {
		argCount++
		query += fmt.Sprintf(` AND (l.alias ILIKE $%d OR l.destination_url ILIKE $%d)`, argCount, argCount)
		args = append(args, "%"+opts.Search+"%")
	}

	// Add status filter
	if opts.Status == "active" {
		query += ` AND (l.expires_at IS NULL OR l.expires_at > NOW())`
	} else if opts.Status == "expired" {
		query += ` AND l.expires_at <= NOW()`
	}

	// Add sorting
	switch opts.SortBy {
	case "created_desc":
		query += ` ORDER BY l.created_at DESC`
	case "created_asc":
		query += ` ORDER BY l.created_at ASC`
	case "clicks_desc":
		query += ` ORDER BY COALESCE(s.total_count, 0) DESC`
	default:
		query += ` ORDER BY l.created_at DESC`
	}

	// Add pagination
	if opts.Limit > 0 {
		argCount++
		query += fmt.Sprintf(` LIMIT $%d`, argCount)
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		argCount++
		query += fmt.Sprintf(` OFFSET $%d`, argCount)
		args = append(args, opts.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*Link
	for rows.Next() {
		link := &Link{Stats: &LinkStats{}}
		err := rows.Scan(
			&link.ID, &link.Alias, &link.DestinationURL, &link.CreatedBy,
			&link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt,
			&link.Stats.DailyCount, &link.Stats.WeeklyCount, &link.Stats.TotalCount,
			&link.Stats.LastAccessedAt,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

// Add new methods for bulk operations
func (r *LinkRepository) BulkDelete(ctx context.Context, ids []int64) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete stats first due to foreign key constraint
	_, err = tx.ExecContext(ctx, 
		"DELETE FROM link_stats WHERE link_id = ANY($1)", 
		pq.Array(ids))
	if err != nil {
		return err
	}

	// Delete the links
	result, err := tx.ExecContext(ctx, 
		"DELETE FROM links WHERE id = ANY($1)", 
		pq.Array(ids))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no links found to delete")
	}

	return tx.Commit()
}

func (r *LinkRepository) BulkUpdateStatus(ctx context.Context, ids []int64, isActive bool) error {
	var expiresAt *time.Time
	if !isActive {
		now := time.Now()
		expiresAt = &now
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE links SET expires_at = $1, updated_at = NOW() WHERE id = ANY($2)",
		expiresAt, pq.Array(ids))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no links found to update")
	}

	return nil
} 