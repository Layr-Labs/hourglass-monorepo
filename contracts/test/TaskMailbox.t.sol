// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Test, console, Vm} from "forge-std/Test.sol";
import {ProxyAdmin} from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import {TransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import {ITransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import {
    IBN254CertificateVerifier,
    IBN254CertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {
    IECDSACertificateVerifier,
    IECDSACertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IECDSACertificateVerifier.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {BN254} from "@eigenlayer-contracts/src/contracts/libraries/BN254.sol";
import {IKeyRegistrarTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IKeyRegistrar.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {TaskMailbox} from "../src/core/TaskMailbox.sol";
import {
    ITaskMailbox,
    ITaskMailboxTypes,
    ITaskMailboxErrors,
    ITaskMailboxEvents
} from "../src/interfaces/core/ITaskMailbox.sol";

import {IAVSTaskHook} from "../src/interfaces/avs/l2/IAVSTaskHook.sol";
import {MockAVSTaskHook} from "./mocks/MockAVSTaskHook.sol";
import {MockBN254CertificateVerifier} from "./mocks/MockBN254CertificateVerifier.sol";
import {MockBN254CertificateVerifierFailure} from "./mocks/MockBN254CertificateVerifierFailure.sol";
import {MockECDSACertificateVerifier} from "./mocks/MockECDSACertificateVerifier.sol";
import {MockECDSACertificateVerifierFailure} from "./mocks/MockECDSACertificateVerifierFailure.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {ReentrantAttacker} from "./mocks/ReentrantAttacker.sol";

contract TaskMailboxUnitTests is Test, ITaskMailboxTypes, ITaskMailboxErrors, ITaskMailboxEvents {
    using OperatorSetLib for OperatorSet;

    // Contracts
    TaskMailbox public taskMailbox;
    ProxyAdmin public proxyAdmin;
    MockAVSTaskHook public mockTaskHook;
    MockBN254CertificateVerifier public mockBN254CertificateVerifier;
    MockECDSACertificateVerifier public mockECDSACertificateVerifier;
    MockERC20 public mockToken;

    // Test addresses
    address public avs = address(0x1);
    address public avs2 = address(0x2);
    address public feeCollector = address(0x3);
    address public refundCollector = address(0x4);
    address public creator = address(0x5);
    address public aggregator = address(0x6);

    // Test operator set IDs
    uint32 public executorOperatorSetId = 1;
    uint32 public executorOperatorSetId2 = 2;

    // Test config values
    uint96 public taskSLA = 60 seconds;
    uint16 public stakeProportionThreshold = 6667; // 66.67%
    uint96 public avsFee = 1 ether;

    function setUp() public virtual {
        // Deploy mock contracts
        mockTaskHook = new MockAVSTaskHook();
        mockBN254CertificateVerifier = new MockBN254CertificateVerifier();
        mockECDSACertificateVerifier = new MockECDSACertificateVerifier();
        mockToken = new MockERC20();

        // Deploy TaskMailbox with proxy pattern
        proxyAdmin = new ProxyAdmin();
        TaskMailbox taskMailboxImpl =
            new TaskMailbox(address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier), "1.0.0");
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskMailboxImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskMailbox.initialize.selector, address(this), 0, address(1))
        );
        taskMailbox = TaskMailbox(address(proxy));

        // Give creator some tokens and approve TaskMailbox
        mockToken.mint(creator, 1000 ether);
        vm.prank(creator);
        mockToken.approve(address(taskMailbox), type(uint256).max);
    }

    function _createValidTaskParams() internal view returns (TaskParams memory) {
        return TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: OperatorSet(avs, executorOperatorSetId),
            payload: bytes("test payload")
        });
    }

    function _createValidExecutorOperatorSetTaskConfig() internal view returns (ExecutorOperatorSetTaskConfig memory) {
        return ExecutorOperatorSetTaskConfig({
            taskHook: IAVSTaskHook(address(mockTaskHook)),
            taskSLA: taskSLA,
            feeToken: IERC20(address(mockToken)),
            curveType: IKeyRegistrarTypes.CurveType.BN254,
            feeCollector: feeCollector,
            consensus: Consensus({
                consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
                value: abi.encode(stakeProportionThreshold)
            }),
            taskMetadata: bytes("test metadata")
        });
    }

    function _createValidBN254Certificate(
        bytes32 messageHash
    ) internal view returns (IBN254CertificateVerifierTypes.BN254Certificate memory) {
        return IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: messageHash,
            signature: BN254.G1Point(1, 2), // Non-zero signature
            apk: BN254.G2Point([uint256(1), uint256(2)], [uint256(3), uint256(4)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });
    }

    function _createValidECDSACertificate(
        bytes32 messageHash
    ) internal view returns (IECDSACertificateVerifierTypes.ECDSACertificate memory) {
        return IECDSACertificateVerifierTypes.ECDSACertificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: messageHash,
            sig: bytes("0x1234567890abcdef") // Non-empty signature
        });
    }
}

contract TaskMailboxUnitTests_Constructor is TaskMailboxUnitTests {
    function test_Constructor_WithCertificateVerifiers() public {
        address bn254Verifier = address(0x1234);
        address ecdsaVerifier = address(0x5678);

        // Deploy with proxy pattern
        ProxyAdmin proxyAdmin = new ProxyAdmin();
        TaskMailbox taskMailboxImpl = new TaskMailbox(bn254Verifier, ecdsaVerifier, "1.0.0");
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskMailboxImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskMailbox.initialize.selector, address(this), 0, address(1))
        );
        TaskMailbox newTaskMailbox = TaskMailbox(address(proxy));

        assertEq(newTaskMailbox.BN254_CERTIFICATE_VERIFIER(), bn254Verifier);
        assertEq(newTaskMailbox.ECDSA_CERTIFICATE_VERIFIER(), ecdsaVerifier);
        assertEq(newTaskMailbox.version(), "1.0.0");
        assertEq(newTaskMailbox.owner(), address(this));
        assertEq(newTaskMailbox.getFeeSplit(), 0);
        assertEq(newTaskMailbox.getFeeSplitCollector(), address(1));
    }
}

// Test contract for registerExecutorOperatorSet
contract TaskMailboxUnitTests_registerExecutorOperatorSet is TaskMailboxUnitTests {
    function testFuzz_registerExecutorOperatorSet(
        address fuzzAvs,
        uint32 fuzzOperatorSetId,
        bool fuzzIsRegistered
    ) public {
        // Skip if fuzzAvs is the proxy admin to avoid proxy admin access issues
        vm.assume(fuzzAvs != address(proxyAdmin));
        OperatorSet memory operatorSet = OperatorSet(fuzzAvs, fuzzOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // Set config first (requirement for registerExecutorOperatorSet)
        vm.prank(fuzzAvs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // If unregistering, expect event
        if (!fuzzIsRegistered) {
            vm.expectEmit(true, true, true, true, address(taskMailbox));
            emit ExecutorOperatorSetRegistered(fuzzAvs, fuzzAvs, fuzzOperatorSetId, fuzzIsRegistered);

            // Register operator set
            vm.prank(fuzzAvs);
            taskMailbox.registerExecutorOperatorSet(operatorSet, fuzzIsRegistered);
        }

        // Verify registration status
        assertEq(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()), fuzzIsRegistered);
    }

    function test_registerExecutorOperatorSet_Unregister() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // Set config (this automatically registers the operator set)
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
        assertTrue(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));

        // Then unregister
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit ExecutorOperatorSetRegistered(avs, avs, executorOperatorSetId, false);

        vm.prank(avs);
        taskMailbox.registerExecutorOperatorSet(operatorSet, false);
        assertFalse(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));
    }

    function test_Revert_registerExecutorOperatorSet_ConfigNotSet() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);

        vm.prank(avs);
        vm.expectRevert(ExecutorOperatorSetTaskConfigNotSet.selector);
        taskMailbox.registerExecutorOperatorSet(operatorSet, true);
    }

    function test_Revert_registerExecutorOperatorSet_InvalidOperatorSetOwner() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // Set config as avs
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Try to register as different address
        vm.prank(avs2);
        vm.expectRevert(InvalidOperatorSetOwner.selector);
        taskMailbox.registerExecutorOperatorSet(operatorSet, false);
    }
}

// Test contract for setExecutorOperatorSetTaskConfig
contract TaskMailboxUnitTests_setExecutorOperatorSetTaskConfig is TaskMailboxUnitTests {
    function test_Revert_InvalidOperatorSetOwner() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // Try to set config as wrong address
        vm.prank(avs2);
        vm.expectRevert(InvalidOperatorSetOwner.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function testFuzz_setExecutorOperatorSetTaskConfig(
        address fuzzCertificateVerifier,
        address fuzzTaskHook,
        address fuzzFeeToken,
        address fuzzFeeCollector,
        uint96 fuzzTaskSLA,
        uint16 fuzzStakeProportionThreshold,
        bytes memory fuzzTaskMetadata
    ) public {
        // Bound inputs
        vm.assume(fuzzCertificateVerifier != address(0));
        vm.assume(fuzzTaskHook != address(0));
        vm.assume(fuzzTaskSLA > 0);
        // Bound stake proportion threshold to valid range (0-10000 basis points)
        fuzzStakeProportionThreshold = uint16(bound(fuzzStakeProportionThreshold, 0, 10_000));

        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);

        ExecutorOperatorSetTaskConfig memory config = ExecutorOperatorSetTaskConfig({
            taskHook: IAVSTaskHook(fuzzTaskHook),
            taskSLA: fuzzTaskSLA,
            feeToken: IERC20(fuzzFeeToken),
            curveType: IKeyRegistrarTypes.CurveType.BN254,
            feeCollector: fuzzFeeCollector,
            consensus: Consensus({
                consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
                value: abi.encode(fuzzStakeProportionThreshold)
            }),
            taskMetadata: fuzzTaskMetadata
        });

        // Since setExecutorOperatorSetTaskConfig always registers if not already registered,
        // we expect both events every time for a new operator set
        // Note: The contract emits config event first, then registration event

        // Expect config event first
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit ExecutorOperatorSetTaskConfigSet(avs, avs, executorOperatorSetId, config);

        // Expect registration event second
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit ExecutorOperatorSetRegistered(avs, avs, executorOperatorSetId, true);

        // Set config
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Verify config was set
        ExecutorOperatorSetTaskConfig memory retrievedConfig = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);

        assertEq(uint8(retrievedConfig.curveType), uint8(IKeyRegistrarTypes.CurveType.BN254));
        assertEq(address(retrievedConfig.taskHook), fuzzTaskHook);
        assertEq(address(retrievedConfig.feeToken), fuzzFeeToken);
        assertEq(retrievedConfig.feeCollector, fuzzFeeCollector);
        assertEq(retrievedConfig.taskSLA, fuzzTaskSLA);
        // Verify consensus configuration
        assertEq(uint8(retrievedConfig.consensus.consensusType), uint8(ConsensusType.STAKE_PROPORTION_THRESHOLD));
        uint16 decodedThreshold = abi.decode(retrievedConfig.consensus.value, (uint16));
        assertEq(decodedThreshold, fuzzStakeProportionThreshold);
        assertEq(retrievedConfig.taskMetadata, fuzzTaskMetadata);

        // Verify operator set is registered
        assertTrue(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));
    }

    function test_setExecutorOperatorSetTaskConfig_AlreadyRegistered() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // First set config (which auto-registers)
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
        assertTrue(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));

        // Update config again
        config.taskSLA = 120;

        // Should not emit registration event since already registered
        vm.recordLogs();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Verify only one event was emitted (config set, not registration)
        Vm.Log[] memory entries = vm.getRecordedLogs();
        assertEq(entries.length, 1);
        assertEq(
            entries[0].topics[0],
            keccak256(
                "ExecutorOperatorSetTaskConfigSet(address,address,uint32,(address,uint96,address,uint8,address,(uint8,bytes),bytes))"
            )
        );

        // Verify the config was updated
        ExecutorOperatorSetTaskConfig memory updatedConfig = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);
        assertEq(updatedConfig.taskSLA, 120);
    }

    function test_Revert_WhenCurveTypeIsNone() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.curveType = IKeyRegistrarTypes.CurveType.NONE;

        // Expecting revert due to accessing zero address certificate verifier
        vm.prank(avs);
        vm.expectRevert();
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenTaskHookIsZero() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskHook = IAVSTaskHook(address(0));

        vm.prank(avs);
        vm.expectRevert(InvalidAddressZero.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenTaskSLAIsZero() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskSLA = 0;

        vm.prank(avs);
        vm.expectRevert(TaskSLAIsZero.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenConsensusValueInvalid_EmptyBytes() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD, value: bytes("")});

        vm.prank(avs);
        vm.expectRevert(InvalidConsensusValue.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenConsensusValueInvalid_WrongLength() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encodePacked(uint8(50)) // Wrong size - should be 32 bytes
        });

        vm.prank(avs);
        vm.expectRevert(InvalidConsensusValue.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenConsensusValueInvalid_ExceedsMaximum() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(10_001)) // Exceeds 10000 basis points
        });

        vm.prank(avs);
        vm.expectRevert(InvalidConsensusValue.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_ConsensusZeroThreshold() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(0)) // Zero threshold is valid
        });

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Verify config was set
        ExecutorOperatorSetTaskConfig memory retrievedConfig = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);
        uint16 decodedThreshold = abi.decode(retrievedConfig.consensus.value, (uint16));
        assertEq(decodedThreshold, 0);
    }

    function test_ConsensusMaxThreshold() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(10_000)) // Maximum 100%
        });

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Verify config was set
        ExecutorOperatorSetTaskConfig memory retrievedConfig = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);
        uint16 decodedThreshold = abi.decode(retrievedConfig.consensus.value, (uint16));
        assertEq(decodedThreshold, 10_000);
    }

    function test_Revert_WhenConsensusTypeIsNone() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.NONE,
            value: bytes("") // Empty value for NONE type
        });

        vm.prank(avs);
        vm.expectRevert(InvalidConsensusType.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }
}

// Test contract for setFeeSplit
contract TaskMailboxUnitTests_setFeeSplit is TaskMailboxUnitTests {
    function test_SetFeeSplit() public {
        uint16 newFeeSplit = 2000; // 20%

        vm.expectEmit(true, true, true, true);
        emit FeeSplitSet(newFeeSplit);

        taskMailbox.setFeeSplit(newFeeSplit);
        assertEq(taskMailbox.getFeeSplit(), newFeeSplit);
    }

    function test_SetFeeSplit_MaxValue() public {
        uint16 maxFeeSplit = 10_000; // 100%

        vm.expectEmit(true, true, true, true);
        emit FeeSplitSet(maxFeeSplit);

        taskMailbox.setFeeSplit(maxFeeSplit);
        assertEq(taskMailbox.getFeeSplit(), maxFeeSplit);
    }

    function test_Revert_SetFeeSplit_NotOwner() public {
        vm.prank(address(0x999));
        vm.expectRevert("Ownable: caller is not the owner");
        taskMailbox.setFeeSplit(1000);
    }

    function test_Revert_SetFeeSplit_ExceedsMax() public {
        vm.expectRevert(InvalidFeeSplit.selector);
        taskMailbox.setFeeSplit(10_001); // > 100%
    }
}

// Test contract for setFeeSplitCollector
contract TaskMailboxUnitTests_setFeeSplitCollector is TaskMailboxUnitTests {
    function test_SetFeeSplitCollector() public {
        address newCollector = address(0x123);

        vm.expectEmit(true, true, true, true);
        emit FeeSplitCollectorSet(newCollector);

        taskMailbox.setFeeSplitCollector(newCollector);
        assertEq(taskMailbox.getFeeSplitCollector(), newCollector);
    }

    function test_Revert_SetFeeSplitCollector_NotOwner() public {
        vm.prank(address(0x999));
        vm.expectRevert("Ownable: caller is not the owner");
        taskMailbox.setFeeSplitCollector(address(0x123));
    }

    function test_Revert_SetFeeSplitCollector_ZeroAddress() public {
        vm.expectRevert(InvalidAddressZero.selector);
        taskMailbox.setFeeSplitCollector(address(0));
    }
}

// Test contract for createTask
contract TaskMailboxUnitTests_createTask is TaskMailboxUnitTests {
    function setUp() public override {
        super.setUp();

        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function testFuzz_createTask(address fuzzRefundCollector, uint96 fuzzAvsFee, bytes memory fuzzPayload) public {
        // Bound inputs
        vm.assume(fuzzPayload.length > 0);
        // We create two tasks in this test, so need at least 2x the fee
        vm.assume(fuzzAvsFee <= mockToken.balanceOf(creator) / 2);

        // Set the mock hook to return the fuzzed fee
        mockTaskHook.setDefaultFee(fuzzAvsFee);

        TaskParams memory taskParams = TaskParams({
            refundCollector: fuzzRefundCollector,
            executorOperatorSet: OperatorSet(avs, executorOperatorSetId),
            payload: fuzzPayload
        });

        // First task will have count 0
        uint256 expectedTaskCount = 0;
        bytes32 expectedTaskHash =
            keccak256(abi.encode(expectedTaskCount, address(taskMailbox), block.chainid, taskParams));

        // Expect event
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit TaskCreated(
            creator,
            expectedTaskHash,
            avs,
            executorOperatorSetId,
            fuzzRefundCollector,
            fuzzAvsFee,
            block.timestamp + taskSLA,
            fuzzPayload
        );

        // Create task
        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Verify task hash
        assertEq(taskHash, expectedTaskHash);

        // Verify global task count incremented by creating another task and checking its hash
        bytes32 nextExpectedTaskHash =
            keccak256(abi.encode(expectedTaskCount + 1, address(taskMailbox), block.chainid, taskParams));
        vm.prank(creator);
        bytes32 nextTaskHash = taskMailbox.createTask(taskParams);
        assertEq(nextTaskHash, nextExpectedTaskHash);

        // Verify task was created
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.creator, creator);
        assertEq(task.creationTime, block.timestamp);
        assertEq(uint8(task.status), uint8(TaskStatus.CREATED));
        assertEq(task.avs, avs);
        assertEq(task.executorOperatorSetId, executorOperatorSetId);
        assertEq(task.refundCollector, fuzzRefundCollector);
        assertEq(task.avsFee, fuzzAvsFee);
        assertEq(task.feeSplit, 0);
        assertEq(task.payload, fuzzPayload);

        // Verify token transfer if fee > 0
        // Note: We created two tasks with the same fee, so balance should be 2 * fuzzAvsFee
        if (fuzzAvsFee > 0) {
            assertEq(mockToken.balanceOf(address(taskMailbox)), fuzzAvsFee * 2);
        }
    }

    function test_createTask_ZeroFee() public {
        // Set the mock hook to return 0 fee
        mockTaskHook.setDefaultFee(0);

        TaskParams memory taskParams = _createValidTaskParams();

        uint256 balanceBefore = mockToken.balanceOf(address(taskMailbox));

        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Verify no token transfer occurred
        assertEq(mockToken.balanceOf(address(taskMailbox)), balanceBefore);

        // Verify task was created with zero fee
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.avsFee, 0);
    }

    function test_createTask_NoFeeToken() public {
        // Set up config without fee token
        OperatorSet memory operatorSet = OperatorSet(avs2, executorOperatorSetId2);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = IERC20(address(0));

        vm.prank(avs2);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        TaskParams memory taskParams = TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        uint256 balanceBefore = mockToken.balanceOf(address(taskMailbox));

        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Verify no token transfer occurred even with non-zero fee
        assertEq(mockToken.balanceOf(address(taskMailbox)), balanceBefore);

        // Verify task was created
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.avsFee, 1 ether);
    }

    function test_Revert_WhenPayloadIsEmpty() public {
        TaskParams memory taskParams = _createValidTaskParams();
        taskParams.payload = bytes("");

        vm.prank(creator);
        vm.expectRevert(PayloadIsEmpty.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_WhenExecutorOperatorSetNotRegistered() public {
        TaskParams memory taskParams = _createValidTaskParams();
        taskParams.executorOperatorSet.id = 99; // Unregistered operator set

        vm.prank(creator);
        vm.expectRevert(ExecutorOperatorSetNotRegistered.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_WhenExecutorOperatorSetTaskConfigNotSet() public {
        // Create an operator set that has never been configured
        OperatorSet memory operatorSet = OperatorSet(avs2, executorOperatorSetId2);

        TaskParams memory taskParams = TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        // Should revert because operator set is not registered (no config set)
        vm.prank(creator);
        vm.expectRevert(ExecutorOperatorSetNotRegistered.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_ReentrancyOnCreateTask() public {
        // Deploy reentrant attacker as task hook
        ReentrantAttacker attacker = new ReentrantAttacker(address(taskMailbox));

        // Set up executor operator set with attacker as hook
        OperatorSet memory operatorSet = OperatorSet(avs2, executorOperatorSetId2);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskHook = IAVSTaskHook(address(attacker));

        vm.prank(avs2);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Give attacker tokens and approve
        mockToken.mint(address(attacker), 1000 ether);
        vm.prank(address(attacker));
        mockToken.approve(address(taskMailbox), type(uint256).max);

        // Set up attack parameters
        TaskParams memory taskParams = TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        attacker.setAttackParams(
            taskParams,
            bytes32(0),
            _createValidBN254Certificate(bytes32(0)),
            bytes(""),
            true, // attack on post
            true // attack createTask
        );

        // Try to create task - should revert on reentrancy
        vm.prank(creator);
        vm.expectRevert("ReentrancyGuard: reentrant call");
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_createTask_InvalidFeeReceiver_RefundCollector() public {
        // Set up operator set with fee token
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = mockToken;
        config.feeCollector = feeCollector;

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task params with zero refund collector
        TaskParams memory taskParams =
            TaskParams({refundCollector: address(0), executorOperatorSet: operatorSet, payload: bytes("test payload")});

        // Should revert with InvalidFeeReceiver when refundCollector is zero
        vm.prank(creator);
        vm.expectRevert(InvalidFeeReceiver.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_createTask_InvalidFeeReceiver_FeeCollector() public {
        // Set up operator set with fee token but zero fee collector
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = mockToken;
        config.feeCollector = address(0); // Zero fee collector

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create valid task params
        TaskParams memory taskParams = _createValidTaskParams();

        // Should revert with InvalidFeeReceiver when feeCollector is zero
        vm.prank(creator);
        vm.expectRevert(InvalidFeeReceiver.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_createTask_ValidWithZeroFeeReceivers_NoFeeToken() public {
        // When there's no fee token, zero addresses should be allowed
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = IERC20(address(0)); // No fee token
        config.feeCollector = address(0); // Zero fee collector is OK when no fee token

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task params with zero refund collector
        TaskParams memory taskParams = TaskParams({
            refundCollector: address(0), // Zero refund collector is OK when no fee token
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        // Should succeed when there's no fee token
        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Verify task was created
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.creator, creator);
        assertEq(task.refundCollector, address(0));
    }

    function test_createTask_CapturesFeeSplitValues() public {
        // Set fee split values
        uint16 feeSplit = 1500; // 15%
        address feeSplitCollector = address(0x456);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Verify task captured current fee split value
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.feeSplit, feeSplit);

        // Change fee split values
        uint16 newFeeSplit = 3000; // 30%
        address newFeeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(newFeeSplit);
        taskMailbox.setFeeSplitCollector(newFeeSplitCollector);

        // Create another task
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Verify new task has new fee split while old task retains old value
        Task memory newTask = taskMailbox.getTaskInfo(newTaskHash);
        assertEq(newTask.feeSplit, newFeeSplit);

        // Verify old task still has old fee split value
        task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.feeSplit, feeSplit);

        // Verify that the global feeSplitCollector is used (not captured in task)
        assertEq(taskMailbox.feeSplitCollector(), newFeeSplitCollector);
    }
}

// Test contract for submitResult
contract TaskMailboxUnitTests_submitResult is TaskMailboxUnitTests {
    bytes32 public taskHash;

    function setUp() public override {
        super.setUp();

        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        taskHash = taskMailbox.createTask(taskParams);
    }

    function testFuzz_submitResult_WithBN254Certificate(
        bytes memory fuzzResult
    ) public {
        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        // Expect event
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit TaskVerified(aggregator, taskHash, avs, executorOperatorSetId, abi.encode(cert), fuzzResult);

        // Submit result
        vm.prank(aggregator);
        taskMailbox.submitResult(taskHash, abi.encode(cert), fuzzResult);

        // Verify task was verified
        TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));

        // Verify result was stored
        bytes memory storedResult = taskMailbox.getTaskResult(taskHash);
        assertEq(storedResult, fuzzResult);

        // Verify certificate was stored
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.executorCert, abi.encode(cert));
    }

    function testFuzz_submitResult_WithECDSACertificate(
        bytes memory fuzzResult
    ) public {
        // Setup executor operator set with ECDSA curve type
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.curveType = IKeyRegistrarTypes.CurveType.ECDSA;

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        // Create ECDSA certificate
        IECDSACertificateVerifierTypes.ECDSACertificate memory cert = _createValidECDSACertificate(newTaskHash);

        // Expect event
        vm.expectEmit(true, true, true, true);
        emit TaskVerified(aggregator, newTaskHash, avs, executorOperatorSetId, abi.encode(cert), fuzzResult);

        // Submit result with ECDSA certificate
        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), fuzzResult);

        // Verify task was verified
        TaskStatus status = taskMailbox.getTaskStatus(newTaskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));

        // Verify result was stored
        bytes memory storedResult = taskMailbox.getTaskResult(newTaskHash);
        assertEq(storedResult, fuzzResult);

        // Verify certificate was stored
        Task memory task = taskMailbox.getTaskInfo(newTaskHash);
        assertEq(task.executorCert, abi.encode(cert));
    }

    function test_Revert_WhenTimestampAtCreation() public {
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        // Don't advance time
        vm.prank(aggregator);
        vm.expectRevert(TimestampAtCreation.selector);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenTaskExpired() public {
        // Advance time past task SLA
        vm.warp(block.timestamp + taskSLA + 1);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        vm.prank(aggregator);
        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.CREATED, TaskStatus.EXPIRED));
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenTaskDoesNotExist() public {
        bytes32 nonExistentHash = keccak256("non-existent");

        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(nonExistentHash);

        vm.prank(aggregator);
        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.CREATED, TaskStatus.NONE));
        taskMailbox.submitResult(nonExistentHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenCertificateVerificationFailed_BN254() public {
        // Create a custom mock that returns false for certificate verification
        MockBN254CertificateVerifierFailure mockFailingVerifier = new MockBN254CertificateVerifierFailure();

        // Deploy a new TaskMailbox with the failing verifier using proxy pattern
        ProxyAdmin proxyAdmin = new ProxyAdmin();
        TaskMailbox taskMailboxImpl =
            new TaskMailbox(address(mockFailingVerifier), address(mockECDSACertificateVerifier), "1.0.0");
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskMailboxImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskMailbox.initialize.selector, address(this), 0, address(1))
        );
        TaskMailbox failingTaskMailbox = TaskMailbox(address(proxy));

        // Give creator tokens and approve the new TaskMailbox
        vm.prank(creator);
        mockToken.approve(address(failingTaskMailbox), type(uint256).max);

        // Set config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        failingTaskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create new task with this config
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = failingTaskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(newTaskHash);

        vm.prank(aggregator);
        vm.expectRevert(CertificateVerificationFailed.selector);
        failingTaskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenCertificateVerificationFailed_ECDSA() public {
        // Create a custom mock that returns false for certificate verification
        MockECDSACertificateVerifierFailure mockECDSACertificateVerifierFailure =
            new MockECDSACertificateVerifierFailure();

        // Deploy a new TaskMailbox with the failing ECDSA verifier using proxy pattern
        ProxyAdmin proxyAdmin = new ProxyAdmin();
        TaskMailbox taskMailboxImpl = new TaskMailbox(
            address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifierFailure), "1.0.0"
        );
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(taskMailboxImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(TaskMailbox.initialize.selector, address(this), 0, address(1))
        );
        TaskMailbox failingTaskMailbox = TaskMailbox(address(proxy));

        // Give creator tokens and approve the new TaskMailbox
        vm.prank(creator);
        mockToken.approve(address(failingTaskMailbox), type(uint256).max);

        // Setup executor operator set with ECDSA curve type
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.curveType = IKeyRegistrarTypes.CurveType.ECDSA;

        vm.prank(avs);
        failingTaskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = failingTaskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Create ECDSA certificate
        IECDSACertificateVerifierTypes.ECDSACertificate memory cert = _createValidECDSACertificate(newTaskHash);

        // Submit should fail
        vm.prank(aggregator);
        vm.expectRevert(CertificateVerificationFailed.selector);
        failingTaskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenInvalidCertificateEncoding() public {
        // Setup executor operator set with ECDSA curve type
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.curveType = IKeyRegistrarTypes.CurveType.ECDSA;

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Create invalid encoded certificate (not properly encoded ECDSA certificate)
        bytes memory invalidCert = abi.encode("invalid", "certificate", "data");
        bytes memory result = bytes("test result");

        // Submit should fail due to decoding error
        vm.prank(aggregator);
        vm.expectRevert(); // Will revert during abi.decode
        taskMailbox.submitResult(newTaskHash, invalidCert, result);
    }

    function test_Revert_AlreadyVerified() public {
        // First submit a valid result
        vm.warp(block.timestamp + 1);
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        vm.prank(aggregator);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));

        // Try to submit again
        vm.prank(aggregator);
        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.CREATED, TaskStatus.VERIFIED));
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("new result"));
    }

    function test_Revert_ReentrancyOnSubmitResult() public {
        // Deploy reentrant attacker as task hook
        ReentrantAttacker attacker = new ReentrantAttacker(address(taskMailbox));

        // Set up executor operator set with attacker as hook
        OperatorSet memory operatorSet = OperatorSet(avs2, executorOperatorSetId2);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskHook = IAVSTaskHook(address(attacker));

        vm.prank(avs2);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        TaskParams memory taskParams = TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        vm.prank(creator);
        bytes32 attackTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Set up attack parameters
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(attackTaskHash);

        attacker.setAttackParams(
            taskParams,
            attackTaskHash,
            cert,
            bytes("result"),
            false, // attack on handleTaskResultSubmission
            false // attack submitResult
        );

        // Try to submit result - should revert on reentrancy
        vm.prank(aggregator);
        vm.expectRevert("ReentrancyGuard: reentrant call");
        taskMailbox.submitResult(attackTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_ReentrancyOnSubmitResult_TryingToCreateTask() public {
        // Deploy reentrant attacker as task hook
        ReentrantAttacker attacker = new ReentrantAttacker(address(taskMailbox));

        // Give attacker tokens and approve
        mockToken.mint(address(attacker), 1000 ether);
        vm.prank(address(attacker));
        mockToken.approve(address(taskMailbox), type(uint256).max);

        // Set up executor operator set with attacker as hook
        OperatorSet memory operatorSet = OperatorSet(avs2, executorOperatorSetId2);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskHook = IAVSTaskHook(address(attacker));

        vm.prank(avs2);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        TaskParams memory taskParams = TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        vm.prank(creator);
        bytes32 attackTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Set up attack parameters
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(attackTaskHash);

        attacker.setAttackParams(
            taskParams,
            attackTaskHash,
            cert,
            bytes("result"),
            false, // attack on handleTaskResultSubmission
            true // attack createTask
        );

        // Try to submit result - should revert on reentrancy
        vm.prank(aggregator);
        vm.expectRevert("ReentrancyGuard: reentrant call");
        taskMailbox.submitResult(attackTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_submitResult_WithZeroStakeThreshold() public {
        // Setup executor operator set with zero stake threshold
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(0)) // Zero threshold
        });

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Submit result with zero threshold should still work
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(newTaskHash);

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("test result"));

        // Verify task was verified
        TaskStatus status = taskMailbox.getTaskStatus(newTaskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));
    }

    function test_submitResult_WithMaxStakeThreshold() public {
        // Setup executor operator set with max stake threshold
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(10_000)) // 100% threshold
        });

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Submit result with max threshold
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(newTaskHash);

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("test result"));

        // Verify task was verified
        TaskStatus status = taskMailbox.getTaskStatus(newTaskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));
    }

    function test_Revert_WhenBN254CertificateHasEmptySignature() public {
        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        // Create BN254 certificate with empty signature (X=0, Y=0)
        IBN254CertificateVerifierTypes.BN254Certificate memory cert = IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: taskHash,
            signature: BN254.G1Point(0, 0), // Empty signature
            apk: BN254.G2Point([uint256(1), uint256(2)], [uint256(3), uint256(4)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });

        // Submit result should fail with EmptyCertificateSignature error
        vm.prank(aggregator);
        vm.expectRevert(EmptyCertificateSignature.selector);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenBN254CertificateHasEmptySignature_OnlyXZero() public {
        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        // Create BN254 certificate with partially empty signature (X=0, Y=1)
        IBN254CertificateVerifierTypes.BN254Certificate memory cert = IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: taskHash,
            signature: BN254.G1Point(0, 1), // Partially empty signature
            apk: BN254.G2Point([uint256(1), uint256(2)], [uint256(3), uint256(4)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });

        // This should now fail since any coordinate being zero is invalid
        vm.prank(aggregator);
        vm.expectRevert(EmptyCertificateSignature.selector);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenBN254CertificateHasEmptySignature_OnlyYZero() public {
        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        // Create BN254 certificate with partially empty signature (X=1, Y=0)
        IBN254CertificateVerifierTypes.BN254Certificate memory cert = IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: taskHash,
            signature: BN254.G1Point(1, 0), // Partially empty signature
            apk: BN254.G2Point([uint256(1), uint256(2)], [uint256(3), uint256(4)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });

        // This should also fail since any coordinate being zero is invalid
        vm.prank(aggregator);
        vm.expectRevert(EmptyCertificateSignature.selector);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenECDSACertificateHasEmptySignature() public {
        // Setup executor operator set with ECDSA curve type
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.curveType = IKeyRegistrarTypes.CurveType.ECDSA;

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        // Create ECDSA certificate with empty signature
        IECDSACertificateVerifierTypes.ECDSACertificate memory cert = IECDSACertificateVerifierTypes.ECDSACertificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: newTaskHash,
            sig: bytes("") // Empty signature
        });

        // Submit result should fail with EmptyCertificateSignature error
        vm.prank(aggregator);
        vm.expectRevert(EmptyCertificateSignature.selector);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenBN254CertificateHasEmptySignature_WithZeroThreshold() public {
        // Setup executor operator set with zero stake threshold
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(0)) // Zero threshold
        });

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Create BN254 certificate with empty signature
        IBN254CertificateVerifierTypes.BN254Certificate memory cert = IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: newTaskHash,
            signature: BN254.G1Point(0, 0), // Empty signature
            apk: BN254.G2Point([uint256(1), uint256(2)], [uint256(3), uint256(4)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });

        // Even with zero threshold, empty signatures should be rejected
        vm.prank(aggregator);
        vm.expectRevert(EmptyCertificateSignature.selector);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_Revert_WhenECDSACertificateHasEmptySignature_WithZeroThreshold() public {
        // Setup executor operator set with ECDSA curve type and zero threshold
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.curveType = IKeyRegistrarTypes.CurveType.ECDSA;
        config.consensus = Consensus({
            consensusType: ConsensusType.STAKE_PROPORTION_THRESHOLD,
            value: abi.encode(uint16(0)) // Zero threshold
        });

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        // Create ECDSA certificate with empty signature
        IECDSACertificateVerifierTypes.ECDSACertificate memory cert = IECDSACertificateVerifierTypes.ECDSACertificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: newTaskHash,
            sig: bytes("") // Empty signature
        });

        // Even with zero threshold, empty signatures should be rejected
        vm.prank(aggregator);
        vm.expectRevert(EmptyCertificateSignature.selector);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("result"));
    }

    function test_submitResult_FeeTransferToCollector() public {
        // Set up operator set with fee token
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = mockToken;
        config.feeCollector = feeCollector;

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create task with fee
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 mailboxBalanceBefore = mockToken.balanceOf(address(taskMailbox));
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        // mailboxBalanceBefore should be 2*avsFee (one from setUp, one from this test)

        // Advance time and submit result
        vm.warp(block.timestamp + 1);

        IBN254CertificateVerifierTypes.BN254Certificate memory cert = IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: newTaskHash,
            signature: BN254.G1Point(1, 2),
            apk: BN254.G2Point([uint256(3), uint256(4)], [uint256(5), uint256(6)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });

        // Mock certificate verification
        vm.mockCall(
            address(mockBN254CertificateVerifier),
            abi.encodeWithSelector(IBN254CertificateVerifier.verifyCertificateProportion.selector),
            abi.encode(true)
        );

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), bytes("result"));

        // Verify fee was transferred to fee collector
        assertEq(mockToken.balanceOf(address(taskMailbox)), mailboxBalanceBefore - avsFee);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + avsFee);

        // Verify task cannot be refunded after verification
        Task memory task = taskMailbox.getTaskInfo(newTaskHash);
        assertEq(uint8(task.status), uint8(TaskStatus.VERIFIED));
        assertFalse(task.isFeeRefunded);
    }

    function test_FeeSplit_10Percent() public {
        // Setup fee split
        uint16 feeSplit = 1000; // 10%
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify fee distribution
        uint256 expectedFeeSplitAmount = (avsFee * feeSplit) / 10_000;
        uint256 expectedFeeCollectorAmount = avsFee - expectedFeeSplitAmount;

        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore + expectedFeeSplitAmount);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + expectedFeeCollectorAmount);
    }

    function test_FeeSplit_50Percent() public {
        // Setup fee split
        uint16 feeSplit = 5000; // 50%
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify fee distribution - should be equal split
        uint256 expectedFeeSplitAmount = avsFee / 2;
        uint256 expectedFeeCollectorAmount = avsFee - expectedFeeSplitAmount;

        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore + expectedFeeSplitAmount);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + expectedFeeCollectorAmount);
    }

    function test_FeeSplit_0Percent() public {
        // Setup fee split - 0% means all fees go to fee collector
        uint16 feeSplit = 0;
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify all fees went to fee collector
        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore); // No change
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + avsFee);
    }

    function test_FeeSplit_100Percent() public {
        // Setup fee split - 100% means all fees go to fee split collector
        uint16 feeSplit = 10_000;
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify all fees went to fee split collector
        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore + avsFee);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore); // No change
    }

    function test_FeeSplit_ZeroFeeAmount() public {
        // Setup fee split
        uint16 feeSplit = 5000; // 50%
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Setup operator set with zero fee
        mockTaskHook.setDefaultFee(0);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify no transfers occurred
        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore);
    }

    function test_FeeSplit_WithSmallFee() public {
        // Setup fee split
        uint16 feeSplit = 3333; // 33.33%
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Setup small fee
        uint96 smallFee = 100; // 100 wei
        mockTaskHook.setDefaultFee(smallFee);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify fee distribution with rounding
        uint256 expectedFeeSplitAmount = (smallFee * feeSplit) / 10_000; // 33 wei
        uint256 expectedFeeCollectorAmount = smallFee - expectedFeeSplitAmount; // 67 wei

        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore + expectedFeeSplitAmount);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + expectedFeeCollectorAmount);
        assertEq(expectedFeeSplitAmount + expectedFeeCollectorAmount, smallFee); // Verify no wei lost
    }

    function test_FeeSplit_Rounding() public {
        // Setup fee split that will cause rounding
        uint16 feeSplit = 1; // 0.01%
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Setup fee that won't divide evenly
        uint96 oddFee = 10_001; // Will result in 1.0001 wei split
        mockTaskHook.setDefaultFee(oddFee);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify rounding down occurs for fee split
        uint256 expectedFeeSplitAmount = (oddFee * feeSplit) / 10_000; // 1 wei (rounded down from 1.0001)
        uint256 expectedFeeCollectorAmount = oddFee - expectedFeeSplitAmount; // 10000 wei

        assertEq(mockToken.balanceOf(feeSplitCollector), expectedFeeSplitAmount);
        assertEq(mockToken.balanceOf(feeCollector), expectedFeeCollectorAmount);
        assertEq(expectedFeeSplitAmount + expectedFeeCollectorAmount, oddFee);
    }

    function testFuzz_FeeSplit(uint16 _feeSplit, uint96 _avsFee) public {
        // Bound inputs
        _feeSplit = uint16(bound(_feeSplit, 0, 10_000));
        vm.assume(_avsFee > 0 && _avsFee <= 1000 ether);
        vm.assume(_avsFee <= mockToken.balanceOf(creator));

        // Setup fee split
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(_feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Setup fee
        mockTaskHook.setDefaultFee(_avsFee);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balances
        uint256 mailboxBalanceBefore = mockToken.balanceOf(address(taskMailbox));
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 feeSplitCollectorBalanceBefore = mockToken.balanceOf(feeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Calculate expected amounts
        uint256 expectedFeeSplitAmount = (uint256(_avsFee) * _feeSplit) / 10_000;
        uint256 expectedAvsAmount = _avsFee - expectedFeeSplitAmount;

        // Verify balances
        assertEq(mockToken.balanceOf(feeSplitCollector), feeSplitCollectorBalanceBefore + expectedFeeSplitAmount);
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + expectedAvsAmount);

        // Verify total distribution equals original fee
        assertEq(expectedFeeSplitAmount + expectedAvsAmount, _avsFee);
    }

    function test_FeeSplit_TaskUsesSnapshotFeeSplitAndCurrentCollector() public {
        // Setup initial fee split
        uint16 initialFeeSplit = 2000; // 20%
        address initialFeeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(initialFeeSplit);
        taskMailbox.setFeeSplitCollector(initialFeeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Change fee split after task creation
        uint16 newFeeSplit = 5000; // 50%
        address newFeeSplitCollector = address(0xABC);
        taskMailbox.setFeeSplit(newFeeSplit);
        taskMailbox.setFeeSplitCollector(newFeeSplitCollector);

        // Check initial balances
        uint256 feeCollectorBalanceBefore = mockToken.balanceOf(feeCollector);
        uint256 initialCollectorBalanceBefore = mockToken.balanceOf(initialFeeSplitCollector);
        uint256 newCollectorBalanceBefore = mockToken.balanceOf(newFeeSplitCollector);

        // Submit result
        vm.warp(block.timestamp + 1);
        bytes memory executorCert = abi.encode(_createValidBN254Certificate(newTaskHash));

        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, executorCert, bytes("result"));

        // Verify fee distribution uses snapshot feeSplit (20%) but current collector (newFeeSplitCollector)
        uint256 expectedFeeSplitAmount = (avsFee * initialFeeSplit) / 10_000;
        uint256 expectedFeeCollectorAmount = avsFee - expectedFeeSplitAmount;

        assertEq(mockToken.balanceOf(initialFeeSplitCollector), initialCollectorBalanceBefore); // No change
        assertEq(mockToken.balanceOf(feeCollector), feeCollectorBalanceBefore + expectedFeeCollectorAmount);
        assertEq(mockToken.balanceOf(newFeeSplitCollector), newCollectorBalanceBefore + expectedFeeSplitAmount); // Gets the fee split
    }
}

// Test contract for refundFee function
contract TaskMailboxUnitTests_refundFee is TaskMailboxUnitTests {
    bytes32 public taskHash;

    function setUp() public override {
        super.setUp();

        // Set up operator set and task config with fee token
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = mockToken;
        config.feeCollector = feeCollector;

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task with fee
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        taskHash = taskMailbox.createTask(taskParams);
    }

    function test_refundFee_Success() public {
        // Move time forward to expire the task
        vm.warp(block.timestamp + taskSLA + 1);

        // Check initial balances
        uint256 mailboxBalanceBefore = mockToken.balanceOf(address(taskMailbox));
        uint256 refundCollectorBalanceBefore = mockToken.balanceOf(refundCollector);

        // Refund fee as refund collector
        vm.expectEmit(true, true, false, true);
        emit FeeRefunded(refundCollector, taskHash, avsFee);

        vm.prank(refundCollector);
        taskMailbox.refundFee(taskHash);

        // Verify balances changed correctly
        assertEq(mockToken.balanceOf(address(taskMailbox)), mailboxBalanceBefore - avsFee);
        assertEq(mockToken.balanceOf(refundCollector), refundCollectorBalanceBefore + avsFee);

        // Verify task state
        Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertTrue(task.isFeeRefunded);
        assertEq(uint8(task.status), uint8(TaskStatus.EXPIRED));
    }

    function test_Revert_refundFee_OnlyRefundCollector() public {
        // Move time forward to expire the task
        vm.warp(block.timestamp + taskSLA + 1);

        // Try to refund as someone else (not refund collector)
        vm.prank(creator);
        vm.expectRevert(OnlyRefundCollector.selector);
        taskMailbox.refundFee(taskHash);

        // Try as a random address
        vm.prank(address(0x1234));
        vm.expectRevert(OnlyRefundCollector.selector);
        taskMailbox.refundFee(taskHash);
    }

    function test_Revert_refundFee_FeeAlreadyRefunded() public {
        // Move time forward to expire the task
        vm.warp(block.timestamp + taskSLA + 1);

        // First refund should succeed
        vm.prank(refundCollector);
        taskMailbox.refundFee(taskHash);

        // Second refund should fail
        vm.prank(refundCollector);
        vm.expectRevert(FeeAlreadyRefunded.selector);
        taskMailbox.refundFee(taskHash);
    }

    function test_Revert_refundFee_TaskNotExpired() public {
        // Try to refund before task expires
        vm.prank(refundCollector);
        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.EXPIRED, TaskStatus.CREATED));
        taskMailbox.refundFee(taskHash);
    }

    function test_Revert_refundFee_TaskAlreadyVerified() public {
        // Submit result to verify the task
        IBN254CertificateVerifierTypes.BN254Certificate memory cert = IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: taskHash,
            signature: BN254.G1Point(1, 2),
            apk: BN254.G2Point([uint256(3), uint256(4)], [uint256(5), uint256(6)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });

        // Mock certificate verification
        vm.mockCall(
            address(mockBN254CertificateVerifier),
            abi.encodeWithSelector(IBN254CertificateVerifier.verifyCertificateProportion.selector),
            abi.encode(true)
        );

        vm.prank(aggregator);
        vm.warp(block.timestamp + 1);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));

        // Move time forward to what would be expiry
        vm.warp(block.timestamp + taskSLA + 1);

        // Try to refund - should fail because task is verified
        vm.prank(refundCollector);
        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.EXPIRED, TaskStatus.VERIFIED));
        taskMailbox.refundFee(taskHash);
    }

    function test_refundFee_NoFeeToken() public {
        // Create a task without fee token
        OperatorSet memory operatorSet = OperatorSet(avs2, executorOperatorSetId2);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.feeToken = IERC20(address(0));

        vm.prank(avs2);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        TaskParams memory taskParams = TaskParams({
            refundCollector: refundCollector,
            executorOperatorSet: operatorSet,
            payload: bytes("test payload")
        });

        vm.prank(creator);
        bytes32 noFeeTaskHash = taskMailbox.createTask(taskParams);

        // Move time forward to expire the task
        vm.warp(block.timestamp + taskSLA + 1);

        // Refund should succeed but no transfer should occur
        uint256 mailboxBalanceBefore = mockToken.balanceOf(address(taskMailbox));
        uint256 refundCollectorBalanceBefore = mockToken.balanceOf(refundCollector);

        vm.prank(refundCollector);
        taskMailbox.refundFee(noFeeTaskHash);

        // Balances should not change since there's no fee token
        assertEq(mockToken.balanceOf(address(taskMailbox)), mailboxBalanceBefore);
        assertEq(mockToken.balanceOf(refundCollector), refundCollectorBalanceBefore);

        // Task should still be marked as refunded
        Task memory task = taskMailbox.getTaskInfo(noFeeTaskHash);
        assertTrue(task.isFeeRefunded);
    }

    function test_refundFee_WithFeeSplit() public {
        // Setup fee split
        uint16 feeSplit = 3000; // 30%
        address feeSplitCollector = address(0x789);
        taskMailbox.setFeeSplit(feeSplit);
        taskMailbox.setFeeSplitCollector(feeSplitCollector);

        // Create task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Check initial balance
        uint256 refundCollectorBalanceBefore = mockToken.balanceOf(refundCollector);

        // Move time forward to expire the task
        vm.warp(block.timestamp + taskSLA + 1);

        // Refund fee
        vm.prank(refundCollector);
        taskMailbox.refundFee(newTaskHash);

        // Verify full fee was refunded (fee split doesn't apply to refunds)
        assertEq(mockToken.balanceOf(refundCollector), refundCollectorBalanceBefore + avsFee);

        // Verify fee split collector got nothing
        assertEq(mockToken.balanceOf(feeSplitCollector), 0);
    }

    function test_refundFee_ZeroFee() public {
        // Set mock to return 0 fee
        mockTaskHook.setDefaultFee(0);

        // Create a task with 0 fee
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 zeroFeeTaskHash = taskMailbox.createTask(taskParams);

        // Move time forward to expire the task
        vm.warp(block.timestamp + taskSLA + 1);

        // Refund should succeed but no transfer should occur
        uint256 mailboxBalanceBefore = mockToken.balanceOf(address(taskMailbox));
        uint256 refundCollectorBalanceBefore = mockToken.balanceOf(refundCollector);

        vm.prank(refundCollector);
        taskMailbox.refundFee(zeroFeeTaskHash);

        // Balances should not change since fee is 0
        assertEq(mockToken.balanceOf(address(taskMailbox)), mailboxBalanceBefore);
        assertEq(mockToken.balanceOf(refundCollector), refundCollectorBalanceBefore);

        // Task should still be marked as refunded
        Task memory task = taskMailbox.getTaskInfo(zeroFeeTaskHash);
        assertTrue(task.isFeeRefunded);
    }
}

// Test contract for view functions
contract TaskMailboxUnitTests_ViewFunctions is TaskMailboxUnitTests {
    bytes32 public taskHash;
    OperatorSet public operatorSet;

    function setUp() public override {
        super.setUp();

        // Set up executor operator set task config
        operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        taskHash = taskMailbox.createTask(taskParams);
    }

    function test_ViewFunctions() public {
        // Test that we can read the immutable certificate verifiers
        assertEq(taskMailbox.BN254_CERTIFICATE_VERIFIER(), address(mockBN254CertificateVerifier));
        assertEq(taskMailbox.ECDSA_CERTIFICATE_VERIFIER(), address(mockECDSACertificateVerifier));
        assertEq(taskMailbox.version(), "1.0.0");
        assertEq(taskMailbox.owner(), address(this));

        // Test fee split getters
        assertEq(taskMailbox.getFeeSplit(), 0); // Default value from initialization
        assertEq(taskMailbox.getFeeSplitCollector(), address(1)); // Default value from initialization
    }

    function test_getExecutorOperatorSetTaskConfig() public {
        ExecutorOperatorSetTaskConfig memory config = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);

        assertEq(uint8(config.curveType), uint8(IKeyRegistrarTypes.CurveType.BN254));
        assertEq(address(config.taskHook), address(mockTaskHook));
        assertEq(address(config.feeToken), address(mockToken));
        assertEq(config.feeCollector, feeCollector);
        assertEq(config.taskSLA, taskSLA);
        assertEq(uint8(config.consensus.consensusType), uint8(ConsensusType.STAKE_PROPORTION_THRESHOLD));
        uint16 decodedThreshold = abi.decode(config.consensus.value, (uint16));
        assertEq(decodedThreshold, stakeProportionThreshold);
        assertEq(config.taskMetadata, bytes("test metadata"));
    }

    function test_getExecutorOperatorSetTaskConfig_Unregistered() public {
        OperatorSet memory unregisteredSet = OperatorSet(avs2, 99);
        ExecutorOperatorSetTaskConfig memory config = taskMailbox.getExecutorOperatorSetTaskConfig(unregisteredSet);

        // Should return empty config
        assertEq(uint8(config.curveType), uint8(IKeyRegistrarTypes.CurveType.NONE));
        assertEq(address(config.taskHook), address(0));
        assertEq(address(config.feeToken), address(0));
        assertEq(config.feeCollector, address(0));
        assertEq(config.taskSLA, 0);
        assertEq(config.consensus.value.length, 0);
        assertEq(config.taskMetadata, bytes(""));
    }

    function test_getTaskInfo() public {
        Task memory task = taskMailbox.getTaskInfo(taskHash);

        assertEq(task.creator, creator);
        assertEq(task.creationTime, block.timestamp);
        assertEq(uint8(task.status), uint8(TaskStatus.CREATED));
        assertEq(task.avs, avs);
        assertEq(task.executorOperatorSetId, executorOperatorSetId);
        assertEq(task.refundCollector, refundCollector);
        assertEq(task.avsFee, avsFee);
        assertEq(task.feeSplit, 0);
        assertEq(task.payload, bytes("test payload"));
        assertEq(task.executorCert, bytes(""));
        assertEq(task.result, bytes(""));
    }

    function test_getTaskInfo_NonExistentTask() public {
        bytes32 nonExistentHash = keccak256("non-existent");
        Task memory task = taskMailbox.getTaskInfo(nonExistentHash);

        // Should return empty task with NONE status (default for non-existent tasks)
        assertEq(task.creator, address(0));
        assertEq(task.creationTime, 0);
        assertEq(uint8(task.status), uint8(TaskStatus.NONE)); // Non-existent tasks show as NONE
        assertEq(task.avs, address(0));
        assertEq(task.executorOperatorSetId, 0);
    }

    function test_getTaskStatus_Created() public {
        TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(TaskStatus.CREATED));
    }

    function test_getTaskStatus_Verified() public {
        vm.warp(block.timestamp + 1);
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        vm.prank(aggregator);
        taskMailbox.submitResult(taskHash, abi.encode(cert), bytes("result"));

        TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));
    }

    function test_getTaskStatus_Expired() public {
        // Advance time past SLA
        vm.warp(block.timestamp + taskSLA + 1);

        TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(TaskStatus.EXPIRED));
    }

    function test_getTaskStatus_None() public {
        // Get status of non-existent task
        bytes32 nonExistentHash = keccak256("non-existent");
        TaskStatus status = taskMailbox.getTaskStatus(nonExistentHash);
        assertEq(uint8(status), uint8(TaskStatus.NONE));
    }

    function test_getTaskInfo_Expired() public {
        // Advance time past SLA
        vm.warp(block.timestamp + taskSLA + 1);

        Task memory task = taskMailbox.getTaskInfo(taskHash);

        // getTaskInfo should return Expired status
        assertEq(uint8(task.status), uint8(TaskStatus.EXPIRED));
    }

    function test_getTaskResult() public {
        // Submit result first
        vm.warp(block.timestamp + 1);
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);
        bytes memory expectedResult = bytes("test result");

        vm.prank(aggregator);
        taskMailbox.submitResult(taskHash, abi.encode(cert), expectedResult);

        // Get result
        bytes memory result = taskMailbox.getTaskResult(taskHash);
        assertEq(result, expectedResult);
    }

    function test_Revert_getTaskResult_NotVerified() public {
        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.VERIFIED, TaskStatus.CREATED));
        taskMailbox.getTaskResult(taskHash);
    }

    function test_Revert_getTaskResult_Expired() public {
        vm.warp(block.timestamp + taskSLA + 1);

        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.VERIFIED, TaskStatus.EXPIRED));
        taskMailbox.getTaskResult(taskHash);
    }

    function test_Revert_getTaskResult_None() public {
        bytes32 nonExistentHash = keccak256("non-existent");

        vm.expectRevert(abi.encodeWithSelector(InvalidTaskStatus.selector, TaskStatus.VERIFIED, TaskStatus.NONE));
        taskMailbox.getTaskResult(nonExistentHash);
    }
}

// Test contract for storage variables
contract TaskMailboxUnitTests_Storage is TaskMailboxUnitTests {
    function test_globalTaskCount() public {
        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create multiple tasks and verify the count through task hashes
        TaskParams memory taskParams = _createValidTaskParams();

        // First task should have count 0
        bytes32 expectedHash0 = keccak256(abi.encode(0, address(taskMailbox), block.chainid, taskParams));
        vm.prank(creator);
        bytes32 taskHash0 = taskMailbox.createTask(taskParams);
        assertEq(taskHash0, expectedHash0);

        // Second task should have count 1
        bytes32 expectedHash1 = keccak256(abi.encode(1, address(taskMailbox), block.chainid, taskParams));
        vm.prank(creator);
        bytes32 taskHash1 = taskMailbox.createTask(taskParams);
        assertEq(taskHash1, expectedHash1);

        // Third task should have count 2
        bytes32 expectedHash2 = keccak256(abi.encode(2, address(taskMailbox), block.chainid, taskParams));
        vm.prank(creator);
        bytes32 taskHash2 = taskMailbox.createTask(taskParams);
        assertEq(taskHash2, expectedHash2);
    }

    function test_isExecutorOperatorSetRegistered() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // Initially not registered
        assertFalse(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));

        // Set config first (requirement for registerExecutorOperatorSet)
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // After setting config, it should be automatically registered
        assertTrue(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));

        // Unregister
        vm.prank(avs);
        taskMailbox.registerExecutorOperatorSet(operatorSet, false);
        assertFalse(taskMailbox.isExecutorOperatorSetRegistered(operatorSet.key()));
    }

    function test_executorOperatorSetTaskConfigs() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        // Set config
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Access config via getExecutorOperatorSetTaskConfig function
        ExecutorOperatorSetTaskConfig memory storedConfig = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);

        assertEq(uint8(storedConfig.curveType), uint8(config.curveType));
        assertEq(address(storedConfig.taskHook), address(config.taskHook));
        assertEq(address(storedConfig.feeToken), address(config.feeToken));
        assertEq(storedConfig.feeCollector, config.feeCollector);
        assertEq(storedConfig.taskSLA, config.taskSLA);
        assertEq(uint8(storedConfig.consensus.consensusType), uint8(config.consensus.consensusType));
        assertEq(storedConfig.consensus.value, config.consensus.value);
        assertEq(storedConfig.taskMetadata, config.taskMetadata);
    }

    function test_tasks() public {
        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Access task via getTaskInfo public function
        Task memory task = taskMailbox.getTaskInfo(taskHash);

        assertEq(task.creator, creator);
        assertEq(task.creationTime, block.timestamp);
        assertEq(uint8(task.status), uint8(TaskStatus.CREATED));
        assertEq(task.avs, avs);
        assertEq(task.executorOperatorSetId, executorOperatorSetId);
        assertEq(task.refundCollector, refundCollector);
        assertEq(task.avsFee, avsFee);
        assertEq(task.feeSplit, 0);
        assertEq(task.payload, bytes("test payload"));
        assertEq(task.executorCert, bytes(""));
        assertEq(task.result, bytes(""));
    }
}

// Test contract for upgradeable functionality
contract TaskMailboxUnitTests_Upgradeable is TaskMailboxUnitTests {
    function test_Initialize_OnlyOnce() public {
        // Try to initialize again, should revert
        vm.expectRevert("Initializable: contract is already initialized");
        taskMailbox.initialize(address(0x9999), 0, address(1));
    }

    function test_Implementation_CannotBeInitialized() public {
        // Deploy a new implementation
        TaskMailbox newImpl =
            new TaskMailbox(address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier), "1.0.1");

        // Try to initialize the implementation directly, should revert
        vm.expectRevert("Initializable: contract is already initialized");
        newImpl.initialize(address(this), 0, address(1));
    }

    function test_ProxyUpgrade() public {
        address newOwner = address(0x1234);

        // Deploy new implementation with different version
        TaskMailbox newImpl =
            new TaskMailbox(address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier), "2.0.0");

        // Check version before upgrade
        assertEq(taskMailbox.version(), "1.0.0");

        // Upgrade proxy to new implementation
        proxyAdmin.upgrade(ITransparentUpgradeableProxy(address(taskMailbox)), address(newImpl));

        // Check version after upgrade
        assertEq(taskMailbox.version(), "2.0.0");

        // Verify state is preserved (owner should still be the same)
        assertEq(taskMailbox.owner(), address(this));
    }

    function test_ProxyAdmin_OnlyOwnerCanUpgrade() public {
        address attacker = address(0x9999);

        // Deploy new implementation
        TaskMailbox newImpl =
            new TaskMailbox(address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier), "2.0.0");

        // Try to upgrade from non-owner, should revert
        vm.prank(attacker);
        vm.expectRevert("Ownable: caller is not the owner");
        proxyAdmin.upgrade(ITransparentUpgradeableProxy(address(taskMailbox)), address(newImpl));
    }

    function test_ProxyAdmin_CannotCallImplementation() public {
        // ProxyAdmin should not be able to call implementation functions
        vm.prank(address(proxyAdmin));
        vm.expectRevert("TransparentUpgradeableProxy: admin cannot fallback to proxy target");
        TaskMailbox(payable(address(taskMailbox))).owner();
    }

    function test_StorageSlotConsistency_AfterUpgrade() public {
        address newOwner = address(0x1234);

        // First, make some state changes
        taskMailbox.transferOwnership(newOwner);
        assertEq(taskMailbox.owner(), newOwner);

        // Set up an executor operator set
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Verify config is set
        ExecutorOperatorSetTaskConfig memory retrievedConfig = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);
        assertEq(address(retrievedConfig.taskHook), address(config.taskHook));

        // Deploy new implementation
        TaskMailbox newImpl =
            new TaskMailbox(address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier), "2.0.0");

        // Upgrade
        vm.prank(address(this)); // proxyAdmin owner
        proxyAdmin.upgrade(ITransparentUpgradeableProxy(address(taskMailbox)), address(newImpl));

        // Verify all state is preserved after upgrade
        assertEq(taskMailbox.owner(), newOwner);
        assertEq(taskMailbox.version(), "2.0.0");

        // Verify the executor operator set config is still there
        ExecutorOperatorSetTaskConfig memory configAfterUpgrade =
            taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);
        assertEq(address(configAfterUpgrade.taskHook), address(config.taskHook));
        assertEq(configAfterUpgrade.taskSLA, config.taskSLA);
        assertEq(uint8(configAfterUpgrade.consensus.consensusType), uint8(config.consensus.consensusType));
        assertEq(configAfterUpgrade.consensus.value, config.consensus.value);
    }

    function test_InitializerModifier_PreventsReinitialization() public {
        // Deploy a new proxy without initialization data
        TransparentUpgradeableProxy uninitializedProxy = new TransparentUpgradeableProxy(
            address(
                new TaskMailbox(address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier), "1.0.0")
            ),
            address(new ProxyAdmin()),
            ""
        );

        TaskMailbox uninitializedTaskMailbox = TaskMailbox(address(uninitializedProxy));

        // Initialize it once
        uninitializedTaskMailbox.initialize(address(this), 0, address(1));
        assertEq(uninitializedTaskMailbox.owner(), address(this));

        // Try to initialize again, should fail
        vm.expectRevert("Initializable: contract is already initialized");
        uninitializedTaskMailbox.initialize(address(0x9999), 0, address(1));
    }
}
