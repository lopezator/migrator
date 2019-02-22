// +build integration

package migrator

import (
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func TestDriver(t *testing.T) {
	txdb.Register("txdb", "postgres", os.Getenv("MIGRATOR_DB_DSN"))
	drv, err := NewDriver("txdb", os.Getenv("MIGRATOR_DB_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	err = drv.Migrate(Up, NewTxMigration("1", func(tx *sql.Tx) error {
		if _, err := tx.Exec("CREATE TABLE migrator (id INT)"); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO migrator (id) VALUES ($1)", 1); err != nil {
			return err
		}
		return nil
	}, nil))
	if err != nil {
		t.Fatal(err)
	}
	got, err := drv.Versions()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"1"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("versions got %s, expected %s", got, expected)
	}
}