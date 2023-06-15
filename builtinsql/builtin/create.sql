INSERT INTO `{{.TableName}}`
({{columns .Columns}})
VALUES
({{args .Columns}})