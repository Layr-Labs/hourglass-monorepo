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

// ITaskMailboxTypesAvsConfig is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesAvsConfig struct {
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
}

// ITaskMailboxTypesExecutorOperatorSetTaskConfig is an auto generated low-level Go binding around an user-defined struct.
type ITaskMailboxTypesExecutorOperatorSetTaskConfig struct {
	CertificateVerifier      common.Address
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
	AggregatorOperatorSetId       uint32
	RefundCollector               common.Address
	AvsFee                        *big.Int
	FeeSplit                      uint16
	ExecutorOperatorSetTaskConfig ITaskMailboxTypesExecutorOperatorSetTaskConfig
	Payload                       []byte
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

// ITaskMailboxMetaData contains all meta data concerning the ITaskMailbox contract.
var ITaskMailboxMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"cancelTask\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createTask\",\"inputs\":[{\"name\":\"taskParams\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.TaskParams\",\"components\":[{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"executorOperatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAvsConfig\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.AvsConfig\",\"components\":[{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExecutorOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskInfo\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.Task\",\"components\":[{\"name\":\"creator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"creationTime\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"feeSplit\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"executorOperatorSetTaskConfig\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskStatus\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isAvsRegistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isExecutorOperatorSetRegistered\",\"inputs\":[{\"name\":\"operatorSetKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerAvs\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsConfig\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.AvsConfig\",\"components\":[{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setExecutorOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AvsConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AvsRegistered\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ExecutorOperatorSetTaskConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"config\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contractIAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCanceled\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCreated\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"},{\"name\":\"taskDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payload\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskVerified\",\"inputs\":[{\"name\":\"aggregator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"result\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AvsNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CertificateVerificationFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DuplicateExecutorOperatorSetId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetTaskConfigNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddressZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAggregatorOperatorSetId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskCreator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskStatus\",\"inputs\":[{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enumITaskMailboxTypes.TaskStatus\"}]},{\"type\":\"error\",\"name\":\"PayloadIsEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TaskSLAIsZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TimestampAtCreation\",\"inputs\":[]}]",
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

// GetAvsConfig is a free data retrieval call binding the contract method 0xa401ba41.
//
// Solidity: function getAvsConfig(address avs) view returns((uint32,uint32[]))
func (_ITaskMailbox *ITaskMailboxCaller) GetAvsConfig(opts *bind.CallOpts, avs common.Address) (ITaskMailboxTypesAvsConfig, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "getAvsConfig", avs)

	if err != nil {
		return *new(ITaskMailboxTypesAvsConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesAvsConfig)).(*ITaskMailboxTypesAvsConfig)

	return out0, err

}

// GetAvsConfig is a free data retrieval call binding the contract method 0xa401ba41.
//
// Solidity: function getAvsConfig(address avs) view returns((uint32,uint32[]))
func (_ITaskMailbox *ITaskMailboxSession) GetAvsConfig(avs common.Address) (ITaskMailboxTypesAvsConfig, error) {
	return _ITaskMailbox.Contract.GetAvsConfig(&_ITaskMailbox.CallOpts, avs)
}

// GetAvsConfig is a free data retrieval call binding the contract method 0xa401ba41.
//
// Solidity: function getAvsConfig(address avs) view returns((uint32,uint32[]))
func (_ITaskMailbox *ITaskMailboxCallerSession) GetAvsConfig(avs common.Address) (ITaskMailboxTypesAvsConfig, error) {
	return _ITaskMailbox.Contract.GetAvsConfig(&_ITaskMailbox.CallOpts, avs)
}

// GetExecutorOperatorSetTaskConfig is a free data retrieval call binding the contract method 0x6bf6fad5.
//
// Solidity: function getExecutorOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,uint96,uint16,bytes))
func (_ITaskMailbox *ITaskMailboxCaller) GetExecutorOperatorSetTaskConfig(opts *bind.CallOpts, operatorSet OperatorSet) (ITaskMailboxTypesExecutorOperatorSetTaskConfig, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "getExecutorOperatorSetTaskConfig", operatorSet)

	if err != nil {
		return *new(ITaskMailboxTypesExecutorOperatorSetTaskConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskMailboxTypesExecutorOperatorSetTaskConfig)).(*ITaskMailboxTypesExecutorOperatorSetTaskConfig)

	return out0, err

}

// GetExecutorOperatorSetTaskConfig is a free data retrieval call binding the contract method 0x6bf6fad5.
//
// Solidity: function getExecutorOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,uint96,uint16,bytes))
func (_ITaskMailbox *ITaskMailboxSession) GetExecutorOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesExecutorOperatorSetTaskConfig, error) {
	return _ITaskMailbox.Contract.GetExecutorOperatorSetTaskConfig(&_ITaskMailbox.CallOpts, operatorSet)
}

// GetExecutorOperatorSetTaskConfig is a free data retrieval call binding the contract method 0x6bf6fad5.
//
// Solidity: function getExecutorOperatorSetTaskConfig((address,uint32) operatorSet) view returns((address,address,address,address,uint96,uint16,bytes))
func (_ITaskMailbox *ITaskMailboxCallerSession) GetExecutorOperatorSetTaskConfig(operatorSet OperatorSet) (ITaskMailboxTypesExecutorOperatorSetTaskConfig, error) {
	return _ITaskMailbox.Contract.GetExecutorOperatorSetTaskConfig(&_ITaskMailbox.CallOpts, operatorSet)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,address,uint32,uint32,address,uint96,uint16,(address,address,address,address,uint96,uint16,bytes),bytes,bytes))
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
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,address,uint32,uint32,address,uint96,uint16,(address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_ITaskMailbox *ITaskMailboxSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _ITaskMailbox.Contract.GetTaskInfo(&_ITaskMailbox.CallOpts, taskHash)
}

// GetTaskInfo is a free data retrieval call binding the contract method 0x4ad52e02.
//
// Solidity: function getTaskInfo(bytes32 taskHash) view returns((address,uint96,uint8,address,uint32,uint32,address,uint96,uint16,(address,address,address,address,uint96,uint16,bytes),bytes,bytes))
func (_ITaskMailbox *ITaskMailboxCallerSession) GetTaskInfo(taskHash [32]byte) (ITaskMailboxTypesTask, error) {
	return _ITaskMailbox.Contract.GetTaskInfo(&_ITaskMailbox.CallOpts, taskHash)
}

// GetTaskResult is a free data retrieval call binding the contract method 0x62fee037.
//
// Solidity: function getTaskResult(bytes32 taskHash) view returns(bytes)
func (_ITaskMailbox *ITaskMailboxCaller) GetTaskResult(opts *bind.CallOpts, taskHash [32]byte) ([]byte, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "getTaskResult", taskHash)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetTaskResult is a free data retrieval call binding the contract method 0x62fee037.
//
// Solidity: function getTaskResult(bytes32 taskHash) view returns(bytes)
func (_ITaskMailbox *ITaskMailboxSession) GetTaskResult(taskHash [32]byte) ([]byte, error) {
	return _ITaskMailbox.Contract.GetTaskResult(&_ITaskMailbox.CallOpts, taskHash)
}

// GetTaskResult is a free data retrieval call binding the contract method 0x62fee037.
//
// Solidity: function getTaskResult(bytes32 taskHash) view returns(bytes)
func (_ITaskMailbox *ITaskMailboxCallerSession) GetTaskResult(taskHash [32]byte) ([]byte, error) {
	return _ITaskMailbox.Contract.GetTaskResult(&_ITaskMailbox.CallOpts, taskHash)
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

// IsAvsRegistered is a free data retrieval call binding the contract method 0xe3d276ab.
//
// Solidity: function isAvsRegistered(address avs) view returns(bool)
func (_ITaskMailbox *ITaskMailboxCaller) IsAvsRegistered(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "isAvsRegistered", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAvsRegistered is a free data retrieval call binding the contract method 0xe3d276ab.
//
// Solidity: function isAvsRegistered(address avs) view returns(bool)
func (_ITaskMailbox *ITaskMailboxSession) IsAvsRegistered(avs common.Address) (bool, error) {
	return _ITaskMailbox.Contract.IsAvsRegistered(&_ITaskMailbox.CallOpts, avs)
}

// IsAvsRegistered is a free data retrieval call binding the contract method 0xe3d276ab.
//
// Solidity: function isAvsRegistered(address avs) view returns(bool)
func (_ITaskMailbox *ITaskMailboxCallerSession) IsAvsRegistered(avs common.Address) (bool, error) {
	return _ITaskMailbox.Contract.IsAvsRegistered(&_ITaskMailbox.CallOpts, avs)
}

// IsExecutorOperatorSetRegistered is a free data retrieval call binding the contract method 0xfa2c0b37.
//
// Solidity: function isExecutorOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool)
func (_ITaskMailbox *ITaskMailboxCaller) IsExecutorOperatorSetRegistered(opts *bind.CallOpts, operatorSetKey [32]byte) (bool, error) {
	var out []interface{}
	err := _ITaskMailbox.contract.Call(opts, &out, "isExecutorOperatorSetRegistered", operatorSetKey)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExecutorOperatorSetRegistered is a free data retrieval call binding the contract method 0xfa2c0b37.
//
// Solidity: function isExecutorOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool)
func (_ITaskMailbox *ITaskMailboxSession) IsExecutorOperatorSetRegistered(operatorSetKey [32]byte) (bool, error) {
	return _ITaskMailbox.Contract.IsExecutorOperatorSetRegistered(&_ITaskMailbox.CallOpts, operatorSetKey)
}

// IsExecutorOperatorSetRegistered is a free data retrieval call binding the contract method 0xfa2c0b37.
//
// Solidity: function isExecutorOperatorSetRegistered(bytes32 operatorSetKey) view returns(bool)
func (_ITaskMailbox *ITaskMailboxCallerSession) IsExecutorOperatorSetRegistered(operatorSetKey [32]byte) (bool, error) {
	return _ITaskMailbox.Contract.IsExecutorOperatorSetRegistered(&_ITaskMailbox.CallOpts, operatorSetKey)
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

// RegisterAvs is a paid mutator transaction binding the contract method 0xef1a14d7.
//
// Solidity: function registerAvs(address avs, bool isRegistered) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) RegisterAvs(opts *bind.TransactOpts, avs common.Address, isRegistered bool) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "registerAvs", avs, isRegistered)
}

// RegisterAvs is a paid mutator transaction binding the contract method 0xef1a14d7.
//
// Solidity: function registerAvs(address avs, bool isRegistered) returns()
func (_ITaskMailbox *ITaskMailboxSession) RegisterAvs(avs common.Address, isRegistered bool) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.RegisterAvs(&_ITaskMailbox.TransactOpts, avs, isRegistered)
}

// RegisterAvs is a paid mutator transaction binding the contract method 0xef1a14d7.
//
// Solidity: function registerAvs(address avs, bool isRegistered) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) RegisterAvs(avs common.Address, isRegistered bool) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.RegisterAvs(&_ITaskMailbox.TransactOpts, avs, isRegistered)
}

// SetAvsConfig is a paid mutator transaction binding the contract method 0x867f1267.
//
// Solidity: function setAvsConfig(address avs, (uint32,uint32[]) config) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) SetAvsConfig(opts *bind.TransactOpts, avs common.Address, config ITaskMailboxTypesAvsConfig) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "setAvsConfig", avs, config)
}

// SetAvsConfig is a paid mutator transaction binding the contract method 0x867f1267.
//
// Solidity: function setAvsConfig(address avs, (uint32,uint32[]) config) returns()
func (_ITaskMailbox *ITaskMailboxSession) SetAvsConfig(avs common.Address, config ITaskMailboxTypesAvsConfig) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SetAvsConfig(&_ITaskMailbox.TransactOpts, avs, config)
}

// SetAvsConfig is a paid mutator transaction binding the contract method 0x867f1267.
//
// Solidity: function setAvsConfig(address avs, (uint32,uint32[]) config) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) SetAvsConfig(avs common.Address, config ITaskMailboxTypesAvsConfig) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SetAvsConfig(&_ITaskMailbox.TransactOpts, avs, config)
}

// SetExecutorOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x4e138f39.
//
// Solidity: function setExecutorOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,uint96,uint16,bytes) config) returns()
func (_ITaskMailbox *ITaskMailboxTransactor) SetExecutorOperatorSetTaskConfig(opts *bind.TransactOpts, operatorSet OperatorSet, config ITaskMailboxTypesExecutorOperatorSetTaskConfig) (*types.Transaction, error) {
	return _ITaskMailbox.contract.Transact(opts, "setExecutorOperatorSetTaskConfig", operatorSet, config)
}

// SetExecutorOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x4e138f39.
//
// Solidity: function setExecutorOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,uint96,uint16,bytes) config) returns()
func (_ITaskMailbox *ITaskMailboxSession) SetExecutorOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesExecutorOperatorSetTaskConfig) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SetExecutorOperatorSetTaskConfig(&_ITaskMailbox.TransactOpts, operatorSet, config)
}

// SetExecutorOperatorSetTaskConfig is a paid mutator transaction binding the contract method 0x4e138f39.
//
// Solidity: function setExecutorOperatorSetTaskConfig((address,uint32) operatorSet, (address,address,address,address,uint96,uint16,bytes) config) returns()
func (_ITaskMailbox *ITaskMailboxTransactorSession) SetExecutorOperatorSetTaskConfig(operatorSet OperatorSet, config ITaskMailboxTypesExecutorOperatorSetTaskConfig) (*types.Transaction, error) {
	return _ITaskMailbox.Contract.SetExecutorOperatorSetTaskConfig(&_ITaskMailbox.TransactOpts, operatorSet, config)
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

// ITaskMailboxAvsConfigSetIterator is returned from FilterAvsConfigSet and is used to iterate over the raw logs and unpacked data for AvsConfigSet events raised by the ITaskMailbox contract.
type ITaskMailboxAvsConfigSetIterator struct {
	Event *ITaskMailboxAvsConfigSet // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxAvsConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxAvsConfigSet)
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
		it.Event = new(ITaskMailboxAvsConfigSet)
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
func (it *ITaskMailboxAvsConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxAvsConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxAvsConfigSet represents a AvsConfigSet event raised by the ITaskMailbox contract.
type ITaskMailboxAvsConfigSet struct {
	Caller                  common.Address
	Avs                     common.Address
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterAvsConfigSet is a free log retrieval operation binding the contract event 0xc5e4272bacf3a88a902bbb2920ed1308c295273ff00838766ed22d5e050087ca.
//
// Solidity: event AvsConfigSet(address indexed caller, address indexed avs, uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds)
func (_ITaskMailbox *ITaskMailboxFilterer) FilterAvsConfigSet(opts *bind.FilterOpts, caller []common.Address, avs []common.Address) (*ITaskMailboxAvsConfigSetIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "AvsConfigSet", callerRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxAvsConfigSetIterator{contract: _ITaskMailbox.contract, event: "AvsConfigSet", logs: logs, sub: sub}, nil
}

// WatchAvsConfigSet is a free log subscription operation binding the contract event 0xc5e4272bacf3a88a902bbb2920ed1308c295273ff00838766ed22d5e050087ca.
//
// Solidity: event AvsConfigSet(address indexed caller, address indexed avs, uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchAvsConfigSet(opts *bind.WatchOpts, sink chan<- *ITaskMailboxAvsConfigSet, caller []common.Address, avs []common.Address) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "AvsConfigSet", callerRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxAvsConfigSet)
				if err := _ITaskMailbox.contract.UnpackLog(event, "AvsConfigSet", log); err != nil {
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

// ParseAvsConfigSet is a log parse operation binding the contract event 0xc5e4272bacf3a88a902bbb2920ed1308c295273ff00838766ed22d5e050087ca.
//
// Solidity: event AvsConfigSet(address indexed caller, address indexed avs, uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds)
func (_ITaskMailbox *ITaskMailboxFilterer) ParseAvsConfigSet(log types.Log) (*ITaskMailboxAvsConfigSet, error) {
	event := new(ITaskMailboxAvsConfigSet)
	if err := _ITaskMailbox.contract.UnpackLog(event, "AvsConfigSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskMailboxAvsRegisteredIterator is returned from FilterAvsRegistered and is used to iterate over the raw logs and unpacked data for AvsRegistered events raised by the ITaskMailbox contract.
type ITaskMailboxAvsRegisteredIterator struct {
	Event *ITaskMailboxAvsRegistered // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxAvsRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxAvsRegistered)
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
		it.Event = new(ITaskMailboxAvsRegistered)
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
func (it *ITaskMailboxAvsRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxAvsRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxAvsRegistered represents a AvsRegistered event raised by the ITaskMailbox contract.
type ITaskMailboxAvsRegistered struct {
	Caller       common.Address
	Avs          common.Address
	IsRegistered bool
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterAvsRegistered is a free log retrieval operation binding the contract event 0x8157f276d267ffc7b002873c20b83d9bd091016e124bf541534269a907029562.
//
// Solidity: event AvsRegistered(address indexed caller, address indexed avs, bool isRegistered)
func (_ITaskMailbox *ITaskMailboxFilterer) FilterAvsRegistered(opts *bind.FilterOpts, caller []common.Address, avs []common.Address) (*ITaskMailboxAvsRegisteredIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "AvsRegistered", callerRule, avsRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxAvsRegisteredIterator{contract: _ITaskMailbox.contract, event: "AvsRegistered", logs: logs, sub: sub}, nil
}

// WatchAvsRegistered is a free log subscription operation binding the contract event 0x8157f276d267ffc7b002873c20b83d9bd091016e124bf541534269a907029562.
//
// Solidity: event AvsRegistered(address indexed caller, address indexed avs, bool isRegistered)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchAvsRegistered(opts *bind.WatchOpts, sink chan<- *ITaskMailboxAvsRegistered, caller []common.Address, avs []common.Address) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "AvsRegistered", callerRule, avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxAvsRegistered)
				if err := _ITaskMailbox.contract.UnpackLog(event, "AvsRegistered", log); err != nil {
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

// ParseAvsRegistered is a log parse operation binding the contract event 0x8157f276d267ffc7b002873c20b83d9bd091016e124bf541534269a907029562.
//
// Solidity: event AvsRegistered(address indexed caller, address indexed avs, bool isRegistered)
func (_ITaskMailbox *ITaskMailboxFilterer) ParseAvsRegistered(log types.Log) (*ITaskMailboxAvsRegistered, error) {
	event := new(ITaskMailboxAvsRegistered)
	if err := _ITaskMailbox.contract.UnpackLog(event, "AvsRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskMailboxExecutorOperatorSetTaskConfigSetIterator is returned from FilterExecutorOperatorSetTaskConfigSet and is used to iterate over the raw logs and unpacked data for ExecutorOperatorSetTaskConfigSet events raised by the ITaskMailbox contract.
type ITaskMailboxExecutorOperatorSetTaskConfigSetIterator struct {
	Event *ITaskMailboxExecutorOperatorSetTaskConfigSet // Event containing the contract specifics and raw log

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
func (it *ITaskMailboxExecutorOperatorSetTaskConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskMailboxExecutorOperatorSetTaskConfigSet)
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
		it.Event = new(ITaskMailboxExecutorOperatorSetTaskConfigSet)
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
func (it *ITaskMailboxExecutorOperatorSetTaskConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskMailboxExecutorOperatorSetTaskConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskMailboxExecutorOperatorSetTaskConfigSet represents a ExecutorOperatorSetTaskConfigSet event raised by the ITaskMailbox contract.
type ITaskMailboxExecutorOperatorSetTaskConfigSet struct {
	Caller                common.Address
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	Config                ITaskMailboxTypesExecutorOperatorSetTaskConfig
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterExecutorOperatorSetTaskConfigSet is a free log retrieval operation binding the contract event 0xb4758fe2b1355bebcbc78c10619457fcaa54e85fb3b994318238b92a097f5425.
//
// Solidity: event ExecutorOperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, (address,address,address,address,uint96,uint16,bytes) config)
func (_ITaskMailbox *ITaskMailboxFilterer) FilterExecutorOperatorSetTaskConfigSet(opts *bind.FilterOpts, caller []common.Address, avs []common.Address, executorOperatorSetId []uint32) (*ITaskMailboxExecutorOperatorSetTaskConfigSetIterator, error) {

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

	logs, sub, err := _ITaskMailbox.contract.FilterLogs(opts, "ExecutorOperatorSetTaskConfigSet", callerRule, avsRule, executorOperatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ITaskMailboxExecutorOperatorSetTaskConfigSetIterator{contract: _ITaskMailbox.contract, event: "ExecutorOperatorSetTaskConfigSet", logs: logs, sub: sub}, nil
}

// WatchExecutorOperatorSetTaskConfigSet is a free log subscription operation binding the contract event 0xb4758fe2b1355bebcbc78c10619457fcaa54e85fb3b994318238b92a097f5425.
//
// Solidity: event ExecutorOperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, (address,address,address,address,uint96,uint16,bytes) config)
func (_ITaskMailbox *ITaskMailboxFilterer) WatchExecutorOperatorSetTaskConfigSet(opts *bind.WatchOpts, sink chan<- *ITaskMailboxExecutorOperatorSetTaskConfigSet, caller []common.Address, avs []common.Address, executorOperatorSetId []uint32) (event.Subscription, error) {

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

	logs, sub, err := _ITaskMailbox.contract.WatchLogs(opts, "ExecutorOperatorSetTaskConfigSet", callerRule, avsRule, executorOperatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskMailboxExecutorOperatorSetTaskConfigSet)
				if err := _ITaskMailbox.contract.UnpackLog(event, "ExecutorOperatorSetTaskConfigSet", log); err != nil {
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

// ParseExecutorOperatorSetTaskConfigSet is a log parse operation binding the contract event 0xb4758fe2b1355bebcbc78c10619457fcaa54e85fb3b994318238b92a097f5425.
//
// Solidity: event ExecutorOperatorSetTaskConfigSet(address indexed caller, address indexed avs, uint32 indexed executorOperatorSetId, (address,address,address,address,uint96,uint16,bytes) config)
func (_ITaskMailbox *ITaskMailboxFilterer) ParseExecutorOperatorSetTaskConfigSet(log types.Log) (*ITaskMailboxExecutorOperatorSetTaskConfigSet, error) {
	event := new(ITaskMailboxExecutorOperatorSetTaskConfigSet)
	if err := _ITaskMailbox.contract.UnpackLog(event, "ExecutorOperatorSetTaskConfigSet", log); err != nil {
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
	Creator               common.Address
	TaskHash              [32]byte
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTaskCanceled is a free log retrieval operation binding the contract event 0x3e701c33cc740e1f61ccdcafcf97e5e65a0d7f4617aed0e8ae51be092ac18a59.
//
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId)
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
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId)
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
// Solidity: event TaskCanceled(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId)
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
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
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
// Solidity: event TaskCreated(address indexed creator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, address refundCollector, uint96 avsFee, uint256 taskDeadline, bytes payload)
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
	Aggregator            common.Address
	TaskHash              [32]byte
	Avs                   common.Address
	ExecutorOperatorSetId uint32
	Result                []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterTaskVerified is a free log retrieval operation binding the contract event 0xd7eb53a86d7419ffc42bf17e0a61b4a2a8ab7f2e62c19368cee7d8822ea9f453.
//
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, bytes result)
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
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, bytes result)
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
// Solidity: event TaskVerified(address indexed aggregator, bytes32 indexed taskHash, address indexed avs, uint32 executorOperatorSetId, bytes result)
func (_ITaskMailbox *ITaskMailboxFilterer) ParseTaskVerified(log types.Log) (*ITaskMailboxTaskVerified, error) {
	event := new(ITaskMailboxTaskVerified)
	if err := _ITaskMailbox.contract.UnpackLog(event, "TaskVerified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
