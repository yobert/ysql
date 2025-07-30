//
// select a from b where id = $1;
// select a from b where id = $id;
// select a from b where $=id;
// insert into a (b) values ($1);
// insert into a ($$b) values ($...);
//

package ysql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoRows = pgx.ErrNoRows

type Handle interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

var (
	// Assert the types we want to work with can satisfy the interface
	_ Handle = &pgxpool.Pool{}
	_ Handle = &pgx.Conn{}
	_ Handle = pgx.Tx(nil)
)

func Query(h Handle, ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	newsql, newargs := munge(sql, args)
	pgxrows, err := h.Query(ctx, newsql, newargs...)
	rows := Rows{pgxrows: pgxrows}
	return &rows, err
}

func QueryRow(h Handle, ctx context.Context, sql string, args ...any) pgx.Row {
	rows, _ := Query(h, ctx, sql, args...)
	return (*Row)(rows.(*Rows))
}
