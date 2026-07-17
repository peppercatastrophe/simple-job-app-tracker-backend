package database

import (
	"context"
	"fmt"
)

func RunMigrations() error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS job_applications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			company VARCHAR(255) NOT NULL,
			category VARCHAR(100),
			description TEXT,
			tech_stack TEXT[],
			status VARCHAR(50) DEFAULT 'applied',
			apply_date_time TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_user_status ON job_applications (user_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_user_apply_date ON job_applications (user_id, apply_date_time)`,
		`CREATE TABLE IF NOT EXISTS job_application_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			job_application_id UUID NOT NULL REFERENCES job_applications(id) ON DELETE CASCADE,
			status VARCHAR(50) NOT NULL,
			note TEXT,
			event_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_job_application_events_app ON job_application_events (job_application_id, event_date)`,
	}

	for _, m := range migrations {
		if _, err := Pool.Exec(context.Background(), m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}

	return nil
}
