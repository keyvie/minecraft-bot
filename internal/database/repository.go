package database

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository[M any] struct {
	pool  *pgxpool.Pool
	table string
}

func NewRepository[M any](pool *pgxpool.Pool, table string) *Repository[M] {
	return &Repository[M]{
		pool:  pool,
		table: table,
	}
}

func (r *Repository[M]) Table() string {
	return r.table
}

type Filter map[string]any

func buildWhere(filter Filter, startIndex int) (string, []any) {
	result := ""
	args := make([]any, 0, len(filter))

	i := 0
	for key, value := range filter {
		if i > 0 {
			result += " AND "
		} else {
			result += " WHERE "
		}

		result += fmt.Sprintf(`"%s" = $%d`, key, startIndex+len(args))
		args = append(args, value)
		i++
	}

	return result, args
}

func (r *Repository[M]) Create(ctx context.Context, model M) (pgconn.CommandTag, error) {

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
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query += fmt.Sprintf(
		" (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return r.pool.Exec(ctx, query, args...)
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
		where, whereArgs := buildWhere(opts.Where, len(args)+1)
		query += where
		args = append(args, whereArgs...)
	}

	if opts.Limit > 0 {
		args = append(args, opts.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	if opts.Offset > 0 {
		args = append(args, opts.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return response, err
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[M])
	if err != nil {
		return response, err
	}
	if len(items) > 0 {
		response.Items = items
		response.Total = len(items)
	}
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
		where, whereArgs := buildWhere(opts.Where, len(args)+1)
		query += where
		args = append(args, whereArgs...)
	}

	query += " LIMIT 1"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return response, err
	}

	item, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByNameLax[M],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response, nil
		}
		return response, err
	}

	response.Item = item
	response.Found = true

	return response, nil
}

type QueryDeleteOptions struct {
	Where Filter
}

func (r *Repository[M]) Delete(
	ctx context.Context,
	opts QueryDeleteOptions,
) (pgconn.CommandTag, error) {
	query := "DELETE FROM " + `"` + r.table + `"`
	args := make([]any, 0)

	if len(opts.Where) > 0 {
		where, whereArgs := buildWhere(opts.Where, len(args)+1)
		query += where
		args = append(args, whereArgs...)
	}

	return r.pool.Exec(ctx, query, args...)
}
