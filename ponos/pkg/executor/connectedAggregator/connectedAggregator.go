package connectedAggregator

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strings"
)

type ConnectedAggregator struct {
	avsAddress          string
	aggregatorAddress   string
	logger              *zap.Logger
	signer              signer.ISigner
	authenticationToken uuid.UUID
	aggregatorPublicKey []byte
}

func NewConnectedAggregator(
	avsAddress string,
	aggregatorAddress string,
	logger *zap.Logger,
	signer signer.ISigner,
) *ConnectedAggregator {
	return &ConnectedAggregator{
		avsAddress:        avsAddress,
		aggregatorAddress: aggregatorAddress,
		logger:            logger,
		signer:            signer,
	}
}

type HandshakeResponse struct {
	NonceSignature     []byte
	AuthToken          string
	AuthTokenSignature []byte
}

func (ca *ConnectedAggregator) GetConnectedAggregatorId() string {
	return BuildIdFromAvsAndAggregatorAddress(ca.avsAddress, ca.aggregatorAddress)
}

func BuildIdFromAvsAndAggregatorAddress(avsAddress string, aggregatorAddress string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(avsAddress), strings.ToLower(aggregatorAddress))
}

// Handshake initiates a handshake with the connected aggregator and returns an auth token
//
// The aggregator sends their address, a nonce (uuid) and a signature of the nonce using their private key.
// The Executor verifies the signature using the public key of the aggregator and sends back:
// - the nonce signed with their key (to correlate the message to the request)
// - an authentication token
// - a signature of the authentication token using the private key of the executor
func (ca *ConnectedAggregator) Handshake(nonce string, aggregatorNonceSignature []byte) (*HandshakeResponse, error) {
	aggregatorPublicKey, err := ca.FetchAggregatorPublicKey(ca.aggregatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch aggregator public key: %w", err)
	}

	// Verify the nonce signature using the public key of the aggregator
	valid, err := ca.signer.VerifyMessage(aggregatorPublicKey, []byte(nonce), []byte(aggregatorNonceSignature))
	if err != nil {
		return nil, fmt.Errorf("failed to verify nonce signature: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid nonce signature")
	}

	authToken, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	nonceSig, err := ca.signer.SignMessage([]byte(nonce))
	if err != nil {
		return nil, fmt.Errorf("failed to sign nonce: %w", err)
	}

	authTokenSig, err := ca.signer.SignMessage([]byte(authToken.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign auth token: %w", err)
	}

	ca.aggregatorPublicKey = aggregatorPublicKey
	ca.authenticationToken = authToken

	return &HandshakeResponse{
		NonceSignature:     nonceSig,
		AuthToken:          authToken.String(),
		AuthTokenSignature: authTokenSig,
	}, nil
}

// AuthenticateConnectionRequest verifies the connection request from the aggregator
// by checking the address and the authentication token
func (ca *ConnectedAggregator) AuthenticateConnectionRequest(aggregatorAddress string, authenticationToken string) error {
	if aggregatorAddress != ca.aggregatorAddress {
		return fmt.Errorf("invalid aggregator address: %s", aggregatorAddress)
	}
	uuidToken, err := uuid.Parse(authenticationToken)
	if err != nil {
		return fmt.Errorf("invalid authentication token format: %w", err)
	}
	if uuidToken != ca.authenticationToken {
		return fmt.Errorf("invalid authentication token: %s", authenticationToken)
	}
	return nil
}

// FetchAggregatorPublicKey fetches the public key of the aggregator from a centralized location
// most likely the AVS contract
func (ca *ConnectedAggregator) FetchAggregatorPublicKey(aggregatorAddress string) ([]byte, error) {
	// TODO(seanmcgary): fetch aggregator public key from...somewhere
	return []byte("totally a real key"), nil
}

func (ca *ConnectedAggregator) Terminate() error {
	ca.logger.Sugar().Infow("Terminating connected aggregator",
		zap.String("avsAddress", ca.avsAddress),
		zap.String("aggregatorAddress", ca.aggregatorAddress),
	)
	return nil
}
