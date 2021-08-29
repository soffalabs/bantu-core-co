package bantu

const (
	MessagingServiceId = "bantu-messaging"
	EventSendEmail     = "bantu.event.messaging.send_email"
)

type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address" binding:"required,email"`
}

type Email struct {
	From         EmailAddress
	Subject      string
	To           EmailAddress
	PlainContent string
	HtmlContent  string
}

type SendEmailInput struct {
	Metas        map[string]interface{} `json:"metas"`
	From         EmailAddress           `json:"from" binding:"required"`
	Subject      string                 `json:"subject" binding:"required"`
	To           EmailAddress           `json:"to" binding:"required"`
	PlainContent string                 `json:"plain_content"`
	HtmlContent  string                 `json:"html_content"`
}

type SendEmailOutout struct {
	Status MessageStatus `json:"status"`
}

type MessageStatus string

const (
	MessageFailed  MessageStatus = "FAILED"
	MessagePending               = "PENDING"
)
