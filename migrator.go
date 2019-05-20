package migrator

import (
	"database/sql"
	"fmt"
)

const tableName = "migrations"

// Migrator is the migrator implementation
type Migrator struct {
	migrations []migration
}

// New creates a new migrator instance
func New(migrations ...migration) *Migrator {
	return &Migrator{migrations: migrations}
}

// Migrate applies all available migrations
func (m *Migrator) Migrate(db *sql.DB) error {
	// create migrations table if doesn't exist
	_, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INT8 NOT NULL,
			version VARCHAR(255) NOT NULL,
			PRIMARY KEY (id)
		);
	`, tableName))
	if err != nil {
		return err
	}

	// count applied migrations
	var count int
	rows, err := db.Query(fmt.Sprintf("SELECT count(*) FROM %s", tableName))
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// plan migrations
	for idx, migration := range m.migrations[count:len(m.migrations)] {
		insertVersion := fmt.Sprintf("INSERT INTO %s (id, version) VALUES (%d, '%s')", tableName, idx+count, migration.String())
		switch m := migration.(type) {
		case *Migration:
			if err := migrate(db, insertVersion, m); err != nil {
				return fmt.Errorf("migrator: error while running migrations: %v", err)
			}
		case *MigrationNoTx:
			if err := migrateNoTx(db, insertVersion, m); err != nil {
				return fmt.Errorf("migrator: error while running migrations: %v", err)
			}
		}
	}

	return nil
}

type migration interface {
	String() string
}

// Migration represents a single migration
type Migration struct {
	Name string
	Func func(*sql.Tx) error
}

// String returns a string representation of the migration
func (m *Migration) String() string {
	return m.Name
}

// MigrationNoTx represents a single not transactional migration
type MigrationNoTx struct {
	Name string
	Func func(*sql.DB) error
}

func (m *MigrationNoTx) String() string {
	return m.Name
}

func migrate(db *sql.DB, insertVersion string, migration *Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if errRb := tx.Rollback(); errRb != nil {
				err = fmt.Errorf("error rolling back: %s\n%s", errRb, err)
			}
			return
		}
		err = tx.Commit()
	}()
	fmt.Println(fmt.Sprintf("migrator: applying migration named '%s'...", migration.Name))
	if err = migration.Func(tx); err != nil {
		return fmt.Errorf("error executing golang migration: %s", err)
	}
	if _, err = tx.Exec(insertVersion); err != nil {
		return fmt.Errorf("error updating migration versions: %s", err)
	}
	fmt.Println(fmt.Sprintf("migrator: applied migration named '%s'", migration.Name))

	return err
}

func migrateNoTx(db *sql.DB, insertVersion string, migration *MigrationNoTx) error {
	fmt.Println(fmt.Sprintf("migrator: applying no tx migration named '%s'...", migration.Name))
	if err := migration.Func(db); err != nil {
		return fmt.Errorf("error executing golang migration: %s", err)
	}
	if _, err := db.Exec(insertVersion); err != nil {
		return fmt.Errorf("error updating migration versions: %s", err)
	}
	fmt.Println(fmt.Sprintf("migrator: applied no tx migration named '%s'...", migration.Name))

	return nil
}
