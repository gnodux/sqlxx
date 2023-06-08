package sqlxx

import (
	"github.com/cookieY/sqlx"
	"reflect"
)

func SelectWith(m *Factory, p reflect.Type, db string, templateList []string, args []any) (any, error) {
	list := reflect.New(reflect.SliceOf(p))
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	var tpl string
	for _, tpl = range templateList {
		if m.template.Lookup(tpl) != nil {
			break
		}
	}
	err = d.SelectTpl(list.Interface(), tpl, args...)
	if err != nil {
		return nil, err
	}

	return list.Elem().Interface(), err
}
func NamedSelectWith(m *Factory, p reflect.Type, db string, templateList []string, arg any) (any, error) {
	list := reflect.New(reflect.SliceOf(p))
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	var tpl string
	for _, tpl = range templateList {
		if m.template.Lookup(tpl) != nil {
			break
		}
	}
	err = d.NamedSelectTpl(list.Interface(), tpl, arg)
	return list.Elem().Interface(), err
}

func NamedGetWith(m *Factory, p reflect.Type, db string, templateList []string, arg any) (any, error) {
	o := reflect.New(p)
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	var tpl string
	for _, tpl = range templateList {
		if m.template.Lookup(tpl) != nil {
			break
		}
	}
	n, err := d.PrepareTplNamed(tpl, arg)
	if err != nil {
		return nil, err
	}
	defer func(n *sqlx.NamedStmt) {
		err = n.Close()
	}(n)
	err = n.Get(o.Interface(), arg)
	return o.Elem().Interface(), err
}

func GetWith(m *Factory, p reflect.Type, db string, templateList []string, args []any) (any, error) {
	o := reflect.New(p)
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	var tpl string
	for _, tpl = range templateList {
		if m.template.Lookup(tpl) != nil {
			break
		}
	}
	n, err := d.PrepareTpl(tpl, args)
	if err != nil {
		return nil, err
	}
	defer func(n *sqlx.Stmt) {
		err = n.Close()
	}(n)
	err = n.Get(o.Interface(), args...)
	return o.Elem().Interface(), err
}
