package hlx

import (
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

const docCount = 1000

type TestDoc struct {
	Field1  string
	Field2  string
	Field3  string
	Field4  string
	Field5  string
	Field6  string
	Field7  string
	Field8  string
	Field9  string
	Field10 string
}

func BenchmarkInsertMap(b *testing.B) {
	// Create in-memory SQLite database
	idx, err := NewIndex(":memory:", TestDoc{})
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// Create 10000 test documents
	docs := make([]Document, docCount)
	for i := 0; i < docCount; i++ {
		doc := make(Document)
		for j, field := range idx.Fields() {
			doc[field] = fmt.Sprintf("value_%d_%d", i, j)
		}
		docs[i] = doc
	}

	// Reset timer before the actual benchmark
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		err := idx.InsertMap(docs...)
		if err != nil {
			b.Fatalf("Failed to insert documents: %v", err)
		}
	}
}

func BenchmarkInsert(b *testing.B) {

	// Define document structure with 10 fields
	// Create in-memory SQLite database
	idx, err := NewIndex(":memory:", TestDoc{})
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// Create 10000 test documents
	docs := make([]interface{}, docCount)
	for i := 0; i < docCount; i++ {
		docs[i] = TestDoc{
			Field1:  fmt.Sprintf("value_%d_1", i),
			Field2:  fmt.Sprintf("value_%d_2", i),
			Field3:  fmt.Sprintf("value_%d_3", i),
			Field4:  fmt.Sprintf("value_%d_4", i),
			Field5:  fmt.Sprintf("value_%d_5", i),
			Field6:  fmt.Sprintf("value_%d_6", i),
			Field7:  fmt.Sprintf("value_%d_7", i),
			Field8:  fmt.Sprintf("value_%d_8", i),
			Field9:  fmt.Sprintf("value_%d_9", i),
			Field10: fmt.Sprintf("value_%d_10", i),
		}
	}

	// Reset timer before the actual benchmark
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		err := idx.Insert(docs...)
		if err != nil {
			b.Fatalf("Failed to insert documents: %v", err)
		}
	}
}
