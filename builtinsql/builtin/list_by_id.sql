SELECT {{allColumns .Columns}}
FROM `{{.TableName}}`
WHERE `{{.PrimaryKey.ColumnName}}` IN (?)
{{if .TenantKey}}
AND {{.TenantKey.ColumnName}}=?
{{end}}