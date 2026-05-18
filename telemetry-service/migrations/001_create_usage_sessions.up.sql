CREATE TABLE IF NOT EXISTS usage_sessions (
    id SERIAL PRIMARY KEY,
    booking_id VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL,
    asset_id VARCHAR(255) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    duration_minutes INT DEFAULT 0,
    email_sent BOOLEAN DEFAULT FALSE
);
