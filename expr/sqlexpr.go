/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

// Package expr package is used to generate sql expression
// 过于复杂，目前不考虑实现
package expr

import "strings"

// TracedBuffer is a buffer that can be used to trace the
type TracedBuffer struct {
	*strings.Builder
}

type Expr interface {
	Format(buffer *TracedBuffer)
}
