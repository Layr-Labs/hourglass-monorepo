package web3signer

import (
	"encoding/json"
	"fmt"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	// Jsonrpc specifies the JSON-RPC version (always "2.0")
	Jsonrpc string `json:"jsonrpc"`
	// Method is the JSON-RPC method name
	Method string `json:"method"`
	// Params contains the method parameters
	Params interface{} `json:"params,omitempty"`
	// ID is a unique identifier for the request
	ID int64 `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	// Jsonrpc specifies the JSON-RPC version (always "2.0")
	Jsonrpc string `json:"jsonrpc"`
	// Result contains the method result (present on success)
	Result interface{} `json:"result,omitempty"`
	// Error contains error information (present on error)
	Error *JSONRPCError `json:"error,omitempty"`
	// ID is the request identifier
	ID int64 `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	// Code is the error code
	Code int `json:"code"`
	// Message is the error message
	Message string `json:"message"`
	// Data contains additional error data
	Data interface{} `json:"data,omitempty"`
}

// EthSignTransactionRequest represents the parameters for eth_signTransaction.
type EthSignTransactionRequest struct {
	// From is the account to sign with
	From string `json:"from"`
	// To is the destination address
	To string `json:"to,omitempty"`
	// Gas is the gas limit
	Gas string `json:"gas,omitempty"`
	// GasPrice is the gas price
	GasPrice string `json:"gasPrice,omitempty"`
	// Value is the value to send
	Value string `json:"value,omitempty"`
	// Data is the transaction data
	Data string `json:"data,omitempty"`
	// Nonce is the transaction nonce
	Nonce string `json:"nonce,omitempty"`
	// ChainID is the chain ID
	ChainID string `json:"chainId,omitempty"`
}

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
	Status string `json:"status"`
	// Checks contains detailed status information for individual components
	Checks []StatusCheck `json:"checks"`
	// Outcome is the final health determination ("UP" or "DOWN")
	Outcome string `json:"outcome"`
}

// StatusCheck represents the status of an individual component within the health check.
type StatusCheck struct {
	// ID is the identifier of the component being checked (e.g., "disk-space", "memory")
	ID string `json:"id"`
	// Status is the status of this component ("UP" or "DOWN")
	Status string `json:"status"`
}

// Web3SignerError represents an error response from the Web3Signer service.
type Web3SignerError struct {
	// Code is the HTTP status code associated with the error
	Code int `json:"code"`
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
	Status string `json:"status,omitempty"`
	// Data contains the response payload
	Data interface{} `json:"data,omitempty"`
	// Error contains error information if the request failed
	Error *Web3SignerError `json:"error,omitempty"`
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
