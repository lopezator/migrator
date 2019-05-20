// +build integration

package migrator

import (
	"database/sql"
	"fmt"
	"os"
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
	fmt.Println("Testing postgres...")
	if err := initMigrator("postgres", os.Getenv("POSTGRES_URL")); err != nil {
		t.Fatal(err)
	}
}

func TestMySQL(t *testing.T) {
	fmt.Println("Testing mysql...")
	if err := initMigrator("mysql", os.Getenv("MYSQL_URL")); err != nil {
		t.Fatal(err)
	}
}
