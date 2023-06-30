/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"database/sql"
	"errors"
	"github.com/cookieY/sqlx"
	"github.com/gnodux/sqlxx/expr"
	"reflect"
	"strings"
	"sync"
)

const (
	TagField      = "dbx"
	MarkPK        = "primaryKey"
	MarkIgnore    = "_"
	MarkTenantKey = "tenantKey"
	MarkIsDeleted = "softDelete"
)

var (
	ErrIdNotFound = errors.New("primary key not found")
)

type NamedEntity interface {
	TableName() string
}

type EntityMeta struct {
	Columns        []*ColumnMeta
	TableName      string
	PrimaryKey     *ColumnMeta
	TenantKey      *ColumnMeta
	LogicDeleteKey *ColumnMeta
}

func (m *EntityMeta) String() string {
	return m.TableName
}

type ColumnMeta struct {
	Name             string
	ColumnName       string
	Type             reflect.Type
	IsPrimaryKey     bool
	IsTenantKey      bool
	IsLogicDeleteKey bool
	Ignore           bool
}

func (c *ColumnMeta) String() string {
	return c.ColumnName
}

func NewEntityMeta(v any) *EntityMeta {
	meta := &EntityMeta{
		TableName: tableName(v),
		Columns:   listValueColumns(v),
	}
	Each(meta.Columns, func(idx int, col *ColumnMeta) bool {
		if col.IsTenantKey {
			meta.TenantKey = col
		}
		if col.IsPrimaryKey {
			meta.PrimaryKey = col
		}
		if col.IsLogicDeleteKey {
			meta.LogicDeleteKey = col
		}
		return true
	})
	return meta
}

func listValueColumns(v any) []*ColumnMeta {
	argv := reflect.TypeOf(v)
	return listColumns(argv)

}
func listColumns(t reflect.Type) []*ColumnMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []*ColumnMeta
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			if field.Anonymous {
				fields = append(fields, listColumns(field.Type)...)
			} else {
				fields = append(fields, NewColumnDefWith(t.Field(i)))
			}
		}
	}
	return fields
}
func parseTags(col *ColumnMeta, tags string) {
	tagList := strings.Split(tags, ",")
	for _, tag := range tagList {
		switch tag {
		case MarkPK:
			col.IsPrimaryKey = true
		case MarkIgnore:
			col.Ignore = true
		case MarkTenantKey:
			col.IsTenantKey = true
		case MarkIsDeleted:
			col.IsLogicDeleteKey = true
		}
	}
}
func NewColumnDefWith(f reflect.StructField) *ColumnMeta {
	col := &ColumnMeta{}
	parseTags(col, f.Tag.Get(TagField))
	col.Name = f.Name
	col.ColumnName = LowerCase(f.Name)
	if col.ColumnName == "tenant_id" {
		col.IsTenantKey = true
	}
	if col.ColumnName == "id" {
		col.IsPrimaryKey = true
	}
	if col.ColumnName == "is_deleted" {
		col.IsLogicDeleteKey = true
	}
	col.Type = f.Type
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
	EraseTx  TxFunc `sql:"builtin/erase_by_id.sql" readonly:"false" tx:"Default"`
}

func (b *BaseMapper[T]) init() {
	b.once.Do(func() {
		var t T
		b.meta = NewEntityMeta(t)
	})
}

func (b *BaseMapper[T]) Meta() *EntityMeta {
	b.init()
	return b.meta
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
func (b *BaseMapper[T]) EraseById(tenantId any, ids any) error {
	b.init()
	if ids == nil {
		return sql.ErrNoRows
	}
	return b.EraseTx(func(tx *Tx) (err error) {
		return tx.RunPrepareNamed(tx.tpl, b.meta, func(stmt *sqlx.NamedStmt) error {
			if _, err = stmt.Exec(map[string]any{
				"tenant_id": tenantId,
				"id":        ids,
			}); err != nil {
				return err
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
func (b *BaseMapper[T]) SimpleQuery(query *expr.SimpleExpr) (result []T, err error) {
	return b.SelectBy(query.Condition, query.SortColumns, query.IsDesc, query.LimitRows, query.OffsetRows)
}
func (b *BaseMapper[T]) SimpleQueryWithCount(query *expr.SimpleExpr) (result []T, count int, err error) {
	count, err = b.CountBy(query.Condition)
	if err != nil {
		return
	}
	result, err = b.SelectBy(query.Condition, query.SortColumns, query.IsDesc, query.LimitRows, query.OffsetRows)
	return
}
func (b *BaseMapper[T]) SelectBy(where map[string]any, orderBy []string, desc bool, limit, offset int) (result []T, err error) {
	b.init()
	argMap := map[string]any{
		"Meta":    b.meta,
		"Where":   where,
		"OrderBy": orderBy,
		"Limit":   limit,
		"Offset":  offset,
		"Desc":    desc,
	}
	err = b.RunPrepareNamed("builtin/select_by.sql", argMap, func(stmt *sqlx.NamedStmt) error {
		queryArgs := map[string]any{}
		for k, v := range where {
			queryArgs[k] = v
		}
		queryArgs["Limit"] = limit
		queryArgs["Offset"] = offset
		return stmt.Select(&result, queryArgs)
	})
	return
}

func (b *BaseMapper[T]) CountBy(where map[string]any) (total int, err error) {
	b.init()
	argm := map[string]any{
		"Meta":  b.meta,
		"Where": where,
	}
	err = b.RunPrepareNamed("builtin/count_by.sql", argm, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&total, where)
	})
	return
}
func (b *BaseMapper[T]) SelectByExample(entity T, orderBy []string, desc bool, limit, offset int) ([]T, error) {
	return b.SelectBy(ToMap(entity), orderBy, desc, limit, offset)
}
func (b *BaseMapper[T]) CountByExample(entity T) (int, error) {
	return b.CountBy(ToMap(entity))
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
		pkf := ev.FieldByName(meta.PrimaryKey.Name)
		if pkf.IsValid() && pkf.CanSet() && pkf.CanInt() {
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
			if ft.Anonymous {
				for k, v := range ToMap(f.Interface()) {
					result[k] = v
				}
			} else {
				result[ft.Name] = f.Interface()
			}
		}
	}

	return result
}
