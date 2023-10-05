package tracerlogger

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/jimxshaw/tracerlogger/logger"

	"go.uber.org/zap"
)

// Error interface for HTTP error responses.
type Error interface {
	CodeError() CodeError
	Error() string
	Respond(w http.ResponseWriter, code int, err error)
}

// CodeError represents custom error codes for HTTP error responses.
type CodeError string

// ResponseError returns the corresponding ResponseError for the CodeError.
// If the CodeError is not found, it defaults to CodeInternalServerError.
func (ce CodeError) ResponseError() (ResponseError, bool) {
	value, exists := codeErrors[ce]
	if !exists {
		return codeErrors[CodeInternalServerError], false
	}
	return value, true
}

// String returns a formatted string representation of the CodeError.
func (ce CodeError) String() string {
	rr, exists := ce.ResponseError()
	if !exists {
		return "Unknown error code."
	}
	return rr.String()
}

// Error returns the error message corresponding to the CodeError.
func (ce CodeError) Error() string {
	rr, exists := ce.ResponseError()
	if !exists {
		return "Unknown error code."
	}
	return rr.Error()
}

// CodeError returns itself.
func (ce CodeError) CodeError() CodeError {
	return ce
}

// Respond sends an HTTP error response corresponding to the CodeError.
func (ce CodeError) Respond(w http.ResponseWriter, code int, err error) {
	response, _ := ce.ResponseError()
	response.Respond(w, code, err)
}

// FieldError represents an error associated with a specific field.
type FieldError struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message,omitempty"`
}

// String returns a formatted string representation of the FieldError.
func (fe FieldError) String() string {
	if fe.Field == "" {
		return fmt.Sprintf("[%s] %s", fe.Code, fe.Message)
	}
	return fmt.Sprintf("[%s] [%s] %s", fe.Code, fe.Field, fe.Message)
}

// Error returns the error message corresponding to the FieldError.
func (fe FieldError) Error() string {
	return fe.String()
}

// ResponseError represents a structured error response.
type ResponseError struct {
	Code    string       `json:"code,omitempty"`
	Title   string       `json:"title,omitempty"`
	Message string       `json:"message,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// String returns a formatted string representation of the ResponseError.
func (re ResponseError) String() string {
	if re.Code == string(CodeFieldsValidation) {
		fieldErrors := []string{}
		for _, fieldError := range re.Errors {
			fieldErrors = append(fieldErrors, fieldError.String())
		}
		re.Message = strings.Join(fieldErrors, ",")
	}

	if re.Message != "" {
		return fmt.Sprintf("[%s] %s", re.Code, re.Message)
	}
	return fmt.Sprintf("[%s] %s", re.Code, re.Title)
}

// Error returns the error message corresponding to the ResponseError.
func (re ResponseError) Error() string {
	return re.String()
}

// CodeError returns the code associated with the ResponseError.
func (re ResponseError) CodeError() CodeError {
	return CodeError(re.Code)
}

// globalErrorResponse merges ResponseError with a general error message.
type globalErrorResponse struct {
	Error string `json:"error"`
	ResponseError
}

// newGlobalErrorResponse creates a new globalErrorResponse from a given ResponseError and error.
func newGlobalErrorResponse(
	errorResponse ResponseError,
	err error,
) globalErrorResponse {
	if err == nil {
		return globalErrorResponse{
			ResponseError: errorResponse,
			Error:         errorResponse.Error(),
		}
	}
	return globalErrorResponse{
		ResponseError: errorResponse,
		Error:         err.Error(),
	}
}

// AddValidationError appends a FieldError to ResponseError's Errors slice.
func (re *ResponseError) AddValidationError(code CodeError, field, message string) {
	validationErr := FieldError{
		Code:    string(code),
		Field:   field,
		Message: message,
	}

	responseError, exists := code.ResponseError()
	if !exists {
		validationErr.Message = "Unknown error code."
	} else if message == "" {
		validationErr.Message = responseError.Message
	}

	re.Errors = append(re.Errors, validationErr)
	re.updateIfValidationError()
}

// Respond sends an HTTP error response using the ResponseError structure.
func (re ResponseError) Respond(w http.ResponseWriter, code int, err error) {
	logErr := err
	if err == nil {
		logErr = re
	}
	log.Error("request with error", zap.Error(logErr))

	response := newGlobalErrorResponse(re, err)
	RespondWithJSON(w, code, response)
}

// updateIfValidationError sets the Code and Title of the ResponseError based on validation errors.
func (re *ResponseError) updateIfValidationError() {
	if len(re.Errors) > 0 {
		responseError, _ := CodeFieldsValidation.ResponseError()
		re.Code = string(CodeFieldsValidation)
		re.Title = responseError.Title
	}
}
