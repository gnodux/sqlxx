/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"github.com/cookieY/sqlx"
	"github.com/gnodux/sqlxx/builtinsql"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

var (
	ErrNilDriver = errors.New("driver is nil")
	ErrNilDB     = errors.New("DB is nil")
)

// DB database wrapper
type DB struct {
	m        *Factory
	template *template.Template
	lock     sync.Mutex
	driver   *Driver
	*sqlx.DB
}

func (d *DB) SetManager(m *Factory) {
	d.m = m
}

//func (d *DB) Parse(tplName string, args any) (string, error) {
//	if d == nil {
//		return "", ErrNilDB
//	}
//	if d.m == nil {
//		return "", ErrNoManager
//	}
//	return d.m.ParseSQL(tplName, args)
//}

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

func (d *DB) RunPrepared(tplName string, arg any, fn func(*sqlx.Stmt) error) (err error) {
	var stmt *sqlx.Stmt
	if stmt, err = d.PrepareTpl(tplName, arg); err != nil {
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
func (d *DB) RunPrepareNamed(tplName string, arg any, fn func(*sqlx.NamedStmt) error) (err error) {
	var stmt *sqlx.NamedStmt
	if stmt, err = d.PrepareTplNamed(tplName, arg); err != nil {
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
	log.Debug("named select tpl:", named.QueryString, args)
	return named.Select(dest, args)
}
func (d *DB) NamedSelect(dest interface{}, sql string, arg any) (err error) {
	if d == nil {
		return ErrNilDB
	}
	var named *sqlx.NamedStmt
	named, err = d.PrepareNamed(sql)
	if err != nil {
		return err
	}
	defer func(named *sqlx.NamedStmt) {
		err = named.Close()

	}(named)
	log.Debug("named select:", named.QueryString, arg)
	return named.Select(dest, arg)
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
	if err = fn(NewTxWith(tx, d, tpl)); err != nil {
		return
	}
	return
}

func (d *DB) SetTemplate(tpl *template.Template) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.template = tpl
}
func (d *DB) Template() *template.Template {
	return d.template
}

// ParseTemplateFS parse template from filesystem。
// 为了保留目录结构，没有直接使用template的ParseFS(template中的ParseFS方法不会保留路径名称)
func (d *DB) ParseTemplateFS(f fs.FS, patterns ...string) error {
	log.Info("parse template from filesystem: ", f, " with patterns:", patterns)
	for _, pattern := range patterns {
		matches, err := fs.Glob(f, pattern)
		if err != nil {
			return err
		}
		for _, mf := range matches {
			buf, err := fs.ReadFile(f, mf)
			if err != nil {
				return err
			}
			log.Info("parse sql:", mf)
			if _, err = d.template.New(strings.ReplaceAll(mf, "\\", "/")).Parse(string(buf)); err != nil {
				return err
			}
		}
	}
	return nil
}
func (d *DB) ParseTemplate(name string, tpl string) (*template.Template, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	t, err := d.template.New(name).Parse(tpl)
	return t, err
}
func (d *DB) Parse(sqlOrTpl string, args any) (query string, err error) {
	if !strings.HasSuffix(sqlOrTpl, sqlSuffix) {
		if strings.Contains(sqlOrTpl, "{{") && strings.Contains(sqlOrTpl, "}}") {
			name := fmt.Sprintf("%x", md5.Sum([]byte(sqlOrTpl)))
			t := d.template.Lookup(name)
			if t == nil {
				t, err = d.ParseTemplate(name, sqlOrTpl)
			}
			if err != nil {
				return
			}
			sb := &strings.Builder{}
			err = t.Execute(sb, args)
			if err == nil {
				query = sb.String()
			}

		} else {
			query = sqlOrTpl
		}

	} else {
		sb := &strings.Builder{}
		err = d.template.ExecuteTemplate(sb, sqlOrTpl, args)
		if err == nil {
			query = sb.String()
		}
	}
	log.Trace("parse sql:", sqlOrTpl, "=>", query, " with args:", args)
	return
}

func Open(driverName, datasource string) (*DB, error) {
	d, _ := Drivers[driverName]
	return OpenWith(StdFactory, d, datasource)
}

func OpenWith(f *Factory, driver *Driver, datasource string) (*DB, error) {
	if driver == nil {
		driver = f.driver
	}
	if driver == nil {
		return nil, ErrNilDriver
	}
	db, err := sqlx.Open(driver.Name, datasource)
	if err != nil {
		return nil, err
	}
	db.MapperFunc(NameFunc)

	newDb := &DB{DB: db, m: f, driver: driver, template: template.New("sql").Funcs(MakeFuncMap(driver))}
	err = newDb.ParseTemplateFS(builtinsql.Builtin, "**/*.sql")
	if err != nil {
		return nil, err
	}
	for _, tfs := range f.templateFS {
		err = newDb.ParseTemplateFS(tfs.FS, tfs.Patterns...)
		if err != nil {
			return nil, err
		}
	}
	return newDb, err
}
