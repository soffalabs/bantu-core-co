package bantu

import (
	"github.com/go-gormigrate/gormigrate/v2"
	sf "github.com/soffa-io/soffa-core-go"
	"gorm.io/gorm"
)

type Opts struct {
	Worker           sf.MessageHandler
	NoDatabase       bool
	Migrations       []*gormigrate.Migration
	TenantMigrations []*gormigrate.Migration
	TenantsDb        bool
}

type Refs struct {
	PrimaryDS sf.DataSource
	TenantsDS sf.DataSource
	Broker    *sf.MessageBroker
}

var (
	TenantMigrationsCounter sf.Counter
)

func ConfigureDefaults(app *sf.Application, opts Opts) Refs {

	refs := Refs{}

	if !opts.NoDatabase {
		databaseUrl := app.Conf("db.url", "DATABASE_URL", true)

		if opts.TenantsDb || opts.TenantMigrations != nil {
			tenantsDbUrl := app.Conf("db.tenants_url", "TENANTS_DATABASE_URL", true)
			if opts.TenantMigrations == nil {
				refs.TenantsDS = *app.AddDataSource("tenants", tenantsDbUrl, nil)
			} else {
				refs.PrimaryDS = *app.AddDataSource("primary", databaseUrl, multitnantFakePrimaryMigrations(app.IsTestEnv()))
				refs.TenantsDS = *app.AddMultitenanDataSource("tenants", tenantsDbUrl, opts.TenantMigrations,
					func() ([]string, error) {
						var tenants []string
						return tenants, refs.PrimaryDS.Pluck("applications", "id", &tenants)
					})

				brokerUrl := app.Conf("amqp.url", "AMQP_URL", true)
				b := app.UseBroker(brokerUrl, app.Name, ExchangeName, wraperMessageHandler(true, opts.Worker, refs.TenantsDS))
				refs.Broker = &b
			}
		}
		if refs.PrimaryDS.Url == "" {
			refs.PrimaryDS = *app.AddDataSource("primary", databaseUrl, opts.Migrations)
		}
	}

	if opts.Worker != nil && refs.Broker == nil {
		brokerUrl := app.Conf("amqp.url", "AMQP_URL", true)
		b := app.UseBroker(brokerUrl, app.Name, ExchangeName, wraperMessageHandler(opts.TenantMigrations != nil, opts.Worker, refs.TenantsDS))
		refs.Broker = &b
	}

	return refs
}

func wraperMessageHandler(multitenant bool, handler sf.MessageHandler, ds sf.DataSource) sf.MessageHandler {
	if !multitenant {
		return handler
	}
	return func(message sf.Message) error {
		if EventAccountApplicationCreated == message.Event {

			var event ApplicationCreated
			if err := sf.FromJson(message.Payload, &event); err != nil {
				return err
			}
			err := ds.Migrate(&event.Application.Id)
			_ = TenantMigrationsCounter.Record(err)
			return err

		} else if handler != nil {
			return handler(message)
		}
		return nil
	}
}

func multitnantFakePrimaryMigrations(isTestEnv bool) []*gormigrate.Migration {
	if !isTestEnv {
		return nil
	}
	type Application struct {
		Id string `gorm:"primaryKey;not null"`
	}
	// This table is not owned by this service so we need to fake it for the tests to go green.
	return []*gormigrate.Migration{
		{
			ID: "0_CreateApplications",
			Migrate: func(tx *gorm.DB) error {
				return tx.Migrator().CreateTable(&Application{})
			},
		},
		{
			ID: "1_InsertDefaultTenants",
			Migrate: func(tx *gorm.DB) error {
				return tx.Create(&Application{Id: "tenant1"}).Error
			},
		},
	}
}
