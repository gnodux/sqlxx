package sqlxx

import "time"

type Tenant struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (m *Tenant) TableName() string {
	return "tenant"
}

type User struct {
	ID       int64     `json:"id"`
	TenantID int64     `json:"tenant_id"`
	Name     string    `json:"name"`
	Password string    `json:"password"`
	Birthday time.Time `json:"birthday"`
	Address  string    `json:"address"`
	Role     string    `json:"role"`
}

func (m *User) TableName() string {
	return "user"
}

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func (m *Role) TableName() string {
	return "role"
}

type AccountBook struct {
	ID       int64   `json:"id"`
	TenantID int64   `json:"tenant_id"` // 租户ID
	CreateBy int64   `json:"create_by"` // 创建人
	Owner    int64   `json:"owner"`     // 账本所有人
	Name     string  `json:"name"`      // 账本名称
	Balance  float64 `json:"balance"`   // 账户余额
	Desc     string  `json:"desc"`      // 账本描述
}

func (m *AccountBook) TableName() string {
	return "account_book"
}

type Transaction struct {
	ID            int64   `json:"id"`
	TenantID      int64   `json:"tenant_id"`       // 租户ID
	AccountBookID int64   `json:"account_book_id"` // 账本ID
	CreateBy      int64   `json:"create_by"`       // 创建人
	CreateTime    int64   `json:"create_time"`     //  创建时间
	Amount        float64 `json:"amount"`          //  交易金额
	Type          string  `json:"type"`            //  交易类型
	Desc          string  `json:"desc"`            //  交易描述
	Status        string  `json:"status"`          //  交易状态
}

func (m *Transaction) TableName() string {
	return "transaction"
}
