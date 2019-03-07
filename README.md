[![CircleCI](https://circleci.com/gh/lopezator/migrator.svg?style=svg)](https://circleci.com/gh/lopezator/migrator)

# migrator

Golang migrations made easy

# Features

* Simple code and usage as a library, embeddable and extensible on your behalf. 
* Any database that is supported by GO's `database/sql` should work good.
* GO code migrations, based transactional or transaction-less, using `*sql.DB` or `*sql.Tx`.
* No need to use `packr`, `gobin` or others, since all migrations are just GO code.

# Usage

The following example assumes:

- A working conn on localhost, with postgres user, no password, and an EMPTY db named migrator, customize this to your 
needs.
	
```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	migratorpkg "github.com/lopezator/migrator"
)

func main() {
	// Initialize migrator with the chosen driver/dsn
	migrator, err := migratorpkg.New("postgres", "postgres://postgres@localhost/migrator?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Migration 1 (up & down, with tx)
	migration1UpFunc := func(tx *sql.Tx) error {
		if _, err := tx.Exec("CREATE TABLE migrator (id INT)"); err != nil {
			return err
		}
		return nil
	}
	migration1DownFunc := func(tx *sql.Tx) error {
		if _, err := tx.Exec("DROP TABLE migrator"); err != nil {
			return err
		}
		return nil
	}
	migration1 := migratorpkg.NewTxMigration("1", migration1UpFunc, migration1DownFunc)

	// Migration 2 (up without tx, no down migration)
	migration2Up := func(db *sql.DB) error {
		if _, err := db.Exec("INSERT INTO migrator (id) VALUES ($1)", 1); err != nil {
			return err
		}
		return nil
	}
	migration2 := migratorpkg.NewDBMigration("2", migration2Up, nil)

	// Add migrations to migrator
	migrator.AddMigrations(migration1, migration2)

	// Migrate up step 1
	err = migrator.Up()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - migrator empty table
	// - schema_migrations table with a single row, version: 1

	// Migrate up step 2 (insert a row into the migrator table, upgrade to version 2)
	err = migrator.Up()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - migrator table with a single row, 1
	// - schema_migrations table with two rows, versions: 1 and 2

	// Migrate down step 2 (esterile in terms of data, downgrade version 1)
	err = migrator.Down()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - migrator table with a single row, 1
	// - schema_migrations table with a single rows, versions: 1

	// Migrate down step 2 (drop the migrator table, downgrade to initial state - no version)
	err = migrator.Down()
	if err != nil {
		log.Fatal(err)
	}
	// Actual contents of the db:
	// - schema_migrations empty table
}
```

Notes on example code: 

- Is your responsibility to the set any sortable list of IDs if you want a correct behavior of the library: numeric, 
unix timestamps, ULIDs...
- Migrator library creates/manages a table named `schema_migrations` to keep the track of the applied versions.
- This example follows a simplistic/verbose approach for demonstration purposes, you can pass directly the functions
without the need of saving them to variables, organize your migrations into single/multiple files, for example inside
a `migrations` folder. As flexible as you want. 

# Motivation

Why another migration library?

* Lightweight dummy implementation with just `database/sql` support, without ORM or heavy library 
dependency/boilerplate.
* Easily embedabble on your application, no need to install/use a separate binary.
* Support of GO migrations, either transactional or transactionless.
* Flexible usage.

# Are not migrator objectives

* For now, to add support to databases outside `database/sql`.
* Complicate the code/logic to add functionality that could be accomplished easily on userland, like a up-to-n, 
down-to-n.
* Add a bunch of dependencies just to provide a CLI/standalone functionality, unless a lot of people see it useful and 
wants to.

# Comparison against other tools

* [rubenv/sql-migrate](https://github.com/rubenv/sql-migrate) doesn't support GO migrations. Sometimes you need GO code 
to accomplish complex tasks that couldn't be done using just SQL.

* [Boostport/migration](https://github.com/Boostport/migration) is a nice tool with support for many libraries. Migrator
code is heavily based on its codebase. It supports both GO and SQL migrations. Unfortunately, when using GO migrations 
you have to write your own logic to retrieve and update version info in the database. Additionally I didn't found a 
nice way to encapsulate both migration and version logic queries inside the same transaction. 

* [golang-migrate/migrate](https://github.com/golang-migrate/migrate) doesn't support GO migrations. Sometimes you need 
GO code to accomplish complex tasks that couldn't be done using just SQL. Additionally it feels a little heavy for the
task.

* [https://github.com/pressly/goose](https://github.com/pressly/goose) it supports both GO and SQL migrations. 
Unfortunately it doesn't support transactionless GO migrations. Sometimes using transactions is either not possible with
the combination of queries you need, and others could be very slow, and you simple don't need them for that specific
migration. It's also pretty big, with internals difficult to follow, too much with things I wasn't going to use or that
could be done in userland fair more easily.

# Contribute

PRs are welcome, this is an early implementation and work is needed in many areas: docs, examples, tests, ci... 

The easiest way to contribute is to use to install golangci-lint and ensure your code comply with linter rules.

```bash
make setup-env
make sanity-check
```
