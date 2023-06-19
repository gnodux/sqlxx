/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"github.com/cookieY/sqlx"
	"reflect"
)

func getTpl(m *Factory, templateList []string) string {
	if len(templateList) == 1 {
		return templateList[0]
	} else {
		for _, tpl := range templateList {
			if m.template.Lookup(tpl) != nil {
				return tpl
			}
		}
	}
	return ""
}
func SelectWith(m *Factory, p reflect.Type, db string, templateList []string, args []any) (any, error) {
	list := reflect.New(reflect.SliceOf(p))
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	tpl := getTpl(m, templateList)
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
	tpl := getTpl(m, templateList)
	err = d.NamedSelectTpl(list.Interface(), tpl, arg)
	return list.Elem().Interface(), err
}

func NamedGetWith(m *Factory, p reflect.Type, db string, templateList []string, arg any) (any, error) {
	var o reflect.Value
	if p.Kind() == reflect.Pointer {
		o = reflect.New(p.Elem())
	} else {
		o = reflect.New(p)
	}
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	tpl := getTpl(m, templateList)
	n, err := d.PrepareTplNamed(tpl, arg)
	if err != nil {
		return nil, err
	}
	defer func(n *sqlx.NamedStmt) {
		err = n.Close()
	}(n)
	err = n.Get(o.Interface(), arg)
	if p.Kind() == reflect.Pointer {
		return o.Interface(), err
	} else {
		return o.Elem().Interface(), err
	}
}

func GetWith(m *Factory, p reflect.Type, db string, templateList []string, args []any) (any, error) {
	var o reflect.Value
	if p.Kind() == reflect.Pointer {
		o = reflect.New(p.Elem())
	} else {
		o = reflect.New(p)
	}
	d, err := m.Get(db)
	if err != nil {
		return nil, err
	}
	tpl := getTpl(m, templateList)
	n, err := d.PrepareTpl(tpl, args)
	if err != nil {
		return nil, err
	}
	defer func(n *sqlx.Stmt) {
		err = n.Close()
	}(n)
	err = n.Get(o.Interface(), args...)
	if p.Kind() == reflect.Pointer {
		return o.Interface(), err
	} else {
		return o.Elem().Interface(), err
	}
}
