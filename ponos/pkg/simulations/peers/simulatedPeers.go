package peers

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

func NewSimulatedPeerFromConfig(simulatedPeer config.SimulatedPeer) (*peering.OperatorPeerInfo, error) {
	pubKey, err := bn254.NewPublicKeyFromHexString(simulatedPeer.PublicKey)
	if err != nil {
		return nil, err
	}
	return &peering.OperatorPeerInfo{
		OperatorAddress: simulatedPeer.OperatorAddress,
		NetworkAddress:  simulatedPeer.NetworkAddress,
		PublicKey:       pubKey,
		OperatorSetIds:  []uint32{simulatedPeer.OperatorSetId},
	}, nil
}

func NewSimulatedPeersFromConfig(simulatedPeers []config.SimulatedPeer) ([]*peering.OperatorPeerInfo, error) {
	peers := make([]*peering.OperatorPeerInfo, len(simulatedPeers))
	for i, simulatedPeer := range simulatedPeers {
		peer, err := NewSimulatedPeerFromConfig(simulatedPeer)
		if err != nil {
			return nil, err
		}
		peers[i] = peer
	}
	return peers, nil
}
