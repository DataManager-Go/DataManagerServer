package handlers

import (
	"errors"
	"net/http"
	"strings"
)

/*
	A wrapper for supported Authorization mechanisms. Currently Only Bearer token
	are supported which are provided by an "Authorization" header inside the request
*/

// ErrorTokenInvalid error if token is invalid
var ErrorTokenInvalid error = errors.New("Token invalid")

// AuthHandler handler for http auth
type AuthHandler struct {
	Request *http.Request
}

// NewAuthHandler returns a new AuthHandler
func NewAuthHandler(request *http.Request) *AuthHandler {
	return &AuthHandler{
		Request: request,
	}
}

// GetBearer return the bearer token
func (authHandler AuthHandler) GetBearer() string {
	authHeader, has := authHandler.Request.Header["Authorization"]
	// Validate bearer token
	if !has || len(authHeader) == 0 || !strings.HasPrefix(authHeader[0], "Bearer") {
		return ""
	}
	return tokenFromBearerHeader(authHeader[0])
}

// Parse the Authorization header and return its real value (the token)
func tokenFromBearerHeader(header string) string {
	return strings.TrimSpace(strings.ReplaceAll(header, "Bearer", ""))
}

// IsInvalid return true if err is invalid
func (authHandler AuthHandler) IsInvalid(err error) bool {
	return err == ErrorTokenInvalid
}
