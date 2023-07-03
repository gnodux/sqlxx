/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func TestTemplateFunc(t *testing.T) {

	type args struct {
		tpl string
		arg any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "where desc 1",
			args: args{
				tpl: "desc.sql",
				arg: map[string]any{
					"cols": []string{
						"d", "e", "f",
					},
					"where": map[string]any{
						"name":      "xudong",
						"isDeleted": false,
					},
				},
			},
		}, {
			name: "where desc 2",
			args: args{tpl: "desc.sql", arg: map[string]any{
				"cols": []string{"a", "c"},
				"where": struct {
					Name string
					Age  int
					Mop  string
				}{
					Name: "xudong",
					Age:  34,
				},
			},
			},
		}, {
			name: "in test",
			args: args{tpl: "in.sql", arg: map[string]any{
				"roles": []any{"admin", "user", "custom"},
			}},
		}, {
			name: "range test",
			args: args{
				tpl: "range.sql",
				arg: map[string]string{
					"tenantId": "10010011",
					"name":     "x",
				},
			},
		}, {
			name: "var define test",
			args: args{
				tpl: "vardefine.sql",
				arg: map[string]any{
					"tenantId": "10010011",
					"name":     "x",
				},
			},
		},
	}
	tpl := template.New("tests").Funcs(DefaultFuncMap)
	_, err := tpl.ParseFS(os.DirFS("testdata/templates"), "*.sql")
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &strings.Builder{}
			err := tpl.ExecuteTemplate(buf, tt.args.tpl, tt.args.arg)
			assert.NoError(t, err)
			if tt.want != "" {
				assert.Equal(t, tt.want, buf.String())
			}
			fmt.Println(buf)
		})
	}
}
func TestMapLength(t *testing.T) {
	m := map[string]any{}
	mv := reflect.ValueOf(m)
	if mv.IsZero() {
		fmt.Println("zero")
	}
	fmt.Println(mv.Len())
}
