package hlx

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func open(uri string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", uri)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initDatabase(ctx context.Context, uri string, fields []string) (*sqlx.DB, error) {
	db, err := open(uri)
	if err != nil {
		return nil, err
	}

	q := `CREATE VIRTUAL TABLE IF NOT EXISTS fulltext_search USING FTS5(`
	for _, k := range fields {
		q += fmt.Sprintf(" %s,", strings.ToLower(k))
	}
	q = q[:len(q)-1] + ");"

	_, err = db.ExecContext(ctx, q)
	return db, err
}
