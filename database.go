package hlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
)

func open(uri string) (*sql.DB, error) {
	if err := validateURI(uri); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func validateURI(uri string) error {
	stat, err := os.Stat(uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(uri)
			if err != nil {
				return err
			}

			return f.Close()
		}

		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", uri)
	}

	return nil
}

func initDatabase(ctx context.Context, uri string, doc DocType) (*sql.DB, error) {
	db, err := open(uri)
	if err != nil {
		return nil, err
	}

	q := `CREATE VIRTUAL TABLE IF NOT EXISTS fulltext_search USING FTS5(_id,`
	for _, k := range doc {
		q += fmt.Sprintf(" %s,", k)
	}
	q = q[:len(q)-1] + ");"

	_, err = db.ExecContext(ctx, q)
	return db, err
}
