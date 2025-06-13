# HLX Examples

This directory contains example programs demonstrating how to use HLX (High-Level indeX) for full-text search operations.

## Examples

### 1. Basic Usage (`main.go`)

A simple example showing basic HLX functionality:
- Creating an index
- Inserting documents
- Performing searches

```bash
go run main.go -tags fts5
```

### 2. Custom Database (`custom_db_example.go`)

Demonstrates how to use HLX with a custom database connection.

```bash
go run custom_db_example.go -tags fts5
```

### 3. Performance Testing (`performance.go`)

A comprehensive performance benchmark that:
- Inserts 100,000 test documents
- Measures insertion performance in batches
- Compares memory vs file-based storage
- Tests search performance
- Provides detailed statistics

```bash
go run performance.go -tags fts5
```

**Sample Output:**
```
HLX Performance Example - Inserting 100,000 documents
============================================================

Memory Database (:memory:)
----------------------------------------
📊 Performance Statistics:
  🚀 Time per document:       152.04 μs
  📈 Throughput:              6,577 docs/sec
  🔍 Search time:             388ms
  💽 Estimated data size:     38.15 MB

File Database (performance_test.db)
----------------------------------------
📊 Performance Statistics:
  🚀 Time per document:       594.18 μs
  📈 Throughput:              1,683 docs/sec
  🔍 Search time:             395ms
  💽 Estimated data size:     38.15 MB
  🗃️  Database file size:      73.53 MB
```

## Prerequisites

### FTS5 Support

All examples require SQLite with FTS5 (Full-Text Search) support. Make sure to use the `-tags fts5` flag when running the examples:

```bash
go run <example>.go -tags fts5
```

### Dependencies

The examples use different SQLite drivers:
- `main.go` and `performance.go`: Uses `modernc.org/sqlite` (pure Go)
- `custom_db_example.go`: Uses `github.com/mattn/go-sqlite3` (CGO)

Make sure your Go environment supports the required driver for each example.

## Document Structure

Most examples use a similar document structure:

```go
type Document struct {
    Id          string  // Auto-generated if empty
    Title       string
    Description string
    Content     string
    Category    string  // Optional field for categorization
}
```

## Search Syntax

HLX supports advanced search queries:

- **Simple search**: `hello world`
- **Field-specific**: `title:"specific title"`
- **Boolean operators**: `content:test AND category:technology`
- **Exclusion**: `- { title content } : "exclude from these fields"`
- **Complex queries**: `(content:test) AND (title:document)`

## Performance Tips

1. **Memory vs File Storage**:
   - Memory (`:memory:`): ~3.9x faster insertion, no persistence
   - File storage: Slower but persistent, suitable for production

2. **Batch Insertions**:
   - Insert documents in batches (1000-5000) for better performance
   - Avoid inserting one document at a time

3. **Search Optimization**:
   - Use field-specific searches when possible
   - FTS5 provides excellent search performance even with large datasets

## Running All Examples

To run all examples sequentially:

```bash
# Basic example
go run main.go -tags fts5

# Custom DB example  
go run custom_db_example.go -tags fts5

# Performance benchmark
go run performance.go -tags fts5
```

## Troubleshooting

### "no such module: FTS5" Error

This error occurs when SQLite doesn't have FTS5 support compiled in. Solutions:

1. Always use the `-tags fts5` flag
2. Ensure your SQLite driver supports FTS5
3. For `go-sqlite3`, you might need: `go run -tags "fts5" example.go`

### Build Issues

If you encounter build issues:

1. Make sure you have a C compiler installed (for CGO-based drivers)
2. Try using the pure Go driver (`modernc.org/sqlite`)
3. Update your Go version to 1.21 or later

## Learn More

- [HLX Documentation](../README.md)
- [SQLite FTS5 Documentation](https://www.sqlite.org/fts5.html)
- [Go SQLite Drivers Comparison](https://github.com/golang/go/wiki/SQLInterface)