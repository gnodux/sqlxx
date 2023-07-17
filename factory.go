/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"fmt"
	"github.com/gnodux/sqlxx/dialect"
	"github.com/gnodux/sqlxx/utils"
	"io/fs"
	"sync"
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

	//SetTemplateFS set sql template from filesystem
	SetTemplateFS = StdFactory.SetTemplateFS

	//ClearTemplateFS clear sql template from filesystem
	ClearTemplateFS = StdFactory.ClearTemplateFS

	//Shutdown manager and close all db
	Shutdown = StdFactory.Shutdown

	////SetTemplate set sql template
	//SetTemplate = StdFactory.SetTemplate

	//ParseTemplateFS set sql template from filesystem
	//ParseTemplateFS = StdFactory.ParseTemplateFS

	//ParseTemplate create a new template
	//ParseTemplate = StdFactory.ParseTemplate
)

type DBConstructor func() (*DB, error)

type TplFS struct {
	FS       fs.FS
	Patterns []string
}

type Factory struct {
	driver       *dialect.Driver
	name         string
	dbs          map[string]*DB
	constructors map[string]DBConstructor
	lock         *sync.RWMutex
	templateFS   []*TplFS
}

func NewFactoryWithDriver(name string, driver *dialect.Driver) *Factory {
	f := &Factory{
		name:         name,
		driver:       driver,
		dbs:          map[string]*DB{},
		constructors: map[string]DBConstructor{},
		lock:         &sync.RWMutex{},
		//template:     template.New("sql").Funcs(MakeFuncMap(driver)),
	}
	//err := f.ParseTemplateFS(builtinsql.Builtin, "**/*.sql")
	//if err != nil {
	//	panic(err)
	//}
	return f
}
func NewFactory(name string) *Factory {
	return NewFactoryWithDriver(name, DefaultDriver)
}

func (m *Factory) SetTemplateFS(f fs.FS, patterns ...string) {
	m.templateFS = append(m.templateFS, &TplFS{
		FS:       f,
		Patterns: patterns,
	})
}
func (m *Factory) ClearTemplateFS() {
	m.templateFS = nil
}

//Get 获取一个数据库连接
//name: 数据库连接名称

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

// MustGet 获取一个数据库连接，如果不存在则panic
func (m *Factory) MustGet(name string) *DB {
	return utils.Must(m.Get(name))
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

// Open 打开一个数据库连接
func (m *Factory) Open(name, driverName, dsn string) (*DB, error) {
	db, err := OpenWith(m, Drivers[driverName], dsn)
	if err != nil {
		return nil, err
	}
	m.Set(name, db)
	return db, nil
}

// Set set a database
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
