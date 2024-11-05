package hlx

import (
	"fmt"
	"testing"
)

func BenchmarkInsert(b *testing.B) {
	b.Run("memory", func(b *testing.B) {
		idx, err := NewIndex[TestDoc](":memory:")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}

		// Create 10000 test documents
		docs := make([]TestDoc, docCount)
		for i := 0; i < docCount; i++ {
			docs[i] = TestDoc{
				Title:       fmt.Sprintf("value_%d_1", i),
				Description: fmt.Sprintf("value_%d_2", i),
				Content:     fmt.Sprintf("value_%d_3", i),
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
	})
	b.Run("tmp db", func(b *testing.B) {
		idx, err := NewIndex[TestDoc]("file:///tmp/hlx_test.db")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}

		// Create 10000 test documents
		docs := make([]TestDoc, docCount)
		for i := 0; i < docCount; i++ {
			docs[i] = TestDoc{
				Title:       fmt.Sprintf("value_%d_1", i),
				Description: fmt.Sprintf("value_%d_2", i),
				Content:     fmt.Sprintf("value_%d_3", i),
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
	})
}
