package bantu

import (
	sf "github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/h"
	"github.com/soffa-io/soffa-core-go/log"
	"github.com/soffa-io/soffa-core-go/rpc"
	"strings"
	"time"
)

const (
	AccountIdPrefix         string = "acc_"
	AccountApiKeyPrefix            = "pk_"
	AccountRootApiKeyPrefix        = "rpk_"
	AccountEventIdPrefix           = "ace_"

	EventAccountCreated = "bantu.event.accounts.account_created"

	AccountServiceId  = "bantu-accounts"
	TestAccountApiKey = "pk_02901920192019201920"

	AccountRpcClientAttribute = "bantu.accounts.rpc.client"
	GetTenantsList            = "bantu.accounts.GetTenantsList"
	FindAccountByKey          = "bantu.accounts.FindByKey"
	FindApplicationByKey      = "bantu.accounts.FindApplicationByKey"

	ApplicationIdPrefix      = "app_"
	ApplicationKeyPrefix     = "sk_"
	ApplicationTestKeyPrefix = "sk_test_"
	ApplicationLiveKeyPrefix = "sk_live_"

	ApplicationServiceId           = "bantu-applications"
	EventAccountApplicationCreated = "bantu.event.accounts.application_created"

	TestApplicationKeyTest = "sk_test_00000000000000000000"
	TestApplicationKeyLive = "sk_live_00000000000000000000"
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

type Application struct {
	Id          string    `json:"id"`
	AccountId   string    `json:"account_id"`
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	ApiKeyTest  string    `json:"api_key_test"`
	ApiKeyLive  string    `json:"api_key_live"`
	CreatedAt   time.Time `json:"created_at"`
}

type ApplicationCreated struct {
	Application Application `json:"application"`
}

type CreateApplicationInput struct {
	AccountId   string `json:"-"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CreateApplicationOutput struct {
	Application Application `json:"application"`
}

type GetApplicationListOutput struct {
	Applications []Application `json:"applications"`
}

type AccountRpc struct {
	client rpc.Client
}

type AccountRpcServer interface {
	FindAccountByKey(key string) (*Account, error)
	FindApplicationByKey(key string) (*Application, error)
	GetTenantsList() ([]string, error)
}

// *********************************************************************************************************************

func GetAccountRpc(context *sf.ApplicationContext) AccountRpc{
	return AccountRpc{client: context.GetRpcClient()}
}


func (a AccountRpc) Serve(impl AccountRpcServer) {

	a.client.ServeAll([]string{FindAccountByKey, FindApplicationByKey}, func(op string, data []byte) (interface{}, error) {
		var input map[string]interface{}
		if err := h.DecodeBytes(data, &input); err != nil {
			return nil, err
		}
		key := input["key"].(string)
		if op == FindAccountByKey {
			return impl.FindAccountByKey(key)
		} else {
			return impl.FindApplicationByKey(key)
		}
	})
	a.client.Serve(GetTenantsList, func(_ string, _ []byte) (interface{}, error) {
		return impl.GetTenantsList()
	})

}

func NewAccountRpcClient(ctx *sf.ApplicationContext) AccountRpc {
	return AccountRpc{client: ctx.GetRpcClient()}
}

func (a AccountRpc) FindAccountByKey(key string) (*Account, error) {
	var result *Account
	err := a.client.Request(FindAccountByKey, sf.H{"key": key}, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a AccountRpc) FindApplicationByKey(key string) (*Application, error) {
	var result *Application
	err := a.client.Request(FindApplicationByKey, sf.H{"key": key}, &result)
	if err != nil {
		log.Error(err)
	}
	if err != nil || h.IsNil(result) {
		return nil, err
	}
	return result, nil
}

func (a AccountRpc) GetTenantsList() ([]string, error) {
	var tenants []string
	err := a.client.Request(GetTenantsList, sf.H{}, &tenants)
	return tenants, err
}

type TestAccounRpcServerImpl struct {
	AccountRpcServer
}

func (s TestAccounRpcServerImpl) FindAccountByKey(key string) (*Account, error) {
	if key == TestAccountApiKey {
		return &Account{
			Id:        "acc_001",
			Name:      "Test",
			ApiKey:    key,
			CreatedAt: time.Now(),
		}, nil
	}
	return nil, nil
}

func (s TestAccounRpcServerImpl) FindApplicationByKey(key string) (*Application, error) {
	if key == TestApplicationKeyTest || key == TestApplicationKeyLive {
		return &Application{
			Id:         "app_001",
			Name:       "Test",
			ApiKeyLive: key,
			ApiKeyTest: key,
			CreatedAt:  time.Now(),
		}, nil
	}
	return nil, nil
}

func (s TestAccounRpcServerImpl) GetTenantsList() ([]string, error) {
	return []string{"app_001"}, nil
}


