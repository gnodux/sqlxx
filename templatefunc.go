/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"
)

const (
	sqlDate = "'2006-01-02 15:04:05'"
)

var (
	DefaultFuncMap = template.FuncMap{
		"where":      where,
		"namedWhere": namedWhere,
		"nwhere":     namedWhere,
		"asc":        asc,
		"desc":       desc,
		"v":          sqlValue,
		"n":          sqlName,
		"list":       sqlValues,
		"columns":    columns,
		"allColumns": allColumns,
		"args":       args,
		"setArgs":    sets,
		"set":        setValue,
		//"tableName":    tableName,
		//"hasTenantKey": hasTenantKey,
		//"tenantKey":    tenantKey,
	}
)

func namedWhere(v any) string {
	return whereWith(v, " AND ", true)
}
func setValue(v any, newV any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rv.Set(reflect.ValueOf(newV))
	return nil
}

func columns(cols []*ColumnMeta) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(sqlName(c.ColumnName))
		pre = ","
	}
	return sb.String()
}
func allColumns(cols []*ColumnMeta) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		sb.WriteString(pre)
		sb.WriteString(sqlName(c.ColumnName))
		pre = ","
	}
	return sb.String()
}

func args(cols []*ColumnMeta) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(":")
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}

func sets(cols []*ColumnMeta) string {
	sb := &strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(sqlName(c.ColumnName))
		sb.WriteString("=:")
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}
func hasTenantKey(cols []*ColumnMeta) bool {
	for _, c := range cols {
		if c.IsTenantKey {
			return true
		}
	}
	return false
}
func tenantKey(cols []*ColumnMeta) string {
	for _, c := range cols {
		if c.IsTenantKey {
			return c.ColumnName
		}
	}
	return ""
}
func where(v any) string {
	return whereWith(v, " AND ", false)
}
func whereOr(v any) string {
	return whereWith(v, " OR ", false)
}
func whereWith(arg any, op string, named bool) string {
	argv := reflect.ValueOf(arg)
	if arg == nil {
		return ""
	}
	if op == "" {
		op = " AND "
	}
	if op[0] != ' ' {
		op = " " + op + " "
	}

	buf := strings.Builder{}
	switch reflect.TypeOf(argv.Interface()).Kind() {
	case reflect.Map:
		comma := " WHERE "
		for _, k := range argv.MapKeys() {
			buf.WriteString(comma)
			buf.WriteString(sqlName(LowerCase(k.String())))
			value := argv.MapIndex(k)
			switch reflect.TypeOf(value.Interface()).Kind() {
			case reflect.String:
				if strings.ContainsAny(value.Interface().(string), "%.?") {
					buf.WriteString(" LIKE ")
				} else {
					buf.WriteString("=")
				}
			default:
				buf.WriteString("=")
			}
			if named {
				buf.WriteString(fmt.Sprintf(":%s", k.String()))
			} else {
				buf.WriteString(sqlValue(value.Interface()))
			}
			comma = op
		}
	case reflect.Struct:
		comma := " WHERE "
		for i := 0; i < argv.NumField(); i++ {
			if argv.Field(i).IsZero() {
				continue
			}
			buf.WriteString(comma)
			buf.WriteString(sqlName(LowerCase(argv.Type().Field(i).Name)))
			buf.WriteString("=")
			buf.WriteString(sqlValue(argv.Field(i).Interface()))
			comma = op
		}

	}
	buf.WriteString(" ")
	return buf.String()

}
func orderBy(direction string, arg any) string {

	if arg == nil {
		return ""
	}
	argv := reflect.ValueOf(arg)
	sb := strings.Builder{}
	switch argv.Kind() {
	case reflect.Slice, reflect.Array:
		if argv.Len() == 0 {
			return ""
		}
		sb.WriteString("ORDER BY ")
		splitter := ""
		for idx := 0; idx < argv.Len(); idx++ {
			sb.WriteString(splitter)
			sb.WriteString(sqlName(argv.Index(idx).Interface()))
			splitter = ","
		}
	}
	if sb.Len() > 0 {
		sb.WriteString(" ")
		sb.WriteString(direction)
	}
	return sb.String()
}
func asc(args any) string {
	return orderBy("ASC", args)
}
func desc(args any) string {
	return orderBy("DESC", args)
}

// sqlValues list of sqlValues
func sqlValues(v any) string {
	value := reflect.ValueOf(v)
	comma := ""
	sb := &strings.Builder{}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < value.Len(); idx++ {
			sb.WriteString(comma)
			sb.WriteString(sqlValue(value.Index(idx).Interface()))
			comma = ","
		}
	default:
		sb.WriteString(sqlValue(value.Interface()))
	}
	return sb.String()
}

// value sql value converter(sql inject process)
func sqlValue(arg any) string {
	var ret string
	switch a := arg.(type) {
	case nil:
		ret = "NULL"
	case string:
		ret = "'" + escape(a) + "'"
	case *string:
		ret = "'" + escape(*a) + "'"
	case time.Time:
		ret = a.Format(sqlDate)
	case *time.Time:
		ret = a.Format(sqlDate)
	case bool:
		if a {
			ret = "TRUE"
		} else {
			ret = "FALSE"
		}
	case uint, uint16, uint32, uint64, int, int16, int32, int64, float32, float64:
		return fmt.Sprintf("%v", a)
	default:
		ret = fmt.Sprintf("'%v'", a)
	}

	return ret
}

var (
	encodeRef = map[byte]byte{
		'\x00': '0',
		'\'':   '\'',
		'"':    '"',
		'\b':   'b',
		'\n':   'n',
		'\r':   'r',
		'\t':   't',
		26:     'Z', // ctl-Z
		'\\':   '\\',
	}
	EncodeMap  [256]byte
	DONTESCAPE byte = 255
)

// escape only support utf-8
func escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))

	for _, w := range []byte(sql) {
		if c := EncodeMap[w]; c == DONTESCAPE {
			dest = append(dest, w)
		} else {
			dest = append(dest, '\\', c)
		}
	}

	return string(dest)
}
func sqlName(name any) string {
	col := ""
	switch n := name.(type) {
	case string:
		col = "`" + n + "`"
	default:
		col = fmt.Sprintf("`%v`", n)
	}
	return col
}
func init() {
	for i := range EncodeMap {
		EncodeMap[i] = DONTESCAPE
	}
	for i := range EncodeMap {
		if to, ok := encodeRef[byte(i)]; ok {
			EncodeMap[byte(i)] = to
		}
	}
}
