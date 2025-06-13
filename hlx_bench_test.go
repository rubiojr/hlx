package hlx

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func BenchmarkInsert(b *testing.B) {
	b.Run("memory", func(b *testing.B) {
		idx, err := NewIndex[TestDoc](":memory:")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}

		// Create 10000 test documents
		docs := make([]TestDoc, docCount)
		for i := range docCount {
			docs[i] = TestDoc{
				Title:       fmt.Sprintf("value_%d_1", i),
				Description: fmt.Sprintf("value_%d_2", i),
				Content:     fmt.Sprintf("value_%d_3", i),
			}
		}

		// Reset timer before the actual benchmark
		b.ResetTimer()
		start := time.Now()

		// Run the benchmark
		for i := 0; i < b.N; i++ {
			err := idx.Insert(docs...)
			if err != nil {
				b.Fatalf("Failed to insert documents: %v", err)
			}
		}

		elapsed := time.Since(start)
		totalDocs := int64(b.N) * int64(docCount)
		timePerDoc := elapsed / time.Duration(totalDocs)
		throughput := float64(totalDocs) / elapsed.Seconds()

		b.ReportMetric(float64(timePerDoc.Nanoseconds()), "ns/doc")
		b.ReportMetric(throughput, "docs/sec")
		b.Logf("Memory DB - Total docs: %d, Time per doc: %v, Throughput: %.2f docs/sec",
			totalDocs, timePerDoc, throughput)
	})
	b.Run("tmp db", func(b *testing.B) {
		idx, err := NewIndex[TestDoc]("file:///tmp/hlx_test.db")
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}

		// Create 10000 test documents
		docs := make([]TestDoc, docCount)
		for i := range docCount {
			docs[i] = TestDoc{
				Title:       fmt.Sprintf("value_%d_1", i),
				Description: fmt.Sprintf("value_%d_2", i),
				Content:     fmt.Sprintf("value_%d_3", i),
			}
		}

		// Reset timer before the actual benchmark
		b.ResetTimer()
		start := time.Now()

		// Run the benchmark
		for i := 0; i < b.N; i++ {
			err := idx.Insert(docs...)
			if err != nil {
				b.Fatalf("Failed to insert documents: %v", err)
			}
		}

		elapsed := time.Since(start)
		totalDocs := int64(b.N) * int64(docCount)
		timePerDoc := elapsed / time.Duration(totalDocs)
		throughput := float64(totalDocs) / elapsed.Seconds()

		b.ReportMetric(float64(timePerDoc.Nanoseconds()), "ns/doc")
		b.ReportMetric(throughput, "docs/sec")
		b.Logf("File DB - Total docs: %d, Time per doc: %v, Throughput: %.2f docs/sec",
			totalDocs, timePerDoc, throughput)
	})
}
