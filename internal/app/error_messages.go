package app

const (
	MsgInvalidDataProvided        = "invalid data provided"
	MsgInvalidLoginPassword       = "invalid login/password"
	MsgInternalServerError        = "internal server error"
	MsgTokenIsExpired             = "token is expired"
	MsgTokenIsExpiredOrInvalid    = "token is expired or invalid"
	MsgNoPrivateDataProvided      = "no private data provided"
	MsgNoDownloadRequestsProvided = "no download requests provided"
	MsgNoUpdateRequestsProvided   = "no update requests provided"
	MsgNoDeleteRequestsProvided   = "no delete requests provided"
	MsgNoUserIDProvided           = "no user ID provided"
	MsgNoClientIDsForSync         = "no client IDs provided for sync"
	MsgEmptyClientIDForSync       = "empty client ID provided for sync"
	MsgAccessDenied               = "access denied"
	MsgVersionIsNotSpecified      = "version is not specified"
	MsgRegistrationFailed         = "registration failed"
	MsgLoginFailed                = "login failed"

	MsgLoginAlreadyExists = "login already exists"
	MsgDataNotFound       = "data not found"
	MsgVersionConflict    = "version conflict, please sync"
)
