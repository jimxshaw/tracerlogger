package tracerlogger

const (
	// General errors 0 - 9999

	// CodeBadRequest - CodeError BadRequest
	CodeBadRequest CodeError = "400"
	// CodeUnauthorized - CodeError Unauthorized
	CodeUnauthorized CodeError = "401"
	// CodeForbidden - CodeError Forbidden
	CodeForbidden CodeError = "403"
	// CodeNotFound - CodeError NotFound
	CodeNotFound CodeError = "404"
	// CodeInternalServerError - CodeError InternalServerError
	CodeInternalServerError CodeError = "500"

	// Hygiene and Validation errors 1XXXX

	// CodeFieldsValidation - CodeError FieldsValidation
	CodeFieldsValidation CodeError = "10000"
	// CodeUniqueFieldValidation - CodeError UniqueFieldValidation
	CodeUniqueFieldValidation CodeError = "10001"
	// CodeFieldMaxLength - CodeError FieldMaxLength
	CodeFieldMaxLength CodeError = "10002"
	// CodeFieldRequired - CodeError FieldRequired
	CodeFieldRequired CodeError = "10003"
	// CodeRouteVariableRequired - CodeError RouteVariableRequired
	CodeRouteVariableRequired CodeError = "10004"
	// CodeFieldMinValue - CodeError FieldMinValue
	CodeFieldMinValue CodeError = "10005"
	// CodeFieldInvalidValue - CodeError FieldInvalidValue
	CodeFieldInvalidValue CodeError = "10006"
	// CodeRequestPayloadMalformed - CodeError RequestPayloadMalformed
	CodeRequestPayloadMalformed CodeError = "10007"
	// CodeFieldNotMatchRegex - CodeError FieldNotMatchRegex
	CodeFieldNotMatchRegex CodeError = "10008"
	// CodeRequestTokenMalformed - CodeError RequestTokenMalformed
	CodeRequestTokenMalformed CodeError = "10009"
	// CodeExpiredRequestToken - CodeError ExpiredRequestToken
	CodeExpiredRequestToken CodeError = "10010"
)

var codeErrors map[CodeError]ResponseError = map[CodeError]ResponseError{
	// General errors 0 - 9999
	CodeBadRequest: {
		Code:    string(CodeBadRequest),
		Title:   "Bad Request",
		Message: "Failed to complete request due to a bad request",
	},
	CodeUnauthorized: {
		Code:    string(CodeUnauthorized),
		Title:   "Unauthorized",
		Message: "The user must be authenticated",
	},
	CodeForbidden: {
		Code:    string(CodeForbidden),
		Title:   "Forbidden",
		Message: "The user does not have sufficient permissions",
	},
	CodeNotFound: {
		Code:    string(CodeNotFound),
		Title:   "Not Found",
		Message: "Failed to find a match for the request",
	},
	CodeInternalServerError: {
		Code:    string(CodeInternalServerError),
		Title:   "Internal Server Error",
		Message: "Something went wrong. Please report the issue to Administrators.",
	},
	// Hygiene and Validation errors 1XXXX
	CodeFieldsValidation: {
		Code:    string(CodeFieldsValidation),
		Title:   "Fields Validation",
		Message: "Multiple fields errors",
	},
	CodeUniqueFieldValidation: {
		Code:    string(CodeUniqueFieldValidation),
		Title:   "Unique Field Validation",
		Message: "Unique field resource already exists",
	},
	CodeFieldMaxLength: {
		Code:    string(CodeFieldMaxLength),
		Title:   "Field Max Length",
		Message: "The field length in the request is greater than maximum length",
	},
	CodeFieldRequired: {
		Code:    string(CodeFieldRequired),
		Title:   "Field Required",
		Message: "The field in the request is required",
	},
	CodeRouteVariableRequired: {
		Code:    string(CodeRouteVariableRequired),
		Title:   "Route Variable Required",
		Message: "The route variable for the request is required",
	},
	CodeFieldMinValue: {
		Code:    string(CodeFieldMinValue),
		Title:   "Field Minimum Value",
		Message: "The field in the request is less than minimum value",
	},
	CodeFieldInvalidValue: {
		Code:    string(CodeFieldInvalidValue),
		Title:   "Field Invalid Value",
		Message: "The field in the request has an invalid value",
	},
	CodeRequestPayloadMalformed: {
		Code:    string(CodeRequestPayloadMalformed),
		Title:   "Payload Malformed",
		Message: "The payload for the request is malformed",
	},
	CodeFieldNotMatchRegex: {
		Code:    string(CodeFieldNotMatchRegex),
		Title:   "Field Not Match Regex",
		Message: "The field in the request does not match regular expression format",
	},
	CodeRequestTokenMalformed: {
		Code:    string(CodeRequestTokenMalformed),
		Title:   "Token Malformed",
		Message: "The token for the request is malformed",
	},
	CodeExpiredRequestToken: {
		Code:    string(CodeExpiredRequestToken),
		Title:   "Expired Token",
		Message: "The request token has expired",
	},
}
