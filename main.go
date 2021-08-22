package bantu

import (
	"github.com/go-gormigrate/gormigrate/v2"
	sf "github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/log"
)

const (
	GlobalMessagingChannel string = "bantu"
	TenantsDS              string = "tenants"
)

type TopicSubscription struct {
	Topic     string
	Broadcast bool
}

type Opts struct {
	MainMigrations   []*gormigrate.Migration
	TenantMigrations []*gormigrate.Migration
}

type bantuConf struct {
	DatabaseUrl        string `json:"db.url" env:"DATABASE_URL"`
	TenantsDatabaseUrl string `json:"db.tenants_url" env:"TENANTS_DATABASE_URL"`
	AmqpUrl            string `json:"amqp.url" env:"AMQP_URL"`
	BroadcastChannel   string `json:"amqp.broadcast_channel" env:"AMQP_BROADCAST_CHANNEL" envDefault:"bantu"`
	// KongAdminUrl       string `json:"kong.admin_url" env:"KONG_ADMIN_URL"`
	// KongUrl            string `json:"kong.url" env:"KONG_URL"`
}

func configureDefaults(app *sf.Application) bantuConf {
	conf := app.GetArg("bantu.config")
	if conf != nil {
		return conf.(bantuConf)
	}
	var bconf = bantuConf{}
	app.LoadConfig(&bconf)
	app.SetArg("bantu.config", bconf)
	return bconf
}

func UseMainDataSource(app *sf.Application, migrations []*gormigrate.Migration) *sf.DataSource {
	conf := configureDefaults(app)
	ds := &sf.DataSource{
		Name:       "primary",
		Url:        conf.DatabaseUrl,
		Migrations: migrations,
	}
	app.RegisterDatasource(ds)
	return ds
}

func UseTenantsDataSource(app *sf.Application, primaryDb *sf.DataSource, migrations []*gormigrate.Migration) *sf.DataSource {
	conf := configureDefaults(app)
	ds := &sf.DataSource{
		Name:       "tenants",
		Url:        conf.DatabaseUrl,
		Migrations: migrations,
		TenantsLoader: func() ([]string, error) {
			var tenants []string
			err := primaryDb.Link.Pluck("applications", "id", &tenants)
			return tenants, err
		},
	}
	app.RegisterDatasource(ds)
	return ds
}

func UserBroker(app *sf.Application, handler sf.MessageHandler) *sf.MessageBroker {
	globalConfig := configureDefaults(app)
	if broker, err := sf.ConnectToBroker(globalConfig.AmqpUrl); err != nil {
		log.FatalErr(err)
		return nil
	} else {
		app.AddBrokerHealthcheck("default", broker)
		app.Subscribe(app.Name, false, handler)
		log.Info("%s subsribed to loopback topic", app.Name)
		app.Subscribe(GlobalMessagingChannel, true, handler)
		log.Info("%s subsribed to broadcast topic", GlobalMessagingChannel)
		return &broker
	}
}
