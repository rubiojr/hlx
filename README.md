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

### Using File-based Storage

```go
// Create a persistent database
idx, err := hlx.NewIndex[Document]("./documents.db")
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

## Important Notes

1. The document struct must have an `Id` field (case-sensitive)
2. If no ID is provided when inserting a document, a UUID will be automatically generated
3. All struct fields will be indexed and searchable
4. Field names are case-insensitive in searches

## License

See [LICENSE](/LICENSE).
