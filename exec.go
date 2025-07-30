package ysql

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

func Exec(h Handle, ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	newsql, newargs := munge(sql, args)
	return h.Exec(ctx, newsql, newargs...)
}
