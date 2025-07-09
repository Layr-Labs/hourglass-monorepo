package web3signer

import (
	"encoding/json"
	"fmt"
)

type SignRequest struct {
	Data string `json:"data"`
}

type SignResponse struct {
	Signature string `json:"signature"`
}

type HealthCheck struct {
	Status  string        `json:"status"`
	Checks  []StatusCheck `json:"checks"`
	Outcome string        `json:"outcome"`
}

type StatusCheck struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type Web3SignerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Web3SignerError) Error() string {
	return fmt.Sprintf("Web3Signer error %d: %s", e.Code, e.Message)
}

type Web3SignerResponse struct {
	Status string           `json:"status,omitempty"`
	Data   interface{}      `json:"data,omitempty"`
	Error  *Web3SignerError `json:"error,omitempty"`
}

func (r *Web3SignerResponse) IsError() bool {
	return r.Error != nil
}

type PublicKeysResponse []string

func (p PublicKeysResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(p))
}

func (p *PublicKeysResponse) UnmarshalJSON(data []byte) error {
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*p = PublicKeysResponse(keys)
	return nil
}
