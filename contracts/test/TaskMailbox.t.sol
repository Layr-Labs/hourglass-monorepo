// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Test, console} from "forge-std/Test.sol";
import {OperatorSet, OperatorSetLib} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {TaskMailbox} from "../src/core/TaskMailbox.sol";
import {ITaskMailbox, ITaskMailboxTypes, ITaskMailboxErrors} from "../src/interfaces/core/ITaskMailbox.sol";
import {IAVSTaskHook} from "../src/interfaces/avs/l2/IAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "../src/interfaces/avs/l2/IBN254CertificateVerifier.sol";
import {MockAVSTaskHook} from "./mocks/MockAVSTaskHook.sol";
import {MockBN254CertificateVerifier} from "./mocks/MockBN254CertificateVerifier.sol";
import {MockBN254CertificateVerifierFailure} from "./mocks/MockBN254CertificateVerifierFailure.sol";
import {MockERC20} from "./mocks/MockERC20.sol";

contract TaskMailboxUnitTests is Test {
    using OperatorSetLib for OperatorSet;

    // Contracts
    TaskMailbox public taskMailbox;
    MockAVSTaskHook public mockTaskHook;
    MockBN254CertificateVerifier public mockCertificateVerifier;
    MockERC20 public mockToken;

    // Test addresses
    address public avs = address(0x1);
    address public avs2 = address(0x2);
    address public feeCollector = address(0x3);
    address public refundCollector = address(0x4);
    address public creator = address(0x5);
    address public aggregator = address(0x6);

    // Test operator set IDs
    uint32 public aggregatorOperatorSetId = 1;
    uint32 public executorOperatorSetId = 2;
    uint32 public executorOperatorSetId2 = 3;

    // Test config values
    uint96 public taskSLA = 1 hours;
    uint16 public stakeProportionThreshold = 6667; // 66.67%
    uint96 public avsFee = 100 ether;

    // Events from ITaskMailbox
    event AvsRegistered(address indexed caller, address indexed avs, bool isRegistered);
    event AvsConfigSet(
        address indexed caller, address indexed avs, uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds
    );
    event ExecutorOperatorSetTaskConfigSet(
        address indexed caller,
        address indexed avs,
        uint32 indexed executorOperatorSetId,
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig config
    );
    event TaskCreated(
        address indexed creator,
        bytes32 indexed taskHash,
        address indexed avs,
        uint32 executorOperatorSetId,
        address refundCollector,
        uint96 avsFee,
        uint256 taskDeadline,
        bytes payload
    );
    event TaskCanceled(
        address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId
    );
    event TaskVerified(
        address indexed aggregator,
        bytes32 indexed taskHash,
        address indexed avs,
        uint32 executorOperatorSetId,
        bytes result
    );

    function setUp() public virtual {
        // Deploy mock contracts
        mockTaskHook = new MockAVSTaskHook();
        mockCertificateVerifier = new MockBN254CertificateVerifier();
        mockToken = new MockERC20();

        // Deploy TaskMailbox
        taskMailbox = new TaskMailbox();

        // Setup initial AVS registration and config
        _registerAndConfigureAvs();

        // Give creator some tokens and approve TaskMailbox
        mockToken.mint(creator, 1000 ether);
        vm.prank(creator);
        mockToken.approve(address(taskMailbox), type(uint256).max);
    }

    function _registerAndConfigureAvs() internal {
        // Register AVS
        vm.prank(avs);
        taskMailbox.registerAvs(avs, true);

        // Set AVS config
        uint32[] memory executorOperatorSetIds = new uint32[](2);
        executorOperatorSetIds[0] = executorOperatorSetId;
        executorOperatorSetIds[1] = executorOperatorSetId2;

        ITaskMailboxTypes.AvsConfig memory avsConfig = ITaskMailboxTypes.AvsConfig({
            aggregatorOperatorSetId: aggregatorOperatorSetId,
            executorOperatorSetIds: executorOperatorSetIds
        });

        vm.prank(avs);
        taskMailbox.setAvsConfig(avs, avsConfig);
    }

    function _createValidTaskParams() internal view returns (ITaskMailboxTypes.TaskParams memory) {
        return ITaskMailboxTypes.TaskParams({
            refundCollector: refundCollector,
            avsFee: avsFee,
            executorOperatorSet: OperatorSet(avs, executorOperatorSetId),
            payload: bytes("test payload")
        });
    }

    function _createValidExecutorOperatorSetTaskConfig()
        internal
        view
        returns (ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory)
    {
        return ITaskMailboxTypes.ExecutorOperatorSetTaskConfig({
            certificateVerifier: address(mockCertificateVerifier),
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
    ) internal view returns (IBN254CertificateVerifier.BN254Certificate memory) {
        return IBN254CertificateVerifier.BN254Certificate({
            referenceTimestamp: uint32(block.timestamp),
            messageHash: messageHash,
            sig: BN254.G1Point(0, 0),
            apk: BN254.G2Point([uint256(0), uint256(0)], [uint256(0), uint256(0)]),
            nonsignerIndices: new uint32[](0),
            nonSignerWitnesses: new IBN254CertificateVerifier.BN254OperatorInfoWitness[](0)
        });
    }
}

// Test contract for setExecutorOperatorSetTaskConfig
contract TaskMailboxUnitTests_setExecutorOperatorSetTaskConfig is TaskMailboxUnitTests {
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

        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = ITaskMailboxTypes.ExecutorOperatorSetTaskConfig({
            certificateVerifier: fuzzCertificateVerifier,
            taskHook: IAVSTaskHook(fuzzTaskHook),
            feeToken: IERC20(fuzzFeeToken),
            feeCollector: fuzzFeeCollector,
            taskSLA: fuzzTaskSLA,
            stakeProportionThreshold: fuzzStakeProportionThreshold,
            taskMetadata: fuzzTaskMetadata
        });

        // Expect event
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit ExecutorOperatorSetTaskConfigSet(avs, avs, executorOperatorSetId, config);

        // Set config
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Verify config was set
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory retrievedConfig =
            taskMailbox.getExecutorOperatorSetTaskConfig(operatorSet);

        assertEq(retrievedConfig.certificateVerifier, fuzzCertificateVerifier);
        assertEq(address(retrievedConfig.taskHook), fuzzTaskHook);
        assertEq(address(retrievedConfig.feeToken), fuzzFeeToken);
        assertEq(retrievedConfig.feeCollector, fuzzFeeCollector);
        assertEq(retrievedConfig.taskSLA, fuzzTaskSLA);
        assertEq(retrievedConfig.stakeProportionThreshold, fuzzStakeProportionThreshold);
        assertEq(retrievedConfig.taskMetadata, fuzzTaskMetadata);
    }

    function testFuzz_Revert_WhenExecutorOperatorSetNotRegistered(
        uint32 unregisteredOperatorSetId
    ) public {
        vm.assume(
            unregisteredOperatorSetId != executorOperatorSetId && unregisteredOperatorSetId != executorOperatorSetId2
        );

        OperatorSet memory unregisteredOperatorSet = OperatorSet(avs, unregisteredOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        vm.expectRevert(ITaskMailboxErrors.ExecutorOperatorSetNotRegistered.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(unregisteredOperatorSet, config);
    }

    function test_Revert_WhenCertificateVerifierIsZero() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.certificateVerifier = address(0);

        vm.prank(avs);
        vm.expectRevert(ITaskMailboxErrors.InvalidAddressZero.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenTaskHookIsZero() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskHook = IAVSTaskHook(address(0));

        vm.prank(avs);
        vm.expectRevert(ITaskMailboxErrors.InvalidAddressZero.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function test_Revert_WhenTaskSLAIsZero() public {
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.taskSLA = 0;

        vm.prank(avs);
        vm.expectRevert(ITaskMailboxErrors.TaskSLAIsZero.selector);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }
}

// Test contract for createTask
contract TaskMailboxUnitTests_createTask is TaskMailboxUnitTests {
    function setUp() public override {
        super.setUp();

        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);
    }

    function testFuzz_createTask(address fuzzRefundCollector, uint96 fuzzAvsFee, bytes memory fuzzPayload) public {
        // Bound inputs
        vm.assume(fuzzPayload.length > 0);
        vm.assume(fuzzAvsFee <= mockToken.balanceOf(creator));

        ITaskMailboxTypes.TaskParams memory taskParams = ITaskMailboxTypes.TaskParams({
            refundCollector: fuzzRefundCollector,
            avsFee: fuzzAvsFee,
            executorOperatorSet: OperatorSet(avs, executorOperatorSetId),
            payload: fuzzPayload
        });

        // Expect event (check all indexed parameters except taskHash)
        vm.expectEmit(true, false, true, true, address(taskMailbox));
        emit TaskCreated(
            creator,
            bytes32(0), // We don't know the exact hash beforehand
            avs,
            executorOperatorSetId,
            fuzzRefundCollector,
            fuzzAvsFee,
            block.timestamp + taskSLA,
            fuzzPayload
        );

        // Create task and capture the returned hash
        vm.prank(creator);
        bytes32 taskHash = taskMailbox.createTask(taskParams);

        // Verify task was created
        ITaskMailboxTypes.Task memory task = taskMailbox.getTaskInfo(taskHash);
        assertEq(task.creator, creator);
        assertEq(task.creationTime, block.timestamp);
        assertEq(uint8(task.status), uint8(ITaskMailboxTypes.TaskStatus.Created));
        assertEq(task.avs, avs);
        assertEq(task.executorOperatorSetId, executorOperatorSetId);
        assertEq(task.refundCollector, fuzzRefundCollector);
        assertEq(task.avsFee, fuzzAvsFee);
        assertEq(task.payload, fuzzPayload);

        // Verify token transfer if fee > 0
        if (fuzzAvsFee > 0) {
            assertEq(mockToken.balanceOf(address(taskMailbox)), fuzzAvsFee);
        }
    }

    function test_Revert_WhenAvsNotRegistered() public {
        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        taskParams.executorOperatorSet.avs = address(0x999); // Unregistered AVS

        vm.prank(creator);
        vm.expectRevert(ITaskMailboxErrors.AvsNotRegistered.selector);
        taskMailbox.createTask(taskParams);
    }

    function testFuzz_Revert_WhenExecutorOperatorSetNotRegistered(
        uint32 unregisteredOperatorSetId
    ) public {
        vm.assume(
            unregisteredOperatorSetId != executorOperatorSetId && unregisteredOperatorSetId != executorOperatorSetId2
        );

        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        taskParams.executorOperatorSet.id = unregisteredOperatorSetId;

        vm.prank(creator);
        vm.expectRevert(ITaskMailboxErrors.ExecutorOperatorSetNotRegistered.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_WhenPayloadIsEmpty() public {
        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        taskParams.payload = bytes("");

        vm.prank(creator);
        vm.expectRevert(ITaskMailboxErrors.PayloadIsEmpty.selector);
        taskMailbox.createTask(taskParams);
    }

    function test_Revert_WhenExecutorOperatorSetTaskConfigNotSet() public {
        // Create a new executor operator set without config
        uint32 newExecutorOperatorSetId = 99;
        uint32[] memory executorOperatorSetIds = new uint32[](1);
        executorOperatorSetIds[0] = newExecutorOperatorSetId;

        ITaskMailboxTypes.AvsConfig memory avsConfig = ITaskMailboxTypes.AvsConfig({
            aggregatorOperatorSetId: aggregatorOperatorSetId,
            executorOperatorSetIds: executorOperatorSetIds
        });

        vm.prank(avs);
        taskMailbox.setAvsConfig(avs, avsConfig);

        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        taskParams.executorOperatorSet.id = newExecutorOperatorSetId;

        vm.prank(creator);
        vm.expectRevert(ITaskMailboxErrors.ExecutorOperatorSetTaskConfigNotSet.selector);
        taskMailbox.createTask(taskParams);
    }
}

// Test contract for cancelTask
contract TaskMailboxUnitTests_cancelTask is TaskMailboxUnitTests {
    bytes32 public taskHash;

    function setUp() public override {
        super.setUp();

        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        taskHash = taskMailbox.createTask(taskParams);
    }

    function test_cancelTask() public {
        // Advance time by 1 second to pass TimestampAtCreation check
        vm.warp(block.timestamp + 1);

        // Expect event
        vm.expectEmit(true, true, true, true, address(taskMailbox));
        emit TaskCanceled(creator, taskHash, avs, executorOperatorSetId);

        // Cancel task
        vm.prank(creator);
        taskMailbox.cancelTask(taskHash);

        // Verify task was canceled
        ITaskMailboxTypes.TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(ITaskMailboxTypes.TaskStatus.Canceled));
    }

    function test_Revert_WhenInvalidTaskStatus() public {
        // Advance time and cancel task first
        vm.warp(block.timestamp + 1);
        vm.prank(creator);
        taskMailbox.cancelTask(taskHash);

        // Try to cancel again
        vm.prank(creator);
        vm.expectRevert(
            abi.encodeWithSelector(
                ITaskMailboxErrors.InvalidTaskStatus.selector,
                ITaskMailboxTypes.TaskStatus.Created,
                ITaskMailboxTypes.TaskStatus.Canceled
            )
        );
        taskMailbox.cancelTask(taskHash);
    }

    function test_Revert_WhenInvalidTaskCreator() public {
        vm.warp(block.timestamp + 1);

        vm.prank(address(0x999)); // Different address
        vm.expectRevert(ITaskMailboxErrors.InvalidTaskCreator.selector);
        taskMailbox.cancelTask(taskHash);
    }

    function test_Revert_WhenTimestampAtCreation() public {
        // Don't advance time
        vm.prank(creator);
        vm.expectRevert(ITaskMailboxErrors.TimestampAtCreation.selector);
        taskMailbox.cancelTask(taskHash);
    }

    function test_Revert_WhenTaskExpired() public {
        // Advance time past task SLA
        vm.warp(block.timestamp + taskSLA + 1);

        vm.prank(creator);
        vm.expectRevert(
            abi.encodeWithSelector(
                ITaskMailboxErrors.InvalidTaskStatus.selector,
                ITaskMailboxTypes.TaskStatus.Created,
                ITaskMailboxTypes.TaskStatus.Expired
            )
        );
        taskMailbox.cancelTask(taskHash);
    }
}

// Test contract for submitResult
contract TaskMailboxUnitTests_submitResult is TaskMailboxUnitTests {
    bytes32 public taskHash;

    function setUp() public override {
        super.setUp();

        // Set up executor operator set task config
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create a task
        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        taskHash = taskMailbox.createTask(taskParams);
    }

    function testFuzz_submitResult(
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
        taskMailbox.submitResult(taskHash, cert, fuzzResult);

        // Verify task was verified
        ITaskMailboxTypes.TaskStatus status = taskMailbox.getTaskStatus(taskHash);
        assertEq(uint8(status), uint8(ITaskMailboxTypes.TaskStatus.Verified));

        // Verify result was stored
        bytes memory storedResult = taskMailbox.getTaskResult(taskHash);
        assertEq(storedResult, fuzzResult);
    }

    function test_Revert_WhenInvalidTaskStatus_NotCreated() public {
        // Cancel task first
        vm.warp(block.timestamp + 1);
        vm.prank(creator);
        taskMailbox.cancelTask(taskHash);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        vm.prank(aggregator);
        vm.expectRevert(
            abi.encodeWithSelector(
                ITaskMailboxErrors.InvalidTaskStatus.selector,
                ITaskMailboxTypes.TaskStatus.Created,
                ITaskMailboxTypes.TaskStatus.Canceled
            )
        );
        taskMailbox.submitResult(taskHash, cert, bytes("result"));
    }

    function test_Revert_WhenTimestampAtCreation() public {
        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        // Don't advance time
        vm.prank(aggregator);
        vm.expectRevert(ITaskMailboxErrors.TimestampAtCreation.selector);
        taskMailbox.submitResult(taskHash, cert, bytes("result"));
    }

    function test_Revert_WhenTaskExpired() public {
        // Advance time past task SLA
        vm.warp(block.timestamp + taskSLA + 1);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(taskHash);

        vm.prank(aggregator);
        vm.expectRevert(
            abi.encodeWithSelector(
                ITaskMailboxErrors.InvalidTaskStatus.selector,
                ITaskMailboxTypes.TaskStatus.Created,
                ITaskMailboxTypes.TaskStatus.Expired
            )
        );
        taskMailbox.submitResult(taskHash, cert, bytes("result"));
    }

    function test_Revert_WhenCertificateVerificationFailed() public {
        // Create a custom mock that returns false for certificate verification
        MockBN254CertificateVerifierFailure mockFailingVerifier = new MockBN254CertificateVerifierFailure();

        // Update the config with failing verifier
        OperatorSet memory operatorSet = OperatorSet(avs, executorOperatorSetId);
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory config = _createValidExecutorOperatorSetTaskConfig();
        config.certificateVerifier = address(mockFailingVerifier);

        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, config);

        // Create new task with this config
        ITaskMailboxTypes.TaskParams memory taskParams = _createValidTaskParams();
        vm.prank(creator);
        bytes32 newTaskHash = taskMailbox.createTask(taskParams);

        // Advance time
        vm.warp(block.timestamp + 1);

        IBN254CertificateVerifier.BN254Certificate memory cert = _createValidBN254Certificate(newTaskHash);

        vm.prank(aggregator);
        vm.expectRevert(ITaskMailboxErrors.CertificateVerificationFailed.selector);
        taskMailbox.submitResult(newTaskHash, cert, bytes("result"));
    }
}
