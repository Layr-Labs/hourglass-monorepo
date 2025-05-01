// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAllocationManager} from
    "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {EIP712} from "@eigenlayer-middleware/lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

import {TaskAVSRegistrarStorage} from "src/avs/l1-contracts/TaskAVSRegistrarStorage.sol";

contract TaskAVSRegistrar is EIP712, TaskAVSRegistrarStorage {
    using BN254 for BN254.G1Point;

    /// @notice Modifier to ensure the caller is the AllocationManager
    modifier onlyAllocationManager() {
        require(msg.sender == address(ALLOCATION_MANAGER), OnlyAllocationManager());
        _;
    }

    constructor(
        address avs,
        IAllocationManager allocationManager
    ) EIP712("TaskAVSRegistrar", "1") TaskAVSRegistrarStorage(avs, allocationManager) {}

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    function registerOperator(
        address operator,
        address avs,
        uint32[] calldata /* operatorSetIds */,
        bytes calldata data
    ) external onlyAllocationManager {
        // TODO: Consider if we want to checkpoint registration params at specific block heights/timestamps and do the quorum apk update within this function.
        require(supportsAVS(avs), InvalidAVS());

        OperatorRegistrationParams memory operatorRegistrationParams = abi.decode(data, (OperatorRegistrationParams));

        /**
         * If the operator has NEVER registered a pubkey before, use `params` to register
         * their pubkey
         *
         * If the operator HAS registered a pubkey, `params` is ignored and the pubkey hash
         * (pubkeyHash) is fetched instead
         */
        bytes32 pubkeyHash = _getOrRegisterOperatorPubkeyHash(
            operator, operatorRegistrationParams.pubkeyRegistrationParams, pubkeyRegistrationMessageHash(operator)
        );

        _setOperatorSocket(operator, pubkeyHash, operatorRegistrationParams.socket);
    }

    function deregisterOperator(
        address, /* operator */
        address avs,
        uint32[] calldata /* operatorSetIds */
    ) external view onlyAllocationManager {
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
    function _getOrRegisterOperatorPubkeyHash(
        address operator,
        PubkeyRegistrationParams memory params,
        BN254.G1Point memory _pubkeyRegistrationMessageHash
    ) internal returns (bytes32 pubkeyHash) {
        pubkeyHash = getOperatorPubkeyHash(operator);
        if (pubkeyHash == bytes32(0)) {
            pubkeyHash = _registerBLSPublicKey(operator, params, _pubkeyRegistrationMessageHash);
        }
        return pubkeyHash;
    }

    function _registerBLSPublicKey(
        address operator,
        PubkeyRegistrationParams memory params,
        BN254.G1Point memory _pubkeyRegistrationMessageHash
    ) internal returns (bytes32 pubkeyHash) {
        pubkeyHash = BN254.hashG1Point(params.pubkeyG1);
        require(pubkeyHash != ZERO_PK_HASH, ZeroPubKey());
        require(getOperatorPubkeyHash(operator) == bytes32(0), OperatorAlreadyRegistered());
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
                    _pubkeyRegistrationMessageHash.X,
                    _pubkeyRegistrationMessageHash.Y
                )
            )
        ) % BN254.FR_MODULUS;

        // e(sigma + P * gamma, [-1]_2) = e(H(m) + [1]_1 * gamma, P')
        require(
            BN254.pairing(
                params.pubkeyRegistrationSignature.plus(params.pubkeyG1.scalar_mul(gamma)),
                BN254.negGeneratorG2(),
                _pubkeyRegistrationMessageHash.plus(BN254.generatorG1().scalar_mul(gamma)),
                params.pubkeyG2
            ),
            InvalidBLSSignatureOrPrivateKey()
        );

        operatorToPubkey[operator] = params.pubkeyG1;
        operatorToPubkeyG2[operator] = params.pubkeyG2;
        operatorToPubkeyHash[operator] = pubkeyHash;
        pubkeyHashToOperator[pubkeyHash] = operator;

        emit NewPubkeyRegistration(operator, pubkeyHash, params.pubkeyG1, params.pubkeyG2);
        return pubkeyHash;
    }

    /**
     * @notice Updates an operator's socket address in the SocketRegistry
     * @param operator The address of the operator
     * @param pubkeyHash The unique identifier of the operator
     * @param socket The new socket address to set for the operator
     * @dev Emits an OperatorSocketUpdate event after updating
     */
    function _setOperatorSocket(address operator, bytes32 pubkeyHash, string memory socket) internal {
        pubkeyHashToSocket[pubkeyHash] = socket;
        operatorToSocket[operator] = socket;
        emit OperatorSocketUpdated(operator, pubkeyHash, socket);
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
        bytes32 pubkeyHash = getOperatorPubkeyHash(operator);

        require(pubkeyHash != bytes32(0), OperatorNotRegistered());

        return (pubkey, pubkeyHash);
    }

    function getOperatorFromPubkeyHash(
        bytes32 pubkeyHash
    ) public view returns (address) {
        return pubkeyHashToOperator[pubkeyHash];
    }

    function getOperatorPubkeyHash(
        address operator
    ) public view returns (bytes32) {
        return operatorToPubkeyHash[operator];
    }

    function getOperatorPubkeyG2(
        address operator
    ) public view returns (BN254.G2Point memory) {
        return operatorToPubkeyG2[operator];
    }

    /**
     * @notice Returns the message hash that an operator must sign to register their BLS public key.
     * @param operator is the address of the operator registering their BLS public key
     */
    function pubkeyRegistrationMessageHash(
        address operator
    ) public view returns (BN254.G1Point memory) {
        return BN254.hashToG1(calculatePubkeyRegistrationMessageHash(operator));
    }

    /**
     * @notice Returns the message hash that an operator must sign to register their BLS public key.
     * @param operator is the address of the operator registering their BLS public key
     */
    function calculatePubkeyRegistrationMessageHash(
        address operator
    ) public view returns (bytes32) {
        return _hashTypedDataV4(keccak256(abi.encode(PUBKEY_REGISTRATION_TYPEHASH, operator)));
    }

    function getOperatorSocketByPubkeyHash(
        bytes32 pubkeyHash
    ) external view returns (string memory) {
        return pubkeyHashToSocket[pubkeyHash];
    }

    function getOperatorSocketByOperator(
        address operator
    ) external view returns (string memory) {
        return operatorToSocket[operator];
    }
}
