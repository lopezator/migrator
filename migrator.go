package migrator

import (
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"
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

// Source holds the defined migrations
type Source struct {
	sync.Mutex
	migrations map[string]*Migration
}

// Migration holds one migration information
type Migration struct {
	ID        string
	Direction Direction
	FuncDb    func(db *sql.DB) error
	FuncTx    func(tx *sql.Tx) error
}

// NewSource creates a source for storing Go functions as migrations.
func NewSource() *Source {
	return &Source{
		migrations: map[string]*Migration{},
	}
}

// AddMigrations adds migrations to the source.
// The file parameter follows the following convention: <number>_<name>
// Examples: 1_email_expand, 2_account_expand, 3_email_contract
func (s *Source) AddMigrations(migrations ...*Migration) {
	for _, migration := range migrations {
		func() {
			s.Lock()
			defer s.Unlock()

			if migration.Direction == Up {
				migration.ID += ".up.go"
			} else if migration.Direction == Down {
				migration.ID += ".down.go"
			}

			s.migrations[migration.ID] = migration
		}()
	}
}

// GetMigration gets a golang migration
func (s *Source) GetMigration(id string) *Migration {
	s.Lock()
	defer s.Unlock()

	return s.migrations[id]
}

// TableName is the table name to be used in the database to hold migration state
const TableName = "schema_migration"

// Config holds the required migration
type Config struct {
	direction Direction
	max       int
}

// NewConfig creates a migrator.Config
func NewConfig(direction Direction, max int) (*Config, error) {
	if direction != Up && direction != Down {
		return nil, errors.New("direction should be either migrator.Up or migrator.Down")
	}
	return &Config{
		direction: direction,
		max:       max,
	}, nil
}

func planMigrations(cfg *Config, src *Source, applied []string) ([]*Migration, error) {
	// Get last migration that was run
	sort.Strings(applied)
	var latest string
	if len(applied) > 0 {
		latest = applied[len(applied)-1]
	}

	// Get migration as a slice of string
	var migrations []string
	for _, migration := range src.migrations {
		migrations = append(migrations, migration.ID)
	}
	sort.Strings(migrations)

	// Figure out which migrations to apply
	var result []*Migration
	var index = -1
	if latest != "" {
		for index < len(migrations)-1 {
			index++
			if migrations[index] == latest {
				break
			}
		}
	}
	var apply []string
	if cfg.direction == Up {
		apply = migrations[index+1:]
	} else if cfg.direction == Down && index != -1 {
		// Add in reverse order
		apply = make([]string, index+1)
		for i := 0; i < index+1; i++ {
			apply[index-i] = migrations[i]
		}
	}
	count := len(apply)
	if cfg.max > 0 && cfg.max < count {
		count = cfg.max
	}

	// Generate slice of migrations from slice of IDs
	for _, ID := range apply[0:count] {
		migration := src.GetMigration(ID)
		result = append(result, &Migration{
			ID:        migration.ID,
			Direction: migration.Direction,
			FuncDb:    migration.FuncDb,
			FuncTx:    migration.FuncTx,
		})
	}

	return result, nil
}

// Migrate runs a migration using a given driver and Source. The direction defines whether
// the migration is up or down, and max is the maximum number of migrations to apply. If max is set to 0,
// then there is no limit on the number of migrations to apply.
func Migrate(cfg *Config, drv *Driver, src *Source) (int, error) {
	count := 0
	applied, err := drv.Versions()
	if err != nil {
		return count, err
	}
	planned, err := planMigrations(cfg, src, applied)
	if err != nil {
		return count, err
	}
	for _, migration := range planned {
		fmt.Println(fmt.Sprintf("migrator: applying migration (%s) named '%s'...", cfg.direction.String(), migration.ID))

		err = drv.Migrate(migration)
		if err != nil {
			errorMessage := "migrator: error while running migration " + migration.ID

			if migration.Direction == Up {
				errorMessage += " (up)"
			} else {
				errorMessage += " (down)"
			}
			return count, fmt.Errorf(errorMessage+": %s", err)
		}

		fmt.Println(fmt.Sprintf("migrator: applied migration (%s) named '%s'", cfg.direction.String(), migration.ID))
		count++
	}
	if len(planned) == 0 {
		fmt.Println("migrator: No more migrations to apply")
	}
	err = drv.Close()

	return count, err
}
