package ysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
)

type Rows struct {
	pgxrows pgx.Rows

	flen   int
	fields []pgproto3.FieldDescription
	names  []string
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

	if r.fields == nil {
		r.fields = r.pgxrows.FieldDescriptions()
		r.flen = len(r.fields)
		r.names = make([]string, r.flen)
		for i, f := range r.fields {
			r.names[i] = string(f.Name)
		}
	}

	newargs := make([]interface{}, len(r.fields))

	for _, arg := range args {
		val := reflect.ValueOf(arg)

		if err := walk(r, val, newargs); err != nil {
			return err
		}
	}

	return r.pgxrows.Scan(newargs...)
}
func (r *Rows) Values() ([]interface{}, error) {
	return r.pgxrows.Values()
}
func (r *Rows) RawValues() [][]byte {
	return r.pgxrows.RawValues()
}

func walk(r *Rows, val reflect.Value, newargs []interface{}) error {

	typ := val.Type()

	// Special treatment for time.Time--- it's a struct, but we don't want to dissect its fields.
	if (typ.Kind() != reflect.Struct || typ.PkgPath() == "time") && (typ.Kind() != reflect.Ptr || val.Elem().Type().Kind() != reflect.Struct || val.Elem().Type().PkgPath() == "time") {
		// Doesn't look like a struct or struct pointer? Pass it down the line to pgx, hopefully
		// it's something scannable the standard way.
		for ii := 0; ii < r.flen; ii++ {
			if newargs[ii] == nil {
				newargs[ii] = val.Interface()
				return nil
			}
		}
		return fmt.Errorf("Scan called with more receivers (%d) than query result fields (%d)", len(newargs), r.flen)
	}

	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	for i := 0; i < typ.NumField(); i++ {
		ftype := typ.Field(i)

		// Another important special case for time.Time. Need to think of a better way to do this.
		if ftype.Type.Kind() == reflect.Struct && ftype.Type.PkgPath() != "time" {
			if err := walk(r, val.Field(i), newargs); err != nil {
				return err
			}
			continue
		}

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

		for ii := 0; ii < r.flen; ii++ {
			if newargs[ii] == nil && r.names[ii] == key {
				val := val.Field(i).Addr()
				newargs[ii] = val.Interface()
				break
			}
		}
	}

	return nil
}
