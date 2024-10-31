package hlx

import (
	"context"
	"database/sql"
	"fmt"
)

func open(uri string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initDatabase(ctx context.Context, uri string, fields []string) (*sql.DB, error) {
	db, err := open(uri)
	if err != nil {
		return nil, err
	}

	q := `CREATE VIRTUAL TABLE IF NOT EXISTS fulltext_search USING FTS5(_id,`
	for _, k := range fields {
		q += fmt.Sprintf(" %s,", k)
	}
	q = q[:len(q)-1] + ");"

	_, err = db.ExecContext(ctx, q)
	return db, err
}
