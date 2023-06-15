SELECT {{allColumns .Meta.Columns}}
FROM `{{.Meta.TableName}}`
{{- if .Where -}}
{{namedWhere .Where}}
{{end}}
{{- if .OrderBy -}}
{{- if .Desc -}}
{{desc .OrderBy}}
{{- else -}}
{{asc .OrderBy}}
{{end}}
{{end}}
{{- if .Limit -}}
LIMIT {{.Limit}}
{{end}}
{{- if .Offset -}}
OFFSET {{.Offset}}
{{end}}
