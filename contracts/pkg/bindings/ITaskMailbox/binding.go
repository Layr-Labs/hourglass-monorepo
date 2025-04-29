// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ITaskMailbox

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

// ITaskMailboxMetaData contains all meta data concerning the ITaskMailbox contract.
var ITaskMailboxMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"cancelTask\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createTask\",\"inputs\":[{\"name\":\"taskParams\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.TaskParams\",\"components\":[{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskInfo\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.Task\",\"components\":[{\"name\":\"creator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"creationTime\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"feeSplit\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"operatorSetTaskConfig\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskStatus\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperatorSet\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"OperatorSetRegistered\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSetTaskConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"config\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structITaskMailboxTypes.OperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"aggregator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCanceled\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCreated\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"},{\"name\":\"taskDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payload\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskVerified\",\"inputs\":[{\"name\":\"aggregator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"result\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CertificateVerificationFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddressZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskAggregator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskCreator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskStatus\",\"inputs\":[{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}]},{\"type\":\"error\",\"name\":\"OperatorSetNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorSetTaskConfigNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PayloadIsEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TaskSLAIsZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TimestampAtCreation\",\"inputs\":[]}]",
}

// ITaskMailboxABI is the input ABI used to generate the binding from.
// Deprecated: Use ITaskMailboxMetaData.ABI instead.
var ITaskMailboxABI = ITaskMailboxMetaData.ABI

// ITaskMailbox is an auto generated Go binding around an Ethereum contract.
type ITaskMailbox struct {
	ITaskMailboxCaller     // Read-only binding to the contract
	ITaskMailboxTransactor // Write-only binding to the contract
	ITaskMailboxFilterer   // Log filterer for contract events
}

// ITaskMailboxCaller is an auto generated read-only Go binding around an Ethereum contract.
type ITaskMailboxCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskMailboxTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ITaskMailboxTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskMailboxFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ITaskMailboxFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskMailboxSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ITaskMailboxSession struct {
	Contract     *ITaskMailbox     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ITaskMailboxCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ITaskMailboxCallerSession struct {
	Contract *ITaskMailboxCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// ITaskMailboxTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ITaskMailboxTransactorSession struct {
	Contract     *ITaskMailboxTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ITaskMailboxRaw is an auto generated low-level Go binding around an Ethereum contract.
type ITaskMailboxRaw struct {
	Contract *ITaskMailbox // Generic contract binding to access the raw methods on
}

// ITaskMailboxCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ITaskMailboxCallerRaw struct {
	Contract *ITaskMailboxCaller // Generic read-only contract binding to access the raw methods on
}

// ITaskMailboxTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ITaskMailboxTransactorRaw struct {
	Contract *ITaskMailboxTransactor // Generic write-only contract binding to access the raw methods on
}

// NewITaskMailbox creates a new instance of ITaskMailbox, bound to a specific deployed contract.
func NewITaskMailbox(address common.Address, backend bind.ContractBackend) (*ITaskMailbox, error) {
	contract, err := bindITaskMailbox(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ITaskMailbox{ITaskMailboxCaller: ITaskMailboxCaller{contract: contract}, ITaskMailboxTransactor: ITaskMailboxTransactor{contract: contract}, ITaskMailboxFilterer: ITaskMailboxFilterer{contract: contract}}, nil
}

// NewITaskMailboxCaller creates a new read-only instance of ITaskMailbox, bound to a specific deployed contract.
func NewITaskMailboxCaller(address common.Address, caller bind.ContractCaller) (*ITaskMailboxCaller, error) {
	contract, err := bindITaskMailbox(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxCaller{contract: contract}, nil
}

// NewITaskMailboxTransactor creates a new write-only instance of ITaskMailbox, bound to a specific deployed contract.
func NewITaskMailboxTransactor(address common.Address, transactor bind.ContractTransactor) (*ITaskMailboxTransactor, error) {
	contract, err := bindITaskMailbox(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxTransactor{contract: contract}, nil
}

// NewITaskMailboxFilterer creates a new log filterer instance of ITaskMailbox, bound to a specific deployed contract.
func NewITaskMailboxFilterer(address common.Address, filterer bind.ContractFilterer) (*ITaskMailboxFilterer, error) {
	contract, err := bindITaskMailbox(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxFilterer{contract: contract}, nil
}

// bindITaskMailbox binds a generic wrapper to an already deployed contract.
func bindITaskMailbox(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ITaskMailboxMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITaskMailbox *ITaskMailboxRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ITaskMailbox.Contract.ITaskMailboxCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITaskMailbox *ITaskMailboxRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.ITaskMailboxTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITaskMailbox *ITaskMailboxRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.ITaskMailboxTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITaskMailbox *ITaskMailboxCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ITaskMailbox.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITaskMailbox *ITaskMailboxTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITaskMailbox *ITaskMailboxTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.contract.Transact(opts, method, params...)
}

// GetOperatorSetTaskConfig is a free data retrieval call binding the contract method 0xca2df5eb.
//
// Solidity: function getOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,address,uint96,uint16,bytes))
func (_ITaskMailbox *ITaskMailboxCaller) GetOperatorSetTaskConfig(opts *bind.CallOpts, operatorSet OperatorSet) (ITaskMailboxTypesOperatorSetTaskConfig, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "getOperatorSetTaskConfig", operatorSet)

	if err != nil {
		return *new(ITaskMailboxTypesOperatorSetTaskConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesOperatorSetTaskConfig)).(*ITaskMailboxTypesOperatorSetTaskConfig)

	return out0, err

}

// GetOperatorSetTaskConfig is a free data retrieval call binding the contract method 0xca2df5eb.
//
// Solidity: function getOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,address,uint96,uint16,bytes))
func (_ITaskMailbox *ITaskMailboxSession) GetOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesOperatorSetTaskConfig, error) {
	return _ITaskMailbox.Contract.GetOperatorSetTaskConfig(&_ITaskMailbox.CallOpts, operatorSet)
}

// GetOperatorSetTaskConfig is a free data retrieval call binding the contract method 0xca2df5eb.
//
// Solidity: function getOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,address,uint96,uint16,bytes))
func (_ITaskMailbox *ITaskMailboxCallerSession) GetOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesOperatorSetTaskConfig, error) {
	return _ITaskMailbox.Contract.GetOperatorSetTaskConfig(&_ITaskMailbox.CallOpts, operatorSet)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,(address,uint32),address,uint96,uint16,(address,address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_ITaskMailbox *ITaskMailboxCaller) GetTaskInfo(opts *bind.CallOpts, taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "getTaskInfo", taskHash)

	if err != nil {
		return *new(ITaskMailboxTypesTask), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesTask)).(*ITaskMailboxTypesTask)

	return out0, err

}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,(address,uint32),address,uint96,uint16,(address,address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_ITaskMailbox *ITaskMailboxSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _ITaskMailbox.Contract.GetTaskInfo(&_ITaskMailbox.CallOpts, taskHash)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,(address,uint32),address,uint96,uint16,(address,address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_ITaskMailbox *ITaskMailboxCallerSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _ITaskMailbox.Contract.GetTaskInfo(&_ITaskMailbox.CallOpts, taskHash)
}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_ITaskMailbox *ITaskMailboxCaller) GetTaskStatus(opts *bind.CallOpts, taskHash [32]byte) (uint8, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "getTaskStatus", taskHash)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_ITaskMailbox *ITaskMailboxSession) GetTaskStatus(taskHash [32]byte) (uint8, error) {
	return _ITaskMailbox.Contract.GetTaskStatus(&_ITaskMailbox.CallOpts, taskHash)
}

// GetTaskStatus is a free data retrieval call binding the contract method 0x2bf6cc79.
//
// Solidity: function getTaskStatus(bytes32 taskHash) view returns(uint8)
func (_ITaskMailbox *ITaskMailboxCallerSession) GetTaskStatus(taskHash [32]byte) (uint8, error) {
	return _ITaskMailbox.Contract.GetTaskStatus(&_ITaskMailbox.CallOpts, taskHash)
}

// CancelTask is a paid mutator transaction binding the contract method 0xee8ca3b5.
//
// Solidity: function cancelTask(bytes32 taskHash) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) CancelTask(opts *bind.TransactOpts, taskHash [32]byte) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "cancelTask", taskHash)
}

// CancelTask is a paid mutator transaction binding the contract method 0xee8ca3b5.
//
// Solidity: function cancelTask(bytes32 taskHash) returns()
func (_ITaskMailbox *ITaskMailboxSession) CancelTask(taskHash [32]byte) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.CancelTask(&_ITaskMailbox.TransactOpts, taskHash)
}

// CancelTask is a paid mutator transaction binding the contract method 0xee8ca3b5.
//
// Solidity: function cancelTask(bytes32 taskHash) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) CancelTask(taskHash [32]byte) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.CancelTask(&_ITaskMailbox.TransactOpts, taskHash)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_ITaskMailbox *ITaskMailboxTransactor) CreateTask(opts *bind.TransactOpts, taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "createTask", taskParams)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_ITaskMailbox *ITaskMailboxSession) CreateTask(taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.CreateTask(&_ITaskMailbox.TransactOpts, taskParams)
}

// CreateTask is a paid mutator transaction binding the contract method 0x0443b7a0.
//
// Solidity: function createTask((address,uint96,(address,uint32),bytes) taskParams) returns(bytes32 taskHash)
func (_ITaskMailbox *ITaskMailboxTransactorSession) CreateTask(taskParams ITaskMailboxTypesTaskParams) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.CreateTask(&_ITaskMailbox.TransactOpts, taskParams)
}

// RegisterOperatorSet is a paid mutator transaction binding the contract method 0xadf87665.
//
// Solidity: function registerOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) RegisterOperatorSet(opts *bind.TransactOpts, operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "registerOperatorSet", operatorSet, isRegistered)
}

// RegisterOperatorSet is a paid mutator transaction binding the contract method 0xadf87665.
//
// Solidity: function registerOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_ITaskMailbox *ITaskMailboxSession) RegisterOperatorSet(operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.RegisterOperatorSet(&_ITaskMailbox.TransactOpts, operatorSet, isRegistered)
}

// RegisterOperatorSet is a paid mutator transaction binding the contract method 0xadf87665.
//
// Solidity: function registerOperatorSet((address,uint32) operatorSet, bool isRegistered) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) RegisterOperatorSet(operatorSet OperatorSet, isRegistered bool) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.RegisterOperatorSet(&_ITaskMailbox.TransactOpts, operatorSet, isRegistered)
}

// SetOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x9ff625d8.
//
// Solidity: function setOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,address,uint96,uint16,bytes) config) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) SetOperatorSetTaskConfig(opts *bind.TransactOpts, operatorSet OperatorSet, config ITaskMailboxTypesOperatorSetTaskConfig) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "setOperatorSetTaskConfig", operatorSet, config)
}

// SetOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x9ff625d8.
//
// Solidity: function setOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,address,uint96,uint16,bytes) config) returns()
func (_ITaskMailbox *ITaskMailboxSession) SetOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesOperatorSetTaskConfig) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SetOperatorSetTaskConfig(&_ITaskMailbox.TransactOpts, operatorSet, config)
}

// SetOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x9ff625d8.
//
// Solidity: function setOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,address,uint96,uint16,bytes) config) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) SetOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesOperatorSetTaskConfig) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SetOperatorSetTaskConfig(&_ITaskMailbox.TransactOpts, operatorSet, config)
}

// SubmitResult is a paid mutator transaction binding the contract method 0x3b433719.
//
// Solidity: function submitResult(bytes32 taskHash, (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, bytes result) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) SubmitResult(opts *bind.TransactOpts, taskHash [32]byte, cert IBN254CertificateVerifierBN254Certificate, result []byte) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "submitResult", taskHash, cert, result)
}

// SubmitResult is a paid mutator transaction binding the contract method 0x3b433719.
//
// Solidity: function submitResult(bytes32 taskHash, (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, bytes result) returns()
func (_ITaskMailbox *ITaskMailboxSession) SubmitResult(taskHash [32]byte, cert IBN254CertificateVerifierBN254Certificate, result []byte) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SubmitResult(&_ITaskMailbox.TransactOpts, taskHash, cert, result)
}

// SubmitResult is a paid mutator transaction binding the contract method 0x3b433719.
//
// Solidity: function submitResult(bytes32 taskHash, (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, bytes result) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) SubmitResult(taskHash [32]byte, cert IBN254CertificateVerifierBN254Certificate, result []byte) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SubmitResult(&_ITaskMailbox.TransactOpts, taskHash, cert, result)
}

// ITaskMailboxOperatorSetRegisteredIterator is returned from FilterOperatorSetRegistered and is used to iterate over the raw logs and unpacked data for OperatorSetRegistered events raised by the ITaskMailbox contract.
type ITaskMailboxOperatorSetRegisteredIterator struct {
	Event *ITaskMailboxOperatorSetRegistered // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxOperatorSetRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxOperatorSetRegistered)
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
		it.Event = new(ITaskMailboxOperatorSetRegistered)
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
func (it *ITaskMailboxOperatorSetRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxOperatorSetRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxOperatorSetRegistered represents a OperatorSetRegistered event raised by the ITaskMailbox contract.
type ITaskMailboxOperatorSetRegistered struct {
	Caller        common.Address
	Avs           common.Address
	OperatorSetId uint32
	IsRegistered  bool
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetRegistered is a free log retrieval operation binding the contract event 0xd5d99994c6b140b6722f9453b46413489373660d9dfd425d769dde6dd330bd58.
//
// Solidity: event OperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed operatorSetId, bool isRegistered)
func (_ITaskMailbox *ITaskMailboxFilterer) FilterOperatorSetRegistered(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, operatorSetId []uint32) (*ITaskMailboxOperatorSetRegisteredIterator, error) {

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

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "OperatorSetRegistered", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxOperatorSetRegisteredIterator{contract: _ITaskMailbox.contract, event: "OperatorSetRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorSetRegistered is a free log subscription operation binding the contract event 0xd5d99994c6b140b6722f9453b46413489373660d9dfd425d769dde6dd330bd58.
//
// Solidity: event OperatorSetRegistered(address indexed caller, address indexed avs, uint32 indexed operatorSetId, bool isRegistered)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchOperatorSetRegistered(opts *bind.WatchOpts, sink chan<- *ITaskMailboxOperatorSetRegistered, caller []common.Address, avs []common.Address, operatorSetId []uint32) (event.Subscription, error) {

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

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "OperatorSetRegistered", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxOperatorSetRegistered)
				if err := _ITaskMailbox.contract.UnpackLog(event, "OperatorSetRegistered", log); err != nil {
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
func (_ITaskMailbox *ITaskMailboxFilterer) ParseOperatorSetRegistered(log types.Log) (*ITaskMailboxOperatorSetRegistered, error) {
	event := new(ITaskMailboxOperatorSetRegistered)
	if err := _ITaskMailbox.contract.UnpackLog(event, "OperatorSetRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskMailboxOperatorSetTaskConfigSetIterator is returned from FilterOperatorSetTaskConfigSet and is used to iterate over the raw logs and unpacked data for OperatorSetTaskConfigSet events raised by the ITaskMailbox contract.
type ITaskMailboxOperatorSetTaskConfigSetIterator struct {
	Event *ITaskMailboxOperatorSetTaskConfigSet // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxOperatorSetTaskConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxOperatorSetTaskConfigSet)
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
		it.Event = new(ITaskMailboxOperatorSetTaskConfigSet)
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
func (it *ITaskMailboxOperatorSetTaskConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxOperatorSetTaskConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxOperatorSetTaskConfigSet represents a OperatorSetTaskConfigSet event raised by the ITaskMailbox contract.
type ITaskMailboxOperatorSetTaskConfigSet struct {
	Caller        common.Address
	Avs           common.Address
	OperatorSetId uint32
	Config        ITaskMailboxTypesOperatorSetTaskConfig
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetTaskConfigSet is a free log retrieval operation binding the contract event 0x6deb0e45caf1a5c5fdab7bfaaf5a8e90a6ad6f5c2c6076e4cd03ba3e4d0ae415.
//
// Solidity: event OperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed operatorSetId, (address,address,address,address,address,uint96,uint16,bytes) config)
func (_ITaskMailbox *ITaskMailboxFilterer) FilterOperatorSetTaskConfigSet(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, operatorSetId []uint32) (*ITaskMailboxOperatorSetTaskConfigSetIterator, error) {

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

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "OperatorSetTaskConfigSet", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxOperatorSetTaskConfigSetIterator{contract: _ITaskMailbox.contract, event: "OperatorSetTaskConfigSet", logs: logs, sub: sub}, nil
}

// WatchOperatorSetTaskConfigSet is a free log subscription operation binding the contract event 0x6deb0e45caf1a5c5fdab7bfaaf5a8e90a6ad6f5c2c6076e4cd03ba3e4d0ae415.
//
// Solidity: event OperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed operatorSetId, (address,address,address,address,address,uint96,uint16,bytes) config)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchOperatorSetTaskConfigSet(opts *bind.WatchOpts, sink chan<- *ITaskMailboxOperatorSetTaskConfigSet, caller []common.Address, avs []common.Address, operatorSetId []uint32) (event.Subscription, error) {

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

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "OperatorSetTaskConfigSet", callerRule, avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxOperatorSetTaskConfigSet)
				if err := _ITaskMailbox.contract.UnpackLog(event, "OperatorSetTaskConfigSet", log); err != nil {
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
func (_ITaskMailbox *ITaskMailboxFilterer) ParseOperatorSetTaskConfigSet(log types.Log) (*ITaskMailboxOperatorSetTaskConfigSet, error) {
	event := new(ITaskMailboxOperatorSetTaskConfigSet)
	if err := _ITaskMailbox.contract.UnpackLog(event, "OperatorSetTaskConfigSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskMailboxTaskCanceledIterator is returned from FilterTaskCanceled and is used to iterate over the raw logs and unpacked data for TaskCanceled events raised by the ITaskMailbox contract.
type ITaskMailboxTaskCanceledIterator struct {
	Event *ITaskMailboxTaskCanceled // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxTaskCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxTaskCanceled)
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
		it.Event = new(ITaskMailboxTaskCanceled)
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
func (it *ITaskMailboxTaskCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxTaskCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxTaskCanceled represents a TaskCanceled event raised by the ITaskMailbox contract.
type ITaskMailboxTaskCanceled struct {
	Creator       common.Address
	TaskHash      [32]byte
	Avs           common.Address
	OperatorSetId uint32
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTaskCanceled is a free log retrieval operation binding the contract event 0x3e701c33cc740e1f61ccdcafcf97e5e65a0d7f4617aed0e8ae51be092ac18a59.
//
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId)
func (_ITaskMailbox *ITaskMailboxFilterer) FilterTaskCanceled(opts *bind.FilterOpts, creator []common.Address, taskHash [][32]byte, avs []common.Address) (*ITaskMailboxTaskCanceledIterator, error) {

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

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "TaskCanceled", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxTaskCanceledIterator{contract: _ITaskMailbox.contract, event: "TaskCanceled", logs: logs, sub: sub}, nil
}

// WatchTaskCanceled is a free log subscription operation binding the contract event 0x3e701c33cc740e1f61ccdcafcf97e5e65a0d7f4617aed0e8ae51be092ac18a59.
//
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchTaskCanceled(opts *bind.WatchOpts, sink chan<- *ITaskMailboxTaskCanceled, creator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "TaskCanceled", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxTaskCanceled)
				if err := _ITaskMailbox.contract.UnpackLog(event, "TaskCanceled", log); err != nil {
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
func (_ITaskMailbox *ITaskMailboxFilterer) ParseTaskCanceled(log types.Log) (*ITaskMailboxTaskCanceled, error) {
	event := new(ITaskMailboxTaskCanceled)
	if err := _ITaskMailbox.contract.UnpackLog(event, "TaskCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskMailboxTaskCreatedIterator is returned from FilterTaskCreated and is used to iterate over the raw logs and unpacked data for TaskCreated events raised by the ITaskMailbox contract.
type ITaskMailboxTaskCreatedIterator struct {
	Event *ITaskMailboxTaskCreated // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxTaskCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxTaskCreated)
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
		it.Event = new(ITaskMailboxTaskCreated)
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
func (it *ITaskMailboxTaskCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxTaskCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxTaskCreated represents a TaskCreated event raised by the ITaskMailbox contract.
type ITaskMailboxTaskCreated struct {
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
func (_ITaskMailbox *ITaskMailboxFilterer) FilterTaskCreated(opts *bind.FilterOpts, creator []common.Address, taskHash [][32]byte, avs []common.Address) (*ITaskMailboxTaskCreatedIterator, error) {

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

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "TaskCreated", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxTaskCreatedIterator{contract: _ITaskMailbox.contract, event: "TaskCreated", logs: logs, sub: sub}, nil
}

// WatchTaskCreated is a free log subscription operation binding the contract event 0x4a09af06a0e08fd1c053a8b400de7833019c88066be8a2d3b3b17174a74fe317.
//
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchTaskCreated(opts *bind.WatchOpts, sink chan<- *ITaskMailboxTaskCreated, creator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "TaskCreated", creatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxTaskCreated)
				if err := _ITaskMailbox.contract.UnpackLog(event, "TaskCreated", log); err != nil {
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
func (_ITaskMailbox *ITaskMailboxFilterer) ParseTaskCreated(log types.Log) (*ITaskMailboxTaskCreated, error) {
	event := new(ITaskMailboxTaskCreated)
	if err := _ITaskMailbox.contract.UnpackLog(event, "TaskCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskMailboxTaskVerifiedIterator is returned from FilterTaskVerified and is used to iterate over the raw logs and unpacked data for TaskVerified events raised by the ITaskMailbox contract.
type ITaskMailboxTaskVerifiedIterator struct {
	Event *ITaskMailboxTaskVerified // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxTaskVerifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxTaskVerified)
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
		it.Event = new(ITaskMailboxTaskVerified)
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
func (it *ITaskMailboxTaskVerifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxTaskVerifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxTaskVerified represents a TaskVerified event raised by the ITaskMailbox contract.
type ITaskMailboxTaskVerified struct {
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
func (_ITaskMailbox *ITaskMailboxFilterer) FilterTaskVerified(opts *bind.FilterOpts, aggregator []common.Address, taskHash [][32]byte, avs []common.Address) (*ITaskMailboxTaskVerifiedIterator, error) {

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

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "TaskVerified", aggregatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxTaskVerifiedIterator{contract: _ITaskMailbox.contract, event: "TaskVerified", logs: logs, sub: sub}, nil
}

// WatchTaskVerified is a free log subscription operation binding the contract event 0xd7eb53a86d7419ffc42bf17e0a61b4a2a8ab7f2e62c19368cee7d8822ea9f453.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 operatorSetId, bytes result)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchTaskVerified(opts *bind.WatchOpts, sink chan<- *ITaskMailboxTaskVerified, aggregator []common.Address, taskHash [][32]byte, avs []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "TaskVerified", aggregatorRule, taskHashRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxTaskVerified)
				if err := _ITaskMailbox.contract.UnpackLog(event, "TaskVerified", log); err != nil {
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
func (_ITaskMailbox *ITaskMailboxFilterer) ParseTaskVerified(log types.Log) (*ITaskMailboxTaskVerified, error) {
	event := new(ITaskMailboxTaskVerified)
	if err := _ITaskMailbox.contract.UnpackLog(event, "TaskVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
