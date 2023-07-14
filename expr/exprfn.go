/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"github.com/gnodux/sqlxx/expr/keywords"
	"strings"
)

type FilterFn func(Expr)

func SelectFn(fn func(s *SelectExpr)) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*SelectExpr); ok {
			fn(s)
		}
	}
}
func DeleteFn(fn func(s *DeleteExpr)) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*DeleteExpr); ok {
			fn(s)
		}
	}
}
func UpdateFn(fn func(s *UpdateExpr)) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*UpdateExpr); ok {
			fn(s)
		}
	}
}

func Limit(limit int) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*SelectExpr); ok {
			s.limit = limit
		}
	}
}
func Offset(offset int) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*SelectExpr); ok {
			s.offset = offset
		}
	}
}

func UseCount(exp Expr) {
	if s, ok := exp.(*SelectExpr); ok {
		s.withCount = true
	}
}

func UseOr(exp Expr) {
	var where Expr
	switch ex := exp.(type) {
	case *SelectExpr:
		where = ex.WhereExpr
	case *DeleteExpr:
		where = ex.WhereExpr
	case *UpdateExpr:
		where = ex.WhereExpr
	}
	if where == nil {
		if lst, ok := where.(*ListExpr); ok {
			lst.Placeholder = keywords.Space
			lst.Separator = keywords.Or
		}
	}
}

func UseAnd(exp Expr) {
	var where Expr
	switch ex := exp.(type) {
	case *SelectExpr:
		where = ex.WhereExpr
	case *DeleteExpr:
		where = ex.WhereExpr
	case *UpdateExpr:
		where = ex.WhereExpr
	}
	if where == nil {
		if lst, ok := where.(*ListExpr); ok {
			lst.Placeholder = keywords.Space
			lst.Separator = keywords.And
		}
	}
}

func UseCondition(exp Expr) FilterFn {
	return func(s Expr) {
		switch ss := s.(type) {
		case *SelectExpr:
			ss.WhereExpr = exp
		case *DeleteExpr:
			ss.WhereExpr = exp
		case *UpdateExpr:
			ss.WhereExpr = exp
		}
	}
}

// AutoFuzzy 自动模糊查询，对于字符类型且包含%.?*等字符的，自动转换为like查询
func AutoFuzzy(exp Expr) {
	switch s := exp.(type) {
	case *SelectExpr:
		fuzzy(s.WhereExpr)
	case *DeleteExpr:
		fuzzy(s.WhereExpr)
	case *UpdateExpr:
		fuzzy(s.WhereExpr)
	}
}

func fuzzy(exp Expr) {
	if exp == nil {
		return
	}
	switch n := exp.(type) {
	case *ListExpr:
		for _, v := range n.ExprList {
			fuzzy(v)
		}
	case *BinaryExpr:
		switch right := n.Right.(type) {
		case *ValueExpr:
			if isStringAndFuzzy(right.Value) {
				n.Space = keywords.Space
				n.Operator = keywords.Like
			}
		case *ConstantExpr:
			if isStringAndFuzzy(right.Value) {
				n.Space = keywords.Space
				n.Operator = keywords.Like
			}
		}
	case *ParenExpr:
		fuzzy(n.Expr)
	}
}
func isStringAndFuzzy(v any) bool {
	switch vv := v.(type) {
	case string:
		if strings.IndexAny(vv, "%?.") >= 0 {
			return true
		}
	}
	return false
}

func Set(exps ...Expr) FilterFn {
	return UpdateFn(func(s *UpdateExpr) {
		s.Values = exps
	})
}

func UseSort(direct string, exprs ...Expr) FilterFn {
	return SelectFn(func(s *SelectExpr) {
		s.OrderByExpr = Sorts(direct, exprs...)
	})
}
func UseOrderBy(exp Expr) FilterFn {
	return SelectFn(func(s *SelectExpr) {
		s.OrderByExpr = exp
	})
}
