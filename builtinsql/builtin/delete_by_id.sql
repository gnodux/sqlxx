DELETE
FROM `{{.TableName}}`
WHERE `{{.PrimaryKey.ColumnName}}` = :{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND `{{.TenantKey.ColumnName}}`=:{{.TenantKey.ColumnName}}
{{end}}
