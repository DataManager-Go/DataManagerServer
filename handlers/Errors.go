package handlers

import (
	"net/http"
	"strings"
)

// WrongInputFormatError wrong user input
const WrongInputFormatError string = "Wrong inputFormat!"

var (
	// RErrNotFound request error if something
	// was requested but wasn't found
	RErrNotFound = NewRequestError("not found", http.StatusNotFound)

	// RErrAlreadyExists request error if something
	// requested already exists
	RErrAlreadyExists = NewRequestError("already exists", http.StatusBadRequest)

	// RErrBadRequest if a request was bad
	RErrBadRequest = NewRequestError("Bad request", http.StatusBadRequest)

	// RErrInvalid if something is invalid
	RErrInvalid = NewRequestError("invalid", http.StatusUnprocessableEntity)

	// RErrTokenInvalid if a token is not valid
	RErrTokenInvalid = RErrInvalid.Prepend("Token").WithCode(http.StatusUnauthorized)

	// RErrNotSupported if an action is not supported
	RErrNotSupported = NewRequestError("not supported", http.StatusUnprocessableEntity)

	// RErrNotAllowed if a request is not allowed for a given user
	RErrNotAllowed = NewRequestError("not allowed", http.StatusForbidden)

	// RErrMissing if something required is missing
	RErrMissing = NewRequestError("missing", http.StatusUnprocessableEntity)

	// RErrTimeout timeout error
	RErrTimeout = NewRequestError("timeout", http.StatusRequestTimeout)

	// RErrPermissionDenied if a user has no permission to run a certain command
	RErrPermissionDenied = NewRequestError("permission denied", http.StatusForbidden)
)

// RequestError error appearing in a request
type RequestError struct {
	Message      string
	ResponseCode int
}

// NewRequestError create a new requestErorr
func NewRequestError(msg string, code int) *RequestError {
	return &RequestError{
		Message:      msg,
		ResponseCode: code,
	}
}

// Prepend text to the error
func (re RequestError) Prepend(txt string) *RequestError {
	// Check if space has to be added
	if strings.HasSuffix(txt, " ") || strings.HasPrefix(re.Message, " ") {
		re.Message = txt + re.Message
	} else {
		re.Message = txt + " " + re.Message
	}

	return &re
}

// Append text to the error
func (re RequestError) Append(txt string) *RequestError {
	// Check if space has to be added
	if strings.HasSuffix(re.Message, " ") || strings.HasPrefix(txt, " ") {
		re.Message += txt
	} else {
		re.Message += " " + txt
	}

	return &re
}

// WithCode uses a different HTTP responsecode
func (re RequestError) WithCode(code int) *RequestError {
	re.ResponseCode = code
	return &re
}

// Implement String for error
func (re RequestError) String() string {
	return re.Message
}

func (re RequestError) Error() string {
	return re.String()
}
