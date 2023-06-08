package sqlxx

import (
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

var (
	//StdFactory default connection manager
	StdFactory = NewFactory("Default")
	//Get a db from
	Get = StdFactory.Get
	//MustGet a db,if db not exists,raise a panic
	MustGet = StdFactory.MustGet
	//Set a db
	Set = StdFactory.Set
	//SetFunc set a db with loader func
	SetFunc = StdFactory.SetFunc
	//New initialize a db
	New = StdFactory.New

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

type LoadFunc func() (*DB, error)

type Factory struct {
	name     string
	dbs      map[string]*DB
	loader   map[string]LoadFunc
	lock     *sync.RWMutex
	template *template.Template
}

func NewFactory(name string) *Factory {
	return &Factory{
		name:     name,
		dbs:      map[string]*DB{},
		loader:   map[string]LoadFunc{},
		lock:     &sync.RWMutex{},
		template: template.New("sql").Funcs(DefaultFuncMap),
	}
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
			if _, err = m.template.New(strings.ReplaceAll(mf, "\\", "/")).Parse(string(buf)); err != nil {
				return err
			}
		}
	}
	return nil
}
func (m *Factory) ParseTemplate(name string, tpl string) (*template.Template, error) {
	t, err := m.template.New(name).Parse(tpl)
	return t, err
}
func (m *Factory) ParseSQL(name string, args any) (query string, err error) {
	sb := &strings.Builder{}
	err = m.template.ExecuteTemplate(sb, name, args)
	query = sb.String()
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
		if loader, lok := m.loader[name]; lok {
			err := func() error {
				m.lock.Lock()
				defer m.lock.Unlock()
				var err error
				conn, err = loader()
				conn.SetManager(m)
				conn.MapperFunc(LowerCase)
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

// New 创建新的数据库连接，New和SetFunc的不同在：New是立即创建数据库实例，但SetFunc会延迟创建实例
func (m *Factory) New(name string, fn LoadFunc) error {
	d, err := fn()
	if err != nil {
		return err
	}
	m.Set(name, d)
	return nil
}
func (m *Factory) Set(name string, db *DB) {
	m.lock.Lock()
	defer m.lock.Unlock()
	db.m = m
	m.dbs[name] = db
}
func (m *Factory) SetFunc(name string, loadFunc LoadFunc) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.loader[name] = loadFunc
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
