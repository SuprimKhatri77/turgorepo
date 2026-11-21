package constants

// Error codes returned in APIResponse.Code. They mirror ErrorCodeSchema in
// packages/types/src/api/error-codes.ts, so change both together. Codes marked
// "starter" aren't returned yet; they're kept for features most projects add
// (pagination, ID params), so the codes exist when those features do.
const (
	// general
	InternalServerError = "INTERNAL_SERVER_ERROR"
	ValidationFailed    = "VALIDATION_FAILED"
	Unauthorized        = "UNAUTHORIZED"
	InvalidPageParam    = "INVALID_PAGE_PARAMETER" // starter
	InvalidIDFormat     = "INVALID_ID_FORMAT"      // starter
	InvalidQueryParam   = "INVALID_QUERY_PARAM"    // starter
	PageNotFound        = "PAGE_NOT_FOUND"         // starter
	RateLimited         = "RATE_LIMITED"
	ServiceUnavailable  = "SERVICE_UNAVAILABLE"
	PayloadTooLarge     = "PAYLOAD_TOO_LARGE"

	// auth
	InvalidCredentials = "INVALID_CREDENTIALS"
	InvalidEmail       = "INVALID_EMAIL"    // starter
	InvalidPassword    = "INVALID_PASSWORD" // starter
	InvalidToken       = "INVALID_TOKEN"
	TokenExpired       = "TOKEN_EXPIRED" // starter
	TokenNotProvided   = "TOKEN_NOT_PROVIDED"
	Forbidden          = "FORBIDDEN"

	// user
	UserNotFound      = "USER_NOT_FOUND"
	UserAlreadyExists = "USER_ALREADY_EXISTS"

	// password reset
	InvalidResetToken = "INVALID_RESET_TOKEN"

	// email verification
	EmailNotVerified         = "EMAIL_NOT_VERIFIED"
	InvalidVerificationToken = "INVALID_VERIFICATION_TOKEN"
)
