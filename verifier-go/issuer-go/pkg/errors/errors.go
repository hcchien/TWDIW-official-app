package errors

import (
	"fmt"
	"net/http"
)

// Error codes matching Java VcException
const (
	// Basic
	Unknown = 99999

	// Credential errors (61xxx)
	ErrCredInvalidCredentialGenerationRequest = 61001
	ErrCredGenerateVCError                    = 61002
	ErrCredPrepareVCError                     = 61003
	ErrCredSignVCError                        = 61004
	ErrCredVerifyVCError                      = 61005
	ErrCredInvalidCredentialID                = 61006
	ErrCredRevokeVCError                      = 61007
	ErrCredPushNotifyError                    = 61008
	ErrCredPushReturnError                    = 61009
	ErrCredCredentialNotFound                 = 61010
	ErrCredQueryVCError                       = 61011
	ErrCredInvalidNonce                       = 61012
	ErrCredDemoGetDataError                   = 61013
	ErrCredDemoReturnError                    = 61014
	ErrCredInvalidCredentialType              = 61015
	ErrCredVCDataSourceNotSet                 = 61016
	ErrCredGetTokenError                      = 61017
	ErrCredGetTokenReturnError                = 61018
	ErrCredTransferVCError                    = 61019
	ErrCredInvalidCredentialTransferRequest   = 61020
	ErrCredConnectVPError                     = 61021
	ErrCredVPReturnError                      = 61022
	ErrCredVPReturnVCSError                   = 61023
	ErrCredVPReturnTypeError                  = 61024
	ErrCredCallCredDataServiceError           = 61025
	ErrCredCredDataServiceReturnError         = 61026
	ErrCredTransferVCNotAllowedError          = 61027
	ErrCredSignIDTError                       = 61028
	ErrCredInvalidCredentialIssuerIdentifier  = 61029
	ErrCredInvalidCredentialSubject           = 61030
	ErrCredConnectOIDVCIError                 = 61031
	ErrCredOIDVCIReturnError                  = 61032
	ErrCredGetHolderDataError                 = 61033
	ErrCredInvalidTransferData                = 61034
	ErrCredInvalidTransferDataCredential      = 61035
	ErrCredInvalidTransferDataCredentialSchema = 61036
	ErrCredCallRevokeServiceError             = 61037
	ErrCredRevokeServiceReturnError           = 61038
	ErrCredVPResponseConvertError             = 61039
	ErrCredVPInvalidError                     = 61040
	ErrCredTxCodeAndVCTransferBothTrueError   = 61041
	ErrCredInvalidExpirationDate              = 61042
	ErrCredInvalidIssuanceDate                = 61043
	ErrCredInvalidIssuanceDateFormat          = 61044
	ErrCredInvalidExpirationDateFormat        = 61045
	ErrCredParseDIDError                      = 61046
	ErrCredInvalidDIDFormat                   = 61047
	ErrCredRevokedCredCannotBeSuspendedError  = 61048
	ErrCredRevokedCredCannotBeRecoveredError  = 61049
	ErrCredentialStatusUnknownError           = 61050
	ErrCredSuspendVCError                     = 61051
	ErrCredRecoverVCError                     = 61052

	// Credential data errors (613xx)
	ErrCredDataInvalidCredentialDataSettingRequest  = 61301
	ErrCredDataCredentialDataConvertError           = 61302
	ErrCredDataInvalidCredentialType                = 61303
	ErrCredDataInvalidIssuerMetadata                = 61304
	ErrCredDataInvalidVCSchema                      = 61305
	ErrCredDataFieldsInSchemaAndMetadataNotIdentical = 61306
	ErrCredDataFieldsInSchemaAndVCDataNotIdentical  = 61307
	ErrCredDataInvalidDataField                     = 61308

	// Sequence errors (614xx)
	ErrSeqInvalidSequenceSettingRequest    = 61401
	ErrSeqInvalidCredentialType            = 61402
	ErrSeqInvalidCredentialIssuerIdentifier = 61403
	ErrSeqInvalidIssuerMetadata            = 61404
	ErrSeqParseIssuerMetadataError         = 61405
	ErrSeqKeyDuplicatedInIssuerMetadata    = 61406
	ErrSeqInvalidSequenceDeletingRequest   = 61407
	ErrSeqInvalidFunctionSwitchSettingRequest = 61408
	ErrSeqTxCodeAndVCTransferBothTrueError = 61409
	ErrSeqInvalidIssuerMetadataDataField   = 61410

	// Setting errors (615xx)
	ErrSettingInvalidUpdateSettingRequest = 61501

	// Status list errors (62xxx)
	ErrSLGenerateStatusListError = 62001
	ErrSLPrepareStatusListError  = 62002
	ErrSLSignStatusListError     = 62003
	ErrSLVerifyStatusListError   = 62004
	ErrSLQueryStatusListError    = 62005
	ErrSLInputStatusListTypeError = 62006
	ErrSLInvalidStatusListOperationRequest = 62007
	ErrSLRetryError              = 62008

	// DID errors (63xxx)
	ErrDIDFrontendGenerateDIDError = 63001
	ErrDIDFrontendRegisterDIDError = 63002
	ErrDIDFrontendReviewDIDError   = 63003
	ErrDIDSignJWTError             = 63004
	ErrDIDParseDIDFromDocumentError = 63005
	ErrDIDRegisterDIDError         = 63009
	ErrDIDRegisterDIDRequest       = 63010
	ErrDIDFrontendGetIssuerDIDInfoError = 63011
	ErrDIDFrontendCreateDIDError   = 63012

	// Public information errors (64xxx)
	ErrInfoInvalidCredentialType = 64001
	ErrInfoInvalidGroupName      = 64002
	ErrInfoStatusListNotFound    = 64003
	ErrInfoInvalidSchemaName     = 64004
	ErrInfoSchemaNotFound        = 64005
	ErrInfoPublicKeyNotFound     = 64006

	// Database errors (68xxx)
	ErrDBQueryError          = 68001
	ErrDBInsertError         = 68002
	ErrDBUpdateError         = 68003
	ErrDBInvalidSequenceName = 68004
	ErrDBDeleteError         = 68005

	// System errors (69xxx)
	ErrSysGenerateKeyError              = 69001
	ErrSysGenerateSchemaError           = 69002
	ErrSysReloadSettingError            = 69003
	ErrSysNotRegisterDIDYetError        = 69004
	ErrSysIssuerDIDNotValidError        = 69005
	ErrSysGetPreloadSettingError        = 69006
	ErrSysCheckSettingError             = 69007
	ErrSysUpdateSettingError            = 69008
	ErrSysNotSetFunctionSwitchYetError  = 69009
	ErrSysNotSetFrontendAccessTokenYetError = 69010
	ErrSysNotSetVCKeyEncError           = 69011
	ErrSysDecryptCipherError            = 69012
	ErrSysDecryptResultNullError        = 69013
	ErrSysCheckKeysValueError           = 69014
	ErrSysEncryptError                  = 69015
	ErrSysNotSetKeyYetError             = 69016
	ErrSysInvalidTimeUnit               = 69017
	ErrSysInvalidTimeDuration           = 69018
)

// VCError represents a verifiable credential error
type VCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *VCError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewVCError creates a new VCError
func NewVCError(code int, message string) *VCError {
	return &VCError{
		Code:    code,
		Message: message,
	}
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *VCError) HTTPStatus() int {
	switch e.Code {
	case ErrCredInvalidCredentialGenerationRequest,
		ErrCredInvalidCredentialID,
		ErrCredInvalidNonce,
		ErrCredInvalidCredentialType,
		ErrCredInvalidCredentialSubject,
		ErrInfoInvalidCredentialType,
		ErrInfoInvalidGroupName,
		ErrInfoInvalidSchemaName,
		ErrDBInvalidSequenceName,
		ErrCredDataInvalidCredentialDataSettingRequest,
		ErrSeqInvalidSequenceSettingRequest,
		ErrDIDRegisterDIDRequest,
		ErrSeqInvalidCredentialType,
		ErrCredDataInvalidCredentialType,
		ErrSeqKeyDuplicatedInIssuerMetadata,
		ErrSLInputStatusListTypeError,
		ErrSeqInvalidSequenceDeletingRequest,
		ErrSeqInvalidFunctionSwitchSettingRequest,
		ErrCredInvalidCredentialTransferRequest,
		ErrCredDataInvalidDataField,
		ErrSeqInvalidIssuerMetadataDataField,
		ErrCredInvalidDIDFormat,
		ErrCredRevokedCredCannotBeSuspendedError,
		ErrCredRevokedCredCannotBeRecoveredError:
		return http.StatusBadRequest

	case ErrCredCredentialNotFound,
		ErrInfoStatusListNotFound,
		ErrInfoSchemaNotFound,
		ErrInfoPublicKeyNotFound:
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}

// Response returns the error as an error response
func (e *VCError) Response() map[string]interface{} {
	return map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	}
}
