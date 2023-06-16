/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	TagDS             = "ds"
	TagSQL            = "sql"
	TagTx             = "tx"
	TagReadonly       = "readonly"
	TxDefault         = "Default"
	TxReadUncommitted = "ReadUncommitted"
	TxReadCommitted   = "ReadCommitted"
	TxWriteCommitted  = "WriteCommitted"
	TxRepeatableRead  = "RepeatableRead"
	TxSnapshot        = "Snapshot"
	TxSerializable    = "Serializable"
	TxLinearizable    = "Linearizable"
)

var (
	ManagerType       = reflect.TypeOf((*Factory)(nil))
	DBType            = reflect.TypeOf((*DB)(nil))
	ExecFuncType      = reflect.TypeOf(ExecFunc(nil))
	NamedExecFuncType = reflect.TypeOf(NamedExecFunc(nil))
	TxFuncType        = reflect.TypeOf(TxFunc(nil))
)

func getTags(field reflect.StructField) (ds, tpl string, level sql.IsolationLevel, readOnly bool) {
	ds = field.Tag.Get(TagDS)
	tpl = field.Tag.Get(TagSQL)
	txtLevel := field.Tag.Get(TagTx)
	switch txtLevel {
	case TxReadCommitted:
		level = sql.LevelReadCommitted
	case TxReadUncommitted:
		level = sql.LevelReadUncommitted
	case TxWriteCommitted:
		level = sql.LevelWriteCommitted
	case TxRepeatableRead:
		level = sql.LevelRepeatableRead
	case TxSnapshot:
		level = sql.LevelSnapshot
	case TxSerializable:
		level = sql.LevelSerializable
	case TxLinearizable:
		level = sql.LevelLinearizable
	default:
		level = sql.LevelDefault
	}
	r := field.Tag.Get(TagReadonly)
	if r != "" && strings.ToLower(r) != "false" {
		readOnly = true
	}
	return
}
func BoostMapper(dest interface{}, m *Factory, ds string) error {
	v := reflect.ValueOf(dest)
	if v.Type().Kind() != reflect.Ptr {
		return errors.New("mapper must a pointer ")
	}
	if v.Elem().Type().Kind() != reflect.Struct {
		return errors.New("mapper can not apply to a struct pointer")
	}
	v = v.Elem()
	for idx := 0; idx < v.Type().NumField(); idx++ {
		field := v.Type().Field(idx)
		fieldDs, sqlTpl, isoLevel, readonly := getTags(field)
		if fieldDs == "" {
			fieldDs = ds
		}
		if field.IsExported() && field.Type.Kind() == reflect.Struct {
			if err := BoostMapper(v.Field(idx).Addr().Interface(), m, fieldDs); err != nil {
				return err
			}
			continue
		}
		switch field.Type.Kind() {
		case reflect.Ptr:
			fv := v.Field(idx)
			if fv.CanSet() && field.Type.AssignableTo(ManagerType) {
				//value is manager,Set a Factory
				fv.Set(reflect.ValueOf(m))
			}
			if fv.CanSet() && field.Type.AssignableTo(DBType) {
				//value is *DB,set datasource
				if db, err := m.Get(ds); err != nil {
					return err
				} else {
					fv.Set(reflect.ValueOf(db))
				}
			}
		case reflect.Func:

			var tplList []string
			if sqlTpl == "" {
				sqlTpl = LowerCase(field.Name) + ".sql"
				pkgPath := LowerCase(v.Type().PkgPath())
				name := LowerCase(v.Type().Name())
				tplList = []string{
					filepath.Join(pkgPath, name, sqlTpl),
					filepath.Join(filepath.Base(pkgPath), name, sqlTpl),
					filepath.Join(name, sqlTpl),
					sqlTpl,
				}
			} else {
				tplList = append(tplList, sqlTpl)
			}

			switch field.Type {
			case ExecFuncType:
				v.Field(idx).Set(reflect.ValueOf(ExecFnWith(m, fieldDs, sqlTpl)))
			case NamedExecFuncType:
				v.Field(idx).Set(reflect.ValueOf(NamedExecFnWith(m, fieldDs, sqlTpl)))
			case TxFuncType:
				v.Field(idx).Set(reflect.ValueOf(TxFnWith(m, fieldDs, sqlTpl, &sql.TxOptions{
					Isolation: isoLevel,
					ReadOnly:  readonly,
				})))
			default:
				name := field.Type.Name()
				quotaIdx := strings.Index(name, "[")
				if quotaIdx > 0 {
					name = name[:quotaIdx]
				}
				var fnVal func([]reflect.Value) []reflect.Value
				switch name {
				case "SelectFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := SelectWith(m, field.Type.Out(0).Elem(), fieldDs, tplList, values[0].Interface().([]any))
						return []reflect.Value{
							valueOrZero(ret, field.Type.Out(0)),
							valueOrZero(err, field.Type.Out(1)),
						}
					}
				case "NamedSelectFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := NamedSelectWith(m, field.Type.Out(0).Elem(), fieldDs, tplList, values[0].Interface())
						return []reflect.Value{
							valueOrZero(ret, field.Type.Out(0)),
							valueOrZero(err, field.Type.Out(1)),
						}
					}
				case "GetFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := GetWith(m, field.Type.Out(0), fieldDs, tplList, values[0].Interface().([]any))
						return []reflect.Value{
							valueOrZero(ret, field.Type.Out(0)),
							valueOrZero(err, field.Type.Out(1)),
						}
					}
				case "NamedGetFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := NamedGetWith(m, field.Type.Out(0), fieldDs, tplList, values[0].Interface())
						return []reflect.Value{
							valueOrZero(ret, field.Type.Out(0)),
							valueOrZero(err, field.Type.Out(1)),
						}
					}
				}
				if fnVal != nil {
					v.Field(idx).Set(reflect.MakeFunc(field.Type, fnVal))
				}
			}
		}
	}
	return nil
}

func valueOrZero(v any, typ reflect.Type) reflect.Value {
	if v == nil {
		return reflect.Zero(typ)
	} else {
		return reflect.ValueOf(v)
	}
}
func Boost(dest interface{}, ds string) error {
	return BoostMapper(dest, StdFactory, ds)
}
func NewMapperWith[T any](factory *Factory, dataSource string) (*T, error) {
	var d T
	err := BoostMapper(&d, factory, dataSource)
	return &d, err
}
func NewMapper[T any](dataSource string) (*T, error) {
	return NewMapperWith[T](StdFactory, dataSource)
}
