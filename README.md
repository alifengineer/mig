## Simple SQL Migration Tool

Only supports for PostgreSQL

### Usage

```go

func MigratePG(dir string, tx *sql.Tx) (err error)

```

### Example

```go

package postgres

import (
    "database/sql"
    "log"

    _ "github.com/lib/pq"
    "github.com/dilmurodov/mig"
)

func StartDBConn() {
    db, err := sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        log.Fatal(err)
    }
    defer tx.Rollback()

    if err := mig.MigratePG("./migrations", tx); err != nil {
        log.Fatal(err)
    }

    if err := tx.Commit(); err != nil {
        log.Fatal(err)
    }
}

```

### Migration File

```sql

-- 20160101_000000_create_table.sql

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

```

### License

MIT

## Author

[Dilmurodov Tolibbek](github.com/dilmurodov)