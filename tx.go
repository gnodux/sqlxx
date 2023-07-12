/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"database/sql"
	"github.com/cookieY/sqlx"
)

// Tx transaction wrapper
type Tx struct {
	*sqlx.Tx
	db  *DB
	tpl string
}

func (t *Tx) Tpl() string {
	return t.tpl
}
func (t *Tx) Parse(tplName string, args any) (string, error) {
	if t.db == nil {
		return "", ErrNilDB
	}
	return t.db.Parse(tplName, args)
}

// ParseAndPrepareNamed use tplName to parse and prepare named statement
func (t *Tx) ParseAndPrepareNamed(tplName string, arg any) (*sqlx.NamedStmt, error) {
	query, err := t.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	return t.PrepareNamed(query)
}

// RunPrepareNamedTpl use tplName to prepare named statement
func (t *Tx) RunPrepareNamedTpl(tplName string, arg any, fn func(*sqlx.NamedStmt) error) (err error) {
	var stmt *sqlx.NamedStmt
	if stmt, err = t.ParseAndPrepareNamed(tplName, arg); err != nil {
		return
	}
	defer func() {
		stErr := stmt.Close()
		if stErr != nil {
			err = stErr
		}
	}()
	return fn(stmt)
}

// RunCurrentPrepareNamed use current tpl to prepare named statement
func (t *Tx) RunCurrentPrepareNamed(arg any, fn func(*sqlx.NamedStmt) error) (err error) {
	return t.RunPrepareNamedTpl(t.tpl, arg, fn)
}

func (t *Tx) ParseAndPrepare(tplName string, arg any) (*sqlx.Stmt, error) {
	query, err := t.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	return t.Preparex(query)
}
func (t *Tx) RunPreparedTpl(tplName string, arg any, fn func(*sqlx.Stmt) error) (err error) {
	var stmt *sqlx.Stmt
	if stmt, err = t.ParseAndPrepare(tplName, arg); err != nil {
		return
	}
	defer func() {
		stErr := stmt.Close()
		if stErr != nil {
			err = stErr
		}
	}()
	return fn(stmt)
}
func (t *Tx) RunCurrentPrepared(arg any, fn func(*sqlx.Stmt) error) (err error) {
	return t.RunPreparedTpl(t.tpl, arg, fn)
}

func (t *Tx) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	log.Debug("prepare named query:", query)
	return t.Tx.PrepareNamed(query)
}
func (t *Tx) Preparex(query string) (*sqlx.Stmt, error) {
	log.Debug("prepare query:", query)
	return t.Tx.Preparex(query)
}

// NamedExecTpl  use tpl to query named statement
func (t *Tx) NamedExecTpl(tplName string, arg interface{}) (sql.Result, error) {
	query, err := t.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named exec tpl:", query, arg)
	return t.NamedExec(query, arg)
}

func (t *Tx) ExecTpl(tplName string, args ...interface{}) (sql.Result, error) {
	query, err := t.Parse(tplName, args)
	if err != nil {
		return nil, err
	}
	log.Debug("exec query:", query, args)
	return t.Exec(query, args...)
}

// ExecCurrent use current tpl to exec
func (t *Tx) ExecCurrent(args ...interface{}) (sql.Result, error) {
	return t.ExecTpl(t.tpl, args...)
}

// NamedExecCurrent use current tpl to exec named statement
func (t *Tx) NamedExecCurrent(arg interface{}) (sql.Result, error) {
	return t.NamedExecTpl(t.tpl, arg)
}

func (t *Tx) GetTpl(dest any, tpl string, args ...any) error {
	query, err := t.Parse(tpl, args)
	if err != nil {
		return err
	}
	log.Debug("get query:", query, args)
	return t.Get(dest, query, args...)
}

func NewTxWith(tx *sqlx.Tx, d *DB, tpl string) *Tx {
	return &Tx{
		Tx:  tx,
		db:  d,
		tpl: tpl,
	}
}
