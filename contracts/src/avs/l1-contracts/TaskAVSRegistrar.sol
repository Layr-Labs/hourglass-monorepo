// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAllocationManager} from
    "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

import {TaskAVSRegistrarStorage} from "src/avs/l1-contracts/TaskAVSRegistrarStorage.sol";

contract TaskAVSRegistrar is TaskAVSRegistrarStorage {
    using BN254 for BN254.G1Point;

    /// @notice Modifier to ensure the caller is the AllocationManager
    modifier onlyAllocationManager() {
        require(msg.sender == address(ALLOCATION_MANAGER), OnlyAllocationManager());
        _;
    }

    constructor(address avs, IAllocationManager allocationManager) TaskAVSRegistrarStorage(avs, allocationManager) {}

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function registerOperator(
        address operator,
        address avs,
        uint32[] calldata operatorSetIds,
        bytes calldata data
    ) external onlyAllocationManager {
        require(supportsAVS(avs), InvalidAVS());

        OperatorRegistrationParams memory operatorRegistrationParams = abi.decode(data, (OperatorRegistrationParams));
        // TODO: Consider if we want to checkpoint registration params at specific block heights/timestamps.
        // TODO: Implement
    }

    function deregisterOperator(
        address, /* operator */
        address avs,
        uint32[] calldata /* operatorSetIds */
    ) external onlyAllocationManager {
        require(supportsAVS(avs), InvalidAVS());
        // TODO: Implement any additional logic for deregistering an operator.
    }

    function supportsAVS(
        address avs
    ) public view returns (bool) {
        return avs == AVS;
    }

    /**
     *
     *                         INTERNAL FUNCTIONS
     *
     */
    function _registerBLSPublicKey(
        address operator,
        PubkeyRegistrationParams calldata params,
        BN254.G1Point calldata pubkeyRegistrationMessageHash
    ) internal returns (bytes32 operatorId) {
        bytes32 pubkeyHash = BN254.hashG1Point(params.pubkeyG1);
        require(pubkeyHash != ZERO_PK_HASH, ZeroPubKey());
        require(getOperatorId(operator) == bytes32(0), OperatorAlreadyRegistered());
        require(pubkeyHashToOperator[pubkeyHash] == address(0), BLSPubkeyAlreadyRegistered());

        // gamma = h(sigma, P, P', H(m))
        uint256 gamma = uint256(
            keccak256(
                abi.encodePacked(
                    params.pubkeyRegistrationSignature.X,
                    params.pubkeyRegistrationSignature.Y,
                    params.pubkeyG1.X,
                    params.pubkeyG1.Y,
                    params.pubkeyG2.X,
                    params.pubkeyG2.Y,
                    pubkeyRegistrationMessageHash.X,
                    pubkeyRegistrationMessageHash.Y
                )
            )
        ) % BN254.FR_MODULUS;

        // e(sigma + P * gamma, [-1]_2) = e(H(m) + [1]_1 * gamma, P')
        require(
            BN254.pairing(
                params.pubkeyRegistrationSignature.plus(params.pubkeyG1.scalar_mul(gamma)),
                BN254.negGeneratorG2(),
                pubkeyRegistrationMessageHash.plus(BN254.generatorG1().scalar_mul(gamma)),
                params.pubkeyG2
            ),
            InvalidBLSSignatureOrPrivateKey()
        );

        operatorToPubkey[operator] = params.pubkeyG1;
        operatorToPubkeyG2[operator] = params.pubkeyG2;
        operatorToPubkeyHash[operator] = pubkeyHash;
        pubkeyHashToOperator[pubkeyHash] = operator;

        emit NewPubkeyRegistration(operator, params.pubkeyG1, params.pubkeyG2);
        return pubkeyHash;
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */

     function getRegisteredPubkey(
        address operator
    ) public view returns (BN254.G1Point memory, bytes32) {
        BN254.G1Point memory pubkey = operatorToPubkey[operator];
        bytes32 pubkeyHash = getOperatorId(operator);

        require(pubkeyHash != bytes32(0), OperatorNotRegistered());

        return (pubkey, pubkeyHash);
    }

    function getOperatorFromPubkeyHash(
        bytes32 pubkeyHash
    ) public view returns (address) {
        return pubkeyHashToOperator[pubkeyHash];
    }

    function getOperatorId(
        address operator
    ) public view returns (bytes32) {
        return operatorToPubkeyHash[operator];
    }

    function getOperatorPubkeyG2(
        address operator
    ) public view returns (BN254.G2Point memory) {
        return operatorToPubkeyG2[operator];
    }
}
