/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

// Package expr package is used to generate sql expression
// 过于复杂，目前不考虑实现
package expr

import (
	"reflect"
	"strings"
)

// TracedBuffer is a buffer that can be used to trace the
type TracedBuffer struct {
	*strings.Builder
}

type Expr interface {
	Format(buffer *TracedBuffer)
}

func toMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
		if vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}
	}
	result := map[string]any{}
	typ := reflect.TypeOf(vv.Interface())
	for idx := 0; idx < vv.NumField(); idx++ {
		f := vv.Field(idx)
		ft := typ.Field(idx)
		if ft.IsExported() && !f.IsZero() {
			if ft.Anonymous {
				for k, v := range toMap(f.Interface()) {
					result[k] = v
				}
			} else {
				result[ft.Name] = f.Interface()
			}
		}
	}

	return result
}
func mergeMap(target map[string]any, sources ...map[string]any) {
	if target == nil {
		return
	}
	for _, source := range sources {
		if source == nil {
			continue
		}
		for k, v := range source {
			target[k] = v
		}
	}
}
