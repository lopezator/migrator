<p align="center"><img src="https://github.com/lopezator/migrator/blob/master/logo.png" width="360"></p>
<p align="center">
    <a href="https://cloud.drone.io/lopezator/migrator"><img src="https://cloud.drone.io/api/badges/lopezator/migrator/status.svg?branch=master" alt="DroneCI" /></a>
    <a href="https://goreportcard.com/report/github.com/lopezator/migrator"><img src="https://goreportcard.com/badge/github.com/lopezator/migrator" alt="Go Report Card" /></a>
    <a href="https://codecov.io/gh/lopezator/migrator"><img src="https://codecov.io/gh/lopezator/migrator/branch/master/graph/badge.svg" alt="codecov" /></a>
    <a href="https://github.com/avelino/awesome-go#database"><img src="https://awesome.re/mentioned-badge.svg" alt="DroneCI" /></a>
    <a href="https://godoc.org/github.com/lopezator/migrator"><img src="https://godoc.org/github.com/lopezator/migrator/go?status.svg" alt="GoDoc" /></a>
    <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0" /></a>
</p>

# migrator

Dead simple Go database migration library.

# Features

* Simple code
* Usage as a library, embeddable and extensible on your behalf
* Support of any database supported by `database/sql`
* Go code migrations, either transactional or transaction-less, using `*sql.Tx` (`migrator.Migration`) or `*sql.DB` (`migrator.MigrationNoTx`)
* No need to use `packr`, `gobin` or others, since all migrations are just Go code

# Compatibility

Although any database supported by [database/sql](https://golang.org/pkg/database/sql/) and one of its recommended drivers [SQLDrivers](https://github.com/golang/go/wiki/SQLDrivers) should work OK, at the moment only `PostgreSQL` and `MySQL` are being explicitly tested.

If you find any issues with any of the databases included under the umbrella of `database/sql`, feel free to [contribute](#Contribute) by opening an issue or sending a pull request.

# Usage

The following example assumes:

- A working `postgres` DB conn on localhost, with a user named `postgres`, empty password, and db named `foo`

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
    m, err := migrator.New(
        migrator.Migrations(
            &migrator.Migration{
                Name: "Create table foo",
                Func: func(tx *sql.Tx) error {
                    if _, err := tx.Exec("CREATE TABLE foo (id INT PRIMARY KEY)"); err != nil {
                        return err
                    }
                    return nil
                },
            },
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
   
    // Migrate up
    db, err := sql.Open("postgres", "postgres://postgres@localhost/foo?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    if err := m.Migrate(db); err != nil {
        log.Fatal(err)
    }
}
```

Notes on examples above:

- Migrator creates/manages a table named `migrations` to keep track of the applied versions. However, if you want to customize the table name `migrator.TableName("my_migrations")` can be passed to `migrator.New` function as an additional option. 

### Logging

By default, migrator prints applying/applied migration info to stdout. 
If that's enough for you, you can skip this section.

If you need some special formatting or want to use a 3rd party logging library, this could be done by using `WithLogger` option as follows:

```go
logger := migrator.WithLogger(migrator.LoggerFunc(func(msg string, args ...interface{}) {
	// Your code here 
})))
```

Then you will only need to pass the logger as an option to `migrator.New`. 

### Looking for more examples?

Just examine the [migrator_test.go](migrator_test.go) file.

### But I don't want to write complex migrations in strings! 😥

You still can use your favorite embedding tool to write your migrations inside `.sql` files and load them into migrator!

I provide a simple example using [esc](https://github.com/mjibson/esc) on the `Using tx, one embedded query` test here: [migrator_test](https://github.com/lopezator/migrator/blob/master/migrator_test.go)

### Erm... Where are the ID's of the migrations to know their order? 🤔

In order to avoid problems with different identifiers, ID collisions, etc... the order of the migrations is just the order being passed to the migrator.

### Wait... no down migrations? 😱

Adding the functionality to reverse a migration introduces complexity to the API, the code, and the risk of losing the synchrony between the defined list of migrations and current state of the database. In addition to this, depending on the case, not all the migrations are easily reversible, or they cannot be reversed.

We also think that it's a good idea to follow an "append-only" philosophy when coming to database migrations, so correcting a defective migration comes in the form of adding a new migration instead of reversing it.

e.g. After a `CREATE TABLE foo` we'll simply add a new `DROP TABLE foo` instead of reverting the first migration, so both states are reflected both in the code and the database.  

### Caveats

- The name of the migrations must be SQL-safe for your engine of choice. Avoiding conflicting characters like `'` is recommended, otherwise, you will have to escape them by yourself e.g. `''` for PostgreSQL and `\'` for MySQL.

# Motivation

Why another migration library?

* Lightweight dummy implementation with just `database/sql` support. Migrator doesn't need any ORM or other heavy libraries as a dependency. It's just made from a [single file](migrator.go) in less than 200 lines of code!
* Easily embedabble into your application, no need to install/use a separate binary
* Supports Go migrations, either transactional or transaction-less
* Flexible usage

# These are not migrator objectives

* Add support to databases outside `database/sql`
* Complicate the code/logic to add functionality that could be accomplished easily on userland, like view current version, list of applied versions, etc.
* Add a bunch of dependencies just to provide a CLI/standalone functionality

# Comparison with other tools

* [rubenv/sql-migrate](https://github.com/rubenv/sql-migrate) doesn't support Go migrations. Sometimes you need Go code to accomplish complex tasks that can't be done using just SQL.

* [Boostport/migration](https://github.com/Boostport/migration) is a nice tool with support for many databases. Migrator code is inspired by its codebase. It supports both Go and SQL migrations. Unfortunately, when using Go migrations you have to write your own logic to retrieve and update version info in the database. Additionally I didn't find a nice way to encapsulate both migration and version logic queries inside the same transaction.

* [golang-migrate/migrate](https://github.com/golang-migrate/migrate) doesn't support Go migrations. Sometimes you need Go code to accomplish complex tasks that couldn't be done using just SQL. Additionally it feels a little heavy for the task.

* [pressly/goose](https://github.com/pressly/goose) supports both Go and SQL migrations. Unfortunately it doesn't support transaction-less Go migrations. Sometimes using transactions is either not possible with the combination of queries you need in a single migration, or others could be very slow and you simply don't need them for that specific case. It's also pretty big, with internals that are difficult to follow. It's crowded with a lot of functionality that could be done in userland pretty fast.

# Contribute

Pull requests are welcome, this is an early implementation and work is needed in all areas: docs, examples, tests, ci...

The easiest way to contribute is by installing [docker](https://docs.docker.com/install/) and [docker-compose](https://docs.docker.com/compose/install/), and ensure you comply with code standards and pass all the tests before submitting a PR by running:

```bash
$> docker-compose up -d --build
$> docker-compose exec migrator make prepare
$> docker-compose exec migrator make sanity-check
$> docker-compose exec migrator make test
$> docker-compose down
```

## Logo

The logo was taken from @ashleymcnamara's [gophers repo](https://github.com/ashleymcnamara/gophers). I've just applied slight modifications to it.
