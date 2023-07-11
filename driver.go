/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

type Driver struct {
	//驱动名称（mysql/mssql）等
	Name string
	//是否使用命名参数
	NamedStatement bool
	//命名参数前缀
	NamedPrefix string
	//SQLNameFunc SQL名称转换函数
	SQLNameFunc func(any) string
	//NameFunc 字段名称转换函数
	NameFunc func(string) string
	//DateFormat 日期格式化
	DateFormat string
	//Keywords 关键字映射
	Keywords map[string]string
}

func (d *Driver) Keyword(name string) string {
	if d.Keywords == nil {
		return name
	}
	if k, ok := d.Keywords[name]; ok {
		return k
	}
	return name
}

var (

	//MySQL MySQL驱动
	MySQL = &Driver{
		Name:           "mysql",
		NamedStatement: true,
		NamedPrefix:    ":",
		DateFormat:     "'2006-01-02 15:04:05'",
		SQLNameFunc:    MakeNameFunc("`", "`"),
		NameFunc:       LowerCase,
	}

	//SQLServer SQLServer驱动
	SQLServer = &Driver{
		Name:           "mssql",
		NamedStatement: true,
		NamedPrefix:    "@",
		DateFormat:     "'2006-01-02 15:04:05'",
		SQLNameFunc:    MakeNameFunc("[", "]"),
		NameFunc:       LowerCase,
	}

	DefaultDriver = MySQL
)
