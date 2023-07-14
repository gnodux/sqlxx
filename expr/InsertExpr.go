/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

// InsertExpr is a struct for insert expression
type InsertExpr struct {
	Table      Expr
	ValueExprs []Expr
}

func (i *InsertExpr) Insert(table Expr) *InsertExpr {
	i.Table = table
	return i
}
func (i *InsertExpr) Values(values ...Expr) *InsertExpr {
	i.ValueExprs = values
	return i
}

// Insert is a function to create InsertExpr
func Insert(table Expr) *InsertExpr {
	return &InsertExpr{Table: table}
}
