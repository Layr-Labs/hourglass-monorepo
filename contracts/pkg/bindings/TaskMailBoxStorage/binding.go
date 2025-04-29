// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package TaskMailBoxStorage

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BN254G1Point is an auto generated low-level Go binding around an user-defined struct.
type BN254G1Point struct {
	X *big.Int
	Y *big.Int
}

// BN254G2Point is an auto generated low-level Go binding around an user-defined struct.
type BN254G2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

// IBN254CertificateVerifierBN254Certificate is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierBN254Certificate struct {
	ReferenceTimestamp uint32
	MessageHash        [32]byte
	Sig                BN254G1Point
	Apk                BN254G2Point
	NonsignerIndices   []uint32
	NonSignerWitnesses []IBN254CertificateVerifierBN254OperatorInfoWitness
}

// IBN254CertificateVerifierBN254OperatorInfo is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierBN254OperatorInfo struct {
	Pubkey  BN254G1Point
	Weights []*big.Int
}

// IBN254CertificateVerifierBN254OperatorInfoWitness is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierBN254OperatorInfoWitness struct {
	OperatorIndex      uint32
	OperatorInfoProofs []byte
	OperatorInfo       IBN254CertificateVerifierBN254OperatorInfo
}

// ITaskMailboxTypesOperatorSetTaskConfig is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesOperatorSetTaskConfig struct {
	CertificateVerifier      common.Address
	TaskHook                 common.Address
	Aggregator               common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}

// ITaskMailboxTypesTask is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesTask struct {
	Creator               common.Address
	CreationTime          *big.Int
	Status                uint8
	OperatorSet           OperatorSet
	RefundCollector       common.Address
	AvsFee                *big.Int
	FeeSplit              uint16
	OperatorSetTaskConfig ITaskMailboxTypesOperatorSetTaskConfig
	Payload               []byte
	Result                []byte
}

// ITaskMailboxTypesTaskParams is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesTaskParams struct {
	RefundCollector common.Address
	AvsFee          *big.Int
	OperatorSet     OperatorSet
	Payload         []byte
}

// OperatorSet is an auto generated low-level Go binding around an user-defined struct.
type OperatorSet struct {
	Avs common.Address
	Id  uint32
}

// TaskMailBoxStorageMetaData contains all meta data concerning the TaskMailBoxStorage contract.
var TaskMailBoxStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"cancelTask\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createTask\",\"inputs\":[{\"name\":\"taskParams\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.TaskParams\",\"components\":[{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskInfo\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.Task\",\"components\":[{\"name\":\"creator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"creationTime\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"feeSplit\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"operatorSetTaskConfig\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskStatus\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorSetRegistered\",\"inputs\":[{\"name\":\"operatorSetKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"registered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSetKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperatorSet\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"OperatorSetRegistered\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSetTaskConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"config\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCanceled\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCreated\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"},{\"name\":\"taskDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payload\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskVerified\",\"inputs\":[{\"name\":\"aggregator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"result\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CertificateVerificationFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddressZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskAggregator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskCreator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskStatus\",\"inputs\":[{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}]},{\"type\":\"error\",\"name\":\"OperatorSetNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorSetTaskConfigNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PayloadIsEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TaskSLAIsZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TimestampAtCreation\",\"inputs\":[]}]",
}

// TaskMailBoxStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use TaskMailBoxStorageMetaData.ABI instead.
var TaskMailBoxStorageABI = TaskMailBoxStorageMetaData.ABI

// TaskMailBoxStorage is an auto generated Go binding around an Ethereum contract.
type TaskMailBoxStorage struct {
	TaskMailBoxStorageCaller     // Read-only binding to the contract
	TaskMailBoxStorageTransactor // Write-only binding to the contract
	TaskMailBoxStorageFilterer   // Log filterer for contract events
}

// TaskMailBoxStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type TaskMailBoxStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskMailBoxStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TaskMailBoxStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskMailBoxStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TaskMailBoxStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskMailBoxStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TaskMailBoxStorageSession struct {
	Contract     *TaskMailBoxStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// TaskMailBoxStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TaskMailBoxStorageCallerSession struct {
	Contract *TaskMailBoxStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// TaskMailBoxStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TaskMailBoxStorageTransactorSession struct {
	Contract     *TaskMailBoxStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// TaskMailBoxStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type TaskMailBoxStorageRaw struct {
	Contract *TaskMailBoxStorage // Generic contract binding to access the raw methods on
}

// TaskMailBoxStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TaskMailBoxStorageCallerRaw struct {
	Contract *TaskMailBoxStorageCaller // Generic read-only contract binding to access the raw methods on
}

// TaskMailBoxStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TaskMailBoxStorageTransactorRaw struct {
	Contract *TaskMailBoxStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTaskMailBoxStorage creates a new instance of TaskMailBoxStorage, bound to a specific deployed contract.
func NewTaskMailBoxStorage(address common.Address, backend bind.ContractBackend) (*TaskMailBoxStorage, error) {
	contract, err := bindTaskMailBoxStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorage{TaskMailBoxStorageCaller: TaskMailBoxStorageCaller{contract: contract}, TaskMailBoxStorageTransactor: TaskMailBoxStorageTransactor{contract: contract}, TaskMailBoxStorageFilterer: TaskMailBoxStorageFilterer{contract: contract}}, nil
}

// NewTaskMailBoxStorageCaller creates a new read-only instance of TaskMailBoxStorage, bound to a specific deployed contract.
func NewTaskMailBoxStorageCaller(address common.Address, caller bind.ContractCaller) (*TaskMailBoxStorageCaller, error) {
	contract, err := bindTaskMailBoxStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageCaller{contract: contract}, nil
}

// NewTaskMailBoxStorageTransactor creates a new write-only instance of TaskMailBoxStorage, bound to a specific deployed contract.
func NewTaskMailBoxStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*TaskMailBoxStorageTransactor, error) {
	contract, err := bindTaskMailBoxStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageTransactor{contract: contract}, nil
}

// NewTaskMailBoxStorageFilterer creates a new log filterer instance of TaskMailBoxStorage, bound to a specific deployed contract.
func NewTaskMailBoxStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*TaskMailBoxStorageFilterer, error) {
	contract, err := bindTaskMailBoxStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageFilterer{contract: contract}, nil
}

// bindTaskMailBoxStorage binds a generic wrapper to an already deployed contract.
func bindTaskMailBoxStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TaskMailBoxStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskMailBoxStorage *TaskMailBoxStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskMailBoxStorage.Contract.TaskMailBoxStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskMailBoxStorage *TaskMailBoxStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.TaskMailBoxStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskMailBoxStorage *TaskMailBoxStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.TaskMailBoxStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskMailBoxStorage *TaskMailBoxStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskMailBoxStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.contract.Transact(opts, method, params...)
}

// GetOperatorSetTaskConfig is a free data retrieval call binding the contract method 0xca2df5eb.
//
// Solidity: function getOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,address,uint96,uint16,bytes))
func (_TaskMailBoxStorage *TaskMailBoxStorageCaller) GetOperatorSetTaskConfig(opts *bind.CallOpts, operatorSet OperatorSet) (ITaskMailboxTypesOperatorSetTaskConfig, error) {
	var out []interface{}
	err := _TaskMailBoxStorage.contract.Call(opts, &out, "getOperatorSetTaskConfig", operatorSet)

	if err != nil {
		return *new(ITaskMailboxTypesOperatorSetTaskConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesOperatorSetTaskConfig)).(*ITaskMailboxTypesOperatorSetTaskConfig)

	return out0, err

}

// GetOperatorSetTaskConfig is a free data retrieval call binding the contract method 0xca2df5eb.
//
// Solidity: function getOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,address,uint96,uint16,bytes))
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) GetOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesOperatorSetTaskConfig, error) {
	return _TaskMailBoxStorage.Contract.GetOperatorSetTaskConfig(&_TaskMailBoxStorage.CallOpts, operatorSet)
}

// GetOperatorSetTaskConfig is a free data retrieval call binding the contract method 0xca2df5eb.
//
// Solidity: function getOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,address,uint96,uint16,bytes))
func (_TaskMailBoxStorage *TaskMailBoxStorageCallerSession) GetOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesOperatorSetTaskConfig, error) {
	return _TaskMailBoxStorage.Contract.GetOperatorSetTaskConfig(&_TaskMailBoxStorage.CallOpts, operatorSet)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,(address,uint32),address,uint96,uint16,(address,address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_TaskMailBoxStorage *TaskMailBoxStorageCaller) GetTaskInfo(opts *bind.CallOpts, taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	var out []interface{}
	err := _TaskMailBoxStorage.contract.Call(opts, &out, "getTaskInfo", taskHash)

	if err != nil {
		return *new(ITaskMailboxTypesTask), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesTask)).(*ITaskMailboxTypesTask)

	return out0, err

}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,(address,uint32),address,uint96,uint16,(address,address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _TaskMailBoxStorage.Contract.GetTaskInfo(&_TaskMailBoxStorage.CallOpts, taskHash)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,(address,uint32),address,uint96,uint16,(address,address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_TaskMailBoxStorage *TaskMailBoxStorageCallerSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _TaskMailBoxStorage.Contract.GetTaskInfo(&_TaskMailBoxStorage.CallOpts, taskHash)
}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_TaskMailBoxStorage *TaskMailBoxStorageCaller) GetTaskStatus(opts *bind.CallOpts, taskHash [32]byte) (uint8, error) {
	var out []interface{}
	err := _TaskMailBoxStorage.contract.Call(opts, &out, "getTaskStatus", taskHash)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) GetTaskStatus(taskHash [32]byte) (uint8, error) {
	return _TaskMailBoxStorage.Contract.GetTaskStatus(&_TaskMailBoxStorage.CallOpts, taskHash)
}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_TaskMailBoxStorage *TaskMailBoxStorageCallerSession) GetTaskStatus(taskHash [32]byte) (uint8, error) {
	return _TaskMailBoxStorage.Contract.GetTaskStatus(&_TaskMailBoxStorage.CallOpts, taskHash)
}

// IsOperatorSetRegistered is a free data retrieval call binding the contract method 0xc4a1ca05.
//
// Solidity: function isOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool registered)
func (_TaskMailBoxStorage *TaskMailBoxStorageCaller) IsOperatorSetRegistered(opts *bind.CallOpts, operatorSetKey [32]byte) (bool, error) {
	var out []interface{}
	err := _TaskMailBoxStorage.contract.Call(opts, &out, "isOperatorSetRegistered", operatorSetKey)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorSetRegistered is a free data retrieval call binding the contract method 0xc4a1ca05.
//
// Solidity: function isOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool registered)
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) IsOperatorSetRegistered(operatorSetKey [32]byte) (bool, error) {
	return _TaskMailBoxStorage.Contract.IsOperatorSetRegistered(&_TaskMailBoxStorage.CallOpts, operatorSetKey)
}

// IsOperatorSetRegistered is a free data retrieval call binding the contract method 0xc4a1ca05.
//
// Solidity: function isOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool registered)
func (_TaskMailBoxStorage *TaskMailBoxStorageCallerSession) IsOperatorSetRegistered(operatorSetKey [32]byte) (bool, error) {
	return _TaskMailBoxStorage.Contract.IsOperatorSetRegistered(&_TaskMailBoxStorage.CallOpts, operatorSetKey)
}

// OperatorSetTaskConfig is a free data retrieval call binding the contract method 0x825c2b8c.
//
// Solidity: function operatorSetTaskConfig(bytes32 operatorSetKey) view returns(address certificateVerifier, address taskHook, address aggregator, address feeToken, address feeCollector, uint96 taskSLA, uint16 stakeProportionThreshold, bytes taskMetadata)
func (_TaskMailBoxStorage *TaskMailBoxStorageCaller) OperatorSetTaskConfig(opts *bind.CallOpts, operatorSetKey [32]byte) (struct {
	CertificateVerifier      common.Address
	TaskHook                 common.Address
	Aggregator               common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}, error) {
	var out []interface{}
	err := _TaskMailBoxStorage.contract.Call(opts, &out, "operatorSetTaskConfig", operatorSetKey)

	outstruct := new(struct {
		CertificateVerifier      common.Address
		TaskHook                 common.Address
		Aggregator               common.Address
		FeeToken                 common.Address
		FeeCollector             common.Address
		TaskSLA                  *big.Int
		StakeProportionThreshold uint16
		TaskMetadata             []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.CertificateVerifier = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.TaskHook = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Aggregator = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.FeeToken = *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	outstruct.FeeCollector = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.TaskSLA = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.StakeProportionThreshold = *abi.ConvertType(out[6], new(uint16)).(*uint16)
	outstruct.TaskMetadata = *abi.ConvertType(out[7], new([]byte)).(*[]byte)

	return *outstruct, err

}

// OperatorSetTaskConfig is a free data retrieval call binding the contract method 0x825c2b8c.
//
// Solidity: function operatorSetTaskConfig(bytes32 operatorSetKey) view returns(address certificateVerifier, address taskHook, address aggregator, address feeToken, address feeCollector, uint96 taskSLA, uint16 stakeProportionThreshold, bytes taskMetadata)
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) OperatorSetTaskConfig(operatorSetKey [32]byte) (struct {
	CertificateVerifier      common.Address
	TaskHook                 common.Address
	Aggregator               common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}, error) {
	return _TaskMailBoxStorage.Contract.OperatorSetTaskConfig(&_TaskMailBoxStorage.CallOpts, operatorSetKey)
}

// OperatorSetTaskConfig is a free data retrieval call binding the contract method 0x825c2b8c.
//
// Solidity: function operatorSetTaskConfig(bytes32 operatorSetKey) view returns(address certificateVerifier, address taskHook, address aggregator, address feeToken, address feeCollector, uint96 taskSLA, uint16 stakeProportionThreshold, bytes taskMetadata)
func (_TaskMailBoxStorage *TaskMailBoxStorageCallerSession) OperatorSetTaskConfig(operatorSetKey [32]byte) (struct {
	CertificateVerifier      common.Address
	TaskHook                 common.Address
	Aggregator               common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}, error) {
	return _TaskMailBoxStorage.Contract.OperatorSetTaskConfig(&_TaskMailBoxStorage.CallOpts, operatorSetKey)
}

// CancelTask is a paid mutator transaction binding the contract method 0xee8ca3b5.
//
// Solidity: function cancelTask(bytes32 taskHash) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactor) CancelTask(opts *bind.TransactOpts, taskHash [32]byte) (*types.Transaction, error) {
	return _TaskMailBoxStorage.contract.Transact(opts, "cancelTask", taskHash)
}

// CancelTask is a paid mutator transaction binding the contract method 0xee8ca3b5.
//
// Solidity: function cancelTask(bytes32 taskHash) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) CancelTask(taskHash [32]byte) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.CancelTask(&_TaskMailBoxStorage.TransactOpts, taskHash)
}

// CancelTask is a paid mutator transaction binding the contract method 0xee8ca3b5.
//
// Solidity: function cancelTask(bytes32 taskHash) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorSession) CancelTask(taskHash [32]byte) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.CancelTask(&_TaskMailBoxStorage.TransactOpts, taskHash)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactor) CreateTask(opts *bind.TransactOpts, taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _TaskMailBoxStorage.contract.Transact(opts, "createTask", taskParams)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) CreateTask(taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.CreateTask(&_TaskMailBoxStorage.TransactOpts, taskParams)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorSession) CreateTask(taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.CreateTask(&_TaskMailBoxStorage.TransactOpts, taskParams)
}

// RegisterOperatorSet is a paid mutator transaction binding the contract method 0xadf87665.
//
// Solidity: function registerOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactor) RegisterOperatorSet(opts *bind.TransactOpts, operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _TaskMailBoxStorage.contract.Transact(opts, "registerOperatorSet", operatorSet, isRegistered)
}

// RegisterOperatorSet is a paid mutator transaction binding the contract method 0xadf87665.
//
// Solidity: function registerOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) RegisterOperatorSet(operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.RegisterOperatorSet(&_TaskMailBoxStorage.TransactOpts, operatorSet, isRegistered)
}

// RegisterOperatorSet is a paid mutator transaction binding the contract method 0xadf87665.
//
// Solidity: function registerOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorSession) RegisterOperatorSet(operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.RegisterOperatorSet(&_TaskMailBoxStorage.TransactOpts, operatorSet, isRegistered)
}

// SetOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x9ff625d8.
//
// Solidity: function setOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,address,uint96,uint16,bytes) config) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactor) SetOperatorSetTaskConfig(opts *bind.TransactOpts, operatorSet OperatorSet, config ITaskMailboxTypesOperatorSetTaskConfig) (*types.Transaction, error) {
	return _TaskMailBoxStorage.contract.Transact(opts, "setOperatorSetTaskConfig", operatorSet, config)
}

// SetOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x9ff625d8.
//
// Solidity: function setOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,address,uint96,uint16,bytes) config) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) SetOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesOperatorSetTaskConfig) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.SetOperatorSetTaskConfig(&_TaskMailBoxStorage.TransactOpts, operatorSet, config)
}

// SetOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x9ff625d8.
//
// Solidity: function setOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,address,uint96,uint16,bytes) config) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorSession) SetOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesOperatorSetTaskConfig) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.SetOperatorSetTaskConfig(&_TaskMailBoxStorage.TransactOpts, operatorSet, config)
}

// SubmitResult is a paid mutator transaction binding the contract method 0x3b433719.
//
// Solidity: function submitResult(bytes32 taskHash, (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, bytes result) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactor) SubmitResult(opts *bind.TransactOpts, taskHash [32]byte, cert IBN254CertificateVerifierBN254Certificate, result []byte) (*types.Transaction, error) {
	return _TaskMailBoxStorage.contract.Transact(opts, "submitResult", taskHash, cert, result)
}

// SubmitResult is a paid mutator transaction binding the contract method 0x3b433719.
//
// Solidity: function submitResult(bytes32 taskHash, (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, bytes result) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageSession) SubmitResult(taskHash [32]byte, cert IBN254CertificateVerifierBN254Certificate, result []byte) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.SubmitResult(&_TaskMailBoxStorage.TransactOpts, taskHash, cert, result)
}

// SubmitResult is a paid mutator transaction binding the contract method 0x3b433719.
//
// Solidity: function submitResult(bytes32 taskHash, (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, bytes result) returns()
func (_TaskMailBoxStorage *TaskMailBoxStorageTransactorSession) SubmitResult(taskHash [32]byte, cert IBN254CertificateVerifierBN254Certificate, result []byte) (*types.Transaction, error) {
	return _TaskMailBoxStorage.Contract.SubmitResult(&_TaskMailBoxStorage.TransactOpts, taskHash, cert, result)
}

// TaskMailBoxStorageOperatorSetRegisteredIterator is returned from FilterOperatorSetRegistered and is used to iterate over the raw logs and unpacked data for OperatorSetRegistered events raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageOperatorSetRegisteredIterator struct {
	Event *TaskMailBoxStorageOperatorSetRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TaskMailBoxStorageOperatorSetRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailBoxStorageOperatorSetRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TaskMailBoxStorageOperatorSetRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TaskMailBoxStorageOperatorSetRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailBoxStorageOperatorSetRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailBoxStorageOperatorSetRegistered represents a OperatorSetRegistered event raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageOperatorSetRegistered struct {
	Caller        common.Address
	Avs           common.Address
	OperatorSetId uint32
	IsRegistered  bool
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetRegistered is a free log retrieval operation binding the contract event 0xd5d99994c6b140b6722f9453b46413489373660d9dfd425d769dde6dd330bd58.
//
// Solidity: event OperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed operatorSetId, bool isRegistered)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) FilterOperatorSetRegistered(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, operatorSetId []uint32) (*TaskMailBoxStorageOperatorSetRegisteredIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.FilterLogs(opts, "OperatorSetRegistered", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageOperatorSetRegisteredIterator{contract: _TaskMailBoxStorage.contract, event: "OperatorSetRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorSetRegistered is a free log subscription operation binding the contract event 0xd5d99994c6b140b6722f9453b46413489373660d9dfd425d769dde6dd330bd58.
//
// Solidity: event OperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed operatorSetId, bool isRegistered)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) WatchOperatorSetRegistered(opts *bind.WatchOpts, sink chan<- *TaskMailBoxStorageOperatorSetRegistered, caller []common.Address, avs []common.Address, operatorSetId []uint32) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.WatchLogs(opts, "OperatorSetRegistered", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailBoxStorageOperatorSetRegistered)
				if err := _TaskMailBoxStorage.contract.UnpackLog(event, "OperatorSetRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOperatorSetRegistered is a log parse operation binding the contract event 0xd5d99994c6b140b6722f9453b46413489373660d9dfd425d769dde6dd330bd58.
//
// Solidity: event OperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed operatorSetId, bool isRegistered)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) ParseOperatorSetRegistered(log types.Log) (*TaskMailBoxStorageOperatorSetRegistered, error) {
	event := new(TaskMailBoxStorageOperatorSetRegistered)
	if err := _TaskMailBoxStorage.contract.UnpackLog(event, "OperatorSetRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailBoxStorageOperatorSetTaskConfigSetIterator is returned from FilterOperatorSetTaskConfigSet and is used to iterate over the raw logs and unpacked data for OperatorSetTaskConfigSet events raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageOperatorSetTaskConfigSetIterator struct {
	Event *TaskMailBoxStorageOperatorSetTaskConfigSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TaskMailBoxStorageOperatorSetTaskConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailBoxStorageOperatorSetTaskConfigSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TaskMailBoxStorageOperatorSetTaskConfigSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TaskMailBoxStorageOperatorSetTaskConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailBoxStorageOperatorSetTaskConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailBoxStorageOperatorSetTaskConfigSet represents a OperatorSetTaskConfigSet event raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageOperatorSetTaskConfigSet struct {
	Caller        common.Address
	Avs           common.Address
	OperatorSetId uint32
	Config        ITaskMailboxTypesOperatorSetTaskConfig
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetTaskConfigSet is a free log retrieval operation binding the contract event 0x6deb0e45caf1a5c5fdab7bfaaf5a8e90a6ad6f5c2c6076e4cd03ba3e4d0ae415.
//
// Solidity: event OperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed operatorSetId, (address,address,address,address,address,uint96,uint16,bytes) config)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) FilterOperatorSetTaskConfigSet(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, operatorSetId []uint32) (*TaskMailBoxStorageOperatorSetTaskConfigSetIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.FilterLogs(opts, "OperatorSetTaskConfigSet", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageOperatorSetTaskConfigSetIterator{contract: _TaskMailBoxStorage.contract, event: "OperatorSetTaskConfigSet", logs: logs, sub: sub}, nil
}

// WatchOperatorSetTaskConfigSet is a free log subscription operation binding the contract event 0x6deb0e45caf1a5c5fdab7bfaaf5a8e90a6ad6f5c2c6076e4cd03ba3e4d0ae415.
//
// Solidity: event OperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed operatorSetId, (address,address,address,address,address,uint96,uint16,bytes) config)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) WatchOperatorSetTaskConfigSet(opts *bind.WatchOpts, sink chan<- *TaskMailBoxStorageOperatorSetTaskConfigSet, caller []common.Address, avs []common.Address, operatorSetId []uint32) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.WatchLogs(opts, "OperatorSetTaskConfigSet", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailBoxStorageOperatorSetTaskConfigSet)
				if err := _TaskMailBoxStorage.contract.UnpackLog(event, "OperatorSetTaskConfigSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOperatorSetTaskConfigSet is a log parse operation binding the contract event 0x6deb0e45caf1a5c5fdab7bfaaf5a8e90a6ad6f5c2c6076e4cd03ba3e4d0ae415.
//
// Solidity: event OperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed operatorSetId, (address,address,address,address,address,uint96,uint16,bytes) config)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) ParseOperatorSetTaskConfigSet(log types.Log) (*TaskMailBoxStorageOperatorSetTaskConfigSet, error) {
	event := new(TaskMailBoxStorageOperatorSetTaskConfigSet)
	if err := _TaskMailBoxStorage.contract.UnpackLog(event, "OperatorSetTaskConfigSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailBoxStorageTaskCanceledIterator is returned from FilterTaskCanceled and is used to iterate over the raw logs and unpacked data for TaskCanceled events raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageTaskCanceledIterator struct {
	Event *TaskMailBoxStorageTaskCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TaskMailBoxStorageTaskCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailBoxStorageTaskCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TaskMailBoxStorageTaskCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TaskMailBoxStorageTaskCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailBoxStorageTaskCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailBoxStorageTaskCanceled represents a TaskCanceled event raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageTaskCanceled struct {
	Creator       common.Address
	TaskHash      [32]byte
	Avs           common.Address
	OperatorSetId uint32
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTaskCanceled is a free log retrieval operation binding the contract event 0x3e701c33cc740e1f61ccdcafcf97e5e65a0d7f4617aed0e8ae51be092ac18a59.
//
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) FilterTaskCanceled(opts *bind.FilterOpts, creator []common.Address, taskHash [][32]byte, avs []common.Address) (*TaskMailBoxStorageTaskCanceledIterator, error) {

	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}
	var taskHashRule []interface{}
	for _, taskHashItem := range taskHash {
		taskHashRule = append(taskHashRule, taskHashItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.FilterLogs(opts, "TaskCanceled", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageTaskCanceledIterator{contract: _TaskMailBoxStorage.contract, event: "TaskCanceled", logs: logs, sub: sub}, nil
}

// WatchTaskCanceled is a free log subscription operation binding the contract event 0x3e701c33cc740e1f61ccdcafcf97e5e65a0d7f4617aed0e8ae51be092ac18a59.
//
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) WatchTaskCanceled(opts *bind.WatchOpts, sink chan<- *TaskMailBoxStorageTaskCanceled, creator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}
	var taskHashRule []interface{}
	for _, taskHashItem := range taskHash {
		taskHashRule = append(taskHashRule, taskHashItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.WatchLogs(opts, "TaskCanceled", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailBoxStorageTaskCanceled)
				if err := _TaskMailBoxStorage.contract.UnpackLog(event, "TaskCanceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskCanceled is a log parse operation binding the contract event 0x3e701c33cc740e1f61ccdcafcf97e5e65a0d7f4617aed0e8ae51be092ac18a59.
//
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) ParseTaskCanceled(log types.Log) (*TaskMailBoxStorageTaskCanceled, error) {
	event := new(TaskMailBoxStorageTaskCanceled)
	if err := _TaskMailBoxStorage.contract.UnpackLog(event, "TaskCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailBoxStorageTaskCreatedIterator is returned from FilterTaskCreated and is used to iterate over the raw logs and unpacked data for TaskCreated events raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageTaskCreatedIterator struct {
	Event *TaskMailBoxStorageTaskCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TaskMailBoxStorageTaskCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailBoxStorageTaskCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TaskMailBoxStorageTaskCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TaskMailBoxStorageTaskCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailBoxStorageTaskCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailBoxStorageTaskCreated represents a TaskCreated event raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageTaskCreated struct {
	Creator         common.Address
	TaskHash        [32]byte
	Avs             common.Address
	OperatorSetId   uint32
	RefundCollector common.Address
	AvsFee          *big.Int
	TaskDeadline    *big.Int
	Payload         []byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterTaskCreated is a free log retrieval operation binding the contract event 0x4a09af06a0e08fd1c053a8b400de7833019c88066be8a2d3b3b17174a74fe317.
//
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) FilterTaskCreated(opts *bind.FilterOpts, creator []common.Address, taskHash [][32]byte, avs []common.Address) (*TaskMailBoxStorageTaskCreatedIterator, error) {

	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}
	var taskHashRule []interface{}
	for _, taskHashItem := range taskHash {
		taskHashRule = append(taskHashRule, taskHashItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.FilterLogs(opts, "TaskCreated", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageTaskCreatedIterator{contract: _TaskMailBoxStorage.contract, event: "TaskCreated", logs: logs, sub: sub}, nil
}

// WatchTaskCreated is a free log subscription operation binding the contract event 0x4a09af06a0e08fd1c053a8b400de7833019c88066be8a2d3b3b17174a74fe317.
//
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) WatchTaskCreated(opts *bind.WatchOpts, sink chan<- *TaskMailBoxStorageTaskCreated, creator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

	var creatorRule []interface{}
	for _, creatorItem := range creator {
		creatorRule = append(creatorRule, creatorItem)
	}
	var taskHashRule []interface{}
	for _, taskHashItem := range taskHash {
		taskHashRule = append(taskHashRule, taskHashItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.WatchLogs(opts, "TaskCreated", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailBoxStorageTaskCreated)
				if err := _TaskMailBoxStorage.contract.UnpackLog(event, "TaskCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskCreated is a log parse operation binding the contract event 0x4a09af06a0e08fd1c053a8b400de7833019c88066be8a2d3b3b17174a74fe317.
//
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) ParseTaskCreated(log types.Log) (*TaskMailBoxStorageTaskCreated, error) {
	event := new(TaskMailBoxStorageTaskCreated)
	if err := _TaskMailBoxStorage.contract.UnpackLog(event, "TaskCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailBoxStorageTaskVerifiedIterator is returned from FilterTaskVerified and is used to iterate over the raw logs and unpacked data for TaskVerified events raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageTaskVerifiedIterator struct {
	Event *TaskMailBoxStorageTaskVerified // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TaskMailBoxStorageTaskVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailBoxStorageTaskVerified)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TaskMailBoxStorageTaskVerified)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TaskMailBoxStorageTaskVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailBoxStorageTaskVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailBoxStorageTaskVerified represents a TaskVerified event raised by the TaskMailBoxStorage contract.
type TaskMailBoxStorageTaskVerified struct {
	Aggregator    common.Address
	TaskHash      [32]byte
	Avs           common.Address
	OperatorSetId uint32
	Result        []byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTaskVerified is a free log retrieval operation binding the contract event 0xd7eb53a86d7419ffc42bf17e0a61b4a2a8ab7f2e62c19368cee7d8822ea9f453.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, bytes result)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) FilterTaskVerified(opts *bind.FilterOpts, aggregator []common.Address, taskHash [][32]byte, avs []common.Address) (*TaskMailBoxStorageTaskVerifiedIterator, error) {

	var aggregatorRule []interface{}
	for _, aggregatorItem := range aggregator {
		aggregatorRule = append(aggregatorRule, aggregatorItem)
	}
	var taskHashRule []interface{}
	for _, taskHashItem := range taskHash {
		taskHashRule = append(taskHashRule, taskHashItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.FilterLogs(opts, "TaskVerified", aggregatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailBoxStorageTaskVerifiedIterator{contract: _TaskMailBoxStorage.contract, event: "TaskVerified", logs: logs, sub: sub}, nil
}

// WatchTaskVerified is a free log subscription operation binding the contract event 0xd7eb53a86d7419ffc42bf17e0a61b4a2a8ab7f2e62c19368cee7d8822ea9f453.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, bytes result)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) WatchTaskVerified(opts *bind.WatchOpts, sink chan<- *TaskMailBoxStorageTaskVerified, aggregator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

	var aggregatorRule []interface{}
	for _, aggregatorItem := range aggregator {
		aggregatorRule = append(aggregatorRule, aggregatorItem)
	}
	var taskHashRule []interface{}
	for _, taskHashItem := range taskHash {
		taskHashRule = append(taskHashRule, taskHashItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _TaskMailBoxStorage.contract.WatchLogs(opts, "TaskVerified", aggregatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailBoxStorageTaskVerified)
				if err := _TaskMailBoxStorage.contract.UnpackLog(event, "TaskVerified", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskVerified is a log parse operation binding the contract event 0xd7eb53a86d7419ffc42bf17e0a61b4a2a8ab7f2e62c19368cee7d8822ea9f453.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, bytes result)
func (_TaskMailBoxStorage *TaskMailBoxStorageFilterer) ParseTaskVerified(log types.Log) (*TaskMailBoxStorageTaskVerified, error) {
	event := new(TaskMailBoxStorageTaskVerified)
	if err := _TaskMailBoxStorage.contract.UnpackLog(event, "TaskVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
