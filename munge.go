package ysql

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	matchValueFieldRe = regexp.MustCompile(`\$=?[\$a-z0-9_]+`)
	matchValueStdRe   = regexp.MustCompile(`\$[0-9]+`)
)

func munge(sql string, args []interface{}) (string, []interface{}) {
	argpass := make([]bool, len(args))
	newargs := make([]interface{}, len(args))
	argmap := make(map[string]string)
	splats := ""

	// leave alone $1, $2, etc --- let pgx behave as usual.
	sql = matchValueStdRe.ReplaceAllStringFunc(sql, func(v string) string {
		i, err := strconv.Atoi(v[1:])
		i--
		if err == nil && i >= 0 && i < len(args) {
			argpass[i] = true
			newargs[i] = args[i]
			argmap[v] = v
		}
		return v
	})

	sql = matchValueFieldRe.ReplaceAllStringFunc(sql, func(v string) string {
		fast, ok := argmap[v]
		if ok {
			return fast
		}

		k := v[1:]

		splat := false
		equals := false
		if len(k) > 0 {
			if k[0] == '$' {
				splat = true
				k = k[1:]
			} else if k[0] == '=' {
				equals = true
				k = k[1:]
			}
		}

		for _, arg := range args {
			val := reflect.ValueOf(arg)
			typ := val.Type()
			if typ.Kind() == reflect.Ptr {
				val = val.Elem()
				typ = val.Type()
			}
			if typ.Kind() != reflect.Struct {
				continue
			}
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

				if k == key {
					holder := ""

					// pick the next unused placeholder
					for ai, used := range argpass {
						if !used {
							argpass[ai] = true
							newargs[ai] = val.Field(i).Interface()
							holder = fmt.Sprintf("$%d", ai+1)
							break
						}
					}

					if holder == "" {
						// allocate another placeholder
						newargs = append(newargs, val.Field(i).Interface())
						holder = fmt.Sprintf("$%d", len(newargs))
					}

					if splat {
						if splats == "" {
							splats = holder
						} else {
							splats += ", " + holder
						}
						argmap[v] = k
						return k
					}

					if equals {
						holder = k + " = " + holder
					}

					argmap[v] = holder
					return holder
				}
			}
		}

		// We couldn't match anything.
		// for the moment just leave it unchanged, which should cause an SQL error.
		argmap[v] = v
		return v
	})

	sql = strings.ReplaceAll(sql, "$...", splats)

	return sql, newargs
}
