package hlx

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
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
