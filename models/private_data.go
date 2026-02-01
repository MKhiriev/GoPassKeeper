package models

const (
	LoginPassword = "login_password"
	Text          = "text"
	Binary        = "binary"
	BankCard      = "bank_card"
)

type PrivateData struct {
	Data     any
	Type     string
	Metadata Metadata
}

type Metadata struct {
	WebSiteAttachment *string
	LoggingAttachment *string
	BankAttachment    *string
	OTP               *string
}

type LoginPasswordData struct {
	Login    string
	Password string
}

type TextData struct {
	Text string
}

type BinaryData struct {
	Text string
}

type BankCardData struct {
	Text string
}
