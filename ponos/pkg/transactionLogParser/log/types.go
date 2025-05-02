// Package parser provides types and functionality for parsing and decoding
// Ethereum transactions and event logs. It helps transform raw blockchain data
// into structured, typed representations.
package log

// DecodedLog represents a decoded Ethereum event log with its arguments and metadata.
// It contains the event name, emitting contract address, and structured argument data.
type DecodedLog struct {
	// LogIndex is the position of the log in the block
	LogIndex uint64 `json:"logIndex"`
	// Address is the contract address that emitted the event
	Address string `json:"address"`
	// Arguments contains the decoded  event parameters
	Arguments []Argument `json:"arguments"`
	// EventName is the name of the emitted event
	EventName string `json:"eventName"`
	// OutputData contains the decoded event data as a map
	OutputData map[string]interface{} `json:"outputData"`
}

// Argument represents a single parameter in a decoded event log or function call.
// It includes the parameter name, type, value, and whether it was indexed in the event.
type Argument struct {
	// Name is the parameter name
	Name string `json:"name"`
	// Type is the Solidity type of the parameter
	Type string `json:"type"`
	// Value is the actual parameter value
	Value interface{} `json:"value"`
	// Indexed indicates whether this was an indexed event parameter
	Indexed bool `json:"indexed"`
}
