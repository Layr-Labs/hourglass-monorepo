package signer

type Signer interface {
	SignMessage(data []byte) ([]byte, error)
	// TODO(seanmcgary): remove this
	VerifyMessage(publicKey []byte, message []byte, signature []byte) (bool, error)
}
