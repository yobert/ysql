package ysql

import (
	"context"

	"github.com/jackc/pgconn"
)

func Exec(h Handle, ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	newsql, newargs := munge(sql, args)
	return h.Exec(ctx, newsql, newargs...)
}
