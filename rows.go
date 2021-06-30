package ysql

import (
	"reflect"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	//"github.com/davecgh/go-spew/spew"
)

type Rows struct {
	pgxrows pgx.Rows

	fields []pgproto3.FieldDescription
}

func (r *Rows) Close() {
	r.pgxrows.Close()
}
func (r *Rows) Err() error {
	return r.pgxrows.Err()
}
func (r *Rows) CommandTag() pgconn.CommandTag {
	return r.pgxrows.CommandTag()
}
func (r *Rows) FieldDescriptions() []pgproto3.FieldDescription {
	return r.pgxrows.FieldDescriptions()
}
func (r *Rows) Next() bool {
	return r.pgxrows.Next()
}
func (r *Rows) Scan(args ...interface{}) error {

	//spew.Dump(args)

	if r.fields == nil {
		r.fields = r.pgxrows.FieldDescriptions()
	}

	newargs := make([]interface{}, len(r.fields))
	min := 0

	fields := make(map[string]int, len(r.fields))
	for i, f := range r.fields {
		fields[string(f.Name)] = i
	}

argloop:
	for _, arg := range args {
		val := reflect.ValueOf(arg)
		typ := val.Type()

		if typ.Kind() != reflect.Ptr || val.Elem().Type().Kind() != reflect.Struct {
			for ii := min; ii < len(newargs); ii++ {
				if newargs[ii] == nil {
					min = ii
					newargs[ii] = arg
					continue argloop
				}
			}
		}

		val = val.Elem()
		typ = val.Type()

		for i := 0; i < typ.NumField(); i++ {
			ftype := typ.Field(i)

			key := strings.ToLower(ftype.Name)

			if tv, ok := ftype.Tag.Lookup("ysql"); ok {
				t, _ := parseTag(tv)
				if t != "" {
					key = t
				}

				if t == "-" {
					continue
				}
			}

			fi, ok := fields[key]
			if ok {
				// Don't overwrite if we already have a destination.
				// This way you can do something like Scan(&id, &record) and
				// if record happens to have a field "id" it won't blow away your &id part.
				if newargs[fi] == nil {
					val := val.Field(i).Addr()
					newargs[fi] = val.Interface()
				}
			}
		}
	}

	//spew.Dump(newargs)

	return r.pgxrows.Scan(newargs...)
}
func (r *Rows) Values() ([]interface{}, error) {
	return r.pgxrows.Values()
}
func (r *Rows) RawValues() [][]byte {
	return r.pgxrows.RawValues()
}
