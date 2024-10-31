package hlx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type Index interface {
	Search(query string) ([]Document, error)
	InsertMap(doc ...Document) error
	Insert(doc ...any) error
	Delete(id string) error
	Get(id string) (Document, error)
	Fields() []string
}

type index struct {
	fields fields
	db     *sql.DB
}

type Document map[string]string
type fields []string

func NewIndex(uri string, doc any) (Index, error) {
	docType := make(fields, 0)
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		docType = append(docType, field.Name)
	}

	db, err := initDatabase(context.Background(), uri, docType)
	if err != nil {
		return nil, err
	}

	return &index{fields: docType, db: db}, nil
}

func (i *index) Fields() []string {
	return i.fields
}

func (i *index) Get(id string) (Document, error) {
	rows, err := i.db.Query("SELECT * FROM fulltext_search WHERE _id = ?", id)
	if err != nil {
		return nil, err
	}

	doc := make(Document)
	if rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		for i, col := range cols {
			doc[col] = values[i].(string)
		}
	}

	return doc, nil
}

func (i *index) Delete(id string) error {
	_, err := i.db.Exec("DELETE FROM fulltext_search WHERE _id = ?", id)
	return err
}

func (i *index) Insert(docs ...interface{}) error {
	if len(docs) == 0 {
		return nil
	}

	q := []string{"?"}
	for range i.fields {
		q = append(q, "?")
	}
	pholder := strings.Join(q, ",") // "?,?"

	for _, doc := range docs {
		v := reflect.ValueOf(doc)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := v.Type()

		cols := []string{"_id"}

		vals := []interface{}{uuid.New().String()}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i).Interface()
			cols = append(cols, field.Name)
			vals = append(vals, value)
		}

		query := fmt.Sprintf("INSERT INTO fulltext_search (%s) VALUES (%s)",
			strings.Join(cols, ","),
			pholder)

		_, err := i.db.Exec(query, vals...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *index) InsertMap(docs ...Document) error {
	if len(docs) == 0 {
		return nil
	}

	q := []string{}
	for range i.fields {
		q = append(q, "?")
	}
	pholder := strings.Join(q, ",") // "?,?"

	for _, doc := range docs {
		cols := make([]string, 0)
		var query string
		vals := []interface{}{uuid.New().String()}
		for k, v := range doc {
			cols = append(cols, k)
			c := strings.Join(cols, ",")
			query = fmt.Sprintf("INSERT INTO fulltext_search (_id,%s) VALUES (?,%s)", c, pholder)
			vals = append(vals, v)
		}

		if _, err := i.db.Exec(query, vals...); err != nil {
			return err
		}
	}

	return nil
}

func (i *index) Search(query string) ([]Document, error) {
	rows, err := i.db.Query("SELECT * FROM fulltext_search WHERE fulltext_search MATCH ?", query)
	if err != nil {
		return nil, err
	}

	var results []Document
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		doc := make(Document)
		for i, col := range cols {
			val := values[i]
			if v, ok := val.([]byte); ok {
				doc[col] = string(v)
			} else if val == nil {
				doc[col] = ""
			} else {
				doc[col] = fmt.Sprintf("%v", val)
			}
		}
		results = append(results, doc)
	}

	return results, nil
}
