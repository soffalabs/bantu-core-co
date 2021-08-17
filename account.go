package bantu

import "time"

// ------------------------------------------------------------------------------------------------------------
// Constants
// ------------------------------------------------------------------------------------------------------------

const (
	AccountIdPrefix          string = "acc_"
	AccountApiKeyPrefix             = "pk_"
	AccountEventIdPrefix            = "ace_"
	ApplicationIdPrefix             = "app_"
	ApplicationTestKeyPrefix        = "sk_test_"
	ApplicationLiveKeyPrefix        = "sk_live_"

	AccountCreatedEvent            = "accounts.account_created"
	AccountApplicationCreatedEvent = "accounts.application_created"

	AccountServiceId      = "bantu-accounts"
	RegistrationServiceId = "bantu-registration"
)

// ------------------------------------------------------------------------------------------------------------
// Tenants
// ------------------------------------------------------------------------------------------------------------

type Tenant struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// ------------------------------------------------------------------------------------------------------------
// Accounts
// ------------------------------------------------------------------------------------------------------------

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
	AccountId string `json:"account_id"`
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

// ------------------------------------------------------------------------------------------------------------
// Applications
// ------------------------------------------------------------------------------------------------------------

type Application struct {
	Id          string    `json:"id"`
	AccountId   string    `json:"account_id"`
	Name        string    `json:"name"`
	Enabled     bool      `gorm:"not null"`
	Description string    `json:"description"`
	ApiKeyTest  string    `json:"api_key_test"`
	ApiKeyLive  string    `json:"api_key_live"`
	CreatedAt   time.Time `json:"created_at"`
}

type ApplicationCreated struct {
	ApplicationId string `json:"application_id"`
	AccountId     string `json:"account_id"`
}

type CreateApplicationInput struct {
	AccountKey  string `json:"-"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CreateApplicationOutput struct {
	Application Application `json:"application"`
}

type GetApplicationListOutput struct {
	Applications []Application `json:"applications"`
}
