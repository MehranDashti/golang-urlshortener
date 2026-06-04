package handler

// ShortenRequest is the request body for POST /client/shorten
type ShortenRequest struct {
	URL       string  `json:"url"        validate:"required,url"`
	ExpiresAt *string `json:"expires_at" validate:"omitempty"`
}

// SignupRequest is the request body for POST /auth/signup
type SignupRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest is the request body for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
