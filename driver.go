package migrator

import (
	"database/sql"
	"fmt"
)

// Driver is the postgres migrator.Driver implementation
type Driver struct {
	db *sql.DB
}

// New creates a new postgres migrator driver
func New(driver, dsn string) (*Driver, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Driver{
		db: db,
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
func (d *Driver) Migrate(migration *Migration) error {
	var insertVersion string
	if migration.Direction == Up {
		insertVersion = "INSERT INTO " + TableName + " (version) VALUES ($1)"
	} else if migration.Direction == Down {
		insertVersion = "DELETE FROM " + TableName + " WHERE version=$1"
	}

	if migration.FuncTx != nil {
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
		err = migration.FuncTx(tx)
		if err != nil {
			return fmt.Errorf("error executing golang migration: %s", err)
		}
		if _, err = tx.Exec(insertVersion, migration.ID); err != nil {
			return fmt.Errorf("error updating migration versions: %s", err)
		}
	} else {
		err := migration.FuncDb(d.db)
		if err != nil {
			return fmt.Errorf("error executing golang migration: %s", err)
		}
		if _, err = d.db.Exec(insertVersion, migration.ID); err != nil {
			return fmt.Errorf("error updating migration versions: %s", err)
		}
	}

	return nil
}
