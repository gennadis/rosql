package db

import (
	"context"
	"fmt"
	"strings"
)

type Result struct {
	Columns []string
	Rows    [][]string
}

func (d *DB) Query(ctx context.Context, sqlQuery string, limit int) (Result, error) {
	sqlQuery = enforceLimit(sqlQuery, limit)

	rows, err := d.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return Result{}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return Result{}, err
	}

	var results [][]string

	for rows.Next() {
		// handle slow query and context cancellation
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}

		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))

		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return Result{}, err
		}

		row := make([]string, len(cols))
		for i, v := range values {
			if v == nil {
				row[i] = "<null>"
				continue
			}
			row[i] = fmt.Sprintf("%v", v)
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return Result{}, err
	}

	return Result{
		Columns: cols,
		Rows:    results,
	}, nil
}

func enforceLimit(sql string, limit int) string {
	lower := strings.ToLower(sql)
	if strings.Contains(lower, " limit ") {
		return sql
	}
	return fmt.Sprintf("%s LIMIT %d", sql, limit)
}
