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

// New creates a new migrator instance
func New(driver, dsn string) (*Migrator, error) {
	drv, err := NewDriver(driver, dsn)
	if err != nil {
		return nil, err
	}
	return &Migrator{drv: drv, migrations: map[string]*Migration{}}, nil
}

// Migration holds one migration information
type Migration struct {
	id     string
	funcDB funcDBMap
	funcTx funcTxMap
}

// Func map types
type funcDBMap map[Direction]funcDB
type funcTxMap map[Direction]funcTx

// Func types
type funcDB func(db *sql.DB) error
type funcTx func(tx *sql.Tx) error

// AddMigrations adds migrations to the source.
func (m *Migrator) AddMigrations(migrations ...*Migration) {
	for _, migration := range migrations {
		m.migrations[migration.id] = migration
	}
}

// NewDBMigration creates a new db migration
func NewDBMigration(id string, funcUp, funcDown func(db *sql.DB) error) *Migration {
	return &Migration{id: id, funcDB: funcDBMap{Up: funcUp, Down: funcDown}}
}

// NewTxMigration instantiates a new tx migration
func NewTxMigration(id string, funcUp, funcDown func(tx *sql.Tx) error) *Migration {
	return &Migration{id: id, funcTx: funcTxMap{Up: funcUp, Down: funcDown}}
}

// Up migrates 1 step up
func (m *Migrator) Up() error {
	if err := m.migrate(Up); err != nil {
		return err
	}
	return nil
}

// Down migrates 1 step down
func (m *Migrator) Down() error {
	if err := m.migrate(Down); err != nil {
		return err
	}
	return nil
}

// Migrate runs a single migration
func (m *Migrator) migrate(direction Direction) error {
	// get applied migrations
	applied, err := m.drv.Versions()
	if err != nil {
		return err
	}

	// plan migration
	planned, err := m.planMigration(direction, applied)
	if err != nil {
		return err
	}

	// apply migration
	fmt.Println(fmt.Sprintf("migrator: applying migration (%s) named '%s'...", direction.String(), planned.id))
	if err := m.drv.Migrate(direction, planned); err != nil {
		return fmt.Errorf("migrator: error while running migration %s (%s): %v", planned.id, direction.String(), err)
	}
	fmt.Println(fmt.Sprintf("migrator: applied migration (%s) named '%s'", direction.String(), planned.id))

	return m.drv.Close()
}

func (m *Migrator) planMigration(direction Direction, applied []string) (*Migration, error) {
	// Get last migration that was run
	sort.Strings(applied)

	// Get migrations as a slice of strings
	var migrations []string
	for _, migration := range m.migrations {
		migrations = append(migrations, migration.id)
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
