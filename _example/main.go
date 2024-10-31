package main

import (
	"fmt"

	"github.com/rubiojr/hlx"
	_ "modernc.org/sqlite"
)

func main() {
	idx, err := hlx.NewIndex("db.sqlite", struct{ Name, Content string }{})
	if err != nil {
		panic(err)
	}

	idx.InsertMap(
		hlx.Document{"name": "greeting", "content": "hello world"},
	)

	idx.Insert(
		struct{ Name, Content string }{Name: "greeting", Content: "alo?"},
		struct{ Content, Name string }{Name: "greeting", Content: "alo?"},
	)

	results, err := idx.Search(`hello OR "alo?"`)
	if err != nil {
		panic(err)
	}
	for _, r := range results {
		fmt.Println(r)
	}
}
