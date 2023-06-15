/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlxx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
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
					"Name":     "user_2%",
					"TenantId": 0,
				}, []string{"Name", "Role"}, true, 100, 10)
				return users, err
			},
		},
		{
			Name: "select by example",
			fn: func() (any, error) {
				users, err := mapper.SelectByExample(User{
					Address: "%57",
					Name:    "user_%",
				}, []string{"role"}, true, 100, 0)
				return users, err
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

func TestBaseMapper_Duck(t *testing.T) {
	m, err := NewMapper[UserMapper](DefaultName)
	assert.NoError(t, err)
	users, err := m.User.SelectByExample(User{
		Name: "user_1%",
	}, []string{"role"}, true, 100, 0)
	assert.NoError(t, err)
	encoder.Encode(users)
}
