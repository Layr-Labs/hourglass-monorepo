// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Test, console} from "forge-std/Test.sol";
import {TaskMailbox} from "../src/core/TaskMailbox.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {MockBN254CertificateVerifier} from "./mocks/MockBN254CertificateVerifier.sol";
import {MockAVSTaskHook} from "./mocks/MockAVSTaskHook.sol";
import {IBN254CertificateVerifier} from "../src/interfaces/avs/l2/IBN254CertificateVerifier.sol";
import {ITaskMailbox, ITaskMailboxTypes} from "../src/interfaces/core/ITaskMailbox.sol";

// Import scripts
import {DeployTaskMailbox} from "../script/local/DeployTaskMailbox.s.sol";
import {SetupAVSTaskMailboxConfig} from "../script/local/SetupAVSTaskMailboxConfig.s.sol";
import {CreateTask} from "../script/local/CreateTask.s.sol";

contract TaskMailboxTest is Test {
    // Contracts
    TaskMailbox public taskMailbox;
    MockBN254CertificateVerifier public certificateVerifier;
    MockAVSTaskHook public taskHook;
    
    // Addresses
    address deployer;
    address avs;
    address app;
    address refundCollector;
    address resultSubmitter;
    
    // Task details
    bytes32 public taskHash;
    
    // Error selectors
    bytes4 private invalidTaskResultSubmitterSelector;
    bytes4 private invalidTaskStatusSelector;
    bytes4 private certificateVerificationFailedSelector;
    
    function setUp() public {
        // Setup error selectors
        invalidTaskResultSubmitterSelector = bytes4(keccak256("InvalidTaskResultSubmitter()"));
        invalidTaskStatusSelector = bytes4(keccak256("InvalidTaskStatus(uint8,uint8)"));
        certificateVerificationFailedSelector = bytes4(keccak256("CertificateVerificationFailed()"));
        
        // Setup private keys similar to anvil defaults
        uint256 deployerPrivateKey = 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80;
        uint256 avsPrivateKey = 0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d;
        uint256 appPrivateKey = 0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a;
        
        // Derive addresses
        deployer = vm.addr(deployerPrivateKey);
        avs = vm.addr(avsPrivateKey);
        app = vm.addr(appPrivateKey);
        resultSubmitter = avs; // AVS acts as result submitter
        refundCollector = address(0x3);
        
        // Set environment variables for scripts
        vm.setEnv("PRIVATE_KEY_DEPLOYER", vm.toString(deployerPrivateKey));
        vm.setEnv("PRIVATE_KEY_AVS", vm.toString(avsPrivateKey));
        vm.setEnv("PRIVATE_KEY_APP", vm.toString(appPrivateKey));
        
        // Step 1: Deploy TaskMailbox using script
        DeployTaskMailbox deployScript = new DeployTaskMailbox();
        vm.prank(deployer);
        deployScript.run();
        
        // Find the deployed TaskMailbox address from logs
        // Since we can't directly get the return value from the script
        // we'll manually deploy it here for our test
        vm.startPrank(deployer);
        taskMailbox = new TaskMailbox();
        vm.stopPrank();
        
        // Step 2-4: Skip AVS L1 deployment for unit tests as we're only testing TaskMailbox
        
        // Step 5: Deploy AVS L2 contracts - MockBN254CertificateVerifier and MockAVSTaskHook
        certificateVerifier = new MockBN254CertificateVerifier();
        taskHook = new MockAVSTaskHook();
        
        // Step 6: Setup AVS Task Mailbox Config
        SetupAVSTaskMailboxConfig setupScript = new SetupAVSTaskMailboxConfig();
        setupScript.run(address(taskMailbox), address(certificateVerifier), address(taskHook));
        
        // Step 7: Create Task
        CreateTask createTaskScript = new CreateTask();
        createTaskScript.run(address(taskMailbox), avs);
        
        // Get task hash from the latest created task
        // For testing purposes, we'll create another task to have control over the task hash
        vm.startPrank(app);
        OperatorSet memory executorOperatorSet = OperatorSet({avs: avs, id: 1});
        ITaskMailboxTypes.TaskParams memory taskParams = ITaskMailboxTypes.TaskParams({
            refundCollector: refundCollector,
            avsFee: 0,
            executorOperatorSet: executorOperatorSet,
            payload: bytes("test payload")
        });
        taskHash = taskMailbox.createTask(taskParams);
        vm.stopPrank();
    }

    function testSubmitResult() public {
        // Set up result data
        bytes memory result = bytes("task result data");
        
        // Create a certificate
        IBN254CertificateVerifier.BN254Certificate memory cert;
        
        // Mock call as resultSubmitter (same as AVS)
        vm.prank(resultSubmitter);
        
        // Submit result for the task
        taskMailbox.submitResult(taskHash, cert, result);
        
        // Get task info and verify state changes
        ITaskMailboxTypes.Task memory task = taskMailbox.getTaskInfo(taskHash);
        
        // Check task status is updated
        assertEq(uint(task.status), uint(ITaskMailboxTypes.TaskStatus.Verified), "Task status should be Verified");
        
        // Check task result is stored
        assertEq(task.result, result, "Task result should be stored");
        
        // Check task result can be retrieved
        bytes memory storedResult = taskMailbox.getTaskResult(taskHash);
        assertEq(storedResult, result, "Task result should match stored result");
    }
    
    function testSubmitResultAsUnauthorized() public {
        // Set up result data
        bytes memory result = bytes("task result data");
        
        // Create a certificate
        IBN254CertificateVerifier.BN254Certificate memory cert;
        
        // Call from unauthorized address (not the resultSubmitter)
        vm.prank(address(0x6));
        
        // Expect revert with InvalidTaskResultSubmitter
        vm.expectRevert(abi.encodeWithSelector(invalidTaskResultSubmitterSelector));
        taskMailbox.submitResult(taskHash, cert, result);
    }
    
    function testSubmitResultForNonExistentTask() public {
        // Set up result data
        bytes memory result = bytes("task result data");
        
        // Create a certificate
        IBN254CertificateVerifier.BN254Certificate memory cert;
        
        // Create a random taskHash that doesn't exist
        bytes32 nonExistentTaskHash = keccak256("non existent task");
        
        // Mock call as resultSubmitter
        vm.prank(resultSubmitter);
        
        // Expect revert with InvalidTaskStatus
        vm.expectRevert(abi.encodeWithSelector(invalidTaskStatusSelector, uint8(ITaskMailboxTypes.TaskStatus.Created), 0));
        taskMailbox.submitResult(nonExistentTaskHash, cert, result);
    }
    
    function testSubmitResultWithInvalidCertificate() public {
        // Set up result data
        bytes memory result = bytes("task result data");
        
        // Create a certificate
        IBN254CertificateVerifier.BN254Certificate memory cert;
        
        // Mock the certificate validation to fail
        MockBN254CertificateVerifier mockFailingVerifier = new MockBN254CertificateVerifier();
        
        // Make the mock return false for verifyCertificateProportion
        vm.mockCall(
            address(mockFailingVerifier),
            abi.encodeWithSelector(IBN254CertificateVerifier.verifyCertificateProportion.selector),
            abi.encode(false)
        );
        
        // Update the task config to use the failing verifier
        OperatorSet memory operatorSet = OperatorSet({
            avs: avs,
            id: 1
        });
        
        ITaskMailboxTypes.ExecutorOperatorSetTaskConfig memory taskConfig = ITaskMailboxTypes.ExecutorOperatorSetTaskConfig({
            certificateVerifier: address(mockFailingVerifier),
            taskHook: taskHook,
            feeToken: IERC20(address(0)),
            feeCollector: address(0),
            taskSLA: 3600,
            stakeProportionThreshold: 6667,
            taskMetadata: bytes("")
        });
        
        vm.prank(avs);
        taskMailbox.setExecutorOperatorSetTaskConfig(operatorSet, taskConfig);
        
        // Mock call as resultSubmitter
        vm.prank(resultSubmitter);
        
        // Expect revert with CertificateVerificationFailed
        vm.expectRevert(abi.encodeWithSelector(certificateVerificationFailedSelector));
        taskMailbox.submitResult(taskHash, cert, result);
    }
}
