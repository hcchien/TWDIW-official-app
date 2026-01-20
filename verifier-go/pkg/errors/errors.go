package errors

import (
	"fmt"
	"net/http"
)

// Error codes matching Java VpException
const (
	// Basic
	Unknown                = 99999
	ErrIllegalArgument     = 70001

	// Presentation
	ErrPresInvalidPresentationValidationRequest = 71001
	ErrPresValidateVPError                      = 71002
	ErrPresValidateVPContentError               = 71003
	ErrPresValidateVPProofError                 = 71004
	ErrPresLackOfHolderPublicKey                = 71005
	ErrPresHolderPublicKeyInconsistent          = 71006

	// Credential
	ErrCredValidateVCContentError   = 72001
	ErrCredValidateVCSchemaError    = 72002
	ErrCredValidateVCProofError     = 72003
	ErrCredValidateVCStatusError    = 72004
	ErrCredLackOfIssuerPublicKey    = 72005
	ErrCredInvalidIssuerDIDFormat   = 72006
	ErrCredInvalidIssuerDIDStatus   = 72007
	ErrCredLackOfSub                = 72008

	// Status List
	ErrSLValidateStatusListError        = 73001
	ErrSLValidateStatusListContentError = 73002
	ErrSLValidateStatusListProofError   = 73003
	ErrSLLackOfIssuerPublicKey          = 73004

	// DID
	ErrDIDFrontendQueryDIDError = 74001

	// Connection
	ErrConnLoadIssuerStatusListError = 77001
	ErrConnLoadIssuerSchemaError     = 77002
	ErrConnLoadIssuerPublicKeyError  = 773003
	ErrConnInvalidIssuerStatusList   = 77004
	ErrConnInvalidIssuerSchema       = 77005
	ErrConnInvalidIssuerPublicKey    = 77006
	ErrConnNoMatchedIssuerPublicKey  = 77007

	// Database
	ErrDBQueryError  = 78001
	ErrDBInsertError = 78002
	ErrDBUpdateError = 78003
)

// VPError represents a verifiable presentation error
type VPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *VPError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewVPError creates a new VPError
func NewVPError(code int, message string) *VPError {
	return &VPError{
		Code:    code,
		Message: message,
	}
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *VPError) HTTPStatus() int {
	switch e.Code {
	case ErrPresInvalidPresentationValidationRequest:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// Response returns the error as an error response
func (e *VPError) Response() map[string]interface{} {
	return map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	}
}
