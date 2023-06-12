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
		"asc":   Asc,
		"desc":  Desc,
		"v":     Value,
		"n":     SQLName,
		"list":  Values,
	}
)

func Where(arg any) string {
	return WhereWith(arg, " AND ")
}
func WhereOr(arg any) string {
	return WhereWith(arg, " OR ")
}
func WhereWith(args any, op string) string {
	if args == nil {
		return ""
	}
	if op == "" {
		op = " AND "
	}
	if op[0] != ' ' {
		op = " " + op + " "
	}

	buf := strings.Builder{}
	argv := reflect.ValueOf(args)
	switch argv.Type().Kind() {
	case reflect.Map:
		comma := " WHERE "
		for _, k := range argv.MapKeys() {
			buf.WriteString(comma)
			buf.WriteString(SQLName(LowerCase(k.String())))
			buf.WriteString("=")
			buf.WriteString(Value(argv.MapIndex(k).Interface()))
			comma = op
		}
	case reflect.Struct:
		comma := " WHERE "
		for i := 0; i < argv.NumField(); i++ {
			if argv.Field(i).IsNil() || argv.Field(i).IsZero() {
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
func OrderBy(direction string, args []any) string {
	if len(args) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString("ORDER BY ")
	splitter := ""
	for _, arg := range args {
		sb.WriteString(splitter)
		sb.WriteString(SQLName(arg))
		splitter = ","
	}
	sb.WriteString(" ")
	sb.WriteString(direction)
	return sb.String()
}
func Asc(args []any) string {
	return OrderBy("ASC", args)
}
func Desc(args []any) string {
	return OrderBy("DESC", args)
}

// Values list of values
func Values(args []any) string {
	comma := ""
	sb := &strings.Builder{}
	for _, arg := range args {
		sb.WriteString(comma)
		sb.WriteString(Value(arg))
		comma = ","
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
