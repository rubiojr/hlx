package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/rubiojr/hlx"
	_ "modernc.org/sqlite"
)

type Doc struct {
	Id      string
	Name    string
	Content string
}

func main() {
	// Create a custom database connection
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create index with custom DB
	idx, err := hlx.NewIndex[Doc]("", hlx.WithDB(db))
	if err != nil {
		log.Fatal(err)
	}

	// Insert some documents
	err = idx.Insert(
		Doc{Name: "greeting", Content: "hello world"},
		Doc{Name: "farewell", Content: "goodbye world"},
		Doc{Name: "question", Content: "how are you?"},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Search for documents
	results, err := idx.Search("world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d documents containing 'world':\n", len(results))
	for _, doc := range results {
		fmt.Printf("  ID: %s, Name: %s, Content: %s\n", doc.Id, doc.Name, doc.Content)
	}

	// You can still use the same database connection for other operations
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM fulltext_search")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nTotal documents in database: %d\n", count)
}