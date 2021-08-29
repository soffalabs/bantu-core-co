package bantu

import (
	"time"
)

const (
	TestAppId1 = "app_001"
	TestAppId2 = "app_002"
)


func (b *Module) EnableAccountRpcMock() {
	b.GetAccountRpc().Serve(TestAccounRpcServerImpl{})
}

type TestAccounRpcServerImpl struct {
	AccountRpcServer
}

func (s TestAccounRpcServerImpl) FindAccountByKey(key string) *Account {
	if key == TestAccountApiKey {
		return &Account{
			Id:        "acc_001",
			Name:      "Test",
			ApiKey:    key,
			CreatedAt: time.Now(),
		}
	}
	return nil
}

func (s TestAccounRpcServerImpl) FindApplicationByKey(key string) *Application {
	if key == TestApplicationKeyTest || key == TestApplicationKeyLive {
		return &Application{
			Id:         TestAppId1,
			Name:       "Test",
			ApiKeyLive: key,
			ApiKeyTest: key,
			CreatedAt:  time.Now(),
		}
	}
	return nil
}

func (s TestAccounRpcServerImpl) GetTenantsList() []string {
	return []string{TestAppId1}
}
