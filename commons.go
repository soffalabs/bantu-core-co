package bantu

const (
	GlobalMessagingChannel string = "bantu"
)


type ServiceConfig struct {
	Port               int    `json:"server.port" env:"PORT" envDefault:"8080"`
	DatabaseUrl        string `json:"db.url" env:"DATABASE_URL"`
	TenantsDatabaseUrl string `json:"db.tenants_url" env:"TENANTS_DATABASE_URL"`
	AmqpUrl            string `json:"amqp.url" env:"AMQP_URL"`
	BroadcastChannel   string `json:"amqp.broadcast_channel" env:"AMQP_BROADCAST_CHANNEL" envDefault:"bantu"`
	KongUrl            string `json:"kong.url" env:"KONG_URL"`
}
