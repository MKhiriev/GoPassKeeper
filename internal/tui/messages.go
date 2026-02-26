package tui

// NavigateTo is a Bubble Tea message sent by any page model to instruct [RootModel]
// to switch the active page.
type NavigateTo struct {
	// Page is the key of the target page in the RootModel pages map.
	Page string
	// Payload is an optional message dispatched to the new page immediately after
	// navigation. May be nil when no initial data is required.
	Payload any
}

// LoginResult is a Bubble Tea message produced by the async login command.
// It is handled both by [LoginModel] (to display errors) and by [RootModel]
// (to capture credentials on success and terminate the login flow).
type LoginResult struct {
	// Err is non-nil when authentication failed.
	Err error
	// Username is the login string submitted by the user.
	Username string
	// UserID is the server-assigned identifier of the authenticated user.
	UserID int64
	// EncryptionKey is the symmetric key derived from the master password,
	// used for client-side encryption of private data.
	EncryptionKey []byte
}

// RegisterResult is a Bubble Tea message produced by the async registration command.
// It is handled by [RegisterModel] to display errors or navigate back to the menu.
type RegisterResult struct {
	// Err is non-nil when registration failed.
	Err error
	// Username is the login string chosen by the user.
	Username string
	// UserID is the server-assigned identifier of the newly created user.
	UserID int64
	// EncryptionKey is the symmetric key derived from the master password.
	EncryptionKey []byte
}

// RegisterSuccessNotice is a Bubble Tea message passed to [MenuModel] as the Payload of
// a [NavigateTo] message after a successful registration, so the menu can display a
// confirmation status line.
type RegisterSuccessNotice struct {
	// Username is the login of the newly registered user, used in the status message.
	Username string
}
