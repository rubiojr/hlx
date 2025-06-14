# hlx

A simple, type-safe, full-text document search library for Go powered by SQLite FTS5.

## Features

- Type-safe document storage and retrieval using Go generics
- Full-text search capabilities using SQLite FTS5
- Automatic UUID generation for documents without IDs
- Support for in-memory and file-based databases
- Simple and intuitive API

## Installation

```bash
go get github.com/rubiojr/hlx
```

## Usage

>[!IMPORTANT]
> When building examples and running tests, `--tags fts5 needs to be passed to the go command, to enable FTS5 support in some SQLite drivers like mattn/go-sqlite3.

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/rubiojr/hlx"
    _ "github.com/mattn/go-sqlite3"
)

// Define your document structure
type Document struct {
    Id      string // Id field is required
    Title   string
    Content string
}

func main() {
    // Create a new index (in-memory database)
    idx, err := hlx.NewIndex[Document](":memory:")
    if err != nil {
        panic(err)
    }

    // Insert documents
    err = idx.Insert(
        Document{
            Title:   "Hello World",
            Content: "This is my first document",
        },
        Document{
            Title:   "Second Post",
            Content: "Another example document",
        },
    )
    if err != nil {
        panic(err)
    }

    // Search documents
    results, err := idx.Search("first OR example")
    if err != nil {
        panic(err)
    }

    // Print results
    for _, doc := range results {
        fmt.Printf("Found document: %s - %s\n", doc.Title, doc.Content)
    }
}
```

### Initialization Examples

#### Basic Initialization (In-Memory)
```go
// Create an in-memory index with default settings
idx, err := hlx.NewIndex[Document](":memory:")
```

#### File-based Storage
```go
// Create a persistent database
idx, err := hlx.NewIndex[Document]("./documents.db")
```

#### Custom Database Connection
```go
import (
    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
)

// Use your own database connection
db, err := sqlx.Open("sqlite3", ":memory:")
if err != nil {
    panic(err)
}

idx, err := hlx.NewIndex[Document]("", hlx.WithDB(db))
```

#### Custom SQLite Pragmas
```go
// Use custom SQLite pragmas for performance tuning
customPragmas := []string{
    "PRAGMA journal_mode=WAL",
    "PRAGMA synchronous=NORMAL",
    "PRAGMA cache_size=20000",
    "PRAGMA temp_store=memory",
    "PRAGMA busy_timeout=5000",
}

idx, err := hlx.NewIndex[Document]("./documents.db", hlx.WithPragmas(customPragmas))
```

#### Combined Options
```go
// Combine multiple options
db, err := sqlx.Open("sqlite3", "./documents.db")
if err != nil {
    panic(err)
}

customPragmas := []string{
    "PRAGMA journal_mode=WAL",
    "PRAGMA cache_size=50000",
}

idx, err := hlx.NewIndex[Document]("", 
    hlx.WithDB(db),
    hlx.WithPragmas(customPragmas),
)
```

#### Custom SQLite Driver
```go
// Use a specific SQLite driver (default is "sqlite3")
idx, err := hlx.NewIndex[Document]("./documents.db", 
    hlx.WithSQLiteDriver("modernc.org/sqlite"),
)
```

### Search Syntax

The search syntax follows SQLite FTS5 query syntax. Here are some examples:

```go
// Simple word search
results, _ := idx.Search("hello")

// Phrase search
results, _ := idx.Search(`"hello world"`)

// Boolean operators
results, _ := idx.Search("hello OR world")
results, _ := idx.Search("hello AND world")

// Column-specific search
results, _ := idx.Search("title:hello")

// Complex queries
results, _ := idx.Search(`(title:"hello world") AND content:example`)

// Exclude columns from search
results, _ := idx.Search(`- { title content } : "hello"`)
```

### Document Operations

```go
// Get document by ID
doc, err := idx.Get("some-id")

// Delete document by ID
err := idx.Delete("some-id")

// Get available fields
fields := idx.Fields()
```

## Performance

See [performance.txt](/performance.txt).

## Important Notes

1. The document struct must have an `Id` field (case-sensitive)
2. If no ID is provided when inserting a document, a UUID will be automatically generated
3. All struct fields will be indexed and searchable
4. Field names are case-insensitive in searches

## License

See [LICENSE](/LICENSE).
