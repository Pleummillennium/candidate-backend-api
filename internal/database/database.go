package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(databaseURL string) (*Database, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Successfully connected to database")

	return &Database{DB: db}, nil
}

func (d *Database) RunMigrations() error {
	// Get all migration files
	files, err := filepath.Glob(filepath.Join("migrations", "*.sql"))
	if err != nil {
		return fmt.Errorf("error reading migration directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found")
	}

	// Execute each migration file in order
	for _, file := range files {
		log.Printf("Running migration: %s", filepath.Base(file))

		migrationSQL, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", file, err)
		}

		if _, err := d.DB.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("error executing migration %s: %w", file, err)
		}
	}

	log.Println("All migrations executed successfully")
	return nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
