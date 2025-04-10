package signer

type Signer interface {
	SignMessage(data []byte) ([]byte, error)
	VerifyMessage(publicKey []byte, message []byte, signature []byte) (bool, error)
}
