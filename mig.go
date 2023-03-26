package mig

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Migration struct {
	Name  string
	Query string
}

func getMigrationFiles(dir string) ([]Migration, error) {

	var files []Migration

	requiredFiles := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".sql" {

			// Read file
			query, err := readFile(path)
			if err != nil {
				return err
			}

			files = append(files, Migration{
				Name:  info.Name(),
				Query: string(query),
			})
		}
		return nil
	}
	err := filepath.Walk(dir, requiredFiles)
	return files, err
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func getMigratedFiles(tx *sql.Tx) ([]Migration, error) {

	// Get all already migrated files

	migratedFiles := make([]Migration, 0)
	rows, err := tx.Query("SELECT * FROM migrations")
	if err != nil {
		return migratedFiles, err
	}

	for rows.Next() {
		var id int
		var name string
		var created_at time.Time
		err = rows.Scan(&id, &name, &created_at)
		if err != nil {
			return migratedFiles, err
		}
		migratedFiles = append(migratedFiles, Migration{
			Name:  name,
			Query: "",
		})
	}

	return migratedFiles, nil
}

func migrateUnmigratedFiles(migrationFiles []Migration, migratedFiles []Migration, tx *sql.Tx) error {

	// Migrate all files that are not migrated yet

	for _, file := range migratedFiles {
		// Check if file is already migrated
		if !contains(migratedFiles, file) {
			// Migrate file
			err := migrateFile(file, tx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func contains(migratedFiles []Migration, file Migration) bool {
	for _, migratedFile := range migratedFiles {
		if migratedFile.Name == file.Name {
			return true
		}
	}
	return false
}

func migrateFile(file Migration, tx *sql.Tx) error {

	// Migrate file
	_, err := tx.Query(file.Query)
	if err != nil {
		return err
	}

	result, err := tx.Exec("INSERT INTO migrations (name, created_at) VALUES ($1, $2)", file, time.Now())
	if err != nil {
		return err
	}
	if afftedRows, err := result.RowsAffected(); err != nil {
		return err
	} else if afftedRows == 0 {
		return errors.New("No rows affected")
	}

	return nil
}

func MigratePG(dir string, tx *sql.Tx) (err error) {

	_, err = tx.Query("CREATE TABLE IF NOT EXISTS migrations (id SERIAL PRIMARY KEY, name VARCHAR(255), created_at TIMESTAMP);")
	if err != nil {
		return err
	}

	// Get all files in the directory
	files, err := getMigrationFiles(dir)
	if err != nil {
		return err
	}

	// Get all files in the database
	migratedFiles, err := getMigratedFiles(tx)
	if err != nil {
		return err
	}

	// Migrate all files that are not in the database
	err = migrateUnmigratedFiles(files, migratedFiles, tx)
	if err != nil {
		return err
	}

	// Get latest migration

	var (
		latestMigration  string
		idLatesMigration int
	)

	err = tx.QueryRow("SELECT id, name FROM migrations ORDER BY id DESC LIMIT 1").Scan(
		&idLatesMigration,
		&latestMigration,
	)

	fmt.Fprint(os.Stdout, "Latest migration: ", latestMigration, " (", idLatesMigration, ")\n")

	return err
}
