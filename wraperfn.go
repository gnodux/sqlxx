/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"context"
	"database/sql"
	"github.com/cookieY/sqlx"
)

const (
	sqlSuffix = ".sql"
)

type SelectFunc[T any] func(args ...any) ([]T, error)
type NamedSelectFunc[T any] func(arg any) ([]T, error)
type GetFunc[T any] func(args ...any) (T, error)
type NamedGetFunc[T any] func(arg any) (T, error)
type ExecFunc func(args ...any) (sql.Result, error)
type NamedExecFunc func(arg any) (sql.Result, error)
type TxFunc func(func(*Tx) error) error

func SelectFnWith[T any](m *Factory, db, tpl string) SelectFunc[T] {
	return func(args ...any) ([]T, error) {
		d, err := m.Get(db)
		if err != nil {
			return nil, err
		}
		var v []T
		err = d.SelectTpl(&v, tpl, args...)
		return v, err
	}
}

func SelectFn[T any](db, tpl string) SelectFunc[T] {
	return SelectFnWith[T](StdFactory, db, tpl)
}

func TxFnWith(db *DB, tpl string, opts *sql.TxOptions) TxFunc {
	return func(fn func(tx *Tx) error) error {
		return db.BatchTpl(context.Background(), opts, tpl, fn)
	}
}

func NamedSelectFnWith[T any](manager *Factory, db, tpl string) NamedSelectFunc[T] {
	return func(arg any) ([]T, error) {
		var v []T
		d, err := manager.Get(db)
		if err != nil {
			return nil, err
		}
		err = d.NamedSelectTpl(&v, tpl, arg)
		return v, err
	}
}

func NamedSelectFn[T any](db, tpl string) NamedSelectFunc[T] {
	return NamedSelectFnWith[T](StdFactory, db, tpl)
}

//func GetFnWith[T any](m *Factory, db, tpl string) GetFunc[T] {
//	return func(args ...any) (v T, err error) {
//		var d *DB
//		d, err = m.Get(db)
//		if err != nil {
//			return
//		}
//		n, err := d.PrepareTpl(tpl, args)
//		if err != nil {
//			return v, err
//		}
//		defer func(n *sqlx.Stmt) {
//			err = n.Close()
//		}(n)
//		err = n.Get(&v, args)
//		return v, err
//	}
//}

//	func GetFn[T any](db, tpl string) GetFunc[T] {
//		return GetFnWith[T](StdFactory, db, tpl)
//	}
func NamedGetFnWith[T any](m *Factory, db, tpl string) NamedGetFunc[T] {
	return func(arg any) (v T, err error) {
		var d *DB
		if d, err = m.Get(db); err != nil {
			return
		}
		var n *sqlx.NamedStmt
		if n, err = d.PrepareTplNamed(tpl, arg); err != nil {
			return v, err
		}
		defer func(n *sqlx.NamedStmt) {
			err = n.Close()
		}(n)
		err = n.Get(&v, arg)
		return v, err
	}
}

//func NamedGetFn[T any](db, tpl string) NamedGetFunc[T] {
//	return NamedGetFnWith[T](StdFactory, db, tpl)
//}

func NamedExecFnWith(db *DB, tpl string) NamedExecFunc {
	return func(arg any) (sql.Result, error) {
		return db.NamedExecTpl(tpl, arg)
	}
}

//func NamedExecFn(db, tpl string) NamedExecFunc {
//	return NamedExecFnWith(StdFactory, db, tpl)
//}

func ExecFnWith(db *DB, tpl string) ExecFunc {
	return func(args ...any) (sql.Result, error) {
		return db.ExecTpl(tpl, args...)
	}
}

//func ExecFn(db, tpl string) ExecFunc {
//	return ExecFnWith(StdFactory, db, tpl)
//}
