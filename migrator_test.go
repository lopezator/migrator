// +build integration

package migrator

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	_ "github.com/lib/pq"              // postgres driver
)

const testTableName = "testMigrationsTable"

func migrateTest(driverName, url string) error {
	migrator, err := New(
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
		&Migration{
			Name: "Using tx, one embedded query",
			Func: func(tx *sql.Tx) error {
				query, err := _escFSString(false, "/testdata/0_bar.sql")
				if err != nil {
					return err
				}
				if _, err := tx.Exec(query); err != nil {
					return err
				}
				return nil
			},
		},
	)
	if err != nil {
		return err
	}

	// Migrate up
	db, err := sql.Open(driverName, url)
	if err != nil {
		return err
	}
	if err := migrator.Migrate(db); err != nil {
		return err
	}

	return nil
}
func migrateNamedTest(driverName, url string) error {
	migrator := NewNamed(testTableName,
		&Migration{
			Name: "Using tx, encapsulate two queries",
			Func: func(tx *sql.Tx) error {
				if _, err := tx.Exec("CREATE TABLE named_foo (id INT PRIMARY KEY)"); err != nil {
					return err
				}
				if _, err := tx.Exec("INSERT INTO named_foo (id) VALUES (1)"); err != nil {
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

func mustMigrator(migrator *Migrator, err error) *Migrator {
	if err != nil {
		panic(err)
	}
	return migrator
}

func TestPostgres(t *testing.T) {
	if err := migrateTest("postgres", os.Getenv("POSTGRES_URL")); err != nil {
		t.Fatal(err)
	}
	if err := migrateNamedTest("postgres", os.Getenv("POSTGRES_URL")); err != nil {
		t.Fatal(err)
	}
}

func TestMySQL(t *testing.T) {
	if err := migrateTest("mysql", os.Getenv("MYSQL_URL")); err != nil {
		t.Fatal(err)
	}
	if err := migrateNamedTest("mysql", os.Getenv("MYSQL_URL")); err != nil {
		t.Fatal(err)
	}
}
func TestMigrationNumber(t *testing.T) {
	testMigrationNumber(t, defaultTableName, 3)
	testMigrationNumber(t, testTableName, 1)
}

func testMigrationNumber(t *testing.T, tableName string, migrationsCount int) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		t.Fatal(err)
	}
	count, err := countApplied(db, tableName)
	if err != nil {
		t.Fatal(err)
	}
	if count != migrationsCount {
		t.Fatalf("db applied migration number should be %d", migrationsCount)
	}
}

func TestDatabaseNotFound(t *testing.T) {
	migrator, err := New(&Migration{})
	if err != nil {
		t.Fatal(err)
	}
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
	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", defaultTableName))
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
			input: mustMigrator(New(&Migration{
				Name: "bad tx migration",
				Func: func(tx *sql.Tx) error {
					if _, err := tx.Exec("FAIL FAST"); err != nil {
						return err
					}
					return nil
				},
			})),
		},
		{
			name: "bad db migration",
			input: mustMigrator(New(&MigrationNoTx{
				Name: "bad db migration",
				Func: func(db *sql.DB) error {
					if _, err := db.Exec("FAIL FAST"); err != nil {
						return err
					}
					return nil
				},
			})),
		},
	}

	for _, tt := range migrators {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Migrate(db)
			if err != nil && !strings.Contains(err.Error(), "pq: syntax error") {
				t.Fatal(err)
			}
		})
	}
}

func TestBadMigrate(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_URL"))
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate(db, "BAD INSERT VERSION", &Migration{Name: "bad insert version", Func: func(tx *sql.Tx) error {
		return nil
	}}); err == nil {
		t.Fatal("BAD INSERT VERSION should fail!")
	}
}

func TestBadMigrateNoTx(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		t.Fatal(err)
	}
	if err := migrateNoTx(db, "BAD INSERT VERSION", &MigrationNoTx{Name: "bad migrate no tx", Func: func(db *sql.DB) error {
		return nil
	}}); err == nil {
		t.Fatal("BAD INSERT VERSION should fail!")
	}
}

func TestBadMigrationNumber(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_URL"))
	if err != nil {
		t.Fatal(err)
	}
	migrator := mustMigrator(New(
		&Migration{
			Name: "bad migration number",
			Func: func(tx *sql.Tx) error {
				if _, err := tx.Exec("CREATE TABLE bar (id INT PRIMARY KEY)"); err != nil {
					return err
				}
				return nil
			},
		},
	))
	if err := migrator.Migrate(db); err == nil {
		t.Fatalf("BAD MIGRATION NUMBER should fail: %v", err)
	}
}

func TestPending(t *testing.T) {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		t.Fatal(err)
	}
	migrator, _ := New()
	pending := migrator.Pending(db)
	if len(pending) != 0 {
		t.Fatalf("pending migrations should be 0")
	}
	migrator = mustMigrator(New(
		&Migration{
			Name: "Using tx, create baz table",
			Func: func(tx *sql.Tx) error {
				if _, err := tx.Exec("CREATE TABLE baz (id INT PRIMARY KEY)"); err != nil {
					return err
				}
				return nil
			},
		},
	))
	pending = migrator.Pending(db)
	if len(pending) != 1 {
		t.Fatalf("pending migrations should be 1")
	}
}
