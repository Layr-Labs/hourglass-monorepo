package web3signer

import (
	"encoding/json"
	"fmt"
)

// SignRequest represents a request to sign data sent to the Web3Signer service.
type SignRequest struct {
	// Data is the hex-encoded data to be signed
	Data string `json:"data"`
}

// SignResponse represents the response from a signing operation.
type SignResponse struct {
	// Signature is the hex-encoded signature returned by the service
	Signature string `json:"signature"`
}

// HealthCheck represents the detailed health status of the Web3Signer service.
type HealthCheck struct {
	// Status is the overall status of the service ("UP" or "DOWN")
	Status  string        `json:"status"`
	// Checks contains detailed status information for individual components
	Checks  []StatusCheck `json:"checks"`
	// Outcome is the final health determination ("UP" or "DOWN")
	Outcome string        `json:"outcome"`
}

// StatusCheck represents the status of an individual component within the health check.
type StatusCheck struct {
	// ID is the identifier of the component being checked (e.g., "disk-space", "memory")
	ID     string `json:"id"`
	// Status is the status of this component ("UP" or "DOWN")
	Status string `json:"status"`
}

// Web3SignerError represents an error response from the Web3Signer service.
type Web3SignerError struct {
	// Code is the HTTP status code associated with the error
	Code    int    `json:"code"`
	// Message is the error message describing what went wrong
	Message string `json:"message"`
}

// Error implements the error interface for Web3SignerError.
func (e *Web3SignerError) Error() string {
	return fmt.Sprintf("Web3Signer error %d: %s", e.Code, e.Message)
}

// Web3SignerResponse represents a generic response structure from the Web3Signer service.
type Web3SignerResponse struct {
	// Status indicates the response status
	Status string           `json:"status,omitempty"`
	// Data contains the response payload
	Data   interface{}      `json:"data,omitempty"`
	// Error contains error information if the request failed
	Error  *Web3SignerError `json:"error,omitempty"`
}

// IsError returns true if the response contains an error.
func (r *Web3SignerResponse) IsError() bool {
	return r.Error != nil
}

// PublicKeysResponse represents a list of public keys returned by the Web3Signer service.
type PublicKeysResponse []string

// MarshalJSON implements the json.Marshaler interface for PublicKeysResponse.
func (p PublicKeysResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(p))
}

// UnmarshalJSON implements the json.Unmarshaler interface for PublicKeysResponse.
func (p *PublicKeysResponse) UnmarshalJSON(data []byte) error {
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*p = PublicKeysResponse(keys)
	return nil
}
