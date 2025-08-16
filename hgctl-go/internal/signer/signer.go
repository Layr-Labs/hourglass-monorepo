package signer

type Signers struct {
	ECDSASigner ISigner
	BLSSigner   ISigner
}

type ISigner interface {
	SignMessage(data []byte) ([]byte, error)
	SignMessageForSolidity(data []byte) ([]byte, error)
}
