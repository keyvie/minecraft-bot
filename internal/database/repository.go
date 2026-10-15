package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository[M any] struct {
	db    *sqlx.DB
	table string
}

func NewRepository[M any](db *sqlx.DB, table string) *Repository[M] {
	return &Repository[M]{
		db:    db,
		table: table,
	}
}

func (r *Repository[M]) Table() string {
	return r.table
}

type Filter map[string]any

func buildWhere(filter Filter) (string, []any) {
	result := ""
	args := make([]any, 0, len(filter))

	i := 0
	for key, value := range filter {
		if i > 0 {
			result += " AND "
		} else {
			result += " WHERE "
		}

		result += `"` + key + `" = ?`
		args = append(args, value)
		i++
	}

	return result, args
}

func (r *Repository[M]) Create(ctx context.Context, model M) (sql.Result, error) {
	var result sql.Result

	query := fmt.Sprintf("INSERT INTO %q", r.table)
	columns := []string{}
	args := []any{}

	v := reflect.ValueOf(model)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		column := t.Field(i).Tag.Get("db")
		if column == "" || column == "-" {
			continue
		}
		columns = append(columns, `"`+column+`"`)
		args = append(args, v.Field(i).Interface())
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query += fmt.Sprintf(
		" (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	return result, err
}

type QueryFindOptions struct {
	Where  Filter
	Limit  int
	Offset int
}

type QueryFindResult[M any] struct {
	Items []M
	Total int
}

func (r *Repository[M]) Find(
	ctx context.Context,
	opts QueryFindOptions,
) (QueryFindResult[M], error) {
	response := QueryFindResult[M]{
		Items: make([]M, 0),
	}

	query := "SELECT * FROM " + `"` + r.table + `"`
	args := make([]any, 0)

	if len(opts.Where) > 0 {
		where, whereArgs := buildWhere(opts.Where)
		query += where
		args = append(args, whereArgs...)
	}

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	} else if opts.Offset > 0 {
		query += " LIMIT -1"
	}

	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	if err := r.db.SelectContext(ctx, &response.Items, query, args...); err != nil {
		return response, err
	}

	response.Total = len(response.Items)
	return response, nil
}

type QueryFindOneOptions struct {
	Where Filter
}

type QueryFindOneResult[M any] struct {
	Item  M
	Found bool
}

func (r *Repository[M]) FindOne(
	ctx context.Context,
	opts QueryFindOneOptions,
) (QueryFindOneResult[M], error) {
	response := QueryFindOneResult[M]{}

	query := "SELECT * FROM " + `"` + r.table + `"`
	args := make([]any, 0)

	if len(opts.Where) > 0 {
		where, whereArgs := buildWhere(opts.Where)
		query += where
		args = append(args, whereArgs...)
	}

	query += " LIMIT 1"

	var items []M
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return response, err
	}

	if len(items) > 0 {
		response.Item = items[0]
		response.Found = true
	}

	return response, nil
}

type QueryDeleteOptions struct {
	Where  Filter
	Limit  int
	Offset int
}

func (r *Repository[M]) Delete(
	ctx context.Context,
	opts QueryDeleteOptions,
) (sql.Result, error) {
	query := "DELETE FROM " + `"` + r.table + `"`
	args := make([]any, 0)

	if len(opts.Where) > 0 {
		where, whereArgs := buildWhere(opts.Where)
		query += where
		args = append(args, whereArgs...)
	}

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	return r.db.ExecContext(ctx, query, args...)
}
