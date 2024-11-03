package hlx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const insertQuery = "INSERT INTO fulltext_search (%s) VALUES (%s)"

type Index[K any] interface {
	Search(query string) ([]K, error)
	Insert(doc ...K) error
	Delete(id string) error
	Get(id string) (K, error)
	Fields() []string
}

type index[K any] struct {
	fields     fields
	db         *sqlx.DB
	insertStmt *sql.Stmt
}

type fields []string

func NewIndex[K any](uri string) (Index[K], error) {
	f := make(fields, 0)
	var zero K
	v := reflect.ValueOf(zero)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	idAdded := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		l := strings.ToLower(field.Name)
		f = append(f, l)
		if l == "id" {
			idAdded = true
		}
	}

	if !idAdded {
		return nil, fmt.Errorf("Id field is missing")
	}

	db, err := initDatabase(context.Background(), uri, f)
	if err != nil {
		return nil, err
	}

	q := []string{}
	for range f {
		q = append(q, "?")
	}
	pholder := strings.Join(q, ",")
	iquery := fmt.Sprintf(insertQuery,
		strings.Join(f, ","),
		pholder)

	stmt, err := db.Prepare(iquery)
	if err != nil {
		return nil, err
	}

	return &index[K]{fields: f, db: db, insertStmt: stmt}, nil
}

func (i *index[K]) Fields() []string {
	return i.fields
}

func (i *index[K]) Get(id string) (K, error) {
	var doc K
	rows, err := i.db.Queryx("SELECT * FROM fulltext_search WHERE id = ?", id)
	if err != nil {
		return doc, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&doc); err != nil {
			return doc, err
		}
	}
	return doc, nil
}

func (i *index[K]) Delete(id string) error {
	_, err := i.db.Exec("DELETE FROM fulltext_search WHERE id = ?", id)
	return err
}

func (i *index[K]) Insert(docs ...K) (err error) {
	vals := make([]any, len(i.fields))

	for _, doc := range docs {
		v := reflect.ValueOf(doc)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := v.Type()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i).Interface()
			if field.Name == "Id" && value == "" {
				value = uuid.New().String()
			}
			vals[i] = value
		}

		_, err = i.insertStmt.Exec(vals...)
	}

	return err
}

func (i *index[K]) Search(query string) ([]K, error) {
	rows, err := i.db.Queryx("SELECT * FROM fulltext_search WHERE fulltext_search MATCH ?", query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []K
	for rows.Next() {
		var result K
		if err := rows.StructScan(&result); err != nil {
			return nil, fmt.Errorf("struct scan failed: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}
