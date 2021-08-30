package bantu

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/broker"
	"github.com/soffa-io/soffa-core-go/conf"
	"github.com/soffa-io/soffa-core-go/db"
	"github.com/soffa-io/soffa-core-go/errors"
	"github.com/soffa-io/soffa-core-go/http"
	"github.com/soffa-io/soffa-core-go/log"
	"strings"
)

var (
	ApplicationAudience = "application"
	AccountAudience     = "account"
)

type Module struct {
	Db         *db.Link
	accountRpc *AccountRpc
	Broker     broker.Client
	Cfg        *conf.Manager
	accountRpcImpl AccountRpcServer
}


func (b *Module) SetAccountRpcImpl(handler AccountRpcServer) {
	b.accountRpcImpl = handler
}

func (b *Module) CreateAccountRpc(handler AccountRpcServer) {
	b.accountRpc = &AccountRpc{client: b.Broker}
	if handler != nil {
		b.accountRpc.Serve(handler)
	}
}

func (b *Module) GetAccountRpc() *AccountRpc {
	return b.accountRpc
}

func (b *Module) TenantsLoader() []string {
	var tenants []string
	t, err := b.accountRpc.GetTenantsList()
	tenants = t
	if tenants == nil {
		tenants = []string{}
	}
	errors.Raise(err)
	return tenants
}

func (b *Module) ApplyTenantMigrations(msg broker.Message) interface{} {
	var application Application
	errors.Raise(msg.Decode(&application))
	b.Db.MigrateTenant(application.Id)
	return nil
}

func CreateDefault(name string, version string, env string, migrations []*gormigrate.Migration) (*soffa.App, *Module) {
	log.Application = name

	cfg := conf.UseDefault(env)

	bm := &Module{Cfg: cfg}

	app := soffa.NewApp(cfg, name, version)

	isAccountService := name == "bantu-accounts"

	app.UseBroker(func(client broker.Client) {
		bm.Broker = client
		if !isAccountService {
			if cfg.IsTestEnv() {
				bm.CreateAccountRpc(new(TestAccounRpcServerImpl))
			} else {
				bm.CreateAccountRpc(nil)
			}
		}
		app.SetArg("bantu.module", bm)
		if !isAccountService {
			client.Subscribe(EventAccountApplicationCreated, bm.ApplyTenantMigrations)
		}
	})

	if migrations != nil {
		prefix := strings.ReplaceAll(strings.TrimPrefix(strings.ToLower(name), "bantu-"), "-", "_")
		app.UseDB(func(m *db.Manager) {
			ds := db.DS{
				Url:         cfg.Require("db.url", "DATABASE_URL"),
				TablePrefix: prefix + "_",
				Migrations:  migrations,
			}
			if !isAccountService {
				ds.TenantsLoader = bm.TenantsLoader
			}
			m.Add(ds)
			bm.Db = m.GetLink()
		})
	}

	return app, bm
}

func (b *Module) EnableJwt(router *http.Router, audience string) {
	router.Use(http.JwtBearerFilter{
		Secret:   b.Cfg.Require("jwt.secret", "JWT_SECRET"),
		Audience: audience,
	})
}
