// +build integration

package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	_ "github.com/lib/pq"              // postgres driver
)

func initMigrator(driverName, url string) error {
	migrator := New(
		&Migration{
			Name: "Using tx, encapsulate two queries",
			Func: func(tx *sql.Tx) error {
				if _, err := tx.Exec("CREATE TABLE foo (id INT PRIMARY KEY)"); err != nil {
					return err
				}
				if _, err := tx.Exec("INSERT INTO foo (id) VALUES (1)"); err != nil {
					return err
				}
				return nil
			},
		},
		&MigrationNoTx{
			Name: "Using db, execute one query",
			Func: func(db *sql.DB) error {
				if _, err := db.Exec("INSERT INTO foo (id) VALUES (2)"); err != nil {
					return err
				}
				return nil
			},
		},
	)

	// Migrate both steps up
	db, err := sql.Open(driverName, url)
	if err != nil {
		return err
	}
	if err := migrator.Migrate(db); err != nil {
		return err
	}

	return nil
}

func TestPostgres(t *testing.T) {
	if err := initMigrator("postgres", os.Getenv("POSTGRES_URL")); err != nil {
		t.Fatal(err)
	}
}

func TestMySQL(t *testing.T) {
	if err := initMigrator("mysql", os.Getenv("MYSQL_URL")); err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseNotFound(t *testing.T) {
	migrator := New(&Migration{})
	db, _ := sql.Open("postgres", "")
	if err := migrator.Migrate(db); err == nil {
		t.Fatal(err)
	}
}

func TestBadMigrations(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		t.Fatal(err)
	}

	var migrators = []struct {
		name  string
		input *Migrator
		want  error
	}{
		{
			name: "bad tx migration",
			input: New(&Migration{
				Name: "bad tx migration",
				Func: func(tx *sql.Tx) error {
					if _, err := tx.Exec("FAIL FAST"); err != nil {
						return err
					}
					return nil
				},
			}),
		},
		{
			name: "bad db migration",
			input: New(&MigrationNoTx{
				Name: "bad db migration",
				Func: func(db *sql.DB) error {
					if _, err := db.Exec("FAIL FAST"); err != nil {
						return err
					}
					return nil
				},
			}),
		},
	}

	for _, tt := range migrators {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New("bla")
			err.Error()
			if err := tt.input.Migrate(db); !strings.Contains(err.Error(), "pq: syntax error")  {
				t.Fatal(err)
			}
		})
	}
}
