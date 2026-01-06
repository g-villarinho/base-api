package model

// RegisterAccountPayload represents the request body for account registration
// @name RegisterAccountPayload
type RegisterAccountPayload struct {
	Name     string `json:"name" validate:"required,max=255" example:"João Silva"`
	Email    string `json:"email" validate:"required,email,max=255" example:"joao.silva@email.com"`
	Password string `json:"password" validate:"min=8,max=255" example:"minhasenha123"`
}

// VerifyEmailPayload represents the query parameters for email verification
// @name VerifyEmailPayload
type VerifyEmailPayload struct {
	Token string `query:"token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// LoginPayload represents the request body for user login
// @name LoginPayload
type LoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"joao.silva@email.com"`
	Password string `json:"password" validate:"min=8,max=255" example:"minhasenha123"`
}

// UpdatePasswordPayload represents the request body for updating user password
// @name UpdatePasswordPayload
type UpdatePasswordPayload struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=255" example:"minhasenha123"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=255" example:"novasenha456"`
}

// ForgotPasswordPayload represents the request body for password reset request
// @name ForgotPasswordPayload
type ForgotPasswordPayload struct {
	Email string `json:"email" validate:"required,email,max=255" example:"joao.silva@email.com"`
}

// ResetPasswordPayload represents the request body for confirming password reset
// @name ResetPasswordPayload
type ResetPasswordPayload struct {
	Token       string `json:"token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	NewPassword string `json:"new_password" validate:"required,min=8,max=255" example:"novasenha456"`
}

// RequestEmailChangePayload represents the request body for requesting email change
// @name RequestEmailChangePayload
type RequestEmailChangePayload struct {
	NewEmail string `json:"new_email" validate:"required,email,max=255" example:"newemail@example.com"`
}

// ConfirmEmailChangePayload represents the request body for confirming email change
// @name ConfirmEmailChangePayload
type ConfirmEmailChangePayload struct {
	Token string `json:"token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LoginInput struct {
	Email      string
	Password   string
	IPAddress  string
	UserAgent  string
	DeviceName string
}

type VerifyEmailInput struct {
	Token      string
	IPAddress  string
	UserAgent  string
	DeviceName string
}

// RequestMagicLinkPayload represents the request body for magic link login request
// @name RequestMagicLinkPayload
type RequestMagicLinkPayload struct {
	Email string `json:"email" validate:"required,email,max=255" example:"joao.silva@email.com"`
}

// VerifyMagicLinkPayload represents the query parameters for magic link verification
// @name VerifyMagicLinkPayload
type VerifyMagicLinkPayload struct {
	Token string `query:"token" validate:"required" example:"abc123..."`
}

// RegisterMagicLinkPayload represents the request body for magic link registration
// @name RegisterMagicLinkPayload
type RegisterMagicLinkPayload struct {
	Name  string `json:"name" validate:"required,max=255" example:"João Silva"`
	Email string `json:"email" validate:"required,email,max=255" example:"joao.silva@email.com"`
}

// AuthConfigResponse represents the authentication configuration
// @name AuthConfigResponse
type AuthConfigResponse struct {
	PasswordEnabled  bool `json:"password_enabled"`
	MagicLinkEnabled bool `json:"magic_link_enabled"`
	PasswordRequired bool `json:"password_required"`
}
