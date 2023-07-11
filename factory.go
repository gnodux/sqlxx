/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"crypto/md5"
	"fmt"
	"github.com/gnodux/sqlxx/builtinsql"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

var (
	DefaultName = "Default"
	//StdFactory default connection manager
	StdFactory = NewFactory(DefaultName)
	//Get a db from
	Get = StdFactory.Get
	//MustGet a db,if db not exists,raise a panic
	MustGet = StdFactory.MustGet
	//Set a db
	Set = StdFactory.Set
	//SetConstructor set a db with constructors func
	SetConstructor = StdFactory.SetConstructor
	//CreateAndSet initialize a db
	CreateAndSet = StdFactory.CreateAndSet

	OpenDB = StdFactory.Open

	//Shutdown manager and close all db
	Shutdown = StdFactory.Shutdown

	//SetTemplate set sql template
	SetTemplate = StdFactory.SetTemplate

	//ParseTemplateFS set sql template from filesystem
	ParseTemplateFS = StdFactory.ParseTemplateFS

	//ParseTemplate create a new template
	ParseTemplate = StdFactory.ParseTemplate

	//ParseSQL parse sql from template
	ParseSQL = StdFactory.ParseSQL
)

type DBConstructor func() (*DB, error)

type Factory struct {
	driver       *Driver
	name         string
	dbs          map[string]*DB
	constructors map[string]DBConstructor
	lock         *sync.RWMutex
	template     *template.Template
}

func NewFactoryWithDriver(name string, driver *Driver) *Factory {
	f := &Factory{
		name:         name,
		driver:       driver,
		dbs:          map[string]*DB{},
		constructors: map[string]DBConstructor{},
		lock:         &sync.RWMutex{},
		template:     template.New("sql").Funcs(MakeFuncMap(driver)),
	}
	err := f.ParseTemplateFS(builtinsql.Builtin, "**/*.sql")
	if err != nil {
		panic(err)
	}
	return f
}
func NewFactory(name string) *Factory {
	return NewFactoryWithDriver(name, DefaultDriver)
}

func (m *Factory) SetTemplate(tpl *template.Template) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.template = tpl
}
func (m *Factory) Template() *template.Template {
	return m.template
}

// ParseTemplateFS parse template from filesystem。
// 为了保留目录结构，没有直接使用template的ParseFS(template中的ParseFS方法不会保留路径名称)
func (m *Factory) ParseTemplateFS(f fs.FS, patterns ...string) error {
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
			if _, err = m.template.New(strings.ReplaceAll(mf, "\\", "/")).Parse(string(buf)); err != nil {
				return err
			}
		}
	}
	return nil
}
func (m *Factory) ParseTemplate(name string, tpl string) (*template.Template, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	t, err := m.template.New(name).Parse(tpl)
	return t, err
}
func (m *Factory) ParseSQL(sqlOrTpl string, args any) (query string, err error) {
	if !strings.HasSuffix(sqlOrTpl, ".sql") {
		if strings.Contains(sqlOrTpl, "{{") && strings.Contains(sqlOrTpl, "}}") {
			name := fmt.Sprintf("%x", md5.Sum([]byte(sqlOrTpl)))
			t := m.template.Lookup(name)
			if t == nil {
				t, err = m.ParseTemplate(name, sqlOrTpl)
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
		err = m.template.ExecuteTemplate(sb, sqlOrTpl, args)
		if err == nil {
			query = sb.String()
		}
	}
	return
}
func (m *Factory) Get(name string) (*DB, error) {
	conn, ok := func() (*DB, bool) {
		m.lock.RLock()
		defer m.lock.RUnlock()
		c, o := m.dbs[name]
		return c, o
	}()
	if !ok {
		if loader, lok := m.constructors[name]; lok {
			err := func() error {
				m.lock.Lock()
				defer func() {
					//无论是否成功，都移除loader，避免反复初始化导致异常
					delete(m.constructors, name)
					m.lock.Unlock()
				}()
				var err error
				conn, err = loader()
				conn.SetManager(m)
				conn.MapperFunc(NameFunc)
				if err != nil {
					return err
				} else {
					m.dbs[name] = conn
				}
				return nil
			}()
			if err != nil {
				return nil, fmt.Errorf("initialize database %s error:%s", name, err)
			}
		} else {
			return nil, fmt.Errorf("database %s not found in %s", name, m.name)
		}
	}
	return conn, nil
}
func (m *Factory) MustGet(name string) *DB {
	d, err := m.Get(name)
	if err != nil {
		panic(err)
	}
	return d
}

// CreateAndSet 创建新的数据库连接并放入管理器中
func (m *Factory) CreateAndSet(name string, fn DBConstructor) (*DB, error) {
	d, err := fn()
	if err != nil {
		return nil, err
	}
	m.Set(name, d)
	return d, nil
}
func (m *Factory) Open(name, dsn string) (*DB, error) {
	db, err := OpenWith(m, dsn)
	if err != nil {
		return nil, err
	}
	m.Set(name, db)
	return db, nil
}

func (m *Factory) Set(name string, db *DB) {
	m.lock.Lock()
	defer m.lock.Unlock()
	db.m = m
	m.dbs[name] = db
}

// SetConstructor set a database constructor(Lazy create DB)
func (m *Factory) SetConstructor(name string, loadFunc DBConstructor) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.constructors[name] = loadFunc
}

func (m *Factory) BoostMapper(dest any, dataSource string) error {
	return BoostMapper(dest, m, dataSource)
}
func (m *Factory) Shutdown() error {
	for _, v := range m.dbs {
		if err := v.Close(); err != nil {
			return err
		}
	}
	return nil
}
func (m *Factory) String() string {
	return fmt.Sprintf("db[%s]", m.name)
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
