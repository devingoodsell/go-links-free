package models

import (
	"context"
	"encoding/json"
	"time"
)

type SystemStats struct {
	DailyActiveUsers  int       `json:"daily_active_users"`
	MonthlyActiveUsers int       `json:"monthly_active_users"`
	TotalLinks        int       `json:"total_links"`
	ActiveLinks       int       `json:"active_links"`
	ExpiredLinks      int       `json:"expired_links"`
	TotalRedirects    int       `json:"total_redirects"`
	LastUpdated       time.Time `json:"last_updated"`
}

type AnalyticsRepository struct {
	db *db.DB
}

func NewAnalyticsRepository(db *db.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{LastUpdated: time.Now()}

	// Get daily active users (users with links accessed in last 24 hours)
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT l.created_by)
		FROM links l
		JOIN link_stats s ON l.id = s.link_id
		WHERE s.last_accessed_at > NOW() - INTERVAL '24 hours'
	`).Scan(&stats.DailyActiveUsers)
	if err != nil {
		return nil, err
	}

	// Get monthly active users
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT l.created_by)
		FROM links l
		JOIN link_stats s ON l.id = s.link_id
		WHERE s.last_accessed_at > NOW() - INTERVAL '30 days'
	`).Scan(&stats.MonthlyActiveUsers)
	if err != nil {
		return nil, err
	}

	// Get link statistics
	err = r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*),
			COUNT(CASE WHEN expires_at IS NULL OR expires_at > NOW() THEN 1 END),
			COUNT(CASE WHEN expires_at <= NOW() THEN 1 END)
		FROM links
	`).Scan(&stats.TotalLinks, &stats.ActiveLinks, &stats.ExpiredLinks)
	if err != nil {
		return nil, err
	}

	// Get total redirects
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_count), 0)
		FROM link_stats
	`).Scan(&stats.TotalRedirects)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

type TimeSeriesData struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int       `json:"value"`
}

func (r *AnalyticsRepository) GetRedirectsOverTime(ctx context.Context, period string) ([]TimeSeriesData, error) {
	var query string
	switch period {
	case "daily":
		query = `
			SELECT DATE_TRUNC('day', s.last_accessed_at) as date,
				   SUM(s.daily_count) as count
			FROM link_stats s
			WHERE s.last_accessed_at > NOW() - INTERVAL '30 days'
			GROUP BY date
			ORDER BY date`
	case "weekly":
		query = `
			SELECT DATE_TRUNC('week', s.last_accessed_at) as date,
				   SUM(s.weekly_count) as count
			FROM link_stats s
			WHERE s.last_accessed_at > NOW() - INTERVAL '12 weeks'
			GROUP BY date
			ORDER BY date`
	default:
		return nil, errors.New("invalid period")
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []TimeSeriesData
	for rows.Next() {
		var item TimeSeriesData
		if err := rows.Scan(&item.Timestamp, &item.Value); err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return data, nil
}

type PopularLink struct {
	ID             int64      `json:"id"`
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destination_url"`
	TotalClicks    int        `json:"total_clicks"`
	DailyClicks    int        `json:"daily_clicks"`
	WeeklyClicks   int        `json:"weekly_clicks"`
	CreatedBy      string     `json:"created_by"` // User email
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type UserActivity struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	LinkCount int       `json:"link_count"`
	LastLogin time.Time `json:"last_login"`
	Stats     struct {
		TotalClicks     int `json:"total_clicks"`
		ActiveLinks     int `json:"active_links"`
		ExpiredLinks    int `json:"expired_links"`
		LinksCreated30d int `json:"links_created_30d"`
	} `json:"stats"`
}

func (r *AnalyticsRepository) GetPopularLinks(ctx context.Context, limit int, period string) ([]PopularLink, error) {
	query := `
		SELECT l.id, l.alias, l.destination_url, 
			   s.total_count, s.daily_count, s.weekly_count,
			   u.email, l.expires_at
		FROM links l
		JOIN link_stats s ON l.id = s.link_id
		JOIN users u ON l.created_by = u.id
		WHERE s.last_accessed_at > NOW() - INTERVAL $1
		ORDER BY 
			CASE $1 
				WHEN '24 hours' THEN s.daily_count
				WHEN '7 days' THEN s.weekly_count
				ELSE s.total_count
			END DESC
		LIMIT $2`

	interval := "24 hours"
	switch period {
	case "weekly":
		interval = "7 days"
	case "monthly":
		interval = "30 days"
	case "all":
		interval = "100 years" // Effectively no time limit
	}

	rows, err := r.db.QueryContext(ctx, query, interval, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []PopularLink
	for rows.Next() {
		var link PopularLink
		err := rows.Scan(
			&link.ID, &link.Alias, &link.DestinationURL,
			&link.TotalClicks, &link.DailyClicks, &link.WeeklyClicks,
			&link.CreatedBy, &link.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

func (r *AnalyticsRepository) GetUserActivity(ctx context.Context, days int) ([]UserActivity, error) {
	query := `
		WITH user_stats AS (
			SELECT 
				u.id,
				u.email,
				u.last_login,
				COUNT(l.id) as link_count,
				COALESCE(SUM(s.total_count), 0) as total_clicks,
				COUNT(CASE WHEN l.expires_at IS NULL OR l.expires_at > NOW() THEN 1 END) as active_links,
				COUNT(CASE WHEN l.expires_at <= NOW() THEN 1 END) as expired_links,
				COUNT(CASE WHEN l.created_at > NOW() - INTERVAL '30 days' THEN 1 END) as links_created_30d
			FROM users u
			LEFT JOIN links l ON u.id = l.created_by
			LEFT JOIN link_stats s ON l.id = s.link_id
			WHERE u.last_login > NOW() - INTERVAL '$1 days'
			GROUP BY u.id, u.email, u.last_login
		)
		SELECT * FROM user_stats
		ORDER BY total_clicks DESC`

	rows, err := r.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []UserActivity
	for rows.Next() {
		var activity UserActivity
		err := rows.Scan(
			&activity.UserID,
			&activity.Email,
			&activity.LastLogin,
			&activity.LinkCount,
			&activity.Stats.TotalClicks,
			&activity.Stats.ActiveLinks,
			&activity.Stats.ExpiredLinks,
			&activity.Stats.LinksCreated30d,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

type DomainStats struct {
	Domain     string `json:"domain"`
	LinkCount  int    `json:"link_count"`
	TotalClicks int   `json:"total_clicks"`
}

func (r *AnalyticsRepository) GetTopDomains(ctx context.Context, limit int) ([]DomainStats, error) {
	query := `
		WITH domain_extract AS (
			SELECT 
				regexp_replace(
					regexp_replace(destination_url, '^https?://([^/]+).*', '\1'),
					'^www\.', ''
				) as domain,
				id
			FROM links
		)
		SELECT 
			d.domain,
			COUNT(DISTINCT d.id) as link_count,
			COALESCE(SUM(s.total_count), 0) as total_clicks
		FROM domain_extract d
		LEFT JOIN link_stats s ON d.id = s.link_id
		GROUP BY d.domain
		ORDER BY total_clicks DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DomainStats
	for rows.Next() {
		var stat DomainStats
		err := rows.Scan(&stat.Domain, &stat.LinkCount, &stat.TotalClicks)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

type PeakUsageStats struct {
	HourlyStats []struct {
		Hour        int `json:"hour"`
		Redirects   int `json:"redirects"`
		UniqueUsers int `json:"unique_users"`
	} `json:"hourly_stats"`
	PeakHour     int       `json:"peak_hour"`
	PeakRedirects int       `json:"peak_redirects"`
	Date         time.Time  `json:"date"`
}

type PerformanceMetrics struct {
	AverageResponseTime float64   `json:"average_response_time_ms"`
	P95ResponseTime    float64   `json:"p95_response_time_ms"`
	P99ResponseTime    float64   `json:"p99_response_time_ms"`
	ErrorRate          float64   `json:"error_rate"`
	RequestsPerSecond  float64   `json:"requests_per_second"`
	TimeWindow         string    `json:"time_window"`
	LastUpdated        time.Time `json:"last_updated"`
}

func (r *AnalyticsRepository) GetPeakUsage(ctx context.Context, date time.Time) (*PeakUsageStats, error) {
	query := `
		WITH hourly_stats AS (
			SELECT 
				EXTRACT(HOUR FROM s.last_accessed_at) as hour,
				COUNT(*) as redirects,
				COUNT(DISTINCT l.created_by) as unique_users
			FROM link_stats s
			JOIN links l ON s.link_id = l.id
			WHERE DATE(s.last_accessed_at) = DATE($1)
			GROUP BY EXTRACT(HOUR FROM s.last_accessed_at)
		),
		peak_stats AS (
			SELECT hour, redirects
			FROM hourly_stats
			ORDER BY redirects DESC
			LIMIT 1
		)
		SELECT 
			json_agg(
				json_build_object(
					'hour', hour,
					'redirects', redirects,
					'unique_users', unique_users
				) ORDER BY hour
			) as hourly_stats,
			(SELECT hour FROM peak_stats) as peak_hour,
			(SELECT redirects FROM peak_stats) as peak_redirects
		FROM hourly_stats`

	stats := &PeakUsageStats{Date: date}
	var hourlyStatsJSON []byte

	err := r.db.QueryRowContext(ctx, query, date).Scan(
		&hourlyStatsJSON,
		&stats.PeakHour,
		&stats.PeakRedirects,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(hourlyStatsJSON, &stats.HourlyStats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *AnalyticsRepository) GetPerformanceMetrics(ctx context.Context, window string) (*PerformanceMetrics, error) {
	// Convert window to PostgreSQL interval
	interval := "1 hour"
	switch window {
	case "day":
		interval = "24 hours"
	case "week":
		interval = "7 days"
	case "month":
		interval = "30 days"
	}

	query := `
		WITH request_stats AS (
			SELECT 
				AVG(response_time) as avg_response_time,
				PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time) as p95_response_time,
				PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY response_time) as p99_response_time,
				COUNT(*) FILTER (WHERE status_code >= 500) * 100.0 / COUNT(*) as error_rate,
				COUNT(*) * 1.0 / EXTRACT(EPOCH FROM ($1::interval)) as requests_per_second
			FROM request_logs
			WHERE timestamp > NOW() - $1::interval
		)`

	metrics := &PerformanceMetrics{
		TimeWindow:  window,
		LastUpdated: time.Now(),
	}

	err := r.db.QueryRowContext(ctx, query, interval).Scan(
		&metrics.AverageResponseTime,
		&metrics.P95ResponseTime,
		&metrics.P99ResponseTime,
		&metrics.ErrorRate,
		&metrics.RequestsPerSecond,
	)
	if err != nil {
		return nil, err
	}

	return metrics, nil
} 