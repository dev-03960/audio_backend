package models

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// SignupRequest represents a user signup request
type SignupRequest struct {
	FullName    string `json:"fullName" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// SendOTPRequest represents an OTP send request
type SendOTPRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

// VerifyOTPRequest represents an OTP verification request
type VerifyOTPRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}
