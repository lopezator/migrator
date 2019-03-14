[![CircleCI](https://circleci.com/gh/lopezator/migrator.svg?style=svg)](https://circleci.com/gh/lopezator/migrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/lopezator/migrator)](https://goreportcard.com/report/github.com/lopezator/migrator)
[![GoDoc](https://godoc.org/github.com/lopezator/migrator/go?status.svg)](https://godoc.org/github.com/lopezator/migrator)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
# migrator

Golang migrations made easy.

Disclaimer: migrator is at a very early stage of development, use it at your own risk.

# Features

* Simple code.
* Usage as a library, embeddable and extensible on your behalf. 
* Support of any database supported by `database/sql`.
* GO code migrations, either transactional or transaction-less, using `*sql.DB` (`migrator.NewDBMigration`) or 
`*sql.Tx` (`migrator.NewTXMigration`).
* No need to use `packr`, `gobin` or others, since all migrations are just GO code.

# Compatibility

Although any database supported by [database/sql](https://golang.org/pkg/database/sql/) and one of its recommended 
drivers [SQLDrivers](https://github.com/golang/go/wiki/SQLDrivers) should work ok, at the moment only `PostgreSQL` and
`MySQL` are being explicitly tested.

If you find any issue with any of the databases included under the umbrella of `database/sql`, feel free to 
[contribute](#Contribute) by opening an issue or sending a pull request.

# Usage

The following example assume:

- A working `postgres` DB conn on localhost, with a user named `postgres`, empty password, and db named `migrator`.

Customize this to your needs by changing the driver and/or connection settings.

### QuickStart:

```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // postgres driver
	"github.com/lopezator/migrator"
)

func main() { 
    m, err := migrator.New("postgres", "postgres://postgres@localhost/migrator?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    m.AddMigrations( // single migration up, nil down migration
        migrator.NewDBMigration("1",
        func(db *sql.DB) error {
           if _, err := db.Exec("CREATE TABLE migrator (id INT)"); err != nil {
               return err
           }
           return nil
        }, nil,
    ))
    if err := m.Up(); err != nil {
        log.Fatal(err)
    }
}
```

### Full explained verbose example:

```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // postgres driver
	"github.com/lopezator/migrator"
)

func main() {
	// Initialize migrator with the chosen driver/dsn.
	m, err := migrator.New("postgres", "postgres://postgres@localhost/migrator?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Define migration 1 up logic, transactional.
	migration1UpFunc := func(tx *sql.Tx) error {
		if _, err := tx.Exec("CREATE TABLE migrator (id INT)"); err != nil {
			return err
		}
		return nil
	}
	// Define migration 1 down logic, transactional.
	migration1DownFunc := func(tx *sql.Tx) error {
		if _, err := tx.Exec("DROP TABLE migrator"); err != nil {
			return err
		}
		return nil
	}
	// Create the migration and pass-in the two functions defined above.
	migration1 := migrator.NewTxMigration("1_version", migration1UpFunc, migration1DownFunc)

	// Define migration 2 up, transaction-less.
	migration2UpFunc := func(db *sql.DB) error {
		if _, err := db.Exec("INSERT INTO migrator (id) VALUES ($1)", 1); err != nil {
			return err
		}
		return nil
	}
	// Create the migration and pass-in the up function defined above, no down function provided (nil).
	migration2 := migrator.NewDBMigration("2_version", migration2UpFunc, nil)

	// Add both migrations to the migrator.
	m.AddMigrations(migration1, migration2)

	// Migrate up step 1.
	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - An empty table named `migrator`.
	// - A table named `schema_migrations` with a single row: `1_version`.

	// Migrate up step 2
	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - A table named `migrator` with a single row: `1`.
	// - A table named `schema_migrations` with two rows, versions: `1_version` and `2_version`.

	// Migrate down step 2 (esterile in terms of data, remember, no down migration defined on `migration2`).
	err = m.Down()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - A tabled named `migrator` with a single row: `1`.
	// - schema_migrations table with a single row, version: `1_version`.

	// Migrate down step 2
	err = m.Down()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - A tabled named `schema_migrations` with no rows.
}
```

Notes on examples above: 

- Is your responsibility to provide any sortable list of IDs to migrator if you want a correct behavior from the 
library: ints, unix timestamps, ULIDs...
- Migrator creates/manages a table named `schema_migrations` to keep the track of the applied versions.
- This example follows a simplistic/verbose approach for demonstration purposes. You can pass directly the functions
without the need of saving them to variables, organize your migrations into single/multiple files... for example inside
a `migrations` folder. As flexible as you want. 

### Looking for more examples?

Just examine [migrator_test.go](migrator_test.go) file.

# Motivation

Why another migration library?

* Lightweight dummy implementation with just `database/sql` support. Migrator doesn't need of any ORM or other heavy 
libraries as a dependency. It's just made from two files!
* Easily embedabble into your application, no need to install/use a separate binary.
* Supports GO migrations, either transactional or transaction-less.
* Flexible usage.

# Are not migrator objectives

* Add support to databases outside `database/sql`.
* Complicate the code/logic to add functionality that could be accomplished easily on userland, like up-to-n, 
down-to-n, view current version, etc.
* Add a bunch of dependencies just to provide a CLI/standalone functionality.

# Comparison against other tools

* [rubenv/sql-migrate](https://github.com/rubenv/sql-migrate) doesn't support GO migrations. Sometimes you need GO code 
to accomplish complex tasks that couldn't be done using just SQL.

* [Boostport/migration](https://github.com/Boostport/migration) is a nice tool with support for many databases. 
Migrator code is inspired on its codebase. It supports both GO and SQL migrations. Unfortunately, when using GO 
migrations you have to write your own logic to retrieve and update version info in the database. Additionally I didn't 
found a nice way to encapsulate both migration and version logic queries inside the same transaction. 

* [golang-migrate/migrate](https://github.com/golang-migrate/migrate) doesn't support GO migrations. Sometimes you need 
GO code to accomplish complex tasks that couldn't be done using just SQL. Additionally it feels a little heavy for the
task.

* [pressly/goose](https://github.com/pressly/goose) it supports both GO and SQL migrations. 
Unfortunately it doesn't support transaction-less GO migrations. Sometimes using transactions is either not possible 
with the combination of queries you need in a single migration, and others could be very slow and you simple don't need 
them for that specific case. It's also pretty big, with internals difficult to follow. It's crowded with a lot of
functionality that could be done in userland pretty fast.

# Contribute

Pull requests are welcome, this is an early implementation and work is needed in all areas: docs, examples, tests, ci... 

The easiest way to contribute is by installing [docker](https://docs.docker.com/install/) and 
[docker-compose](https://docs.docker.com/compose/install/), and ensure you comply with code standards and pass all the 
tests before submitting a PR by running:

```bash
$> docker-compose up -d --build
$> docker-compose exec migrator make prepare
$> docker-compose exec migrator make sanity-check
$> docker-compose exec migrator make test
$> docker-compose down
```