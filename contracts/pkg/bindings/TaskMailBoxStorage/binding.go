// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package TaskMailboxStorage

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

// IBN254CertificateVerifierTypesBN254Certificate is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierTypesBN254Certificate struct {
	ReferenceTimestamp uint32
	MessageHash        [32]byte
	Signature          BN254G1Point
	Apk                BN254G2Point
	NonSignerWitnesses []IBN254CertificateVerifierTypesBN254OperatorInfoWitness
}

// IBN254CertificateVerifierTypesBN254OperatorInfoWitness is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierTypesBN254OperatorInfoWitness struct {
	OperatorIndex     uint32
	OperatorInfoProof []byte
	OperatorInfo      IOperatorTableCalculatorTypesBN254OperatorInfo
}

// IECDSACertificateVerifierTypesECDSACertificate is an auto generated low-level Go binding around an user-defined struct.
type IECDSACertificateVerifierTypesECDSACertificate struct {
	ReferenceTimestamp uint32
	MessageHash        [32]byte
	Sig                []byte
}

// IOperatorTableCalculatorTypesBN254OperatorInfo is an auto generated low-level Go binding around an user-defined struct.
type IOperatorTableCalculatorTypesBN254OperatorInfo struct {
	Pubkey  BN254G1Point
	Weights []*big.Int
}

// ITaskMailboxTypesExecutorOperatorSetTaskConfig is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesExecutorOperatorSetTaskConfig struct {
	CurveType                uint8
	TaskHook                 common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}

// ITaskMailboxTypesTask is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesTask struct {
	Creator                       common.Address
	CreationTime                  *big.Int
	Status                        uint8
	Avs                           common.Address
	ExecutorOperatorSetId         uint32
	RefundCollector               common.Address
	AvsFee                        *big.Int
	FeeSplit                      uint16
	ExecutorOperatorSetTaskConfig ITaskMailboxTypesExecutorOperatorSetTaskConfig
	Payload                       []byte
	ExecutorCert                  []byte
	Result                        []byte
}

// ITaskMailboxTypesTaskParams is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesTaskParams struct {
	RefundCollector     common.Address
	AvsFee              *big.Int
	ExecutorOperatorSet OperatorSet
	Payload             []byte
}

// OperatorSet is an auto generated low-level Go binding around an user-defined struct.
type OperatorSet struct {
	Avs common.Address
	Id  uint32
}

// TaskMailboxStorageMetaData contains all meta data concerning the TaskMailboxStorage contract.
var TaskMailboxStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"BN254_CERTIFICATE_VERIFIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ECDSA_CERTIFICATE_VERIFIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"createTask\",\"inputs\":[{\"name\":\"taskParams\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.TaskParams\",\"components\":[{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"executorOperatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"executorOperatorSetTaskConfigs\",\"inputs\":[{\"name\":\"operatorSetKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"curveType\",\"type\":\"uint8\",\"internalType\":\"enumIKeyRegistrarTypes.CurveType\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getBN254CertificateBytes\",\"inputs\":[{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifierTypes.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"signature\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifierTypes.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProof\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIOperatorTableCalculatorTypes.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}]}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getECDSACertificateBytes\",\"inputs\":[{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIECDSACertificateVerifierTypes.ECDSACertificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getExecutorOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"curveType\",\"type\":\"uint8\",\"internalType\":\"enumIKeyRegistrarTypes.CurveType\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskInfo\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.Task\",\"components\":[{\"name\":\"creator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"creationTime\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"feeSplit\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"executorOperatorSetTaskConfig\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"curveType\",\"type\":\"uint8\",\"internalType\":\"enumIKeyRegistrarTypes.CurveType\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"executorCert\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskStatus\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isExecutorOperatorSetRegistered\",\"inputs\":[{\"name\":\"operatorSetKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerExecutorOperatorSet\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setExecutorOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"curveType\",\"type\":\"uint8\",\"internalType\":\"enumIKeyRegistrarTypes.CurveType\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"cert\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ExecutorOperatorSetRegistered\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ExecutorOperatorSetTaskConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"config\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"curveType\",\"type\":\"uint8\",\"internalType\":\"enumIKeyRegistrarTypes.CurveType\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCreated\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"},{\"name\":\"taskDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payload\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskVerified\",\"inputs\":[{\"name\":\"aggregator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"executorCert\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CertificateVerificationFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetTaskConfigNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddressZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCurveType\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOperatorSetOwner\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskCreator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskStatus\",\"inputs\":[{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}]},{\"type\":\"error\",\"name\":\"PayloadIsEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TaskSLAIsZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TimestampAtCreation\",\"inputs\":[]}]",
}

// TaskMailboxStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use TaskMailboxStorageMetaData.ABI instead.
var TaskMailboxStorageABI = TaskMailboxStorageMetaData.ABI

// TaskMailboxStorage is an auto generated Go binding around an Ethereum contract.
type TaskMailboxStorage struct {
	TaskMailboxStorageCaller     // Read-only binding to the contract
	TaskMailboxStorageTransactor // Write-only binding to the contract
	TaskMailboxStorageFilterer   // Log filterer for contract events
}

// TaskMailboxStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type TaskMailboxStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskMailboxStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TaskMailboxStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskMailboxStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TaskMailboxStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskMailboxStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TaskMailboxStorageSession struct {
	Contract     *TaskMailboxStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// TaskMailboxStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TaskMailboxStorageCallerSession struct {
	Contract *TaskMailboxStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// TaskMailboxStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TaskMailboxStorageTransactorSession struct {
	Contract     *TaskMailboxStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// TaskMailboxStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type TaskMailboxStorageRaw struct {
	Contract *TaskMailboxStorage // Generic contract binding to access the raw methods on
}

// TaskMailboxStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TaskMailboxStorageCallerRaw struct {
	Contract *TaskMailboxStorageCaller // Generic read-only contract binding to access the raw methods on
}

// TaskMailboxStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TaskMailboxStorageTransactorRaw struct {
	Contract *TaskMailboxStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTaskMailboxStorage creates a new instance of TaskMailboxStorage, bound to a specific deployed contract.
func NewTaskMailboxStorage(address common.Address, backend bind.ContractBackend) (*TaskMailboxStorage, error) {
	contract, err := bindTaskMailboxStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorage{TaskMailboxStorageCaller: TaskMailboxStorageCaller{contract: contract}, TaskMailboxStorageTransactor: TaskMailboxStorageTransactor{contract: contract}, TaskMailboxStorageFilterer: TaskMailboxStorageFilterer{contract: contract}}, nil
}

// NewTaskMailboxStorageCaller creates a new read-only instance of TaskMailboxStorage, bound to a specific deployed contract.
func NewTaskMailboxStorageCaller(address common.Address, caller bind.ContractCaller) (*TaskMailboxStorageCaller, error) {
	contract, err := bindTaskMailboxStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageCaller{contract: contract}, nil
}

// NewTaskMailboxStorageTransactor creates a new write-only instance of TaskMailboxStorage, bound to a specific deployed contract.
func NewTaskMailboxStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*TaskMailboxStorageTransactor, error) {
	contract, err := bindTaskMailboxStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageTransactor{contract: contract}, nil
}

// NewTaskMailboxStorageFilterer creates a new log filterer instance of TaskMailboxStorage, bound to a specific deployed contract.
func NewTaskMailboxStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*TaskMailboxStorageFilterer, error) {
	contract, err := bindTaskMailboxStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageFilterer{contract: contract}, nil
}

// bindTaskMailboxStorage binds a generic wrapper to an already deployed contract.
func bindTaskMailboxStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TaskMailboxStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskMailboxStorage *TaskMailboxStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskMailboxStorage.Contract.TaskMailboxStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskMailboxStorage *TaskMailboxStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.TaskMailboxStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskMailboxStorage *TaskMailboxStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.TaskMailboxStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskMailboxStorage *TaskMailboxStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskMailboxStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskMailboxStorage *TaskMailboxStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskMailboxStorage *TaskMailboxStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.contract.Transact(opts, method, params...)
}

// BN254CERTIFICATEVERIFIER is a free data retrieval call binding the contract method 0xf7424fc9.
//
// Solidity: function BN254_CERTIFICATE_VERIFIER() view returns(address)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) BN254CERTIFICATEVERIFIER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "BN254_CERTIFICATE_VERIFIER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BN254CERTIFICATEVERIFIER is a free data retrieval call binding the contract method 0xf7424fc9.
//
// Solidity: function BN254_CERTIFICATE_VERIFIER() view returns(address)
func (_TaskMailboxStorage *TaskMailboxStorageSession) BN254CERTIFICATEVERIFIER() (common.Address, error) {
	return _TaskMailboxStorage.Contract.BN254CERTIFICATEVERIFIER(&_TaskMailboxStorage.CallOpts)
}

// BN254CERTIFICATEVERIFIER is a free data retrieval call binding the contract method 0xf7424fc9.
//
// Solidity: function BN254_CERTIFICATE_VERIFIER() view returns(address)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) BN254CERTIFICATEVERIFIER() (common.Address, error) {
	return _TaskMailboxStorage.Contract.BN254CERTIFICATEVERIFIER(&_TaskMailboxStorage.CallOpts)
}

// ECDSACERTIFICATEVERIFIER is a free data retrieval call binding the contract method 0x54743ad2.
//
// Solidity: function ECDSA_CERTIFICATE_VERIFIER() view returns(address)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) ECDSACERTIFICATEVERIFIER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "ECDSA_CERTIFICATE_VERIFIER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ECDSACERTIFICATEVERIFIER is a free data retrieval call binding the contract method 0x54743ad2.
//
// Solidity: function ECDSA_CERTIFICATE_VERIFIER() view returns(address)
func (_TaskMailboxStorage *TaskMailboxStorageSession) ECDSACERTIFICATEVERIFIER() (common.Address, error) {
	return _TaskMailboxStorage.Contract.ECDSACERTIFICATEVERIFIER(&_TaskMailboxStorage.CallOpts)
}

// ECDSACERTIFICATEVERIFIER is a free data retrieval call binding the contract method 0x54743ad2.
//
// Solidity: function ECDSA_CERTIFICATE_VERIFIER() view returns(address)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) ECDSACERTIFICATEVERIFIER() (common.Address, error) {
	return _TaskMailboxStorage.Contract.ECDSACERTIFICATEVERIFIER(&_TaskMailboxStorage.CallOpts)
}

// ExecutorOperatorSetTaskConfigs is a free data retrieval call binding the contract method 0x1c7edb17.
//
// Solidity: function executorOperatorSetTaskConfigs(bytes32 operatorSetKey) view returns(uint8 curveType, address taskHook, address feeToken, address feeCollector, uint96 taskSLA, uint16 stakeProportionThreshold, bytes taskMetadata)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) ExecutorOperatorSetTaskConfigs(opts *bind.CallOpts, operatorSetKey [32]byte) (struct {
	CurveType                uint8
	TaskHook                 common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "executorOperatorSetTaskConfigs", operatorSetKey)

	outstruct := new(struct {
		CurveType                uint8
		TaskHook                 common.Address
		FeeToken                 common.Address
		FeeCollector             common.Address
		TaskSLA                  *big.Int
		StakeProportionThreshold uint16
		TaskMetadata             []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.CurveType = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.TaskHook = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.FeeToken = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.FeeCollector = *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	outstruct.TaskSLA = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.StakeProportionThreshold = *abi.ConvertType(out[5], new(uint16)).(*uint16)
	outstruct.TaskMetadata = *abi.ConvertType(out[6], new([]byte)).(*[]byte)

	return *outstruct, err

}

// ExecutorOperatorSetTaskConfigs is a free data retrieval call binding the contract method 0x1c7edb17.
//
// Solidity: function executorOperatorSetTaskConfigs(bytes32 operatorSetKey) view returns(uint8 curveType, address taskHook, address feeToken, address feeCollector, uint96 taskSLA, uint16 stakeProportionThreshold, bytes taskMetadata)
func (_TaskMailboxStorage *TaskMailboxStorageSession) ExecutorOperatorSetTaskConfigs(operatorSetKey [32]byte) (struct {
	CurveType                uint8
	TaskHook                 common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}, error) {
	return _TaskMailboxStorage.Contract.ExecutorOperatorSetTaskConfigs(&_TaskMailboxStorage.CallOpts, operatorSetKey)
}

// ExecutorOperatorSetTaskConfigs is a free data retrieval call binding the contract method 0x1c7edb17.
//
// Solidity: function executorOperatorSetTaskConfigs(bytes32 operatorSetKey) view returns(uint8 curveType, address taskHook, address feeToken, address feeCollector, uint96 taskSLA, uint16 stakeProportionThreshold, bytes taskMetadata)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) ExecutorOperatorSetTaskConfigs(operatorSetKey [32]byte) (struct {
	CurveType                uint8
	TaskHook                 common.Address
	FeeToken                 common.Address
	FeeCollector             common.Address
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}, error) {
	return _TaskMailboxStorage.Contract.ExecutorOperatorSetTaskConfigs(&_TaskMailboxStorage.CallOpts, operatorSetKey)
}

// GetBN254CertificateBytes is a free data retrieval call binding the contract method 0x1ae370eb.
//
// Solidity: function getBN254CertificateBytes((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),(uint32,bytes,((uint256,uint256),uint256[]))[]) cert) pure returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) GetBN254CertificateBytes(opts *bind.CallOpts, cert IBN254CertificateVerifierTypesBN254Certificate) ([]byte, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "getBN254CertificateBytes", cert)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetBN254CertificateBytes is a free data retrieval call binding the contract method 0x1ae370eb.
//
// Solidity: function getBN254CertificateBytes((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),(uint32,bytes,((uint256,uint256),uint256[]))[]) cert) pure returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageSession) GetBN254CertificateBytes(cert IBN254CertificateVerifierTypesBN254Certificate) ([]byte, error) {
	return _TaskMailboxStorage.Contract.GetBN254CertificateBytes(&_TaskMailboxStorage.CallOpts, cert)
}

// GetBN254CertificateBytes is a free data retrieval call binding the contract method 0x1ae370eb.
//
// Solidity: function getBN254CertificateBytes((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),(uint32,bytes,((uint256,uint256),uint256[]))[]) cert) pure returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) GetBN254CertificateBytes(cert IBN254CertificateVerifierTypesBN254Certificate) ([]byte, error) {
	return _TaskMailboxStorage.Contract.GetBN254CertificateBytes(&_TaskMailboxStorage.CallOpts, cert)
}

// GetECDSACertificateBytes is a free data retrieval call binding the contract method 0x1270a892.
//
// Solidity: function getECDSACertificateBytes((uint32,bytes32,bytes) cert) pure returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) GetECDSACertificateBytes(opts *bind.CallOpts, cert IECDSACertificateVerifierTypesECDSACertificate) ([]byte, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "getECDSACertificateBytes", cert)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetECDSACertificateBytes is a free data retrieval call binding the contract method 0x1270a892.
//
// Solidity: function getECDSACertificateBytes((uint32,bytes32,bytes) cert) pure returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageSession) GetECDSACertificateBytes(cert IECDSACertificateVerifierTypesECDSACertificate) ([]byte, error) {
	return _TaskMailboxStorage.Contract.GetECDSACertificateBytes(&_TaskMailboxStorage.CallOpts, cert)
}

// GetECDSACertificateBytes is a free data retrieval call binding the contract method 0x1270a892.
//
// Solidity: function getECDSACertificateBytes((uint32,bytes32,bytes) cert) pure returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) GetECDSACertificateBytes(cert IECDSACertificateVerifierTypesECDSACertificate) ([]byte, error) {
	return _TaskMailboxStorage.Contract.GetECDSACertificateBytes(&_TaskMailboxStorage.CallOpts, cert)
}

// GetExecutorOperatorSetTaskConfig is a free data retrieval call binding the contract method 0x6bf6fad5.
//
// Solidity: function getExecutorOperatorSetTaskConfig((address,uint32) operatorSet) view returns((uint8,address,address,address,uint96,uint16,bytes))
func (_TaskMailboxStorage *TaskMailboxStorageCaller) GetExecutorOperatorSetTaskConfig(opts *bind.CallOpts, operatorSet OperatorSet) (ITaskMailboxTypesExecutorOperatorSetTaskConfig, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "getExecutorOperatorSetTaskConfig", operatorSet)

	if err != nil {
		return *new(ITaskMailboxTypesExecutorOperatorSetTaskConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesExecutorOperatorSetTaskConfig)).(*ITaskMailboxTypesExecutorOperatorSetTaskConfig)

	return out0, err

}

// GetExecutorOperatorSetTaskConfig is a free data retrieval call binding the contract method 0x6bf6fad5.
//
// Solidity: function getExecutorOperatorSetTaskConfig((address,uint32) operatorSet) view returns((uint8,address,address,address,uint96,uint16,bytes))
func (_TaskMailboxStorage *TaskMailboxStorageSession) GetExecutorOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesExecutorOperatorSetTaskConfig, error) {
	return _TaskMailboxStorage.Contract.GetExecutorOperatorSetTaskConfig(&_TaskMailboxStorage.CallOpts, operatorSet)
}

// GetExecutorOperatorSetTaskConfig is a free data retrieval call binding the contract method 0x6bf6fad5.
//
// Solidity: function getExecutorOperatorSetTaskConfig((address,uint32) operatorSet) view returns((uint8,address,address,address,uint96,uint16,bytes))
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) GetExecutorOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesExecutorOperatorSetTaskConfig, error) {
	return _TaskMailboxStorage.Contract.GetExecutorOperatorSetTaskConfig(&_TaskMailboxStorage.CallOpts, operatorSet)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,address,uint32,address,uint96,uint16,(uint8,address,address,address,uint96,uint16,bytes),bytes,bytes,bytes))
func (_TaskMailboxStorage *TaskMailboxStorageCaller) GetTaskInfo(opts *bind.CallOpts, taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "getTaskInfo", taskHash)

	if err != nil {
		return *new(ITaskMailboxTypesTask), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesTask)).(*ITaskMailboxTypesTask)

	return out0, err

}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,address,uint32,address,uint96,uint16,(uint8,address,address,address,uint96,uint16,bytes),bytes,bytes,bytes))
func (_TaskMailboxStorage *TaskMailboxStorageSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _TaskMailboxStorage.Contract.GetTaskInfo(&_TaskMailboxStorage.CallOpts, taskHash)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,address,uint32,address,uint96,uint16,(uint8,address,address,address,uint96,uint16,bytes),bytes,bytes,bytes))
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _TaskMailboxStorage.Contract.GetTaskInfo(&_TaskMailboxStorage.CallOpts, taskHash)
}

// GetTaskResult is a free data retrieval call binding the contract method 0x62fee037.
//
// Solidity: function getTaskResult(bytes32 taskHash) view returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) GetTaskResult(opts *bind.CallOpts, taskHash [32]byte) ([]byte, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "getTaskResult", taskHash)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetTaskResult is a free data retrieval call binding the contract method 0x62fee037.
//
// Solidity: function getTaskResult(bytes32 taskHash) view returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageSession) GetTaskResult(taskHash [32]byte) ([]byte, error) {
	return _TaskMailboxStorage.Contract.GetTaskResult(&_TaskMailboxStorage.CallOpts, taskHash)
}

// GetTaskResult is a free data retrieval call binding the contract method 0x62fee037.
//
// Solidity: function getTaskResult(bytes32 taskHash) view returns(bytes)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) GetTaskResult(taskHash [32]byte) ([]byte, error) {
	return _TaskMailboxStorage.Contract.GetTaskResult(&_TaskMailboxStorage.CallOpts, taskHash)
}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) GetTaskStatus(opts *bind.CallOpts, taskHash [32]byte) (uint8, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "getTaskStatus", taskHash)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_TaskMailboxStorage *TaskMailboxStorageSession) GetTaskStatus(taskHash [32]byte) (uint8, error) {
	return _TaskMailboxStorage.Contract.GetTaskStatus(&_TaskMailboxStorage.CallOpts, taskHash)
}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) GetTaskStatus(taskHash [32]byte) (uint8, error) {
	return _TaskMailboxStorage.Contract.GetTaskStatus(&_TaskMailboxStorage.CallOpts, taskHash)
}

// IsExecutorOperatorSetRegistered is a free data retrieval call binding the contract method 0xfa2c0b37.
//
// Solidity: function isExecutorOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool isRegistered)
func (_TaskMailboxStorage *TaskMailboxStorageCaller) IsExecutorOperatorSetRegistered(opts *bind.CallOpts, operatorSetKey [32]byte) (bool, error) {
	var out []interface{}
	err := _TaskMailboxStorage.contract.Call(opts, &out, "isExecutorOperatorSetRegistered", operatorSetKey)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExecutorOperatorSetRegistered is a free data retrieval call binding the contract method 0xfa2c0b37.
//
// Solidity: function isExecutorOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool isRegistered)
func (_TaskMailboxStorage *TaskMailboxStorageSession) IsExecutorOperatorSetRegistered(operatorSetKey [32]byte) (bool, error) {
	return _TaskMailboxStorage.Contract.IsExecutorOperatorSetRegistered(&_TaskMailboxStorage.CallOpts, operatorSetKey)
}

// IsExecutorOperatorSetRegistered is a free data retrieval call binding the contract method 0xfa2c0b37.
//
// Solidity: function isExecutorOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool isRegistered)
func (_TaskMailboxStorage *TaskMailboxStorageCallerSession) IsExecutorOperatorSetRegistered(operatorSetKey [32]byte) (bool, error) {
	return _TaskMailboxStorage.Contract.IsExecutorOperatorSetRegistered(&_TaskMailboxStorage.CallOpts, operatorSetKey)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_TaskMailboxStorage *TaskMailboxStorageTransactor) CreateTask(opts *bind.TransactOpts, taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _TaskMailboxStorage.contract.Transact(opts, "createTask", taskParams)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_TaskMailboxStorage *TaskMailboxStorageSession) CreateTask(taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.CreateTask(&_TaskMailboxStorage.TransactOpts, taskParams)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_TaskMailboxStorage *TaskMailboxStorageTransactorSession) CreateTask(taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.CreateTask(&_TaskMailboxStorage.TransactOpts, taskParams)
}

// RegisterExecutorOperatorSet is a paid mutator transaction binding the contract method 0x49acd884.
//
// Solidity: function registerExecutorOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_TaskMailboxStorage *TaskMailboxStorageTransactor) RegisterExecutorOperatorSet(opts *bind.TransactOpts, operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _TaskMailboxStorage.contract.Transact(opts, "registerExecutorOperatorSet", operatorSet, isRegistered)
}

// RegisterExecutorOperatorSet is a paid mutator transaction binding the contract method 0x49acd884.
//
// Solidity: function registerExecutorOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_TaskMailboxStorage *TaskMailboxStorageSession) RegisterExecutorOperatorSet(operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.RegisterExecutorOperatorSet(&_TaskMailboxStorage.TransactOpts, operatorSet, isRegistered)
}

// RegisterExecutorOperatorSet is a paid mutator transaction binding the contract method 0x49acd884.
//
// Solidity: function registerExecutorOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_TaskMailboxStorage *TaskMailboxStorageTransactorSession) RegisterExecutorOperatorSet(operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.RegisterExecutorOperatorSet(&_TaskMailboxStorage.TransactOpts, operatorSet, isRegistered)
}

// SetExecutorOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0xdc024c93.
//
// Solidity: function setExecutorOperatorSetTaskConfig((address,uint32) operatorSet, (uint8,address,address,address,uint96,uint16,bytes) config) returns()
func (_TaskMailboxStorage *TaskMailboxStorageTransactor) SetExecutorOperatorSetTaskConfig(opts *bind.TransactOpts, operatorSet OperatorSet, config ITaskMailboxTypesExecutorOperatorSetTaskConfig) (*types.Transaction, error) {
	return _TaskMailboxStorage.contract.Transact(opts, "setExecutorOperatorSetTaskConfig", operatorSet, config)
}

// SetExecutorOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0xdc024c93.
//
// Solidity: function setExecutorOperatorSetTaskConfig((address,uint32) operatorSet, (uint8,address,address,address,uint96,uint16,bytes) config) returns()
func (_TaskMailboxStorage *TaskMailboxStorageSession) SetExecutorOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesExecutorOperatorSetTaskConfig) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.SetExecutorOperatorSetTaskConfig(&_TaskMailboxStorage.TransactOpts, operatorSet, config)
}

// SetExecutorOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0xdc024c93.
//
// Solidity: function setExecutorOperatorSetTaskConfig((address,uint32) operatorSet, (uint8,address,address,address,uint96,uint16,bytes) config) returns()
func (_TaskMailboxStorage *TaskMailboxStorageTransactorSession) SetExecutorOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesExecutorOperatorSetTaskConfig) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.SetExecutorOperatorSetTaskConfig(&_TaskMailboxStorage.TransactOpts, operatorSet, config)
}

// SubmitResult is a paid mutator transaction binding the contract method 0xa5fabc81.
//
// Solidity: function submitResult(bytes32 taskHash, bytes cert, bytes result) returns()
func (_TaskMailboxStorage *TaskMailboxStorageTransactor) SubmitResult(opts *bind.TransactOpts, taskHash [32]byte, cert []byte, result []byte) (*types.Transaction, error) {
	return _TaskMailboxStorage.contract.Transact(opts, "submitResult", taskHash, cert, result)
}

// SubmitResult is a paid mutator transaction binding the contract method 0xa5fabc81.
//
// Solidity: function submitResult(bytes32 taskHash, bytes cert, bytes result) returns()
func (_TaskMailboxStorage *TaskMailboxStorageSession) SubmitResult(taskHash [32]byte, cert []byte, result []byte) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.SubmitResult(&_TaskMailboxStorage.TransactOpts, taskHash, cert, result)
}

// SubmitResult is a paid mutator transaction binding the contract method 0xa5fabc81.
//
// Solidity: function submitResult(bytes32 taskHash, bytes cert, bytes result) returns()
func (_TaskMailboxStorage *TaskMailboxStorageTransactorSession) SubmitResult(taskHash [32]byte, cert []byte, result []byte) (*types.Transaction, error) {
	return _TaskMailboxStorage.Contract.SubmitResult(&_TaskMailboxStorage.TransactOpts, taskHash, cert, result)
}

// TaskMailboxStorageExecutorOperatorSetRegisteredIterator is returned from FilterExecutorOperatorSetRegistered and is used to iterate over the raw logs and unpacked data for ExecutorOperatorSetRegistered events raised by the TaskMailboxStorage contract.
type TaskMailboxStorageExecutorOperatorSetRegisteredIterator struct {
	Event *TaskMailboxStorageExecutorOperatorSetRegistered // Event containing the contract specifics and raw log

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
func (it *TaskMailboxStorageExecutorOperatorSetRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailboxStorageExecutorOperatorSetRegistered)
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
		it.Event = new(TaskMailboxStorageExecutorOperatorSetRegistered)
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
func (it *TaskMailboxStorageExecutorOperatorSetRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailboxStorageExecutorOperatorSetRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailboxStorageExecutorOperatorSetRegistered represents a ExecutorOperatorSetRegistered event raised by the TaskMailboxStorage contract.
type TaskMailboxStorageExecutorOperatorSetRegistered struct {
	Caller                common.Address
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	IsRegistered          bool
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterExecutorOperatorSetRegistered is a free log retrieval operation binding the contract event 0x48b63f21a1eb9dd6880e196de6d7db3fbd0c282b74f1298dcb4cf53472298f39.
//
// Solidity: event ExecutorOperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, bool isRegistered)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) FilterExecutorOperatorSetRegistered(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, executorOperatorSetId []uint32) (*TaskMailboxStorageExecutorOperatorSetRegisteredIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var executorOperatorSetIdRule []interface{}
	for _, executorOperatorSetIdItem := range executorOperatorSetId {
		executorOperatorSetIdRule = append(executorOperatorSetIdRule, executorOperatorSetIdItem)
	}

	logs, sub, err := _TaskMailboxStorage.contract.FilterLogs(opts, "ExecutorOperatorSetRegistered", callerRule, avsRule, executorOperatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageExecutorOperatorSetRegisteredIterator{contract: _TaskMailboxStorage.contract, event: "ExecutorOperatorSetRegistered", logs: logs, sub: sub}, nil
}

// WatchExecutorOperatorSetRegistered is a free log subscription operation binding the contract event 0x48b63f21a1eb9dd6880e196de6d7db3fbd0c282b74f1298dcb4cf53472298f39.
//
// Solidity: event ExecutorOperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, bool isRegistered)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) WatchExecutorOperatorSetRegistered(opts *bind.WatchOpts, sink chan<- *TaskMailboxStorageExecutorOperatorSetRegistered, caller []common.Address, avs []common.Address, executorOperatorSetId []uint32) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var executorOperatorSetIdRule []interface{}
	for _, executorOperatorSetIdItem := range executorOperatorSetId {
		executorOperatorSetIdRule = append(executorOperatorSetIdRule, executorOperatorSetIdItem)
	}

	logs, sub, err := _TaskMailboxStorage.contract.WatchLogs(opts, "ExecutorOperatorSetRegistered", callerRule, avsRule, executorOperatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailboxStorageExecutorOperatorSetRegistered)
				if err := _TaskMailboxStorage.contract.UnpackLog(event, "ExecutorOperatorSetRegistered", log); err != nil {
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

// ParseExecutorOperatorSetRegistered is a log parse operation binding the contract event 0x48b63f21a1eb9dd6880e196de6d7db3fbd0c282b74f1298dcb4cf53472298f39.
//
// Solidity: event ExecutorOperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, bool isRegistered)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) ParseExecutorOperatorSetRegistered(log types.Log) (*TaskMailboxStorageExecutorOperatorSetRegistered, error) {
	event := new(TaskMailboxStorageExecutorOperatorSetRegistered)
	if err := _TaskMailboxStorage.contract.UnpackLog(event, "ExecutorOperatorSetRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator is returned from FilterExecutorOperatorSetTaskConfigSet and is used to iterate over the raw logs and unpacked data for ExecutorOperatorSetTaskConfigSet events raised by the TaskMailboxStorage contract.
type TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator struct {
	Event *TaskMailboxStorageExecutorOperatorSetTaskConfigSet // Event containing the contract specifics and raw log

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
func (it *TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailboxStorageExecutorOperatorSetTaskConfigSet)
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
		it.Event = new(TaskMailboxStorageExecutorOperatorSetTaskConfigSet)
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
func (it *TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailboxStorageExecutorOperatorSetTaskConfigSet represents a ExecutorOperatorSetTaskConfigSet event raised by the TaskMailboxStorage contract.
type TaskMailboxStorageExecutorOperatorSetTaskConfigSet struct {
	Caller                common.Address
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	Config                ITaskMailboxTypesExecutorOperatorSetTaskConfig
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterExecutorOperatorSetTaskConfigSet is a free log retrieval operation binding the contract event 0xd24112b0542208bfc95bf644ada32b4e8d1be6e896d62fe5ce1d756656850a4a.
//
// Solidity: event ExecutorOperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, (uint8,address,address,address,uint96,uint16,bytes) config)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) FilterExecutorOperatorSetTaskConfigSet(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, executorOperatorSetId []uint32) (*TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var executorOperatorSetIdRule []interface{}
	for _, executorOperatorSetIdItem := range executorOperatorSetId {
		executorOperatorSetIdRule = append(executorOperatorSetIdRule, executorOperatorSetIdItem)
	}

	logs, sub, err := _TaskMailboxStorage.contract.FilterLogs(opts, "ExecutorOperatorSetTaskConfigSet", callerRule, avsRule, executorOperatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageExecutorOperatorSetTaskConfigSetIterator{contract: _TaskMailboxStorage.contract, event: "ExecutorOperatorSetTaskConfigSet", logs: logs, sub: sub}, nil
}

// WatchExecutorOperatorSetTaskConfigSet is a free log subscription operation binding the contract event 0xd24112b0542208bfc95bf644ada32b4e8d1be6e896d62fe5ce1d756656850a4a.
//
// Solidity: event ExecutorOperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, (uint8,address,address,address,uint96,uint16,bytes) config)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) WatchExecutorOperatorSetTaskConfigSet(opts *bind.WatchOpts, sink chan<- *TaskMailboxStorageExecutorOperatorSetTaskConfigSet, caller []common.Address, avs []common.Address, executorOperatorSetId []uint32) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var executorOperatorSetIdRule []interface{}
	for _, executorOperatorSetIdItem := range executorOperatorSetId {
		executorOperatorSetIdRule = append(executorOperatorSetIdRule, executorOperatorSetIdItem)
	}

	logs, sub, err := _TaskMailboxStorage.contract.WatchLogs(opts, "ExecutorOperatorSetTaskConfigSet", callerRule, avsRule, executorOperatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailboxStorageExecutorOperatorSetTaskConfigSet)
				if err := _TaskMailboxStorage.contract.UnpackLog(event, "ExecutorOperatorSetTaskConfigSet", log); err != nil {
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

// ParseExecutorOperatorSetTaskConfigSet is a log parse operation binding the contract event 0xd24112b0542208bfc95bf644ada32b4e8d1be6e896d62fe5ce1d756656850a4a.
//
// Solidity: event ExecutorOperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, (uint8,address,address,address,uint96,uint16,bytes) config)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) ParseExecutorOperatorSetTaskConfigSet(log types.Log) (*TaskMailboxStorageExecutorOperatorSetTaskConfigSet, error) {
	event := new(TaskMailboxStorageExecutorOperatorSetTaskConfigSet)
	if err := _TaskMailboxStorage.contract.UnpackLog(event, "ExecutorOperatorSetTaskConfigSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailboxStorageTaskCreatedIterator is returned from FilterTaskCreated and is used to iterate over the raw logs and unpacked data for TaskCreated events raised by the TaskMailboxStorage contract.
type TaskMailboxStorageTaskCreatedIterator struct {
	Event *TaskMailboxStorageTaskCreated // Event containing the contract specifics and raw log

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
func (it *TaskMailboxStorageTaskCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailboxStorageTaskCreated)
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
		it.Event = new(TaskMailboxStorageTaskCreated)
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
func (it *TaskMailboxStorageTaskCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailboxStorageTaskCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailboxStorageTaskCreated represents a TaskCreated event raised by the TaskMailboxStorage contract.
type TaskMailboxStorageTaskCreated struct {
	Creator               common.Address
	TaskHash              [32]byte
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	RefundCollector       common.Address
	AvsFee                *big.Int
	TaskDeadline          *big.Int
	Payload               []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTaskCreated is a free log retrieval operation binding the contract event 0x4a09af06a0e08fd1c053a8b400de7833019c88066be8a2d3b3b17174a74fe317.
//
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) FilterTaskCreated(opts *bind.FilterOpts, creator []common.Address, taskHash [][32]byte, avs []common.Address) (*TaskMailboxStorageTaskCreatedIterator, error) {

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

	logs, sub, err := _TaskMailboxStorage.contract.FilterLogs(opts, "TaskCreated", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageTaskCreatedIterator{contract: _TaskMailboxStorage.contract, event: "TaskCreated", logs: logs, sub: sub}, nil
}

// WatchTaskCreated is a free log subscription operation binding the contract event 0x4a09af06a0e08fd1c053a8b400de7833019c88066be8a2d3b3b17174a74fe317.
//
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) WatchTaskCreated(opts *bind.WatchOpts, sink chan<- *TaskMailboxStorageTaskCreated, creator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _TaskMailboxStorage.contract.WatchLogs(opts, "TaskCreated", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailboxStorageTaskCreated)
				if err := _TaskMailboxStorage.contract.UnpackLog(event, "TaskCreated", log); err != nil {
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
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) ParseTaskCreated(log types.Log) (*TaskMailboxStorageTaskCreated, error) {
	event := new(TaskMailboxStorageTaskCreated)
	if err := _TaskMailboxStorage.contract.UnpackLog(event, "TaskCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskMailboxStorageTaskVerifiedIterator is returned from FilterTaskVerified and is used to iterate over the raw logs and unpacked data for TaskVerified events raised by the TaskMailboxStorage contract.
type TaskMailboxStorageTaskVerifiedIterator struct {
	Event *TaskMailboxStorageTaskVerified // Event containing the contract specifics and raw log

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
func (it *TaskMailboxStorageTaskVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskMailboxStorageTaskVerified)
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
		it.Event = new(TaskMailboxStorageTaskVerified)
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
func (it *TaskMailboxStorageTaskVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskMailboxStorageTaskVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskMailboxStorageTaskVerified represents a TaskVerified event raised by the TaskMailboxStorage contract.
type TaskMailboxStorageTaskVerified struct {
	Aggregator            common.Address
	TaskHash              [32]byte
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	ExecutorCert          []byte
	Result                []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTaskVerified is a free log retrieval operation binding the contract event 0x659f23b2e7edf490e5fd6561c5148691ed0375ed7ddd3ab1bcfcfdbec4f209a9.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, bytes executorCert, bytes result)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) FilterTaskVerified(opts *bind.FilterOpts, aggregator []common.Address, taskHash [][32]byte, avs []common.Address) (*TaskMailboxStorageTaskVerifiedIterator, error) {

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

	logs, sub, err := _TaskMailboxStorage.contract.FilterLogs(opts, "TaskVerified", aggregatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &TaskMailboxStorageTaskVerifiedIterator{contract: _TaskMailboxStorage.contract, event: "TaskVerified", logs: logs, sub: sub}, nil
}

// WatchTaskVerified is a free log subscription operation binding the contract event 0x659f23b2e7edf490e5fd6561c5148691ed0375ed7ddd3ab1bcfcfdbec4f209a9.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, bytes executorCert, bytes result)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) WatchTaskVerified(opts *bind.WatchOpts, sink chan<- *TaskMailboxStorageTaskVerified, aggregator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _TaskMailboxStorage.contract.WatchLogs(opts, "TaskVerified", aggregatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskMailboxStorageTaskVerified)
				if err := _TaskMailboxStorage.contract.UnpackLog(event, "TaskVerified", log); err != nil {
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

// ParseTaskVerified is a log parse operation binding the contract event 0x659f23b2e7edf490e5fd6561c5148691ed0375ed7ddd3ab1bcfcfdbec4f209a9.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, bytes executorCert, bytes result)
func (_TaskMailboxStorage *TaskMailboxStorageFilterer) ParseTaskVerified(log types.Log) (*TaskMailboxStorageTaskVerified, error) {
	event := new(TaskMailboxStorageTaskVerified)
	if err := _TaskMailboxStorage.contract.UnpackLog(event, "TaskVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
