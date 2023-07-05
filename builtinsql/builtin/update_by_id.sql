UPDATE `{{.TableName}}`
SET {{setArgs .Columns}}
WHERE `{{.PrimaryKey.ColumnName}}`=:{{.PrimaryKey.ColumnName}}