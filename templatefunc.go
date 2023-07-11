/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"fmt"
	"github.com/gnodux/sqlxx/expr"
	"reflect"
	"strings"
	"text/template"
	"time"
)

const (
	sqlDate = "'2006-01-02 15:04:05'"
)

func MakeFuncMap(driver *Driver) template.FuncMap {
	return template.FuncMap{
		"where":      func(v any) string { return where(v, driver) },
		"namedWhere": func(v any) string { return namedWhere(v, driver) },
		"nwhere":     func(v any) string { return namedWhere(v, driver) },
		"asc":        func(cols []string) string { return orderByMap(expr.Asc(cols...), driver) },
		"desc":       func(cols []string) string { return orderByMap(expr.Desc(cols...), driver) },
		"v":          func(v any) string { return sqlValue(v, driver) },
		"n":          driver.SQLNameFunc,
		"sqlName":    driver.SQLNameFunc,
		"list":       func(v []any) string { return sqlValues(v, driver) },
		"columns":    func(v []*ColumnMeta) string { return columns(v, driver) },
		"allColumns": func(v []*ColumnMeta) string { return allColumns(v, driver) },
		"args":       func(v []*ColumnMeta) string { return args(v, driver) },
		"setArgs":    func(v []*ColumnMeta) string { return sets(v, driver) },
		"orderBy":    func(v map[string]string) string { return orderByMap(v, driver) },
	}
}

func namedWhere(v any, driver *Driver) string {
	return whereWith(v, " "+driver.Keyword("AND")+" ", true, driver)
}

//func setValue(v any, newV any) any {
//	rv := reflect.ValueOf(v)
//	if rv.Kind() == reflect.Ptr {
//		rv = rv.Elem()
//	}
//	rv.Set(reflect.ValueOf(newV))
//	return nil
//}

func orderByMap(order map[string]string, driver *Driver) string {
	if len(order) == 0 {
		return ""
	}
	sb := strings.Builder{}
	pre := " ORDER BY "
	for k, v := range order {
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(driver.NameFunc(k)))
		sb.WriteString(" ")
		sb.WriteString(v)
		pre = ","
	}
	if sb.Len() > 0 {
		sb.WriteString(" ")
	}
	return sb.String()
}

func columns(cols []*ColumnMeta, driver *Driver) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(c.ColumnName))
		pre = ","
	}
	return sb.String()
}
func allColumns(cols []*ColumnMeta, driver *Driver) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(c.ColumnName))
		pre = ","
	}
	return sb.String()
}

func args(cols []*ColumnMeta, driver *Driver) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(driver.NamedPrefix)
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}

func sets(cols []*ColumnMeta, driver *Driver) string {
	sb := &strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(c.ColumnName))
		sb.WriteString("=" + driver.NamedPrefix)
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}

func where(v any, driver *Driver) string {
	return whereWith(v, " "+driver.Keyword("AND")+" ", false, driver)
}
func whereOr(v any, driver *Driver) string {
	return whereWith(v, " "+driver.Keyword("OR")+" ", false, driver)
}
func whereWith(arg any, op string, named bool, driver *Driver) string {
	argv := reflect.ValueOf(arg)
	if arg == nil {
		return ""
	}
	if op == "" {
		op = " " + driver.Keyword("AND") + " "
	}
	if op[0] != ' ' {
		op = " " + op + " "
	}

	buf := strings.Builder{}
	switch reflect.TypeOf(argv.Interface()).Kind() {
	case reflect.Map:
		comma := " " + driver.Keyword("WHERE") + " "
		for _, k := range argv.MapKeys() {
			buf.WriteString(comma)
			buf.WriteString(driver.SQLNameFunc(driver.NameFunc(k.String())))
			value := argv.MapIndex(k)
			switch reflect.TypeOf(value.Interface()).Kind() {
			case reflect.String:
				if strings.ContainsAny(value.Interface().(string), "%.?") {
					buf.WriteString(" " + driver.Keyword("LIKE") + " ")
				} else {
					buf.WriteString("=")
				}
			default:
				buf.WriteString("=")
			}
			if named {
				buf.WriteString(fmt.Sprintf(":%s", k.String()))
			} else {
				buf.WriteString(sqlValue(value.Interface(), driver))
			}
			comma = op
		}
	case reflect.Struct:
		comma := " " + driver.Keyword("WHERE") + " "
		for i := 0; i < argv.NumField(); i++ {
			if argv.Field(i).IsZero() {
				continue
			}
			buf.WriteString(comma)
			buf.WriteString(driver.SQLNameFunc(driver.NameFunc(argv.Type().Field(i).Name)))
			buf.WriteString("=")
			buf.WriteString(sqlValue(argv.Field(i).Interface(), driver))
			comma = op
		}

	}
	buf.WriteString(" ")
	return buf.String()

}

// sqlValues list of sqlValues
func sqlValues(v any, driver *Driver) string {
	value := reflect.ValueOf(v)
	comma := ""
	sb := &strings.Builder{}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < value.Len(); idx++ {
			sb.WriteString(comma)
			sb.WriteString(sqlValue(value.Index(idx).Interface(), driver))
			comma = ","
		}
	default:
		sb.WriteString(sqlValue(value.Interface(), driver))
	}
	return sb.String()
}

// value sql value converter(sql inject process)
func sqlValue(arg any, driver *Driver) string {
	var ret string
	switch a := arg.(type) {
	case nil:
		ret = driver.Keyword("NULL")
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
			ret = driver.Keyword("TRUE")
		} else {
			ret = driver.Keyword("FALSE")
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
func mysqlName(name any) string {
	return quotedName(name, "`", "`")
}

func MakeNameFunc(prefix, suffix string) func(any) string {
	return func(name any) string {
		return quotedName(name, prefix, suffix)
	}
}

func quotedName(name any, prefix, suffix string) string {
	col := ""
	switch n := name.(type) {
	case string:
		col = prefix + n + suffix
	case fmt.Stringer:
		col = prefix + n.String() + suffix
	default:
		col = fmt.Sprintf("%s%v%s", prefix, n, suffix)
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
