package bantu

import (
	"github.com/go-gormigrate/gormigrate/v2"
	sf "github.com/soffa-io/soffa-core-go"
	"github.com/soffa-io/soffa-core-go/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

type Opts struct {
	MessageHandler   sf.MessageHandler
	NoDatabase       bool
	Migrations       []*gormigrate.Migration
	TenantMigrations []*gormigrate.Migration
	TenantsDb        bool
}

type Refs struct {
	PrimaryDS  *sf.DataSource
	TenantsDS  *sf.DataSource
	Broker     *sf.MessageBroker
	NatsClient *sf.NatsClient
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
				refs.TenantsDS = app.AddDataSource("tenants", tenantsDbUrl, nil)
			} else {
				refs.PrimaryDS = app.AddDataSource("primary", databaseUrl, testMigrations(app))
				refs.TenantsDS = app.AddMultitenanDataSource("tenants", tenantsDbUrl, opts.TenantMigrations,
					func() ([]string, error) {
						var tenants []string
						return tenants, refs.PrimaryDS.Pluck("applications", "id", &tenants)
					})

				brokerUrl := app.Conf("amqp.url", "AMQP_URL", true)
				b := app.UseBroker(brokerUrl, app.Name, ExchangeName, wraperMessageHandler(true, opts.MessageHandler, refs.TenantsDS))
				refs.Broker = &b
			}
		}
		if refs.PrimaryDS == nil {
			if opts.Migrations == nil {
				refs.PrimaryDS = app.AddDataSource("primary", databaseUrl, testMigrations(app))
			}else {
				refs.PrimaryDS = app.AddDataSource("primary", databaseUrl, opts.Migrations)
			}
		}
	}

	if opts.MessageHandler != nil && refs.Broker == nil {
		brokerUrl := app.Conf("amqp.url", "AMQP_URL", true)
		b := app.UseBroker(brokerUrl, app.Name, ExchangeName, wraperMessageHandler(opts.TenantMigrations != nil, opts.MessageHandler, refs.TenantsDS))
		refs.Broker = &b
	}

	natsUrl := app.Conf("nats.url", "NATS_URL", false)

	if !sf.IsStrEmpty(natsUrl) {
		nc := sf.ConnectToNats(natsUrl, app.Name)
		if opts.MessageHandler != nil {
			log.FatalErr(nc.Subscribe(app.Name, opts.MessageHandler))
			log.FatalErr(nc.Subscribe(ExchangeName, opts.MessageHandler))
		}
		refs.NatsClient = nc
	}

	jwtSecret := app.Conf("jwt.secret", "JWT_SECRET", true)
	if app.Name == "bantu-accounts" {
		app.Router().SetJwtSettings(jwtSecret, "account")
	} else {
		app.Router().SetJwtSettings(jwtSecret, "application")
	}
	if refs.PrimaryDS != nil {
		accountRepo := NewAccountRepo(refs.PrimaryDS)
		applicationRepo := NewApplicationRepo(refs.PrimaryDS)

		app.Router().SetAuthenticator(func(username string, _ string) (*sf.Authentication, error) {
			if app.Name == "bantu-accounts" {
				a, err := accountRepo.FindByKey(username)
				if err != nil || a == nil {
					return nil, nil
				}
				return &sf.Authentication{Username: username, Principal: a}, nil
			} else {
				a, err := applicationRepo.FindByKey(username)
				if err != nil || a == nil {
					return nil, nil
				}
				return &sf.Authentication{Username: username, Principal: a}, nil
			}
		})
	}

	return refs
}

func wraperMessageHandler(multitenant bool, handler sf.MessageHandler, ds *sf.DataSource) sf.MessageHandler {
	if !multitenant {
		return handler
	}
	return func(message sf.Message) error {
		if EventAccountApplicationCreated == message.Event {

			var event ApplicationCreated
			if err := sf.Convert(message.Payload, &event); err != nil {
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

func testMigrations(app *sf.Application) []*gormigrate.Migration {
	if !app.IsTestEnv() {
		return nil
	}

	// This table is not owned by this service so we need to fake it for the tests to go green.
	return []*gormigrate.Migration{
		{
			ID: "0_Accounts",
			Migrate: func(tx *gorm.DB) error {
				if app.Name == "bantu-accounts" {
					return nil
				}
				type Account struct {
					Id     string `gorm:"primaryKey;not null"`
					ApiKey string `gorm:"not null"`
				}
				err := tx.Migrator().CreateTable(&Account{})
				if err != nil {
					return err
				}
				return tx.Create(&Account{Id: "ac_test_01", ApiKey: TestAccountApiKey}).Error
			},
		},
		{
			ID: "1_CreateApplications",
			Migrate: func(tx *gorm.DB) error {
				type Application struct {
					Id         string `gorm:"primaryKey;not null"`
					ApiKeyTest string `gorm:"not null"`
					ApiKeyLive string `gorm:"not null"`
				}
				if err := tx.Migrator().CreateTable(&Application{}); err != nil {
					return nil
				}
				return tx.Create(&Application{
					Id:         "app01",
					ApiKeyTest: TestApplicationKeyTest,
					ApiKeyLive: TestApplicationKeyLive,
				}).Error
			},
		},
	}
}

func CreateTestBearerToken(t *testing.T, subject string, audience string) string {
	secret := os.Getenv("JWT_SECRET")
	token, err := sf.CreateJwt(secret, "bantu", subject, audience, sf.H{})
	assert.Nil(t, err)
	return "bearer " + token
}
