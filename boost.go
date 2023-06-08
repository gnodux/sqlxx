package sqlxx

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	TagDS  = "ds"
	TagSQL = "sql"

	ManagerType       = reflect.TypeOf((*Factory)(nil))
	DBType            = reflect.TypeOf((*DB)(nil))
	ExecFuncType      = reflect.TypeOf(ExecFunc(nil))
	NamedExecFuncType = reflect.TypeOf(NamedExecFunc(nil))
)

func GetTags(field reflect.StructField) (ds, sql string) {
	ds = field.Tag.Get(TagDS)
	sql = field.Tag.Get(TagSQL)
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
		fieldDs, sql := GetTags(field)
		if fieldDs == "" {
			fieldDs = ds
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
			if sql == "" {
				sql = LowerCase(field.Name) + ".sql"
				pkgPath := LowerCase(v.Type().PkgPath())
				name := LowerCase(v.Type().Name())
				tplList = []string{
					filepath.Join(pkgPath, name, sql),
					filepath.Join(filepath.Base(pkgPath), name, sql),
					filepath.Join(name, sql),
					sql,
				}
			} else {
				tplList = append(tplList, sql)
			}

			if field.Type == ExecFuncType {
				v.Field(idx).Set(reflect.ValueOf(ExecFnWith(m, fieldDs, sql)))
			} else if field.Type == NamedExecFuncType {
				v.Field(idx).Set(reflect.ValueOf(NamedExecFnWith(m, fieldDs, sql)))
			} else {
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
func NewMapperWith[T any](m *Factory, dbName string) (T, error) {
	var d T
	err := BoostMapper(&d, m, dbName)
	return d, err
}
func NewMapper[T any](ds string) (T, error) {
	return NewMapperWith[T](StdFactory, ds)
}
