/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

const (
	SortAsc  = "ASC"
	SortDesc = "DESC"
)

// SimpleExpr is a simple query
type SimpleExpr struct {
	Condition  map[string]any
	Sort       map[string]string
	LimitRows  int
	OffsetRows int
}

func (s *SimpleExpr) With(column string, value any) *SimpleExpr {
	if s.Condition == nil {
		s.Condition = map[string]any{}
	}
	s.Condition[column] = value
	return s
}
func (s *SimpleExpr) WithMap(m map[string]any) *SimpleExpr {
	if s.Condition == nil {
		s.Condition = map[string]any{}
	}
	for k, v := range m {
		s.Condition[k] = v
	}
	return s
}
func (s *SimpleExpr) OrderBy(sortMap map[string]string) *SimpleExpr {
	for k, v := range sortMap {
		s.Sort[k] = v
	}
	return s
}

func (s *SimpleExpr) Asc(columns ...string) *SimpleExpr {
	for _, col := range columns {
		s.Sort[col] = SortAsc
	}

	return s
}
func (s *SimpleExpr) Desc(columns ...string) *SimpleExpr {
	for _, col := range columns {
		s.Sort[col] = SortDesc
	}
	return s
}
func (s *SimpleExpr) Limit(limit int) *SimpleExpr {
	s.LimitRows = limit
	return s
}
func (s *SimpleExpr) Offset(offset int) *SimpleExpr {
	s.OffsetRows = offset
	return s
}
func Simple(examples ...any) *SimpleExpr {
	query := &SimpleExpr{
		Condition: map[string]any{},
		Sort:      map[string]string{},
	}
	for _, example := range examples {
		switch w := example.(type) {
		case map[string]any:
			mergeMap(query.Condition, w)
		default:
			mergeMap(query.Condition, toMap(w))
		}
	}
	return query
}

func Asc(columns ...string) map[string]string {
	if len(columns) == 0 {
		return nil
	}
	result := map[string]string{}
	for _, col := range columns {
		result[col] = SortAsc
	}
	return result
}
func Desc(columns ...string) map[string]string {
	if len(columns) == 0 {
		return nil
	}
	result := map[string]string{}
	for _, col := range columns {
		result[col] = SortDesc
	}
	return result
}
