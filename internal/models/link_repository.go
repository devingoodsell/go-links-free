package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/devingoodsell/go-links-free/internal/db"
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
		INSERT INTO links (alias, destination_url, created_by, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		link.Alias,
		link.DestinationURL,
		link.CreatedBy,
		link.ExpiresAt,
		link.IsActive,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		// Check for duplicate alias
		if isPgDuplicateError(err) {
			return ErrDuplicate
		}
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
			   COALESCE(s.daily_count, 0), COALESCE(s.weekly_count, 0),
			   COALESCE(s.total_count, 0), s.last_accessed_at as "lastAccessedAt"
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
		return nil, ErrNotFound
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
		WHERE id = $3 AND created_by = $4
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		link.DestinationURL,
		link.ExpiresAt,
		link.ID,
		link.CreatedBy,
	).Scan(&link.UpdatedAt)

	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

func (r *LinkRepository) Delete(ctx context.Context, id int64, userID int64) error {
	log.Printf("Delete called with id=%d, userID=%d", id, userID)

	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete stats first due to foreign key constraint
	_, err = tx.ExecContext(ctx,
		"DELETE FROM link_stats WHERE link_id = $1",
		id)
	if err != nil {
		return err
	}

	// Then delete the link
	result, err := tx.ExecContext(ctx,
		"DELETE FROM links WHERE id = $1 AND created_by = $2",
		id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return tx.Commit()
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

	// Add domain filter
	if opts.Domain != "" {
		argCount++
		query += fmt.Sprintf(` AND l.destination_url ILIKE $%d`, argCount)
		args = append(args, "%"+opts.Domain+"%")
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

// Update these method signatures
func (r *LinkRepository) BulkDelete(ctx context.Context, userID int64, ids []int64) error {
	// First verify ownership of all links
	for _, id := range ids {
		link, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link.CreatedBy != userID {
			return ErrUnauthorized
		}
	}

	// Then perform the bulk delete
	query := `DELETE FROM links WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, pq.Array(ids))
	return err
}

func (r *LinkRepository) BulkUpdateStatus(ctx context.Context, userID int64, ids []int64, isActive bool) error {
	// First verify ownership of all links
	for _, id := range ids {
		link, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if link.CreatedBy != userID {
			return ErrUnauthorized
		}
	}

	// Then perform the bulk update
	query := `UPDATE links SET is_active = $1 WHERE id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, isActive, pq.Array(ids))
	return err
}

// Helper function to check for Postgres duplicate key error
func isPgDuplicateError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

func (r *LinkRepository) GetByID(ctx context.Context, id int64) (*Link, error) {
	query := `
		SELECT l.id, l.alias, l.destination_url, l.created_by, l.expires_at,
			   l.created_at, l.updated_at,
			   COALESCE(s.daily_count, 0), COALESCE(s.weekly_count, 0),
			   COALESCE(s.total_count, 0), s.last_accessed_at as "lastAccessedAt"
		FROM links l
		LEFT JOIN link_stats s ON l.id = s.link_id
		WHERE l.id = $1`

	link := &Link{Stats: &LinkStats{}}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&link.ID, &link.Alias, &link.DestinationURL, &link.CreatedBy,
		&link.ExpiresAt, &link.CreatedAt, &link.UpdatedAt,
		&link.Stats.DailyCount, &link.Stats.WeeklyCount, &link.Stats.TotalCount,
		&link.Stats.LastAccessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (r *LinkRepository) ListForUser(ctx context.Context, userID int64, page, pageSize int) ([]Link, int64, error) {
	offset := page * pageSize

	log.Printf("Listing links for user %d, page %d, pageSize %d", userID, page, pageSize)

	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM links WHERE created_by = $1",
		userID,
	).Scan(&total)
	if err != nil {
		log.Printf("Error getting total count: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d total links", total)

	// Get paginated links
	query := `
		SELECT id, alias, destination_url, created_by, expires_at, created_at, updated_at, is_active 
		FROM links 
		WHERE created_by = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	log.Printf("Executing query: %s with userID=%d, pageSize=%d, offset=%d",
		query, userID, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		err := rows.Scan(
			&link.ID,
			&link.Alias,
			&link.DestinationURL,
			&link.CreatedBy,
			&link.ExpiresAt,
			&link.CreatedAt,
			&link.UpdatedAt,
			&link.IsActive,
		)
		if err != nil {
			return nil, 0, err
		}
		link.Stats = &LinkStats{} // Initialize empty stats
		links = append(links, link)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return links, total, nil
}
