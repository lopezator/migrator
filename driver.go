package migrator

import (
	"database/sql"
	"errors"
	"fmt"
)

// Driver is the postgres migrator.Driver implementation
type Driver struct {
	db          *sql.DB
	placeHolder string
}

// NewDriver creates a new migrator driver
func NewDriver(name, dsn string) (*Driver, error) {
	var placeHolder string
	switch name {
	case "postgres":
		placeHolder = "$1"
	case "mysql":
		placeHolder = "?"
	default:
		return nil, errors.New("driver not supported, valid values are: (postgres, mysql)")
	}
	db, err := sql.Open(name, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Driver{
		db:          db,
		placeHolder: placeHolder,
	}
	_, err = d.db.Exec("CREATE TABLE IF NOT EXISTS " + TableName + " (version varchar(255) not null primary key)")
	if err != nil {
		return nil, err
	}

	return d, nil
}

// Close is the migrator.Driver implementation of io.Closer
func (d *Driver) Close() error {
	return nil
}

// Versions lists all the applied versions
func (d *Driver) Versions() ([]string, error) {
	rows, err := d.db.Query("SELECT version FROM " + TableName + " ORDER BY version DESC")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var versions []string
	for rows.Next() {
		var version string
		err = rows.Scan(&version)
		if err != nil {
			return versions, err
		}
		versions = append(versions, version)
	}
	err = rows.Err()

	return versions, err
}

// Migrate executes a planned migration using the postgres driver
func (d *Driver) Migrate(direction Direction, migration *Migration) error {
	var insertVersion string
	if direction == Up {
		insertVersion = "INSERT INTO " + TableName + " (version) VALUES (" + d.placeHolder + ")"
	} else if direction == Down {
		insertVersion = "DELETE FROM " + TableName + " WHERE version=" + d.placeHolder
	}
	var err error
	if migration.funcTx != nil {
		var err error
		tx, err := d.db.Begin()
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
		if funcTx, ok := migration.funcTx[direction]; ok {
			if funcTx != nil {
				if err = funcTx(tx); err != nil {
					return fmt.Errorf("error executing golang migration: %s", err)
				}
			}
		}
		if _, err = tx.Exec(insertVersion, migration.id); err != nil {
			return fmt.Errorf("error updating migration versions: %s", err)
		}
	} else if migration.funcDB != nil {
		if funcDB, ok := migration.funcDB[direction]; ok {
			if funcDB != nil {
				if err := funcDB(d.db); err != nil {
					return fmt.Errorf("error executing golang migration: %s", err)
				}
			}
		}
		if _, err := d.db.Exec(insertVersion, migration.id); err != nil {
			return fmt.Errorf("error updating migration versions: %s", err)
		}
	}

	return err
}
