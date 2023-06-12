/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"github.com/stretchr/testify/assert"
	"os"
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
			name: "desc function",
			args: args{
				tpl: "desc.sql",
				arg: map[string]any{
					"cols": []any{
						"d", "e", "f",
					},
					"where": map[string]any{
						"name":      "xudong",
						"isDeleted": false,
					},
				},
			},
		},
	}
	tpl := template.New("tests").Funcs(DefaultFuncMap)
	_, err := tpl.ParseFS(os.DirFS("test_templates"), "*.sql")
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &strings.Builder{}
			err := tpl.ExecuteTemplate(buf, tt.args.tpl, tt.args.arg)
			assert.NoError(t, err)
			if tt.want != "" {
				assert.Equal(t, tt.want, buf.String())
			}
			os.Stdout.WriteString(buf.String())
		})
	}
}
