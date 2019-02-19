package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

// Direction type up/down
type Direction int

// String returns a string representation of the direction
func (d Direction) String() string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	}
	return ""
}

// Constants for direction
const (
	Up Direction = iota
	Down
)

// TableName is the table name to be used in the database to hold migration state
const TableName = "schema_migration"

// Migrator is the migrator implementation
type Migrator struct {
	migrations map[string]*Migration
	drv        *Driver
}

// Config holds the required migration
type Config struct {
	Driver string
	Dsn    string
}

// New creates a new migrator instance
func New(cfg *Config) (*Migrator, error) {
	drv, err := NewDriver(cfg.Driver, cfg.Dsn)
	if err != nil {
		return nil, err
	}
	return &Migrator{drv: drv, migrations: map[string]*Migration{}}, nil
}

// Migration holds one migration information
type Migration struct {
	ID     string
	Func   MigrationFuncMap
	FuncTx MigrationFuncTxMap
}

type MigrationFuncMap map[Direction]MigrationFunc
type MigrationFuncTxMap map[Direction]MigrationFuncTx

type MigrationFunc func(db *sql.DB) error
type MigrationFuncTx func(tx *sql.Tx) error

// AddMigrations adds migrations to the source.
func (m *Migrator) AddMigrations(migrations ...*Migration) {
	for _, migration := range migrations {
		m.migrations[migration.ID] = migration
	}
}

func (m *Migrator) planMigration(direction Direction, applied []string) (*Migration, error) {
	// Get last migration that was run
	sort.Strings(applied)

	// Get migrations as a slice of strings
	var migrations []string
	for _, migration := range m.migrations {
		migrations = append(migrations, migration.ID)
	}
	sort.Strings(migrations)
	count := len(applied)

	// Figure out which migration to apply
	var apply string
	if direction == Up {
		if count >= len(migrations) {
			return nil, errors.New("migrator: no more (up) migrations to apply")
		}
		apply = migrations[count]
	} else if direction == Down {
		if count == 0 {
			return nil, errors.New("migrator: no more (down) migrations to apply")
		}
		apply = migrations[count-1]
	}

	// Get migration to apply
	migration, ok := m.migrations[apply]
	if !ok {
		return nil, errors.New("migrator: migration not found")
	}
	return migration, nil
}

// Migrate runs a single migration
func (m *Migrator) Migrate(direction Direction) (int, error) {
	count := 0
	if direction != Up && direction != Down {
		return count, errors.New("direction should be either migrator.Up or migrator.Down")
	}

	// get applied migrations
	applied, err := m.drv.Versions()
	if err != nil {
		return count, err
	}

	// plan migration
	planned, err := m.planMigration(direction, applied)
	if err != nil {
		return count, err
	}

	// apply migration
	fmt.Println(fmt.Sprintf("migrator: applying migration (%s) named '%s'...", direction.String(), planned.ID))
	if err := m.drv.Migrate(direction, planned); err != nil {
		return count, fmt.Errorf("migrator: error while running migration %s (%s): %v", planned.ID, direction.String(), err)
	}
	fmt.Println(fmt.Sprintf("migrator: applied migration (%s) named '%s'", direction.String(), planned.ID))
	count++

	return count, m.drv.Close()
}
