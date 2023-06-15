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
	SQLDate = "'2006-01-02 15:04:05'"
)

var (
	DefaultFuncMap = template.FuncMap{
		"where": Where,
		"namedWhere": func(v any) string {
			return WhereWith(v, " AND ", true)
		},
		"asc":        Asc,
		"desc":       Desc,
		"v":          Value,
		"n":          SQLName,
		"list":       Values,
		"columns":    columns,
		"allColumns": allColumns,
		"args":       args,
		"setArgs":    sets,
		//"tableName":    tableName,
		//"hasTenantKey": hasTenantKey,
		//"tenantKey":    tenantKey,
	}
)

func listValueColumns(v any) []*ColumnDef {
	argv := reflect.TypeOf(v)
	return listColumns(argv)

}
func listColumns(t reflect.Type) []*ColumnDef {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []*ColumnDef
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).IsExported() {
			fields = append(fields, NewColumnDefWith(t.Field(i)))
		}
	}
	return fields
}

func columns(cols []*ColumnDef) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(SQLName(c.ColumnName))
		pre = ","
	}
	return sb.String()
}
func allColumns(cols []*ColumnDef) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		sb.WriteString(pre)
		sb.WriteString(SQLName(c.ColumnName))
		pre = ","
	}
	return sb.String()
}

func args(cols []*ColumnDef) string {
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

func sets(cols []*ColumnDef) string {
	sb := &strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(SQLName(c.ColumnName))
		sb.WriteString("=:")
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}
func hasTenantKey(cols []*ColumnDef) bool {
	for _, c := range cols {
		if c.IsTenantKey {
			return true
		}
	}
	return false
}
func tenantKey(cols []*ColumnDef) string {
	for _, c := range cols {
		if c.IsTenantKey {
			return c.ColumnName
		}
	}
	return ""
}
func Where(v any) string {
	return WhereWith(v, " AND ", false)
}
func WhereOr(v any) string {
	return WhereWith(v, " OR ", false)
}
func WhereWith(arg any, op string, named bool) string {
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
			buf.WriteString(SQLName(LowerCase(k.String())))
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
				buf.WriteString(Value(value.Interface()))
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
			buf.WriteString(SQLName(LowerCase(argv.Type().Field(i).Name)))
			buf.WriteString("=")
			buf.WriteString(Value(argv.Field(i).Interface()))
			comma = op
		}

	}
	buf.WriteString(" ")
	return buf.String()

}
func OrderBy(direction string, arg any) string {

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
			sb.WriteString(SQLName(argv.Index(idx).Interface()))
			splitter = ","
		}
	}
	if sb.Len() > 0 {
		sb.WriteString(" ")
		sb.WriteString(direction)
	}
	return sb.String()
}
func Asc(args any) string {
	return OrderBy("ASC", args)
}
func Desc(args any) string {
	return OrderBy("DESC", args)
}

// Values list of values
func Values(v any) string {
	value := reflect.ValueOf(v)
	comma := ""
	sb := &strings.Builder{}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < value.Len(); idx++ {
			sb.WriteString(comma)
			sb.WriteString(Value(value.Index(idx).Interface()))
			comma = ","
		}
	default:
		sb.WriteString(Value(value.Interface()))
	}
	return sb.String()
}

// Value sql value converter(sql inject process)
func Value(arg any) string {
	var ret string
	switch a := arg.(type) {
	case nil:
		ret = "NULL"
	case string:
		ret = "'" + Escape(a) + "'"
	case *string:
		ret = "'" + Escape(*a) + "'"
	case time.Time:
		ret = a.Format(SQLDate)
	case *time.Time:
		ret = a.Format(SQLDate)
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

// Escape only support utf-8
func Escape(sql string) string {
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
func SQLName(name any) string {
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
