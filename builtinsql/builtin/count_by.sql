SELECT count(1) as `total`
FROM `{{.Meta.TableName}}`
{{- if .Where -}}
{{namedWhere .Where}}
{{end}}