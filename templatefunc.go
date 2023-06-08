package sqlxx

import (
	"fmt"
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
	if op == "" {
		op = " AND "
	}
	if op[0] != ' ' {
		op = " " + op + " "
	}

	switch v := args.(type) {
	case nil:
		return ""
	case map[string]any:
		buf := strings.Builder{}
		comma := "WHERE "
		for k, vv := range v {
			if vv == nil {
				continue
			}
			buf.WriteString(comma)
			buf.WriteString(k)
			buf.WriteString("=:")
			buf.WriteString(k)
			comma = op
		}
		return buf.String()
	}
	return fmt.Sprintf("%v", args)
}
func OrderBy(direction string, args []string) string {
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
func Asc(args []string) string {
	return OrderBy("ASC", args)
}
func Desc(args ...string) string {
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
func SQLName(name string) string {
	return "`" + name + "`"
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
