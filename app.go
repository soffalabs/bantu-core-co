package bantu

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/broker"
	"github.com/soffa-io/soffa-core-go/conf"
	sf "github.com/soffa-io/soffa-core-go/counters"
	"github.com/soffa-io/soffa-core-go/db"
	"github.com/soffa-io/soffa-core-go/errors"
	"github.com/soffa-io/soffa-core-go/http"
	"strings"
)

var (
	ApplicationAudience = "application"
	AccountAudience     = "account"
)

var (
	TenantMigrationsCounter sf.Counter
)

type Module struct {
	Db         *db.Link
	accountRpc *AccountRpc
	Broker     broker.Client
	Cfg        *conf.Manager
}

func NewModule(db *db.Link, broker broker.Client) *Module {
	return &Module{Db: db, Broker: broker}
}

func (b *Module) CreateAccountRpc(handler AccountRpcServer) {
	b.accountRpc = &AccountRpc{client: b.Broker}
	b.accountRpc.Serve(handler)
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
	TenantMigrationsCounter.OK()
	return nil
}

func CreateDefault(name string, version string, tenantService bool, env string, migrations []*gormigrate.Migration) (*soffa.App, *Module) {
	cfg := conf.UseDefault(env)

	bm := &Module{Cfg: cfg}

	app := soffa.NewApp(cfg, name, version)

	app.UseBroker(func(client broker.Client) {
		bm.Broker = client
		if cfg.IsTestEnv() {
			bm.CreateAccountRpc(new(TestAccounRpcServerImpl))
		}else {
			bm.CreateAccountRpc(nil)
		}
		app.SetArg("bantu.module", bm)
		if tenantService {
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
			if tenantService {
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
