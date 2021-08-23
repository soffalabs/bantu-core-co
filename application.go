package bantu

import "time"

const (
	ApplicationIdPrefix      = "app_"
	ApplicationTestKeyPrefix = "sk_test_"
	ApplicationLiveKeyPrefix = "sk_live_"

	ApplicationServiceId           = "bantu-applications"
	EventAccountApplicationCreated = "bantu.event.accounts.application_created"
)

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
