UPDATE `{{.TableName}}`
SET {{setArgs .Columns}}
WHERE `{{.PrimaryKey.ColumnName}}`=:{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND `{{.TenantKey.ColumnName}}`=:{{.TenantKey.ColumnName}}
{{end}}