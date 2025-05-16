// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IAllocationManager} from "@eigenlayer-contracts/src/contracts/interfaces/IAllocationManager.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";
import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";

import {ITaskAVSRegistrar} from "../interfaces/avs/l1/ITaskAVSRegistrar.sol";
import {TaskAVSRegistrarBaseStorage} from "./TaskAVSRegistrarBaseStorage.sol";

/**
 * @title TaskAVSRegistrarBase
 * @notice Base contract for registering operators and managing BLS public keys for AVS tasks
 * @dev Extends EIP712 for signature validation and TaskAVSRegistrarBaseStorage for state variables
 */
abstract contract TaskAVSRegistrarBase is EIP712, TaskAVSRegistrarBaseStorage {
    // TODO: Decide if we want to make contract a transparent proxy with owner set up. And add Pausable and Ownable.

    using BN254 for BN254.G1Point;

    /// @notice Modifier to ensure the caller is the AllocationManager
    modifier onlyAllocationManager() {
        require(msg.sender == address(ALLOCATION_MANAGER), OnlyAllocationManager());
        _;
    }

    /**
     * @notice Constructs the TaskAVSRegistrarBase contract
     * @param avs The address of the AVS
     * @param allocationManager The AllocationManager contract address
     */
    constructor(
        address avs,
        IAllocationManager allocationManager
    ) EIP712("TaskAVSRegistrar", "v0.1.0") TaskAVSRegistrarBaseStorage(avs, allocationManager) {}

    /**
     *
     *                         EXTERNAL FUNCTIONS
     *
     */
    
    /**
     * @notice Registers an operator with the AVS
     * @param operator The address of the operator to register
     * @param avs The AVS address the operator is registering with
     * @param operatorSetIds The IDs of the operator sets to register the operator with
     * @param data Encoded registration parameters including pubkey and socket
     * @dev Only callable by the AllocationManager
     */
    function registerOperator(
        address operator,
        address avs,
        uint32[] calldata operatorSetIds,
        bytes calldata data
    ) external onlyAllocationManager {
        require(supportsAVS(avs), InvalidAVS());

        OperatorRegistrationParams memory operatorRegistrationParams = abi.decode(data, (OperatorRegistrationParams));

        // Pubkey can only be registered once, so we check if the operator has already registered a pubkey
        // TODO: Support updating pubkey
        bytes32 pubkeyHash = operatorToPubkeyHash[operator];
        if (pubkeyHash == bytes32(0)) {
            _registerBLSPublicKey(
                operator, operatorRegistrationParams.pubkeyRegistrationParams, pubkeyRegistrationMessageHash(operator)
            );
        }

        // Set the operator's socket
        _setOperatorSocket(operator, operatorRegistrationParams.socket);

        // Update current APK for each operatorSetId
        _processOperatorSetApkUpdate(
            operator, operatorSetIds, operatorRegistrationParams.pubkeyRegistrationParams.pubkeyG1
        );
    }

    /**
     * @notice Deregisters an operator from the AVS
     * @param operator The address of the operator to deregister
     * @param avs The AVS address the operator is deregistering from
     * @param operatorSetIds The IDs of the operator sets to deregister the operator from
     * @dev Only callable by the AllocationManager
     */
    function deregisterOperator(
        address operator,
        address avs,
        uint32[] calldata operatorSetIds
    ) external onlyAllocationManager {
        require(supportsAVS(avs), InvalidAVS());

        // Update current APK for each operatorSetId
        _processOperatorSetApkUpdate(operator, operatorSetIds, operatorToPubkey[operator].negate());
    }

    /// @inheritdoc ITaskAVSRegistrar
    function updateOperatorSocket(
        string memory socket
    ) external {
        // TODO: Should we check for UAM permissions here?
        OperatorSet[] memory operatorSets = ALLOCATION_MANAGER.getRegisteredSets(msg.sender);
        bool isRegisteredToAVS = false;
        for (uint256 i = 0; i < operatorSets.length; i++) {
            if (supportsAVS(operatorSets[i].avs)) {
                isRegisteredToAVS = true;
                break;
            }
        }
        require(isRegisteredToAVS, OperatorNotRegistered());
        _setOperatorSocket(msg.sender, socket);
    }

    /**
     *
     *                         INTERNAL FUNCTIONS
     *
     */
    
    /**
     * @notice Registers a BLS public key for an operator
     * @param operator The address of the operator
     * @param params The parameters for registering the pubkey
     * @param _pubkeyRegistrationMessageHash The message hash that should be signed
     * @dev Verifies the signature and registers the pubkey
     */
    function _registerBLSPublicKey(
        address operator,
        PubkeyRegistrationParams memory params,
        BN254.G1Point memory _pubkeyRegistrationMessageHash
    ) internal {
        bytes32 pubkeyHash = BN254.hashG1Point(params.pubkeyG1);
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
    }

    /**
     * @notice Updates an operator's socket address
     * @param operator The address of the operator
     * @param socket The new socket address to set for the operator
     * @dev Emits an OperatorSocketUpdate event after updating
     */
    function _setOperatorSocket(address operator, string memory socket) internal {
        bytes32 pubkeyHash = operatorToPubkeyHash[operator];
        // TODO: Do we need pubkeyHashToSocket storage mapping?
        pubkeyHashToSocket[pubkeyHash] = socket;
        operatorToSocket[operator] = socket;
        emit OperatorSocketUpdated(operator, pubkeyHash, socket);
    }

    /**
     * @notice Updates the aggregate public key (APK) for one or more operator sets
     * @param operator The address of the operator
     * @param operatorSetIds The IDs of the operator sets to update
     * @param point The BLS public key point to add or remove from the APK
     * @dev For registration, adds the point; for deregistration, adds the negation of the point
     */
    function _processOperatorSetApkUpdate(
        address operator,
        uint32[] memory operatorSetIds,
        BN254.G1Point memory point
    ) internal {
        bytes32 pubkeyHash = operatorToPubkeyHash[operator];

        BN254.G1Point memory newApk;
        for (uint256 i = 0; i < operatorSetIds.length; i++) {
            // Update aggregate public key for this operatorSet
            newApk = currentApk[operatorSetIds[i]].plus(point);
            currentApk[operatorSetIds[i]] = newApk;
            emit OperatorSetApkUpdated(operator, pubkeyHash, operatorSetIds[i], newApk);
        }
    }

    /**
     *
     *                         VIEW FUNCTIONS
     *
     */
    
    /**
     * @notice Checks if the contract supports a specific AVS
     * @param avs The address of the AVS to check
     * @return True if the AVS is supported, false otherwise
     */
    function supportsAVS(
        address avs
    ) public view returns (bool) {
        return avs == AVS;
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getApk(
        uint8 operatorSetId
    ) public view returns (BN254.G1Point memory) {
        return currentApk[operatorSetId];
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getRegisteredPubkeyInfo(
        address operator
    ) public view returns (PubkeyInfo memory) {
        BN254.G1Point memory pubkey = operatorToPubkey[operator];
        BN254.G2Point memory pubkeyG2 = operatorToPubkeyG2[operator];

        bytes32 pubkeyHash = getOperatorPubkeyHash(operator);
        require(pubkeyHash != bytes32(0), OperatorNotRegistered());

        return PubkeyInfo({pubkeyG1: pubkey, pubkeyG2: pubkeyG2, pubkeyHash: pubkeyHash});
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getRegisteredPubkey(
        address operator
    ) public view returns (BN254.G1Point memory, bytes32) {
        // TODO: Deprecate this function. Only added for backwards compatibility with BLSApkRegistry.
        BN254.G1Point memory pubkey = operatorToPubkey[operator];

        bytes32 pubkeyHash = getOperatorPubkeyHash(operator);
        require(pubkeyHash != bytes32(0), OperatorNotRegistered());

        return (pubkey, pubkeyHash);
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getOperatorPubkeyG2(
        address operator
    ) public view override returns (BN254.G2Point memory) {
        // TODO: Deprecate this function. Only added for backwards compatibility with BLSApkRegistry.
        return operatorToPubkeyG2[operator];
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getOperatorFromPubkeyHash(
        bytes32 pubkeyHash
    ) public view returns (address) {
        return pubkeyHashToOperator[pubkeyHash];
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getOperatorPubkeyHash(
        address operator
    ) public view returns (bytes32) {
        return operatorToPubkeyHash[operator];
    }

    /// @inheritdoc ITaskAVSRegistrar
    function pubkeyRegistrationMessageHash(
        address operator
    ) public view returns (BN254.G1Point memory) {
        return BN254.hashToG1(calculatePubkeyRegistrationMessageHash(operator));
    }

    /// @inheritdoc ITaskAVSRegistrar
    function calculatePubkeyRegistrationMessageHash(
        address operator
    ) public view returns (bytes32) {
        return _hashTypedDataV4(keccak256(abi.encode(PUBKEY_REGISTRATION_TYPEHASH, operator)));
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getOperatorSocketByPubkeyHash(
        bytes32 pubkeyHash
    ) public view returns (string memory) {
        return pubkeyHashToSocket[pubkeyHash];
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getOperatorSocketByOperator(
        address operator
    ) public view returns (string memory) {
        return operatorToSocket[operator];
    }

    /// @inheritdoc ITaskAVSRegistrar
    function getBatchOperatorPubkeyInfoAndSocket(
        address[] calldata operators
    ) public view returns (PubkeyInfoAndSocket[] memory) {
        PubkeyInfoAndSocket[] memory pubkeyInfosAndSockets = new PubkeyInfoAndSocket[](operators.length);
        for (uint256 i = 0; i < operators.length; i++) {
            pubkeyInfosAndSockets[i] = PubkeyInfoAndSocket({
                pubkeyInfo: getRegisteredPubkeyInfo(operators[i]),
                socket: getOperatorSocketByOperator(operators[i])
            });
        }
        return pubkeyInfosAndSockets;
    }

    /// @inheritdoc ITaskAVSRegistrar
    function packRegisterPayload(
        string memory socket,
        PubkeyRegistrationParams memory pubkeyRegistrationParams
    ) public pure returns (bytes memory) {
        return
            abi.encode(OperatorRegistrationParams({socket: socket, pubkeyRegistrationParams: pubkeyRegistrationParams}));
    }
}
