CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    is_admin BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    alias VARCHAR(100) UNIQUE NOT NULL,
    destination_url TEXT NOT NULL,
    created_by INTEGER REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS link_stats (
    id SERIAL PRIMARY KEY,
    link_id INTEGER REFERENCES links(id),
    daily_count INTEGER DEFAULT 0,
    weekly_count INTEGER DEFAULT 0,
    total_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS request_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    path VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time FLOAT NOT NULL,
    user_id INTEGER REFERENCES users(id),
    error_message TEXT,
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    request_size INTEGER,
    response_size INTEGER,
    host VARCHAR(255),
    protocol VARCHAR(10),
    query_params TEXT,
    request_headers JSONB,
    trace_id UUID
);

CREATE TABLE IF NOT EXISTS request_log_aggregates (
    date DATE PRIMARY KEY,
    total_requests INTEGER NOT NULL,
    avg_response_time FLOAT NOT NULL,
    error_count INTEGER NOT NULL,
    status_2xx INTEGER NOT NULL,
    status_3xx INTEGER NOT NULL,
    status_4xx INTEGER NOT NULL,
    status_5xx INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_links_alias ON links(alias);
CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links(expires_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_timestamp ON request_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_request_logs_status_code ON request_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_request_logs_ip_address ON request_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_request_logs_trace_id ON request_logs(trace_id);
CREATE INDEX IF NOT EXISTS idx_request_log_aggregates_date ON request_log_aggregates(date); 