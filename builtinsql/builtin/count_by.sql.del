SELECT count(1) as {{n "total"}}
FROM {{n .Meta.TableName}}
{{- if .Where -}}
{{namedWhere .Where}}
{{end}}