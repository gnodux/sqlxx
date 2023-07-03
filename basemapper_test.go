/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"fmt"
	"github.com/gnodux/sqlxx/expr"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestBaseMapper_Pointer(t *testing.T) {
	mapper, err := NewMapper[BaseMapper[*User]](DefaultName)
	assert.NoError(t, err)

	user := User{
		TenantID: 10011002,
		Name:     "test user1",
		Password: fmt.Sprintf("%d", rand.Int63n(99999)),
		Birthday: time.Now(),
		Address:  "Room 1103, Building 17,JIANWAI SOHO EAST Area,ChaoYang, BeiJing",
		Role:     "user",
	}
	tests := []struct {
		Name string
		fn   func() (any, error)
	}{
		{
			Name: "create user",
			fn: func() (any, error) {
				err = mapper.Create(&user)
				t.Log(user.ID, user.TenantID)
				return user, err
			},
		}, {
			Name: "list by id",
			fn: func() (any, error) {
				return mapper.ListById(0, 1, 2, 3, 4, 5)
			},
		}, {
			Name: "delete by id",
			fn: func() (any, error) {
				nu := user
				nu.Name = "test user2"
				err = mapper.Create(&nu)
				encoder.Encode(nu)
				err = mapper.DeleteById(nu.TenantID, nu.ID)
				if err != nil {
					return nil, err
				}
				err = mapper.EraseById(nu.TenantID, nu.ID)
				return nil, err
			},
		}, {
			Name: "update user",
			fn: func() (any, error) {
				if err = mapper.Create(&user); err != nil {
					return nil, err
				}
				user.Role = "test2"
				if err = mapper.Update(&user); err != nil {
					return user, err
				}
				if err = mapper.DeleteById(user.TenantID, user.ID); err != nil {
					return nil, err
				}
				return user, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := tt.fn()
			_ = encoder.Encode(result)
			assert.NoError(t, err)
		})
	}
}

func TestBaseMapper_Struct(t *testing.T) {
	mapper, err := NewMapper[BaseMapper[User]](DefaultName)
	assert.NoError(t, err)
	user := User{
		TenantID: 10011002,
		Name:     "test user1",
		Password: fmt.Sprintf("%d", rand.Int63n(99999)),
		Birthday: time.Now(),
		Address:  "Room 1103, Building 17,JIANWAI SOHO EAST Area,ChaoYang, BeiJing",
		Role:     "user",
	}
	tests := []struct {
		Name string
		fn   func() (any, error)
	}{
		{
			Name: "create user",
			fn: func() (any, error) {
				err = mapper.Create(user)
				t.Log(user.ID, user.TenantID)
				mapper.DeleteById(user.TenantID, user.ID)
				return user, err
			},
		}, {
			Name: "list by id",
			fn: func() (any, error) {
				return mapper.ListById(0, 1, 2, 3, 4, 5)
			},
		}, {
			Name: "delete by id",
			fn: func() (any, error) {
				nu := user
				nu.Name = "test user2"
				err = mapper.Create(nu)
				encoder.Encode(nu)
				err = mapper.DeleteById(nu.TenantID, nu.ID)
				return nil, err
			},
		}, {
			Name: "update user",
			fn: func() (any, error) {
				if err = mapper.Create(user); err != nil {
					return nil, err
				}
				user.Role = "test2"
				if err = mapper.Update(user); err != nil {
					return user, err
				}
				if err = mapper.DeleteById(user.TenantID, user.ID); err != nil {
					return nil, err
				}
				return user, nil
			},
		}, {
			Name: "select by name",
			fn: func() (any, error) {
				users, err := mapper.SelectBy(map[string]any{
					"Name": "user_2%",
				}, expr.Desc("Name", "Role"), 100, 10)
				return users, err
			},
		}, {
			Name: "count by name",
			fn: func() (any, error) {
				total, err := mapper.CountBy(map[string]any{
					"Name": "user_2%",
				})
				return total, err
			},
		},
		{
			Name: "select by example",
			fn: func() (any, error) {
				users, err := mapper.SelectByExample(User{
					Address: "%57",
					Name:    "user_%",
				}, expr.Desc("role"), 100, 0)
				return users, err
			},
		}, {
			Name: "count by example",
			fn: func() (any, error) {
				users, err := mapper.CountByExample(User{
					Address: "%57",
					Name:    "user_%",
				})
				return users, err
			},
		}, {
			Name: "SimpleExpr select with count",
			fn: func() (any, error) {
				result, total, err := mapper.SimpleQueryWithCount(expr.Simple(User{Name: "user_%"}).Desc("role").Limit(100).Offset(0))
				assert.Greater(t, total, 0)
				t.Log("total", total)
				return result, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := tt.fn()
			_ = encoder.Encode(result)
			assert.NoError(t, err)
		})
	}
}

type UserMapper struct {
	User BaseMapper[User]
	Role BaseMapper[Role]
}

type RoleExtMapper struct {
	BaseMapper[Role]
	ListByRoleName SelectFunc[Role]       `sql:"select * from role where name = ?"`
	ListByRole     NamedSelectFunc[*Role] `sql:"select * from role where name = :name"`
	ListUsers      NamedSelectFunc[*User] `sql:"select * from user where role = :name"`
}

func TestExtMapper(t *testing.T) {
	m, err := NewMapper[RoleExtMapper](DefaultName)
	assert.NoError(t, err, "create mapper error")
	users, err := m.ListByRoleName("admin")
	assert.NoError(t, err, "list by role error")
	encoder.Encode(users)
	roles, err := m.ListByRole(&Role{Name: "admin"})
	assert.NoError(t, err, "list by role error")
	encoder.Encode(roles)

	roles, err = m.ListByRole(&Role{Name: "admin"})
	assert.NoError(t, err, "list by role error")
	encoder.Encode(roles)
	adminUsers, err := m.ListUsers(&Role{Name: "admin"})
	assert.NoError(t, err, "list by role error")
	encoder.Encode(adminUsers)
}

func TestBaseMapper_Duck(t *testing.T) {
	m, err := NewMapper[UserMapper](DefaultName)
	assert.NoError(t, err)
	users, err := m.User.SelectByExample(User{
		Name: "user_1%",
	}, expr.Desc("role"), 100, 0)
	assert.NoError(t, err)
	encoder.Encode(users)
}

type Std struct {
	CreateTime time.Time
	CreateBy   int64
	ModifyTime time.Time
	ModifyBy   int64
	IsDeleted  bool
}

type MyUser struct {
	UserName string
	Password string
	Std
	Transaction
}

func Test_listColumns(t *testing.T) {
	my := MyUser{}
	cols := listColumns(reflect.TypeOf(my))
	err := encoder.Encode(cols)
	if err != nil {
		t.Fatal(err)
	}
}

type UserExt struct {
	User
	Nation string `json:"nation"`
	Phone  string `json:"phone"`
}

func Test_ToMap(t *testing.T) {
	m := map[string]interface{}{
		"Name":    "test",
		"ID":      int64(9000912),
		"Address": "Room 1103, Building 17,JIANWAI SOHO EAST Area,ChaoYang, BeiJing",
		"Nation":  "China",
		"Phone":   "123456789",
	}
	ue := &UserExt{
		User: User{
			Name:    "test",
			ID:      9000912,
			Address: "Room 1103, Building 17,JIANWAI SOHO EAST Area,ChaoYang, BeiJing",
		},
		Nation: "China",
		Phone:  "123456789",
	}
	assert.Equal(t, m, ToMap(ue))
	encoder.Encode(ToMap(ue))
}
