package models

import "time"

// CredentialRequestDTO represents a request to generate a credential
type CredentialRequestDTO struct {
	IssuerDID            string                 `json:"issuer_did"`
	CredentialType       string                 `json:"credential_type"`
	CredentialSubjectID  string                 `json:"credential_subject_id,omitempty"`
	CredentialSubject    map[string]interface{} `json:"credential_subject"`
	IssuanceDate         *time.Time             `json:"issuance_date,omitempty"`
	ExpirationDate       *time.Time             `json:"expiration_date,omitempty"`
	Nonce                string                 `json:"nonce,omitempty"`
}

// CredentialResponseDTO represents the response from credential generation
type CredentialResponseDTO struct {
	CID        string `json:"cid"`
	Credential string `json:"credential"`
	Nonce      string `json:"nonce,omitempty"`
}

// Credential represents a credential entity
type Credential struct {
	CID                 string    `json:"cid"`
	CredentialType      string    `json:"credential_type"`
	CredentialSubjectID string    `json:"credential_subject_id"`
	IssuanceDate        time.Time `json:"issuance_date"`
	ExpirationDate      time.Time `json:"expiration_date"`
	Content             string    `json:"content"`
	TicketNumber        int       `json:"ticket_number"`
	LastUpdateTime      time.Time `json:"last_update_time"`
	CredentialStatus    string    `json:"credential_status"`
	Nonce               string    `json:"nonce"`
}

// CredentialPolicyEntity represents credential policy configuration
type CredentialPolicyEntity struct {
	CredentialType       string `json:"credential_type"`
	IssuerIdentifier     string `json:"issuer_identifier"`
	IssuerMetadata       string `json:"issuer_metadata"`
	VCSchema             string `json:"vc_schema"`
	VCDataSource         string `json:"vc_data_source"`
	IssuanceDateDuration int    `json:"issuance_date_duration"`
	IssuanceDateTimeUnit string `json:"issuance_date_time_unit"`
	ExpirationDuration   int    `json:"expiration_duration"`
	ExpirationTimeUnit   string `json:"expiration_time_unit"`
	FuncSwitch           string `json:"func_switch"`
}

// Ticket represents a credential ticket entity
type Ticket struct {
	CredentialType string    `json:"credential_type"`
	TicketNumber   int       `json:"ticket_number"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

// StatusList represents a status list entity
type StatusList struct {
	GroupName      string    `json:"group_name"`
	Content        string    `json:"content"`
	LastUpdateTime time.Time `json:"last_update_time"`
	StatusListType string    `json:"status_list_type"`
}

// QueryCredentialRequest represents a request to query a credential
type QueryCredentialRequest struct {
	CID   string `json:"cid,omitempty"`
	Nonce string `json:"nonce,omitempty"`
}

// RevokeCredentialRequest represents a request to revoke a credential
type RevokeCredentialRequest struct {
	CID string `json:"cid"`
}

// SuspendCredentialRequest represents a request to suspend a credential
type SuspendCredentialRequest struct {
	CID string `json:"cid"`
}

// RecoverCredentialRequest represents a request to recover a credential
type RecoverCredentialRequest struct {
	CID string `json:"cid"`
}

// StatusListRequest represents a request to generate/query status list
type StatusListRequest struct {
	GroupName      string `json:"group_name"`
	StatusListType string `json:"status_list_type"`
}

// StatusListResponse represents the response from status list operations
type StatusListResponse struct {
	GroupName   string `json:"group_name"`
	StatusList  string `json:"status_list"`
	ContentType string `json:"content_type"`
}

// CredentialStatus represents credential status constants
const (
	CredentialStatusActive    = "ACTIVE"
	CredentialStatusRevoked   = "REVOKED"
	CredentialStatusSuspended = "SUSPENDED"
)

// StatusListType represents status list type constants
const (
	StatusListTypeRevocation = "revocation"
	StatusListTypeSuspension = "suspension"
)

// TimeUnit represents time unit constants
const (
	TimeUnitSecond = "SECOND"
	TimeUnitMinute = "MINUTE"
	TimeUnitHour   = "HOUR"
	TimeUnitDay    = "DAY"
	TimeUnitMonth  = "MONTH"
	TimeUnitYear   = "YEAR"
)
