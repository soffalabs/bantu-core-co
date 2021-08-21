package bantu

import (
	"strings"
	"time"
)

const (
	AccountIdPrefix         string = "acc_"
	AccountApiKeyPrefix            = "pk_"
	AccountRootApiKeyPrefix        = "rpk_"
	AccountEventIdPrefix           = "ace_"

	EventAccountCreated = "bantu.event.accounts.account_created"

	AccountServiceId      = "bantu-accounts"
)

type Account struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
	Email     string    `json:"email"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type AccountEvent struct {
	Id    string    `json:"id"`
	Event string    `json:"event"`
	Metas string    `json:"metas"`
	Date  time.Time `json:"created_at"`
}

type AccountCreated struct {
	Account Account `json:"account"`
}

type CreateAccountInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type CreateAccountOutput struct {
	AccountId string `json:"account_id"`
	ApiKey    string `json:"api_key"`
}

type GetAccountListOutput struct {
	Accounts []Account `json:"accounts"`
}


func (a Account) IsRoot() bool {
	return strings.HasPrefix(a.ApiKey, AccountRootApiKeyPrefix)
}

