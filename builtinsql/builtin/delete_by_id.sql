{{if .LogicDeleteKey -}}
UPDATE `{{.TableName}}`
SET `{{.LogicDeleteKey.ColumnName}}` = 1
WHERE `{{.PrimaryKey.ColumnName}}` = :{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND `{{.TenantKey.ColumnName}}`=:{{.TenantKey.ColumnName}}
{{end}}
{{else}}
DELETE
FROM `{{.TableName}}`
WHERE `{{.PrimaryKey.ColumnName}}` = :{{.PrimaryKey.ColumnName}}
{{if .TenantKey}}
AND `{{.TenantKey.ColumnName}}`=:{{.TenantKey.ColumnName}}
{{end}}
{{end}}