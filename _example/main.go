package main

import (
	"fmt"

	"github.com/rubiojr/hlx"
	_ "modernc.org/sqlite"
)

func main() {
	doc := hlx.DocType{"name", "content"}
	idx, err := hlx.NewIndex("db.sqlite", doc)
	if err != nil {
		panic(err)
	}

	idx.InsertMap(
		hlx.Document{"name": "greeting", "content": "hello world"},
	)

	doc2 := struct {
		Name    string
		Content string
	}{"greeting", "alo?"}

	idx.Insert(doc2)

	results, err := idx.Search(`hello OR "alo?"`)
	if err != nil {
		panic(err)
	}
	for _, r := range results {
		fmt.Println(r)
	}
}
