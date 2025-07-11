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

type Options struct {
	DB      *sqlx.DB
	driver  string
	pragmas []string
}

type Option func(*Options)

func WithDB(db *sqlx.DB) Option {
	return func(o *Options) {
		o.DB = db
	}
}

func WithSQLiteDriver(drv string) Option {
	return func(o *Options) {
		o.driver = drv
	}
}

func WithPragmas(pragmas []string) Option {
	return func(o *Options) {
		o.pragmas = pragmas
	}
}

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

var DefaultPragmas = []string{
	"PRAGMA journal_mode=WAL",
	"PRAGMA synchronous=NORMAL",
	"PRAGMA cache_size=10000",
	"PRAGMA temp_store=memory",
	"PRAGMA mmap_size=268435456",
	"PRAGMA busy_timeout=5000",
}

func NewIndex[K any](uri string, opts ...Option) (Index[K], error) {
	options := &Options{pragmas: DefaultPragmas}
	for _, opt := range opts {
		opt(options)
	}

	if options.driver == "" {
		options.driver = "sqlite3"
	}

	f := make(fields, 0)
	var zero K
	v := reflect.ValueOf(zero)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	idAdded := false
	for i := range t.NumField() {
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

	var db *sqlx.DB
	var err error
	if options.DB != nil {
		db, err = initDatabase(context.Background(), options.DB, uri, f, options.pragmas)
	} else {
		db, err = open(options.driver, uri)
		if err != nil {
			return nil, err
		}
		db, err = initDatabase(context.Background(), db, uri, f, options.pragmas)
		if err != nil {
			return nil, err
		}
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
		err = rows.StructScan(&doc)
		return doc, err
	}

	return doc, ErrDocumentNotFound
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

		for i := range t.NumField() {
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
