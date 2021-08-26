package bantu

import (
	"github.com/go-gormigrate/gormigrate/v2"
	sf "github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/h"
	"github.com/soffa-io/soffa-core-go/log"
	"github.com/soffa-io/soffa-core-go/rpc"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

type Opts struct {
	//MessageHandler   sf.MessageHandler
	NoDatabase       bool
	Migrations       []*gormigrate.Migration
	TenantMigrations []*gormigrate.Migration
	TenantsDb        bool
	//ConfigureRpc     func(client *rpc.client)
	//CreateBroker     bool
}

type Refs struct {
	EntityManager        *sf.DbLink
	TenantsEntityManager *sf.DbLink
	Broker               *sf.MessageBroker
	RpcClient            *rpc.Client
}

var (
	TenantMigrationsCounter sf.Counter
	//primary                      sf.DbLink
)


func CreateDbLinks(context *sf.ApplicationContext, mainMigrations []*gormigrate.Migration, tenantsMigrations []*gormigrate.Migration) []*sf.DbLink {

	var dbLinks []*sf.DbLink

	if mainMigrations != nil {
		databaseUrl := context.Conf("db.url", "DATABASE_URL", true)

		primary := &sf.DbLink{
			ServiceName: strings.TrimPrefix(context.Name, "bantu-"),
			Name:        PrimaryDbId,
			Url:         databaseUrl,
			Migrations:  mainMigrations,
			//TenantsLoader: loader,
		}
		dbLinks = append(dbLinks, primary)
	}

	tenantsDbUrl := context.Conf("db.tenants_url", "TENANTS_DATABASE_URL", context.Name != "bantu-gateway")

	if !h.IsEmpty(tenantsDbUrl) {
		var tenantsLoader sf.TenantsLoaderFn

		if tenantsMigrations != nil {
			tenantsLoader = func() ([]string, error) {
				accountRpc := GetAccountRpc(context)
				return accountRpc.GetTenantsList()
			}
		}

		tenants := &sf.DbLink{
			ServiceName:   strings.TrimPrefix(context.Name, "bantu-"),
			Name:          TenantsDbId,
			Url:           tenantsDbUrl,
			Migrations:    tenantsMigrations,
			TenantsLoader: tenantsLoader,
		}
		dbLinks = append(dbLinks, tenants)
	}

	return dbLinks

}

func Init(context *sf.ApplicationContext, router *sf.Router) {
	rpcUrl := context.Conf("rpc.server.url", "RPC_SERVER_URL", true)
	nc := rpc.NewClient(rpcUrl, context.Name)
	context.UserRpcClient(nc)
	if context.IsTestEnv() && context.Name != AccountServiceId {
		srv := AccountRpc{client: context.GetRpcClient()}
		srv.Serve(new(TestAccounRpcServerImpl))
	}
	jwtSecret := router.App.Conf("jwt.secret", "JWT_SECRET", true)

	if router.App.Name == AccountServiceId{
		router.SetJwtSettings(jwtSecret, "account")
	} else {
		router.JwtAuth()
		router.SetJwtSettings(jwtSecret, "application")
	}
}


func BrokerInfo(context *sf.ApplicationContext, handler sf.MessageHandler, tenant bool) sf.BrokerInfo {
	brokerUrl := context.Conf("amqp.url", "AMQP_URL", true)
	if tenant {
		handler = WrapMessageHandler(handler)
	}
	if h.IsNil(handler) {
		log.Fatal("No handler was provided")
	}
	return sf.BrokerInfo{
		Url:      brokerUrl,
		Queue:    context.Name,
		Exchange: ExchangeName,
		Handler:  handler,
	}
}

func CreateTestBearerToken(t *testing.T, subject string, audience string) string {
	secret := os.Getenv("JWT_SECRET")
	sf.AssertNotEmpty(secret, "JWT_SECRET is missing")
	token, err := sf.CreateJwt(secret, "bantu", subject, audience, sf.H{})
	assert.Nil(t, err)
	return token
}

func WithActiveTenant(req sf.Request, res sf.Response, cb func(ds sf.DbLink)) {
	err := req.Context.GetDbLink(TenantsDbId).WithTenant(req.Auth().Username, func(ds sf.DbLink) error {
		cb(ds)
		return nil
	})
	if err != nil {
		res.Send(nil, err)
	}
}

func WrapMessageHandler(handler sf.MessageHandler) sf.MessageHandler {
	return func(context *sf.ApplicationContext, message sf.Message) error {
		if EventAccountApplicationCreated == message.Event {

			var event ApplicationCreated
			if err := sf.Convert(message.Payload, &event); err != nil {
				return err
			}
			db := context.GetDbLink(TenantsDbId)
			err := db.Migrate(&event.Application.Id)
			_ = TenantMigrationsCounter.Record(err)
			return err

		} else if handler != nil {
			return handler(context, message)
		}
		return nil
	}
}
