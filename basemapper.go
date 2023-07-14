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
	. "github.com/gnodux/sqlxx/meta"
	. "github.com/gnodux/sqlxx/utils"
	"reflect"
	"sync"
)

var (
	ErrIdNotFound = errors.New("primary key not found")
)

type SelectExprBuilder func(*expr.SelectExpr)

type BaseMapper[T any] struct {
	*DB
	once            sync.Once
	meta            *Entity
	CreateTx        TxFunc `sql:"builtin/create.sql" readonly:"false" tx:"Default"`
	UpdateTx        TxFunc `sql:"builtin/update_by_id_tenant_id.sql" readonly:"false" tx:"Default"`
	UpdateByIdTx    TxFunc `sql:"builtin/update_by_id.sql" readonly:"false" tx:"Default"`
	PartialUpdateTx TxFunc `sql:"builtin/partial_update_by_id_tenant_id.sql" readonly:"false" tx:"Default"`
	DeleteTx        TxFunc `sql:"builtin/delete_by_id.sql" readonly:"false" tx:"Default"`
	EraseTx         TxFunc `sql:"builtin/erase_by_id.sql" readonly:"false" tx:"Default"`
}

func (b *BaseMapper[T]) init() {
	b.once.Do(func() {
		var t T
		b.meta = NewEntity(t)
	})
}

func (b *BaseMapper[T]) Meta() *Entity {
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
	if query, err = b.ParseSQL("builtin/list_by_id.sql", b.meta); err != nil {
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

// Update 更新所有列.(如果包含租户ID,则会自动添加租户ID作为更新条件)
// useTenantId 是否使用租户ID作为更新条件
// 如果需要更新部分列,请使用PartialUpdate
func (b *BaseMapper[T]) Update(useTenantId bool, entities ...T) error {
	b.init()
	if len(entities) == 0 {
		return sql.ErrNoRows
	}
	return b.UpdateTx(func(tx *Tx) (err error) {
		return tx.RunCurrentPrepareNamed(map[string]any{
			"Meta":         b.meta,
			"UserTenantId": useTenantId,
		}, func(stmt *sqlx.NamedStmt) error {
			for _, entity := range entities {
				if _, err = stmt.Exec(entity); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

// PartialUpdate 更新指定列.(如果包含租户ID,则会自动添加租户ID作为更新条件)
// useTenantId 是否使用租户ID作为更新条件
// specifiedField 指定需要更新的列,如果为空则自动从实体中获取"非空"列进行更新（指定更新列名时，使用字段名而不是列名）
// entities 实体列表
func (b *BaseMapper[T]) PartialUpdate(useTenantId bool, specifiedField []string, entities ...T) error {
	b.init()
	if len(entities) == 0 {
		return sql.ErrNoRows
	}
	var excludes []string
	if b.meta.TenantKey != nil {
		excludes = append(excludes, b.meta.TenantKey.Name)
	}
	if b.meta.PrimaryKey != nil {
		excludes = append(excludes, b.meta.PrimaryKey.Name)
	}
	var metaCols []*Column
	if len(specifiedField) > 0 {
		metaCols = Search(b.meta.Columns, func(col *Column) bool {
			return Contains(specifiedField, func(s string) bool {
				return col.Name == s
			})
		})
	}
	return b.PartialUpdateTx(func(tx *Tx) (err error) {
		for _, entity := range entities {
			if specifiedField == nil {
				data := ToMap(entity, excludes...)
				metaCols = Search(b.meta.Columns, func(col *Column) bool {
					_, ok := data[col.Name]
					return ok
				})
			}
			if err = tx.RunCurrentPrepareNamed(map[string]any{
				"Meta":        b.meta,
				"Columns":     metaCols,
				"UseTenantId": useTenantId,
			}, func(stmt *sqlx.NamedStmt) (stErr error) {
				_, stErr = stmt.Exec(entity)
				return
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// AutoPartialUpdate 更新指定列.(不会自动添加租户ID作为更新条件)
// useTenantId 是否使用租户ID作为更新条件
func (b *BaseMapper[T]) AutoPartialUpdate(useTenantId bool, entities ...T) error {
	return b.PartialUpdate(useTenantId, nil, entities...)
}

// DeleteById will set the is_deleted flag to true(if is_deleted column exists) or really delete record(if is_deleted column not exists).
func (b *BaseMapper[T]) DeleteById(tenantId any, ids ...any) error {
	b.init()
	if len(ids) == 0 {
		return sql.ErrNoRows
	}
	return b.DeleteTx(func(tx *Tx) (err error) {
		return tx.RunCurrentPrepareNamed(b.meta, func(stmt *sqlx.NamedStmt) error {
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

// EraseById will REALLY delete the record from database.
func (b *BaseMapper[T]) EraseById(tenantId any, ids ...any) error {
	b.init()
	if ids == nil {
		return sql.ErrNoRows
	}
	return b.EraseTx(func(tx *Tx) (err error) {
		return tx.RunCurrentPrepareNamed(b.meta, func(stmt *sqlx.NamedStmt) error {
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

func (b *BaseMapper[T]) Create(entities ...T) error {
	b.init()
	if len(entities) == 0 {
		return sql.ErrNoRows
	}
	return b.CreateTx(func(tx *Tx) (err error) {
		return tx.RunCurrentPrepareNamed(b.meta, func(stmt *sqlx.NamedStmt) error {
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

//	func (b *BaseMapper[T]) SimpleQuery(query *expr.SimpleExpr) (result []T, err error) {
//		return b.SelectBy(query.Condition, query.Sort, query.LimitRows, query.OffsetRows)
//	}
//
//	func (b *BaseMapper[T]) SimpleQueryWithCount(query *expr.SimpleExpr) (result []T, count int, err error) {
//		count, err = b.CountBy(query.Condition)
//		if err != nil {
//			return
//		}
//		result, err = b.SelectBy(query.Condition, query.Sort, query.LimitRows, query.OffsetRows)
//		return
//	}
//func (b *BaseMapper[T]) SelectBy(where map[string]any, orderBy map[string]string, limit, offset int) (result []T, err error) {
//	b.init()
//	argMap := map[string]any{
//		"Meta":    b.meta,
//		"Where":   where,
//		"OrderBy": orderBy,
//		"Limit":   limit,
//		"Offset":  offset,
//	}
//	err = b.RunPrepareNamed("builtin/select_by.sql", argMap, func(stmt *sqlx.NamedStmt) error {
//		queryArgs := map[string]any{}
//		for k, v := range where {
//			queryArgs[k] = v
//		}
//		queryArgs["Limit"] = limit
//		queryArgs["Offset"] = offset
//		return stmt.Select(&result, queryArgs)
//	})
//	return
//}

// Select 使用SelectExprBuilder构建查询
// 默认限制100条,如果需要更多,请使用builder中的Limit方法
func (b *BaseMapper[T]) Select(builders ...expr.FilterFn) (result []T, total int64, err error) {
	b.init()
	//默认Limit 100
	queryExpr := expr.Select(b.meta.ColumnExprs()...).From(b.meta).Limit(100)
	for _, fn := range builders {
		fn(queryExpr)
	}
	err = b.SelectExpr(&result, queryExpr)
	if err != nil {
		return
	}
	if queryExpr.UseCount() {
		countExpr := queryExpr.BuildCountExpr()
		err = b.GetExpr(&total, countExpr)
	}
	return
}

//func (b *BaseMapper[T]) CountBy(where map[string]any) (total int, err error) {
//	b.init()
//	argm := map[string]any{
//		"Meta":  b.meta,
//		"Where": where,
//	}
//	err = b.RunPrepareNamed("builtin/count_by.sql", argm, func(stmt *sqlx.NamedStmt) error {
//		return stmt.Get(&total, where)
//	})
//	return
//}

func (b *BaseMapper[T]) CountBy(where map[string]any, fns ...expr.FilterFn) (total int64, err error) {
	b.init()
	queryExpr := expr.Select(expr.Count).From(b.meta)
	var whereColumns []expr.Expr
	for name, val := range where {
		col := b.Meta().Column(name)
		if col != nil {
			whereColumns = append(whereColumns, expr.Eq(col, expr.Var(name, val)))
		}
	}
	if len(whereColumns) > 0 {
		fns = append([]expr.FilterFn{expr.UseCondition(expr.And(whereColumns...))}, fns...)
	}
	err = b.GetExpr(&total, queryExpr, fns...)
	return
}
func (b *BaseMapper[T]) CountByExample(entity T, filters ...expr.FilterFn) (total int64, err error) {
	return b.CountBy(ToMap(entity), filters...)
}

func (b *BaseMapper[T]) SelectByExample(entity T, builders ...expr.FilterFn) ([]T, int64, error) {
	valMap := ToMap(entity)
	var whereColumns []expr.Expr
	for name, val := range valMap {
		col := b.Meta().Column(name)
		if col != nil {
			whereColumns = append(whereColumns, expr.Eq(col, expr.Var(name, val)))
		}
	}
	if len(whereColumns) > 0 {
		builders = append([]expr.FilterFn{expr.UseCondition(expr.And(whereColumns...))}, builders...)
	}
	return b.Select(builders...)
}

func (b *BaseMapper[T]) UpdateBy(builders ...expr.FilterFn) (effect int64, err error) {
	b.init()
	updateExpr := expr.Update(b.meta)
	for _, fn := range builders {
		fn(updateExpr)
	}
	var result sql.Result
	result, err = b.ExecExpr(updateExpr)
	if err != nil {
		return 0, err
	}
	effect, err = result.RowsAffected()
	return
}
func (b *BaseMapper[T]) UpdateByExample(newValue T, example T, builders ...expr.FilterFn) (effect int64, err error) {
	valMap := ToMap(example)
	var whereColumns []expr.Expr
	for name, val := range valMap {
		col := b.Meta().Column(name)
		if col != nil {
			whereColumns = append(whereColumns, expr.Eq(col, expr.Var(name, val)))
		}
	}
	newValMap := ToMap(newValue)
	var updateColumns []expr.Expr
	for name, val := range newValMap {
		col := b.Meta().Column(name)
		if col != nil {
			updateColumns = append(updateColumns, expr.Eq(col, expr.Var(name, val)))
		}
	}
	builders = append([]expr.FilterFn{expr.Set(updateColumns...)}, builders...)
	if len(whereColumns) > 0 {
		builders = append([]expr.FilterFn{expr.UseCondition(expr.And(whereColumns...))}, builders...)
	}
	return b.UpdateBy(builders...)
}
func (b *BaseMapper[T]) DeleteBy(builders ...expr.DeleteExprFn) (effect int64, err error) {
	b.init()
	if len(builders) == 0 {
		return 0, errors.New("delete by must have one builder")
	}
	deleteExpr := expr.Delete(b.meta)
	for _, fn := range builders {
		fn(deleteExpr)
	}
	var result sql.Result
	result, err = b.ExecExpr(deleteExpr)
	if err != nil {
		return 0, err
	}
	effect, err = result.RowsAffected()
	return
}
func (b *BaseMapper[T]) DeleteByExample(example T, builders ...expr.DeleteExprFn) (effect int64, err error) {
	valMap := ToMap(example)
	var whereColumns []expr.Expr
	for name, val := range valMap {
		col := b.Meta().Column(name)
		if col != nil {
			whereColumns = append(whereColumns, expr.Eq(col, expr.Var(name, val)))
		}
	}
	if len(whereColumns) > 0 {
		builders = append([]expr.DeleteExprFn{expr.UseDeleteCondition(expr.And(whereColumns...))}, builders...)
	}
	return b.DeleteBy(builders...)
}
func setPrimaryKey(entity any, meta *Entity, result sql.Result) error {
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
