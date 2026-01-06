package notification

import "time"

type SendWelcomeEmailParams struct {
	UserRegistration time.Time
	Name             string
	VerificationLink string
	Email            string
}

type SendVerifyEmailParams struct {
	UserName         string
	VerificationLink string
	UserEmail        string
}

type SendChangeEmailParams struct {
	UserName         string
	NewEmail         string
	ConfirmationLink string
	UserEmail        string
}

type SendResetPasswordEmailParams struct {
	UserName  string
	ResetLink string
	UserEmail string
}

type SendMagicLinkEmailParams struct {
	UserName  string
	LoginLink string
	UserEmail string
}
