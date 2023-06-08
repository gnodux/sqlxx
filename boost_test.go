package sqlxx

import (
	"fmt"
	"os"
	"testing"
	"text/template"
)

type MyMapper struct {
	*DB
	ListAll    SelectFunc[User]      `sql:"examples/select_users.sql"`
	ListUserBy NamedSelectFunc[User] `sql:"examples/select_user_where.sql"`
	//没有tag,自动根据名称寻找模版
	ListUserByIds  NamedSelectFunc[User]
	GetById        GetFunc[User]         `sql:"examples/get_user_by_id.sql"`
	GetByNamedId   NamedGetFunc[User]    `sql:"examples/get_user_by_id_name.sql"`
	ListUserByName NamedSelectFunc[User] `sql:"examples/select_user_by_name.sql"`
	AddBy          ExecFunc              `sql:"examples/insert_users.sql"`
	Add            NamedExecFunc         `sql:"examples/insert_users.sql"`
}

func TestMapper(t *testing.T) {
	d1, err := NewMapper[MyMapper]("default")
	var testUser []User
	err = d1.Select(&testUser, "select * from `user` limit 1")
	if err != nil {
		t.Fatal(err)
	}
	encoder.Encode(testUser)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("list all")
	fmt.Println(d1.ListAll())
	fmt.Println("list user by")
	fmt.Println(d1.ListUserBy(nil))
	fmt.Println("get by id")
	fmt.Println(d1.GetById(1))
	fmt.Println("get by id ")
	fmt.Println(d1.GetById(-1))
	fmt.Println("get by named id")
	fmt.Println(d1.GetByNamedId(map[string]any{
		"id": 1,
	}))
	fmt.Println(d1.ListUserByName(&User{Name: "%user_1%"}))
	fmt.Println(d1.ListUserByIds(map[string]any{
		"ids": []interface{}{
			12, 18,
		},
	}))
}

func TestName(t *testing.T) {
	t.Log(LowerCase("ListAll"))
}

func TestParseTemplate(t *testing.T) {
	tpl := template.New("sql").Funcs(DefaultFuncMap)
	_, err := tpl.ParseFS(os.DirFS("./queries/"), "**/*.sql")
	if err != nil {
		t.Fatal(err)
		return
	}
	tpl.New("test/job").Parse("{{.Name}}:{{.Cat}}({{list tags}})")
	for _, tp := range tpl.Templates() {
		t.Log(tp.Name())
	}

	t.Log(tpl.ExecuteTemplate(os.Stdout, "get_user_by_id_name.sql", map[string]any{
		"id": 1000,
	}))
}
