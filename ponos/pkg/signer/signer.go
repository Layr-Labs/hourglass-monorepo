package signer

type ISigner interface {
	SignMessage(data []byte) ([]byte, error)
	SignMessageForSolidity(data []byte) ([]byte, error)
}
