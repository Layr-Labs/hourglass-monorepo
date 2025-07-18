package signer

type ISigner interface {
	SignMessage(data []byte) ([]byte, error)
	SignMessageForSolidity(data [32]byte) ([]byte, error)
}
