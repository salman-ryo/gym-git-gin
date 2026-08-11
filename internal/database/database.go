package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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

// RunMigrations executes embedded DDL SQL migrations in sorted alphabetical order
func RunMigrations(db *sql.DB) error {
	entries, err := migrations.UpMigrationsFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed reading embedded migrations directory: %w", err)
	}

	// Iterate through migration files in sorted order (ReadDir returns them sorted by filename)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".up.sql") {
			continue
		}

		log.Printf("Executing database migration: %s", filename)
		content, err := migrations.UpMigrationsFS.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed reading migration file %s: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed executing migration script %s: %w", filename, err)
		}
	}

	log.Println("Database auto-migrations executed successfully (all tables & seed data verified)")
	return nil
}

