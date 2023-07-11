SELECT {{allColumns .Meta.Columns}}
FROM {{n .Meta.TableName}}
{{namedWhere .Where}}
{{orderBy .OrderBy}}
{{- if .Limit -}}
 LIMIT {{.Limit}}
{{end}}
{{- if .Offset -}}
 OFFSET {{.Offset}}
{{end}}
