#!/bin/bash
set -e

# Create databases
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE assets;
    CREATE DATABASE telemetry;
EOSQL

# Load assets schema & seed into assets database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "assets" <<-EOSQL
    CREATE TABLE IF NOT EXISTS assets (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL,
        type VARCHAR(50) NOT NULL,
        status VARCHAR(50) NOT NULL,
        health_score INT DEFAULT 100,
        location VARCHAR(255),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        last_maintained_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);
    CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);

    INSERT INTO assets (name, type, status, health_score, location) VALUES
    ('Peloton Bike #1', 'cardio', 'available', 95, 'Zone A'),
    ('Peloton Bike #2', 'cardio', 'available', 88, 'Zone A'),
    ('Peloton Tread #1', 'cardio', 'available', 92, 'Zone A'),
    ('Security Camera Entry', 'security', 'available', 100, 'Entrance'),
    ('Security Camera Gym Floor', 'security', 'available', 100, 'Main Floor'),
    ('Smith Machine', 'strength', 'available', 85, 'Zone B'),
    ('Bench Press #1', 'strength', 'available', 90, 'Zone B'),
    ('Dumbbell Set 5-50lbs', 'strength', 'available', 98, 'Zone B'),
    ('Yoga Mat #1', 'yoga', 'available', 75, 'Studio 1'),
    ('Yoga Mat #2', 'yoga', 'available', 80, 'Studio 1'),
    ('Kettlebell Set', 'strength', 'available', 95, 'Zone B'),
    ('Rowing Machine', 'cardio', 'available', 82, 'Zone A'),
    ('Elliptical #1', 'cardio', 'available', 70, 'Zone A'),
    ('Leg Press', 'strength', 'available', 88, 'Zone B'),
    ('Climbing Wall Wall #1', 'climbing', 'available', 99, 'Zone C')
    ON CONFLICT DO NOTHING;
EOSQL

# Load telemetry schema into telemetry database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "telemetry" <<-EOSQL
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
EOSQL
