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
	// Read migration file
	migrationPath := filepath.Join("migrations", "001_init_schema.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("error reading migration file: %w", err)
	}

	// Execute migration
	if _, err := d.DB.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("error executing migration: %w", err)
	}

	log.Println("Migrations executed successfully")
	return nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
