package bantu


// ------------------------------------------------------------------------------------------------------------
// Constants
// ------------------------------------------------------------------------------------------------------------


const (
	MessagingServiceId = "bantu-messaging"
	MessagingSendEmailEvent = "bantu.messaging.send_email"
)

// ------------------------------------------------------------------------------------------------------------
// Email
// ------------------------------------------------------------------------------------------------------------

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
	From         EmailAddress `json:"from" binding:"required"`
	Subject      string       `json:"subject" binding:"required"`
	To           EmailAddress `json:"to" binding:"required"`
	PlainContent string       `json:"plain_content" binding:"required_if=html_content ''"`
	HtmlContent  string       `json:"html_content"`
}


type SendEmailOutout struct {
	Status MessageStatus `json:"status"`
}

type MessageStatus string

const (
	MessageFailed  MessageStatus = "FAILED"
	MessagePending               = "PENDING"
)

