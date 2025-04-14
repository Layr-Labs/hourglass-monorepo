package fauxSigner

type FauxSigner struct {
}

func NewFauxSigner() *FauxSigner {
	return &FauxSigner{}
}

func (fs *FauxSigner) SignMessage(data []byte) ([]byte, error) {
	return []byte("Totally signed message"), nil
}

func (fs *FauxSigner) VerifyMessage(publicKey []byte, message []byte, signature []byte) (bool, error) {
	return true, nil
}
