package peers

import (
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
)

func NewSimulatedPeerFromConfig(simulatedPeer config.SimulatedPeer) (*peering.OperatorPeerInfo, error) {
	pubKey, err := bn254.NewPublicKeyFromHexString(simulatedPeer.PublicKey)
	if err != nil {
		return nil, err
	}
	return &peering.OperatorPeerInfo{
		OperatorAddress: simulatedPeer.OperatorAddress,
		OperatorSets: []*peering.OperatorSet{
			{
				OperatorSetID:  simulatedPeer.OperatorSetId,
				PublicKey:      pubKey,
				NetworkAddress: simulatedPeer.NetworkAddress,
			},
		},
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
