package auth

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrEmailNotVerified   = errors.New("email not verified")

	// Token errors
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidSignature   = errors.New("invalid token signature")
	ErrMissingToken       = errors.New("authorization token required")
	ErrTokenRevoked       = errors.New("token has been revoked")

	// Password errors
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must be less than 128 characters")
	ErrPasswordTooWeak    = errors.New("password must contain uppercase, lowercase, number, and special character")

	// 2FA errors
	ErrInvalid2FACode     = errors.New("invalid 2FA code")
	Err2FARequired        = errors.New("2FA verification required")
	Err2FANotEnabled      = errors.New("2FA is not enabled for this account")

	// Permission errors
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrTenantMismatch     = errors.New("resource does not belong to your tenant")
)
