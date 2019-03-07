// +build integration

package migrator

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq" // postgres driver
	_ "github.com/go-sql-driver/mysql" // mariadb/mysql driver
)

func TestPostgres(t *testing.T) {
	postgres, err := New("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		t.Fatal(err)
	}
	err = migrate(postgres, "$1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMariaDB(t *testing.T) {
	mysql, err := New("mysql", os.Getenv("MARIA_DB_URL"))
	if err != nil {
		t.Fatal(err)
	}
	err = migrate(mysql, "?")
	if err != nil {
		t.Fatal(err)
	}
}

func migrate(migrator *Migrator, placeholder string) error {
	// configure migrations
	var migrations []*Migration
	migrations = append(migrations,

		// migration 1, using tx, encapsulate two queries in an up migration, no down migration
		NewTxMigration("1",
			// up migration
			func(tx *sql.Tx) error {
				if _, err := tx.Exec("CREATE TABLE migrator (id INT)"); err != nil {
					return err
				}
				if _, err := tx.Exec(fmt.Sprintf("INSERT INTO migrator (id) VALUES (%s)", placeholder), 1); err != nil {
					return err
				}
				return nil
			},
			func(tx *sql.Tx) error {
				// empty down migration
				if _, err := tx.Exec("DROP TABLE migrator"); err != nil {
					return err
				}
				return nil
			},
		),

		// migration 2, using db, execute one query in an up migration, revert that query in a down migration
		NewTxMigration("2",
			// up migration
			func(tx *sql.Tx) error {
				if _, err := tx.Exec(fmt.Sprintf("INSERT INTO migrator (id) VALUES (%s)", placeholder), 2); err != nil {
					return err
				}
				return nil
			},
			// down migration
			func(tx *sql.Tx) error {
				if _, err := tx.Exec(fmt.Sprintf("DELETE FROM migrator WHERE id = %s", placeholder), 2); err != nil {
					return err
				}
				return nil
			},
		),
	)
	migrator.AddMigrations(migrations...)

	// Migrate two steps up
	for i := 1; i <= 2; i++ {
		if err := migrator.Up(); err != nil {
			return err
		}
	}

	// Migrate two steps down
	for i := 1; i <= 2; i++ {
		if err := migrator.Down(); err != nil {
			return err
		}
	}

	return nil
}
