package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gymgit/backend/migrations"

	_ "github.com/lib/pq"
)

// ConnectDB establishes and configures a PostgreSQL connection pool and runs auto-migrations
func ConnectDB(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is empty")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("Warning: Database ping failed: %v", err)
		return db, err
	}

	log.Println("Successfully connected to the database")

	// Automatically run migrations on startup if tables/seeds do not exist
	if err := RunMigrations(db); err != nil {
		log.Printf("Warning: Auto-migration failed: %v", err)
	}

	return db, nil
}

// RunMigrations executes embedded DDL SQL migrations idempotently
func RunMigrations(db *sql.DB) error {
	if migrations.MigrationSQL == "" {
		return fmt.Errorf("migration SQL payload is empty")
	}

	if _, err := db.Exec(migrations.MigrationSQL); err != nil {
		return fmt.Errorf("failed executing database migrations: %w", err)
	}

	log.Println("Database auto-migrations executed successfully (tables & seed data verified)")
	return nil
}
