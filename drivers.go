/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import "github.com/gnodux/sqlxx/dialect"

var (
	DefaultDriver = dialect.MySQL
	MySQL         = dialect.MySQL
	SQLServer     = dialect.SQLServer
	Drivers       = map[string]*dialect.Driver{
		"mysql": MySQL,
		"mssql": SQLServer,
	}
)
