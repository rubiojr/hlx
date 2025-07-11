package hlx

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

const docCount = 10000

type TestDoc struct {
	Id          string
	Title       string
	Description string
	Content     string
}

func TestNewIndex(t *testing.T) {
	type badDoc struct{}

	_, err := NewIndex[badDoc](":memory:")
	assert.NotNil(t, err, "Expected error, got nil")
	assert.Equal(t, err.Error(), "Id field is missing")

	td := t.TempDir()
	dbfile := filepath.Join(td, "test.db")

	type doc struct {
		Id string
	}

	_, err = NewIndex[doc](":memory:")
	assert.NoError(t, err)

	_, err = NewIndex[doc](fmt.Sprintf("file://%s", dbfile))
	assert.NoError(t, err)
	assert.FileExists(t, dbfile)
}

func TestNewIndexWithCustomDB(t *testing.T) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	type doc struct {
		Id string
	}

	idx, err := NewIndex[doc]("", WithDB(db))
	assert.NoError(t, err)
	assert.NotNil(t, idx)

	testDoc := doc{Id: "test-id"}
	err = idx.Insert(testDoc)
	assert.NoError(t, err)

	result, err := idx.Get("test-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-id", result.Id)
}

func TestNewIndexWithCustomPragmas(t *testing.T) {
	customPragmas := []string{
		"PRAGMA journal_mode=DELETE",
		"PRAGMA synchronous=FULL",
		"PRAGMA cache_size=5000",
		"PRAGMA temp_store=file",
		"PRAGMA busy_timeout=10000",
	}

	type doc struct {
		Id    string
		Title string
	}

	idx, err := NewIndex[doc](":memory:", WithPragmas(customPragmas))
	assert.NoError(t, err)
	assert.NotNil(t, idx)

	testDoc := doc{Id: "test-id", Title: "Test Title"}
	err = idx.Insert(testDoc)
	assert.NoError(t, err)

	result, err := idx.Get("test-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-id", result.Id)
	assert.Equal(t, "Test Title", result.Title)

	results, err := idx.Search("Test Title")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-id", results[0].Id)
}

func TestNewIndexWithCustomDBAndPragmas(t *testing.T) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	customPragmas := []string{
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=20000",
		"PRAGMA temp_store=memory",
		"PRAGMA busy_timeout=15000",
	}

	type doc struct {
		Id      string
		Title   string
		Content string
	}

	idx, err := NewIndex[doc]("", WithDB(db), WithPragmas(customPragmas))
	assert.NoError(t, err)
	assert.NotNil(t, idx)

	testDoc := doc{Id: "pragma-test", Title: "Pragma Test", Content: "Testing custom pragmas"}
	err = idx.Insert(testDoc)
	assert.NoError(t, err)

	result, err := idx.Get("pragma-test")
	assert.NoError(t, err)
	assert.Equal(t, "pragma-test", result.Id)
	assert.Equal(t, "Pragma Test", result.Title)
	assert.Equal(t, "Testing custom pragmas", result.Content)

	results, err := idx.Search("custom pragmas")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "pragma-test", results[0].Id)

	// Verify some pragma settings were applied by querying SQLite
	var synchronous int
	err = db.Get(&synchronous, "PRAGMA synchronous")
	assert.NoError(t, err)
	assert.Equal(t, 1, synchronous) // NORMAL = 1

	var cacheSize int
	err = db.Get(&cacheSize, "PRAGMA cache_size")
	assert.NoError(t, err)
	assert.Equal(t, 20000, cacheSize)

	var busyTimeout int
	err = db.Get(&busyTimeout, "PRAGMA busy_timeout")
	assert.NoError(t, err)
	assert.Equal(t, 15000, busyTimeout)
}

func TestInsert(t *testing.T) {
	idx, err := NewIndex[TestDoc](":memory:")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	t.Run("WithId", func(t *testing.T) {
		doc := TestDoc{
			Id:          "1",
			Title:       "Test Document",
			Description: "This is a test description",
			Content:     "This is the main content of the test document",
		}
		err = idx.Insert(doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}

		// Test if the document was inserted
		res, err := idx.Get("1")
		if err != nil {
			t.Fatalf("Failed to get document: %v", err)
		}

		if res.Id != "1" {
			t.Fatalf("Id should be 1, was %s", res.Id)
		}
	})

	t.Run("WithoutId", func(t *testing.T) {
		idx, err := NewIndex[TestDoc](":memory:")
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}

		doc := TestDoc{
			Title:       "Test Document",
			Description: "This is a test description",
			Content:     "This is the main content of the test document",
		}
		err = idx.Insert(doc)
		if err != nil {
			t.Fatalf("Failed to insert document: %v", err)
		}

		res, err := idx.Search("Test Document")
		if err != nil {
			t.Fatalf("Failed to search document: %v", err)
		}

		if _, err := uuid.Parse(res[0].Id); err != nil {
			t.Fatal("Id is not a valid UUID")
		}
	})

	t.Run("auto-generates UUIDs", func(t *testing.T) {
		doc := TestDoc{
			Title: "Foobar",
		}
		err := idx.Insert(doc)
		assert.NoError(t, err)
		docs, err := idx.Search("title:Foobar")
		assert.NoError(t, err)
		_, err = uuid.Parse(docs[0].Id)
		assert.NoError(t, err)
	})
}

func TestSearch(t *testing.T) {
	// Create a test document
	doc := TestDoc{
		Title:       "Test Document",
		Description: "This is a test description",
		Content:     "This is the main content of the test document",
	}

	// Initialize the index with SQLite in-memory database
	idx, err := NewIndex[TestDoc](":memory:")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Insert test document
	err = idx.Insert(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Test cases
	tests := []struct {
		name          string
		searchQuery   string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Search exact match",
			searchQuery:   "Test Document",
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Search partial match",
			searchQuery:   "test",
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Search no match",
			searchQuery:   "nonexistent",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Search with content",
			searchQuery:   "main content",
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Search with title",
			searchQuery:   `title: "Test Document"`,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Exclude title and content columns",
			searchQuery:   `- { title content } : "Test Document"`,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Search only content",
			searchQuery:   `content: Test AND Document`,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Search with title and content",
			searchQuery:   `(content : Test) AND (title: Document)`,
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := idx.Search(tt.searchQuery)

			// Check error expectations
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check results count
			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			// If we expect results, verify the content
			if tt.expectedCount > 0 && len(results) > 0 {
				result := results[0]
				if result.Title != doc.Title {
					t.Errorf("Expected title %q, got %q", doc.Title, result.Title)
				}
				if result.Description != doc.Description {
					t.Errorf("Expected description %q, got %q", doc.Description, result.Description)
				}
				if result.Content != doc.Content {
					t.Errorf("Expected content %q, got %q", doc.Content, result.Content)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	idx, err := NewIndex[TestDoc](":memory:")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Create and insert a test document
	doc := TestDoc{
		Id:          "test-id",
		Title:       "Test Document",
		Description: "This is a test description",
		Content:     "This is the main content of the test document",
	}

	err = idx.Insert(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	t.Run("Get existing document", func(t *testing.T) {
		result, err := idx.Get("test-id")
		assert.NoError(t, err)
		assert.Equal(t, doc.Id, result.Id)
		assert.Equal(t, doc.Title, result.Title)
		assert.Equal(t, doc.Description, result.Description)
		assert.Equal(t, doc.Content, result.Content)
	})

	t.Run("Get non-existing document", func(t *testing.T) {
		_, err := idx.Get("non-existing-id")
		assert.Error(t, err)
		assert.Equal(t, "document not found", err.Error())
	})

	t.Run("auto-generates UUIDs", func(t *testing.T) {
		doc := TestDoc{
			Title: "Foobar",
		}
		err := idx.Insert(doc)
		assert.NoError(t, err)
		docs, err := idx.Search("title:Foobar")
		assert.NoError(t, err)
		u := docs[0].Id
		_, err = uuid.Parse(u)
		assert.NoError(t, err)
		doc, err = idx.Get(u)
		assert.NoError(t, err)
		assert.Equal(t, u, doc.Id)
	})
}
