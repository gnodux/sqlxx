/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

// SimpleExpr is a simple query
type SimpleExpr struct {
	Condition   map[string]any
	SortColumns []string
	IsDesc      bool
	LimitRows   int
	OffsetRows  int
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
func (s *SimpleExpr) OrderBy(columns ...string) *SimpleExpr {
	s.SortColumns = columns
	return s
}
func (s *SimpleExpr) Asc(columns ...string) *SimpleExpr {
	if len(columns) > 0 {
		s.SortColumns = columns
	}
	s.IsDesc = false
	return s
}
func (s *SimpleExpr) Desc(columns ...string) *SimpleExpr {
	if len(columns) > 0 {
		s.SortColumns = columns
	}
	s.IsDesc = true
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
