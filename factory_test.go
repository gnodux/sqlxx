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
	"github.com/gnodux/sqlxx/utils"
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
	encoder.SetIndent("", "  ")
	SetConstructor(DefaultName, func() (*DB, error) {
		return Open("mysql", "xxtest:xxtest@tcp(localhost)/sqlxx?charset=utf8&parseTime=true&multiStatements=true")
	})
	SetTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
	initData()
	m.Run()
}
func initData() {
	err := MustGet(DefaultName).Batch(context.Background(), &sql.TxOptions{ReadOnly: false, Isolation: sql.LevelReadCommitted}, func(tx *Tx) error {
		if _, err := tx.Execxx("initialize/create_tables.sql"); err != nil {
			return err
		}
		count := 0

		if err := tx.Get(&count, "SELECT COUNT(1) FROM tenant"); err != nil {
			return err
		}
		if count == 0 {
			if _, err := tx.NamedExecxx("initialize/insert_tenant.sql", Tenant{
				Name: "test tenant",
			}); err != nil {
				return err
			}
		}

		if err := tx.Getxx(&count, "examples/count_user.sql"); err != nil {
			return err
		}
		if count == 0 {
			//add users
			for i := 0; i < 100; i++ {
				if _, err := tx.NamedExecxx("initialize/insert_user.sql", User{
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
		if err := tx.Getxx(&count, "examples/count_role.sql"); err != nil {
			return err
		}
		if count == 0 {
			//add roles
			utils.Must(tx.NamedExecxx("initialize/insert_role.sql", Role{
				Name: "admin",
				Desc: "system administrator",
			}))

			utils.Must(tx.NamedExecxx("initialize/insert_role.sql", Role{
				Name: "user",
				Desc: "normal user",
			}))
			utils.Must(tx.NamedExecxx("initialize/insert_role.sql", Role{
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
	fn := NewSelectFunc[User](DefaultName, "examples/select_users.sql")
	users, err := fn()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(users)
}
func TestSelectUserLikeName(t *testing.T) {
	fn := NewNamedSelectFunc[User](DefaultName, "examples/select_user_where.sql")
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
	err := d.Selectxx(&u, "select_users.sql")
	fmt.Println(err)
}
func TestMustGet(t *testing.T) {
	utils.Must(Get(DefaultName))
}
