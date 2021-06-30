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

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNoRows = pgx.ErrNoRows

type Handle interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}

var (
	_ Handle = &pgxpool.Pool{}
	_ Handle = &pgx.Conn{}
	_ Handle = pgx.Tx(nil)
)

func Query(h Handle, ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	newsql, newargs := munge(sql, args)
	pgxrows, err := h.Query(ctx, newsql, newargs...)
	rows := Rows{pgxrows: pgxrows}
	return &rows, err
}

func QueryRow(h Handle, ctx context.Context, sql string, args ...interface{}) pgx.Row {
	rows, _ := Query(h, ctx, sql, args...)
	return (*Row)(rows.(*Rows))
}
