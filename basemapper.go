/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"database/sql"
	"errors"
	"github.com/cookieY/sqlx"
	"reflect"
	"strings"
	"sync"
)

const (
	TagField      = "colx"
	MarkPK        = "primaryKey"
	MarkIgnore    = "_"
	MarkTenantKey = "tenantKey"
)

var (
	ErrIdNotFound = errors.New("primary key not found")
)

type NamedEntity interface {
	TableName() string
}

type EntityMeta struct {
	Columns    []*ColumnDef
	TableName  string
	PrimaryKey *ColumnDef
	TenantKey  *ColumnDef
}
type ColumnDef struct {
	Name         string
	ColumnName   string
	IsPrimaryKey bool
	IsTenantKey  bool
	Ignore       bool
}

func NewEntityMeta(v any) *EntityMeta {
	meta := &EntityMeta{
		TableName: tableName(v),
		Columns:   listValueColumns(v),
	}
	Each(meta.Columns, func(idx int, col *ColumnDef) bool {
		if col.IsTenantKey {
			meta.TenantKey = col
		}
		if col.IsPrimaryKey {
			meta.PrimaryKey = col
		}
		return true
	})
	return meta
}

func parseTags(col *ColumnDef, tags string) {
	tagList := strings.Split(tags, ",")
	for _, tag := range tagList {
		switch tag {
		case MarkPK:
			col.IsPrimaryKey = true
		case MarkIgnore:
			col.Ignore = true
		case MarkTenantKey:
			col.IsTenantKey = true
		}
	}
}
func NewColumnDefWith(f reflect.StructField) *ColumnDef {
	col := &ColumnDef{}
	parseTags(col, f.Tag.Get(TagField))
	col.Name = f.Name
	col.ColumnName = LowerCase(f.Name)
	if col.ColumnName == "tenant_id" {
		col.IsTenantKey = true
	}
	if strings.ToLower(f.Name) == "id" {
		col.IsPrimaryKey = true
	}
	return col
}

func tableName(v any) string {
	if ni, ok := v.(NamedEntity); ok {
		return ni.TableName()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return LowerCase(t.Name())
}

type BaseMapper[T any] struct {
	*DB
	once     sync.Once
	meta     *EntityMeta
	CreateTx TxFunc `sql:"builtin/create.sql" readonly:"false" tx:"Default"`
	UpdateTx TxFunc `sql:"builtin/update_by_id.sql" readonly:"false" tx:"Default"`
	DeleteTx TxFunc `sql:"builtin/delete_by_id.sql" readonly:"false" tx:"Default"`
}

func (b *BaseMapper[T]) init() {
	b.once.Do(func() {
		var t T
		b.meta = NewEntityMeta(t)
	})
}

func (b *BaseMapper[T]) ListById(tenantId any, ids ...any) (entities []T, err error) {
	b.init()
	if len(ids) == 0 {
		return nil, sql.ErrNoRows
	}
	var (
		query   string
		argList []any
		stmt    *sqlx.Stmt
	)
	if query, err = b.Parse("builtin/list_by_id.sql", b.meta); err != nil {
		return
	}
	if b.meta.TenantKey != nil {
		query, argList, err = sqlx.In(query, ids, tenantId)
	} else {
		query, argList, err = sqlx.In(query, ids)
	}
	if err != nil {
		return
	}
	if stmt, err = b.Preparex(query); err != nil {
		return
	}
	defer func() {
		if stErr := stmt.Close(); stErr != nil {
			err = stErr
		}
	}()
	err = stmt.Select(&entities, argList...)
	return entities, err
}
func (b *BaseMapper[T]) Update(entities ...T) error {
	b.init()
	if len(entities) == 0 {
		return sql.ErrNoRows
	}
	return b.UpdateTx(func(tx *Tx) (err error) {
		return tx.RunPrepareNamed(tx.tpl, b.meta, func(stmt *sqlx.NamedStmt) error {
			for _, entity := range entities {
				if _, err = stmt.Exec(entity); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
func (b *BaseMapper[T]) DeleteById(tenantId any, ids ...any) error {
	b.init()
	if len(ids) == 0 {
		return sql.ErrNoRows
	}
	return b.DeleteTx(func(tx *Tx) (err error) {
		return tx.RunPrepareNamed(tx.tpl, b.meta, func(stmt *sqlx.NamedStmt) error {
			for _, id := range ids {
				if _, err = stmt.Exec(map[string]any{
					"tenant_id": tenantId,
					"id":        id,
				}); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
func (b *BaseMapper[T]) UpdateById(entities ...T) error {
	b.init()
	if len(entities) == 0 {
		return sql.ErrNoRows
	}
	return b.UpdateTx(func(tx *Tx) (err error) {
		return tx.RunPrepareNamed(tx.tpl, b.meta, func(stmt *sqlx.NamedStmt) error {
			for _, entity := range entities {
				if _, err = stmt.Exec(entity); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
func (b *BaseMapper[T]) Create(entities ...T) error {
	b.init()
	if len(entities) == 0 {
		return sql.ErrNoRows
	}
	return b.CreateTx(func(tx *Tx) (err error) {
		return tx.RunPrepareNamed(tx.tpl, b.meta, func(stmt *sqlx.NamedStmt) error {
			var result sql.Result
			for _, entity := range entities {
				if result, err = stmt.Exec(entity); err != nil {
					return err
				} else {
					err = setPrimaryKey(&entity, b.meta, result)
				}
			}
			return nil
		})
	})
}

func (b *BaseMapper[T]) SelectBy(where map[string]any, orderBy []string, desc bool, limit, offset int) (result []T, err error) {
	b.init()
	argm := map[string]any{
		"Meta":    b.meta,
		"Where":   where,
		"OrderBy": orderBy,
		"Limit":   limit,
		"Offset":  offset,
		"Desc":    desc,
	}
	err = b.RunPrepareNamed("builtin/select_by.sql", argm, func(stmt *sqlx.NamedStmt) error {
		argd := map[string]any{}
		for k, v := range where {
			argd[k] = v
		}
		argd["Limit"] = limit
		argd["Offset"] = offset
		return stmt.Select(&result, argd)
	})
	return
}
func (b *BaseMapper[T]) SelectByExample(entity T, orderBy []string, desc bool, limit, offset int) ([]T, error) {
	return b.SelectBy(ToMap(entity), orderBy, desc, limit, offset)
}

func setPrimaryKey(entity any, meta *EntityMeta, result sql.Result) error {
	if meta.PrimaryKey == nil {
		return nil
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	} else {
		ev := reflect.ValueOf(entity)
		if ev.Kind() == reflect.Pointer {
			ev = ev.Elem()
			if ev.Kind() == reflect.Pointer {
				ev = ev.Elem()
			}
		}
		pkf := ev.FieldByNameFunc(func(s string) bool {
			if s == meta.PrimaryKey.Name {
				return true
			}
			return false
		})
		if pkf.IsValid() {
			pkf.SetInt(id)
		}
	}
	return nil
}

func Each[T any](lst []T, fn func(int, T) bool) {
	for idx, itm := range lst {
		if !fn(idx, itm) {
			break
		}
	}
}
func ToMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
		if vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}
	}
	result := map[string]any{}
	typ := reflect.TypeOf(vv.Interface())
	for idx := 0; idx < vv.NumField(); idx++ {
		f := vv.Field(idx)
		ft := typ.Field(idx)
		if ft.IsExported() && !f.IsZero() {
			result[ft.Name] = f.Interface()
		}
	}

	return result
}
