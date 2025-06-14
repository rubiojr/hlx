package main

import (
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rubiojr/hlx"
)

type Document struct {
	Id          string
	Title       string
	Description string
	Content     string
	Category    string
}

func main() {
	fmt.Println("HLX Performance Example - Inserting 100,000 documents")
	fmt.Println("============================================================")
	fmt.Println("This example demonstrates HLX performance by:")
	fmt.Println("  • Creating 100,000 test documents")
	fmt.Println("  • Inserting them in batches of 1,000")
	fmt.Println("  • Measuring insertion and search performance")
	fmt.Println("  • Comparing memory vs file-based storage")
	fmt.Println("  • Testing file storage with synchronous=off for maximum speed")
	fmt.Println()
	fmt.Println("Usage: go run --tags fts5 performance.go")
	fmt.Println()

	// Database files to clean up
	dbFiles := []string{
		"performance_test.db", "performance_test.db-shm", "performance_test.db-wal",
		"performance_test_fast.db", "performance_test_fast.db-shm", "performance_test_fast.db-wal",
	}

	// Ensure cleanup happens even if program panics
	defer func() {
		fmt.Print("\n🧹 Cleaning up test databases... ")
		cleaned := 0
		for _, file := range dbFiles {
			if err := os.Remove(file); err == nil {
				cleaned++
			}
		}
		if cleaned > 0 {
			fmt.Printf("Removed %d files\n", cleaned)
		} else {
			fmt.Println("No files to clean")
		}
	}()

	// Test memory, file, and file with synchronous=off
	runPerformanceTest("Memory Database", ":memory:", nil)
	runPerformanceTest("File Database", "performance_test.db", nil)

	// Test with synchronous=off for maximum write performance
	fastPragmas := []string{
		"PRAGMA synchronous=OFF",
		"PRAGMA journal_mode=WAL",
		"PRAGMA cache_size=20000",
		"PRAGMA temp_store=memory",
	}
	runPerformanceTest("File Database (Fast Mode)", "performance_test_fast.db", fastPragmas)

	fmt.Println("\n" + "============================================================")
	fmt.Println("Performance comparison completed!")
	fmt.Println()
	fmt.Println("📋 Summary:")
	fmt.Println("  • Memory database: Fastest, but not persistent")
	fmt.Println("  • File database: Safe and persistent, moderate speed")
	fmt.Println("  • File database (Fast Mode): Maximum speed with synchronous=OFF")
	fmt.Println()
	fmt.Println("⚠️  WARNING: synchronous=OFF trades safety for speed!")
	fmt.Println("   Data may be lost if the system crashes before writes are flushed to disk.")
}

func runPerformanceTest(name, dbPath string, pragmas []string) {
	fmt.Printf("\n%s (%s)\n", name, dbPath)
	fmt.Println("----------------------------------------")

	// Create index
	start := time.Now()
	var idx hlx.Index[Document]
	var err error

	if pragmas != nil {
		idx, err = hlx.NewIndex[Document](dbPath, hlx.WithSQLiteDriver("sqlite3"), hlx.WithPragmas(pragmas))
	} else {
		idx, err = hlx.NewIndex[Document](dbPath, hlx.WithSQLiteDriver("sqlite3"))
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to create index: %v", err))
	}
	indexCreationTime := time.Since(start)

	// Generate test documents
	fmt.Print("Generating 100,000 test documents... ")
	start = time.Now()
	docs := generateDocuments(100000)
	generationTime := time.Since(start)
	fmt.Printf("Done in %v\n", generationTime)

	// Insert documents in batches
	batchSize := 1000
	totalBatches := len(docs) / batchSize

	fmt.Printf("Inserting documents in batches of %d...\n", batchSize)

	insertStart := time.Now()
	var totalInsertTime time.Duration

	for i := 0; i < totalBatches; i++ {
		batchStart := time.Now()
		startIdx := i * batchSize
		endIdx := startIdx + batchSize

		err := idx.Insert(docs[startIdx:endIdx]...)
		if err != nil {
			panic(fmt.Sprintf("Failed to insert batch %d: %v", i+1, err))
		}

		batchTime := time.Since(batchStart)
		totalInsertTime += batchTime

		// Show progress every 10 batches
		if (i+1)%10 == 0 {
			progress := float64(i+1) / float64(totalBatches) * 100
			avgBatchTime := totalInsertTime / time.Duration(i+1)
			fmt.Printf("  Progress: %.1f%% (%d/%d batches) - Avg batch time: %v\n",
				progress, i+1, totalBatches, avgBatchTime)
		}
	}

	totalTime := time.Since(insertStart)

	// Calculate statistics
	totalDocs := len(docs)
	timePerDoc := totalTime / time.Duration(totalDocs)
	throughput := float64(totalDocs) / totalTime.Seconds()

	// Test search performance
	fmt.Print("Testing search performance... ")
	searchStart := time.Now()
	results, err := idx.Search("content:test")
	if err != nil {
		panic(fmt.Sprintf("Search failed: %v", err))
	}
	searchTime := time.Since(searchStart)
	fmt.Printf("Found %d results in %v\n", len(results), searchTime)

	// Display comprehensive statistics
	fmt.Printf("\n📊 Performance Statistics:\n")
	fmt.Printf("  ⏱️  Index creation time:     %v\n", indexCreationTime)
	fmt.Printf("  📝 Document generation:     %v\n", generationTime)
	fmt.Printf("  💾 Total insertion time:    %v\n", totalTime)
	fmt.Printf("  🚀 Time per document:       %v (%.2f μs)\n", timePerDoc, float64(timePerDoc.Nanoseconds())/1000)
	fmt.Printf("  📈 Throughput:              %.2f docs/sec\n", throughput)
	fmt.Printf("  🔍 Search time:             %v\n", searchTime)
	fmt.Printf("  📊 Total documents:         %d\n", totalDocs)
	fmt.Printf("  📦 Batch size:              %d\n", batchSize)
	fmt.Printf("  🔢 Total batches:           %d\n", totalBatches)

	if pragmas != nil {
		fmt.Printf("  ⚙️  Custom pragmas:         %v\n", pragmas)
	}

	// Memory usage estimation
	avgDocSize := estimateDocumentSize()
	totalDataSize := float64(totalDocs*avgDocSize) / (1024 * 1024) // MB
	fmt.Printf("  💽 Estimated data size:     %.2f MB\n", totalDataSize)

	if dbPath != ":memory:" {
		if stat, err := os.Stat(dbPath); err == nil {
			dbSize := float64(stat.Size()) / (1024 * 1024) // MB
			fmt.Printf("  🗃️  Database file size:      %.2f MB\n", dbSize)
			compression := (1 - (dbSize / totalDataSize)) * 100
			if compression > 0 {
				fmt.Printf("  📦 Storage efficiency:      %.1f%% (%.2fx compression)\n", compression, totalDataSize/dbSize)
			}
		}
	}
}

func generateDocuments(count int) []Document {
	docs := make([]Document, count)
	categories := []string{"Technology", "Science", "Business", "Health", "Sports", "Entertainment"}

	for i := 0; i < count; i++ {
		docs[i] = Document{
			Title:       fmt.Sprintf("Document Title %d", i+1),
			Description: fmt.Sprintf("This is a detailed description for document number %d. It contains various keywords for testing search functionality.", i+1),
			Content:     fmt.Sprintf("This is the main content of document %d. It includes test data, sample information, and various terms that can be searched. The content is designed to be realistic while remaining simple for performance testing.", i+1),
			Category:    categories[i%len(categories)],
		}
	}

	return docs
}

func estimateDocumentSize() int {
	// Rough estimation of average document size in bytes
	// Title: ~20 chars, Description: ~150 chars, Content: ~200 chars, Category: ~12 chars
	// Plus some overhead for struct and string headers
	return 400 // bytes per document estimate
}
