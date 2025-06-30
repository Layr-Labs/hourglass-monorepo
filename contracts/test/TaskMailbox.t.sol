// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Test, console, Vm} from "forge-std/Test.sol";
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

        // Deploy TaskMailbox
        taskMailbox =
            new TaskMailbox(address(this), address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifier));

        // Give creator some tokens and approve TaskMailbox
        mockToken.mint(creator, 1000 ether);
        vm.prank(creator);
        mockToken.approve(address(taskMailbox), type(uint256).max);
    }

    function _createValidTaskParams() internal view returns (TaskParams memory) {
        return TaskParams({
            refundCollector: refundCollector,
            avsFee: avsFee,
            executorOperatorSet: OperatorSet(avs, executorOperatorSetId),
            payload: bytes("test payload")
        });
    }

    function _createValidExecutorOperatorSetTaskConfig() internal view returns (ExecutorOperatorSetTaskConfig memory) {
        return ExecutorOperatorSetTaskConfig({
            curveType: IKeyRegistrarTypes.CurveType.BN254,
            taskHook: IAVSTaskHook(address(mockTaskHook)),
            feeToken: IERC20(address(mockToken)),
            feeCollector: feeCollector,
            taskSLA: taskSLA,
            stakeProportionThreshold: stakeProportionThreshold,
            taskMetadata: bytes("test metadata")
        });
    }

    function _createValidBN254Certificate(
        bytes32 messageHash
    ) internal view returns (IBN254CertificateVerifierTypes.BN254Certificate memory) {
        return IBN254CertificateVerifierTypes.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: messageHash,
            signature: BN254.G1Point(0, 0),
            apk: BN254.G2Point([uint256(0), uint256(0)], [uint256(0), uint256(0)]),
            nonSignerWitnesses: new IBN254CertificateVerifierTypes.BN254OperatorInfoWitness[](0)
        });
    }

    function _createValidECDSACertificate(
        bytes32 messageHash
    ) internal view returns (IECDSACertificateVerifierTypes.ECDSACertificate memory) {
        return IECDSACertificateVerifierTypes.ECDSACertificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: messageHash,
            sig: bytes("")
        });
    }
}

contract TaskMailboxUnitTests_Constructor is TaskMailboxUnitTests {
    function test_Constructor_WithCertificateVerifiers() public {
        address bn254Verifier = address(0x1234);
        address ecdsaVerifier = address(0x5678);

        TaskMailbox newTaskMailbox = new TaskMailbox(address(this), bn254Verifier, ecdsaVerifier);

        assertEq(newTaskMailbox.BN254_CERTIFICATE_VERIFIER(), bn254Verifier);
        assertEq(newTaskMailbox.ECDSA_CERTIFICATE_VERIFIER(), ecdsaVerifier);
        assertEq(newTaskMailbox.owner(), address(this));
    }
}

// Test contract for registerExecutorOperatorSet
contract TaskMailboxUnitTests_registerExecutorOperatorSet is TaskMailboxUnitTests {
    function testFuzz_registerExecutorOperatorSet(
        address fuzzAvs,
        uint32 fuzzOperatorSetId,
        bool fuzzIsRegistered
    ) public {
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

        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);

        ExecutorOperatorSetTaskConfig memory config = ExecutorOperatorSetTaskConfig({
            curveType: IKeyRegistrarTypes.CurveType.BN254,
            taskHook: IAVSTaskHook(fuzzTaskHook),
            feeToken: IERC20(fuzzFeeToken),
            feeCollector: fuzzFeeCollector,
            taskSLA: fuzzTaskSLA,
            stakeProportionThreshold: fuzzStakeProportionThreshold,
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
        assertEq(retrievedConfig.stakeProportionThreshold, fuzzStakeProportionThreshold);
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
                "ExecutorOperatorSetTaskConfigSet(address,address,uint32,(uint8,address,address,address,uint96,uint16,bytes))"
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

        TaskParams memory taskParams = TaskParams({
            refundCollector: fuzzRefundCollector,
            avsFee: fuzzAvsFee,
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
        TaskParams memory taskParams = _createValidTaskParams();
        taskParams.avsFee = 0;

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
            avsFee: 1 ether,
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
            avsFee: avsFee,
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
            avsFee: avsFee,
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
        emit TaskVerified(aggregator, taskHash, avs, executorOperatorSetId, fuzzResult);

        // Submit result
        vm.prank(aggregator);
        taskMailbox.submitResult(taskHash, abi.encode(cert), fuzzResult);

        // Verify task was verified
        TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));

        // Verify result was stored
        bytes memory storedResult = taskMailbox.getTaskResult(taskHash);
        assertEq(storedResult, fuzzResult);
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
        emit TaskVerified(aggregator, newTaskHash, avs, executorOperatorSetId, fuzzResult);

        // Submit result with ECDSA certificate
        vm.prank(aggregator);
        taskMailbox.submitResult(newTaskHash, abi.encode(cert), fuzzResult);

        // Verify task was verified
        TaskStatus status = taskMailbox.getTaskStatus(newTaskHash);
        assertEq(uint8(status), uint8(TaskStatus.VERIFIED));

        // Verify result was stored
        bytes memory storedResult = taskMailbox.getTaskResult(newTaskHash);
        assertEq(storedResult, fuzzResult);
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

        // Deploy a new TaskMailbox with the failing verifier
        TaskMailbox failingTaskMailbox =
            new TaskMailbox(address(this), address(mockFailingVerifier), address(mockECDSACertificateVerifier));

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

        // Deploy a new TaskMailbox with the failing ECDSA verifier
        TaskMailbox failingTaskMailbox = new TaskMailbox(
            address(this), address(mockBN254CertificateVerifier), address(mockECDSACertificateVerifierFailure)
        );

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
            avsFee: avsFee,
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
            avsFee: avsFee,
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
    }

    function test_getExecutorOperatorSetTaskConfig() public {
        ExecutorOperatorSetTaskConfig memory config = taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);

        assertEq(uint8(config.curveType), uint8(IKeyRegistrarTypes.CurveType.BN254));
        assertEq(address(config.taskHook), address(mockTaskHook));
        assertEq(address(config.feeToken), address(mockToken));
        assertEq(config.feeCollector, feeCollector);
        assertEq(config.taskSLA, taskSLA);
        assertEq(config.stakeProportionThreshold, stakeProportionThreshold);
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
        assertEq(config.stakeProportionThreshold, 0);
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
        assertEq(storedConfig.stakeProportionThreshold, config.stakeProportionThreshold);
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
        assertEq(task.result, bytes(""));
    }
}
