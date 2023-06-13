/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cookieY/sqlx"
)

var (
	ErrNoManager = errors.New("no manager")
	ErrNilDB     = errors.New("DB is nil")
)

// DB database wrapper
type DB struct {
	m *Factory
	*sqlx.DB
}

func (d *DB) SetManager(m *Factory) {
	d.m = m
}

func (d *DB) Parse(tplName string, args any) (string, error) {
	if d == nil {
		return "", ErrNilDB
	}
	if d.m == nil {
		return "", ErrNoManager
	}
	return d.m.ParseSQL(tplName, args)
}

func (d *DB) PrepareTpl(tplName string, args any) (*sqlx.Stmt, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.Parse(tplName, args)
	if err != nil {
		return nil, err
	}
	return d.Preparex(query)
}
func (d *DB) PrepareTplNamed(tplName string, args any) (*sqlx.NamedStmt, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.Parse(tplName, args)
	if err != nil {
		return nil, err
	}
	return d.PrepareNamed(query)
}
func (d *DB) SelectTpl(dest interface{}, tplName string, args ...any) error {
	if d == nil {
		return ErrNilDB
	}
	query, err := d.Parse(tplName, args)
	if err != nil {
		return err
	}
	log.Debug("select:", query, args)
	return d.DB.Select(dest, query, args...)
}
func (d *DB) NamedSelectTpl(dest interface{}, tplName string, args interface{}) (err error) {
	if d == nil {
		return ErrNilDB
	}
	var named *sqlx.NamedStmt
	named, err = d.PrepareTplNamed(tplName, args)
	if err != nil {
		return err
	}
	defer func(named *sqlx.NamedStmt) {
		err = named.Close()

	}(named)
	if args == nil {
		args = map[string]any{}
	}
	log.Debug("named select:", named.QueryString, args)
	return named.Select(dest, args)
}
func (d *DB) NamedExecTpl(tplName string, arg interface{}) (sql.Result, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named exec:", query, arg)
	return d.NamedExec(query, arg)
}

func (d *DB) ExecTpl(tplName string, args ...interface{}) (sql.Result, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.Parse(tplName, args)
	if err != nil {
		return nil, err
	}
	log.Debug("exec:", query, args)
	return d.Exec(query, args...)
}
func (d *DB) NamedQueryTpl(tplName string, arg interface{}) (*sqlx.Rows, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named query:", query, arg)
	return d.NamedQuery(query, arg)
}
func (d *DB) Batch(ctx context.Context, opts *sql.TxOptions, fn func(tx *Tx) error) (err error) {
	return d.BatchTpl(ctx, opts, "", fn)
}
func (d *DB) BatchTpl(ctx context.Context, opts *sql.TxOptions, tpl string, fn func(tx *Tx) error) (err error) {
	if d == nil {
		return ErrNilDB
	}
	var tx *sqlx.Tx
	tx, err = d.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}
	}()
	if err = fn(NewTx(tx, d.m, tpl)); err != nil {
		return
	}
	return
}
func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	db.MapperFunc(LowerCase)
	return &DB{DB: db}, err
}
