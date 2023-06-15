/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

var (
	encoder = json.NewEncoder(os.Stdout)
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.TraceLevel)
	SetConstructor(DefaultName, func() (*DB, error) {
		return Open("mysql", "xxtest:xxtest@tcp(localhost)/sqlxx?charset=utf8&parseTime=true&multiStatements=true")
	})
	err := ParseTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
	if err != nil {
		panic(err)
	}
	initData()
	m.Run()
}
func initData() {
	err := MustGet(DefaultName).Batch(context.Background(), &sql.TxOptions{ReadOnly: false, Isolation: sql.LevelReadCommitted}, func(tx *Tx) error {
		if _, err := tx.ExecTpl("initialize/create_tables.sql"); err != nil {
			return err
		}
		count := 0

		if err := tx.Get(&count, "SELECT COUNT(1) FROM tenant"); err != nil {
			return err
		}
		if count == 0 {
			if _, err := tx.NamedExecTpl("initialize/insert_tenant.sql", Tenant{
				Name: "test tenant",
			}); err != nil {
				return err
			}
		}

		if err := tx.GetTpl(&count, "examples/count_user.sql"); err != nil {
			return err
		}
		if count == 0 {
			//add users
			for i := 0; i < 100; i++ {
				if _, err := tx.NamedExecTpl("initialize/insert_user.sql", User{
					Name:     fmt.Sprintf("user_%d", i),
					TenantID: 1,
					Password: "password",
					Birthday: time.Now(),
					Address:  fmt.Sprintf("address %d", i),
					Role:     "admin",
				}); err != nil {
					return err
				}
			}
		}
		if err := tx.GetTpl(&count, "examples/count_role.sql"); err != nil {
			return err
		}
		if count == 0 {
			//add roles
			Must(tx.NamedExecTpl("initialize/insert_role.sql", Role{
				Name: "admin",
				Desc: "system administrator",
			}))

			Must(tx.NamedExecTpl("initialize/insert_role.sql", Role{
				Name: "user",
				Desc: "normal user",
			}))
			Must(tx.NamedExecTpl("initialize/insert_role.sql", Role{
				Name: "customer",
				Desc: "customer",
			}))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
func TestSelectUsers(t *testing.T) {
	fn := SelectFn[User](DefaultName, "examples/select_users.sql")
	users, err := fn()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(users)
}
func TestSelectUserLikeName(t *testing.T) {
	fn := NamedSelectFn[User](DefaultName, "examples/select_user_where.sql")
	users, err := fn(User{
		Name: "user_6",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(users)
}

func TestNilDB(t *testing.T) {
	var d *DB
	var u []User
	err := d.SelectTpl(&u, "select_users.sql")
	fmt.Println(err)
}
func TestMustGet(t *testing.T) {
	Must(Get(DefaultName))
}