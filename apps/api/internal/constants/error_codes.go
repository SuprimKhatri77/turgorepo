package constants

const (
	// general
	InternalServerError = "INTERNAL_SERVER_ERROR"
	ValidationFailed    = "VALIDATION_FAILED"
	Unauthorized        = "UNAUTHORIZED"
	InvalidPageParam    = "INVALID_PAGE_PARAMETER"
	InvalidIDFormat     = "INVALID_ID_FORMAT"
	InvalidQueryParam   = "INVALID_QUERY_PARAM"
	PageNotFound        = "PAGE_NOT_FOUND"

	// auth
	InvalidCredentials = "INVALID_CREDENTIALS"
	InvalidEmail       = "INVALID_EMAIL"
	InvalidPassword    = "INVALID_PASSWORD"
	InvalidToken       = "INVALID_TOKEN"
	TokenExpired       = "TOKEN_EXPIRED"
	TokenNotProvided   = "TOKEN_NOT_PROVIDED"
	TokenInvalid       = "TOKEN_INVALID"
	Forbidden          = "FORBIDDEN"

	// user
	UserNotFound      = "USER_NOT_FOUND"
	UserAlreadyExists = "USER_ALREADY_EXISTS"
)
