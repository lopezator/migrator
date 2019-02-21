// +build integration

package migrator

import (
	"testing"

	txdb "github.com/DATA-DOG/go-txdb"
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
)

func TestPostgres(t *testing.T) {
	dsn := "postgres://migrator:migrator@postgres/migrator?sslmode=disable"
	txdb.Register("txdb_postgres", "postgres", dsn)
	_, err := NewDriver("txdb_postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMySQL(t *testing.T) {
	dsn := "migrator:migrator@tcp(mysql)/migrator"
	txdb.Register("txdb_mysql", "mysql", dsn)
	_, err := NewDriver("txdb_mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
}

