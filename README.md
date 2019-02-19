# migrator

Golang migrations made easy

Support for PostgreSQL ([lib/pq](https://github.com/lib/pq)) 
and MySQL ([go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)) 

# Usage

## PostgreSQL


```go
package main

import (
    "database/sql"
	"log"
    
    _ "github.com/lib/pq" // postgres driver
    "github.com/lopezator/migrator"
)

func main() {
	// Configure the migrator
    cfg, err := migrator.NewConfig(migrator.Up, 1)
    if err != nil {
        log.Fatal(err)
    }

    // Create migration driver
    drv, err := migrator.New("postgres", "")
    if err != nil {
        log.Fatal(err)
    }

    // Add migrations
    src := migrator.NewSource()
    src.AddMigrations(&migrator.Migration{
    	ID: "0_create_table",
    	Direction: migrator.Up,
    	FuncDb: func(db *sql.DB) error {
    	    _, err := db.Exec("CREATE TABLE test")
    	    return err
    	},
    })

    // Run the migration
    _, err = migrator.Migrate(cfg, drv, src)
    if err != nil {
    	log.Fatal(err)
    }
}
```      

