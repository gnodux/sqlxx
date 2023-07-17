/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

// Package expr package 用于生成sql语句
package expr

//go:generate go run genindx.go

import (
	"github.com/gnodux/sqlxx/dialect"
	"strings"
)

// TracedBuffer is a buffer that can be used to trace the
type TracedBuffer struct {
	//namedArgs 命名参数
	namedArgs map[string]any
	//args 位置参数
	args     []any
	NamedVar bool
	*dialect.Driver
	strings.Builder
}

func NewTracedBuffer(driver *dialect.Driver) *TracedBuffer {
	return &TracedBuffer{Driver: driver, NamedVar: true}
}

func (t *TracedBuffer) AppendNamedArg(name string, value any) *TracedBuffer {
	if t.namedArgs == nil {
		t.namedArgs = map[string]any{}
	}
	t.namedArgs[name] = value
	return t
}
func (t *TracedBuffer) AppendArg(value any) *TracedBuffer {
	t.args = append(t.args, value)
	return t
}
func (t *TracedBuffer) Append(buf []byte) *TracedBuffer {
	t.Builder.Write(buf)
	return t
}

func (t *TracedBuffer) NewLine() *TracedBuffer {
	t.Builder.WriteString("\n")
	return t
}
func (t *TracedBuffer) AppendString(s string) *TracedBuffer {
	t.Builder.WriteString(s)
	return t
}
func (t *TracedBuffer) AppendExprs(exprs ...Expr) *TracedBuffer {
	for _, expr := range exprs {
		expr.Format(t)
	}
	return t
}

func (t *TracedBuffer) AppendKeyword(keyword string) *TracedBuffer {
	t.Builder.WriteString(t.Keyword(keyword))
	return t
}
func (t *TracedBuffer) AppendKeywordWithSpace(keyword string) *TracedBuffer {
	t.Builder.WriteString(t.KeywordWithSpace(keyword))
	return t
}
func (t *TracedBuffer) Build(exp Expr) (string, []any, error) {
	t.NamedVar = false
	t.Builder.Reset()
	exp.Format(t)
	return t.Builder.String(), t.args, nil
}
func (t *TracedBuffer) BuildNamed(exp Expr) (string, map[string]any, error) {
	t.NamedVar = true
	t.Builder.Reset()
	exp.Format(t)
	return t.Builder.String(), t.namedArgs, nil
}

type Expr interface {
	Format(buffer *TracedBuffer)
}
