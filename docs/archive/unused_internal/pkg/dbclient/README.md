# Database Client Package

This package provides a client for interacting with the database-service, making it easy for services to manage their databases and migrations.

## Features

- **Database Management**: Create, configure, and manage databases through the database-service
- **Migrations**: Run database migrations using the standard `golang-migrate` library
- **Connection Pooling**: Built-in connection pooling and health checks
- **Automatic Cleanup**: Clean up resources when your service shuts down
- **Support for Multiple Databases**: PostgreSQL and MySQL support out of the box

## Installation

Add the package to your Go module:

```bash
go get github.com/your-org/your-repo/internal/pkg/dbclient
```

## Usage

### Basic Example

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/your-org/your-repo/internal/pkg/dbclient"
)

func main() {
	// Create a new database client
	client := dbclient.NewClient(
		"http://database-service:8089", // URL of the database-service
		"your-auth-token",              // Optional authentication token
	)

	// Create a new database manager
	manager := dbclient.NewDatabaseManager(
		client,
		"example-service",         // Name of your service
		dbclient.DatabaseTypePostgreSQL, // Database type (postgresql or mysql)
		"./migrations",            // Path to migrations directory
		nil,                       // Optional logger (uses default if nil)
	)

	// Initialize the database
	db, err := manager.Initialize(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer manager.Close()

	// Use the database connection
	// ...

	// Clean up when done
	if err := manager.Cleanup(context.Background()); err != nil {
		log.Printf("Warning: failed to clean up database: %v", err)
	}
}
```

### Configuration

You can configure the database client using environment variables or a configuration file. Here's an example configuration file:

```yaml
# config.yaml
database_service:
  base_url: "http://database-service:8089"
  auth_token: "your-auth-token"
  database:
    type: "postgresql"
    ssl_mode: "disable"
  migrations:
    path: "./migrations"
    table: "schema_migrations"
```

### Running Migrations

Place your migration files in the specified migrations directory. The migrations should follow the `golang-migrate` naming convention:

```
migrations/
├── 000001_init.up.sql
├── 000001_init.down.sql
├── 000002_add_users_table.up.sql
└── 000002_add_users_table.down.sql
```

The migrations will be run automatically when you call `manager.Initialize()`.

### Using the Database Connection

Once initialized, you can use the standard `database/sql` package to interact with your database:

```go
// Query example
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return fmt.Errorf("failed to query users: %w", err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return fmt.Errorf("failed to scan row: %w", err)
    }
    log.Printf("User: ID=%d, Name=%s", id, name)
}

if err := rows.Err(); err != nil {
    return fmt.Errorf("error iterating rows: %w", err)
}

// Exec example
result, err := db.ExecContext(ctx, 
    "INSERT INTO users (name, email) VALUES ($1, $2)", 
    "John Doe", 
    "john@example.com",
)
if err != nil {
    return fmt.Errorf("failed to insert user: %w", err)
}

id, err := result.LastInsertId()
if err != nil {
    return fmt.Errorf("failed to get last insert ID: %w", err)
}

log.Printf("Created user with ID: %d", id)
```

## Error Handling

The package returns standard Go errors that you can handle in your application. For database errors, you can use the standard `database/sql` error handling:

```go
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        // Handle no rows in result set
    } else if errors.Is(err, sql.ErrConnDone) {
        // Handle connection done
    } else {
        // Handle other errors
    }
}
```

## Best Practices

1. **Connection Pooling**: The `DatabaseManager` manages a connection pool for you. Don't create multiple instances for the same database.

2. **Context Usage**: Always pass a context to database operations to support timeouts and cancellation.

3. **Migrations**: Keep your migration files idempotent and test them thoroughly.

4. **Error Handling**: Always check and handle errors from database operations.

5. **Cleanup**: Call `manager.Cleanup()` when your service shuts down to release resources.

## License

This package is part of your project and is licensed under the same terms as your project.
