package main

import (
	"fmt"

	"github.com/rubiojr/hlx"
	_ "modernc.org/sqlite"
)

type Doc struct {
	Id      string
	Name    string
	Content string
}

func main() {
	idx, err := hlx.NewIndex[Doc]("db.sqlite", hlx.WithSQLiteDriver("sqlite"))
	if err != nil {
		panic(err)
	}

	err = idx.Insert(
		Doc{Name: "greeting", Content: "hello"},
		Doc{Name: "greeting", Content: "alo?"},
	)
	if err != nil {
		panic(err)
	}

	results, err := idx.Search(`hello OR "alo?"`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Results found:", len(results))
	for _, r := range results {
		fmt.Println(r)
	}
}
