package tui

// NavigateTo is sent by pages to switch active page in RootModel.
type NavigateTo struct {
	Page    string
	Payload any
}

// LoginResult is sent when login command finishes.
type LoginResult struct {
	Err           error
	Username      string
	UserID        int64
	EncryptionKey []byte
}

// RegisterResult is sent when register command finishes.
type RegisterResult struct {
	Err           error
	Username      string
	UserID        int64
	EncryptionKey []byte
}

// RegisterSuccessNotice is passed to menu after successful registration.
type RegisterSuccessNotice struct {
	Username string
}
