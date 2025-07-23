// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ITaskAVSRegistrarBase

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

// ITaskAVSRegistrarBaseTypesAvsConfig is an auto generated low-level Go binding around an user-defined struct.
type ITaskAVSRegistrarBaseTypesAvsConfig struct {
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
}

// ITaskAVSRegistrarBaseMetaData contains all meta data concerning the ITaskAVSRegistrarBase contract.
var ITaskAVSRegistrarBaseMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAVS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAvsConfig\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarBaseTypes.AvsConfig\",\"components\":[{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsConfig\",\"inputs\":[{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarBaseTypes.AvsConfig\",\"components\":[{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AvsConfigSet\",\"inputs\":[{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorDeregistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketSet\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CallerNotOperator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DataLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DuplicateExecutorOperatorSetId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetIdsEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAggregatorOperatorSetId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"KeyNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAllocationManager\",\"inputs\":[]}]",
}

// ITaskAVSRegistrarBaseABI is the input ABI used to generate the binding from.
// Deprecated: Use ITaskAVSRegistrarBaseMetaData.ABI instead.
var ITaskAVSRegistrarBaseABI = ITaskAVSRegistrarBaseMetaData.ABI

// ITaskAVSRegistrarBase is an auto generated Go binding around an Ethereum contract.
type ITaskAVSRegistrarBase struct {
	ITaskAVSRegistrarBaseCaller     // Read-only binding to the contract
	ITaskAVSRegistrarBaseTransactor // Write-only binding to the contract
	ITaskAVSRegistrarBaseFilterer   // Log filterer for contract events
}

// ITaskAVSRegistrarBaseCaller is an auto generated read-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarBaseCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskAVSRegistrarBaseTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarBaseTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskAVSRegistrarBaseFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ITaskAVSRegistrarBaseFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskAVSRegistrarBaseSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ITaskAVSRegistrarBaseSession struct {
	Contract     *ITaskAVSRegistrarBase // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// ITaskAVSRegistrarBaseCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ITaskAVSRegistrarBaseCallerSession struct {
	Contract *ITaskAVSRegistrarBaseCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// ITaskAVSRegistrarBaseTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ITaskAVSRegistrarBaseTransactorSession struct {
	Contract     *ITaskAVSRegistrarBaseTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// ITaskAVSRegistrarBaseRaw is an auto generated low-level Go binding around an Ethereum contract.
type ITaskAVSRegistrarBaseRaw struct {
	Contract *ITaskAVSRegistrarBase // Generic contract binding to access the raw methods on
}

// ITaskAVSRegistrarBaseCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarBaseCallerRaw struct {
	Contract *ITaskAVSRegistrarBaseCaller // Generic read-only contract binding to access the raw methods on
}

// ITaskAVSRegistrarBaseTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarBaseTransactorRaw struct {
	Contract *ITaskAVSRegistrarBaseTransactor // Generic write-only contract binding to access the raw methods on
}

// NewITaskAVSRegistrarBase creates a new instance of ITaskAVSRegistrarBase, bound to a specific deployed contract.
func NewITaskAVSRegistrarBase(address common.Address, backend bind.ContractBackend) (*ITaskAVSRegistrarBase, error) {
	contract, err := bindITaskAVSRegistrarBase(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBase{ITaskAVSRegistrarBaseCaller: ITaskAVSRegistrarBaseCaller{contract: contract}, ITaskAVSRegistrarBaseTransactor: ITaskAVSRegistrarBaseTransactor{contract: contract}, ITaskAVSRegistrarBaseFilterer: ITaskAVSRegistrarBaseFilterer{contract: contract}}, nil
}

// NewITaskAVSRegistrarBaseCaller creates a new read-only instance of ITaskAVSRegistrarBase, bound to a specific deployed contract.
func NewITaskAVSRegistrarBaseCaller(address common.Address, caller bind.ContractCaller) (*ITaskAVSRegistrarBaseCaller, error) {
	contract, err := bindITaskAVSRegistrarBase(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseCaller{contract: contract}, nil
}

// NewITaskAVSRegistrarBaseTransactor creates a new write-only instance of ITaskAVSRegistrarBase, bound to a specific deployed contract.
func NewITaskAVSRegistrarBaseTransactor(address common.Address, transactor bind.ContractTransactor) (*ITaskAVSRegistrarBaseTransactor, error) {
	contract, err := bindITaskAVSRegistrarBase(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseTransactor{contract: contract}, nil
}

// NewITaskAVSRegistrarBaseFilterer creates a new log filterer instance of ITaskAVSRegistrarBase, bound to a specific deployed contract.
func NewITaskAVSRegistrarBaseFilterer(address common.Address, filterer bind.ContractFilterer) (*ITaskAVSRegistrarBaseFilterer, error) {
	contract, err := bindITaskAVSRegistrarBase(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseFilterer{contract: contract}, nil
}

// bindITaskAVSRegistrarBase binds a generic wrapper to an already deployed contract.
func bindITaskAVSRegistrarBase(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ITaskAVSRegistrarBaseMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ITaskAVSRegistrarBase.Contract.ITaskAVSRegistrarBaseCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.ITaskAVSRegistrarBaseTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.ITaskAVSRegistrarBaseTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ITaskAVSRegistrarBase.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.contract.Transact(opts, method, params...)
}

// GetAVS is a free data retrieval call binding the contract method 0xf62b9a54.
//
// Solidity: function getAVS() view returns(address)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCaller) GetAVS(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ITaskAVSRegistrarBase.contract.Call(opts, &out, "getAVS")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAVS is a free data retrieval call binding the contract method 0xf62b9a54.
//
// Solidity: function getAVS() view returns(address)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) GetAVS() (common.Address, error) {
	return _ITaskAVSRegistrarBase.Contract.GetAVS(&_ITaskAVSRegistrarBase.CallOpts)
}

// GetAVS is a free data retrieval call binding the contract method 0xf62b9a54.
//
// Solidity: function getAVS() view returns(address)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCallerSession) GetAVS() (common.Address, error) {
	return _ITaskAVSRegistrarBase.Contract.GetAVS(&_ITaskAVSRegistrarBase.CallOpts)
}

// GetAvsConfig is a free data retrieval call binding the contract method 0x41f548f0.
//
// Solidity: function getAvsConfig() view returns((uint32,uint32[]))
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCaller) GetAvsConfig(opts *bind.CallOpts) (ITaskAVSRegistrarBaseTypesAvsConfig, error) {
	var out []interface{}
	err := _ITaskAVSRegistrarBase.contract.Call(opts, &out, "getAvsConfig")

	if err != nil {
		return *new(ITaskAVSRegistrarBaseTypesAvsConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskAVSRegistrarBaseTypesAvsConfig)).(*ITaskAVSRegistrarBaseTypesAvsConfig)

	return out0, err

}

// GetAvsConfig is a free data retrieval call binding the contract method 0x41f548f0.
//
// Solidity: function getAvsConfig() view returns((uint32,uint32[]))
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) GetAvsConfig() (ITaskAVSRegistrarBaseTypesAvsConfig, error) {
	return _ITaskAVSRegistrarBase.Contract.GetAvsConfig(&_ITaskAVSRegistrarBase.CallOpts)
}

// GetAvsConfig is a free data retrieval call binding the contract method 0x41f548f0.
//
// Solidity: function getAvsConfig() view returns((uint32,uint32[]))
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCallerSession) GetAvsConfig() (ITaskAVSRegistrarBaseTypesAvsConfig, error) {
	return _ITaskAVSRegistrarBase.Contract.GetAvsConfig(&_ITaskAVSRegistrarBase.CallOpts)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCaller) GetOperatorSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _ITaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) GetOperatorSocket(operator common.Address) (string, error) {
	return _ITaskAVSRegistrarBase.Contract.GetOperatorSocket(&_ITaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCallerSession) GetOperatorSocket(operator common.Address) (string, error) {
	return _ITaskAVSRegistrarBase.Contract.GetOperatorSocket(&_ITaskAVSRegistrarBase.CallOpts, operator)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCaller) SupportsAVS(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _ITaskAVSRegistrarBase.contract.Call(opts, &out, "supportsAVS", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) SupportsAVS(avs common.Address) (bool, error) {
	return _ITaskAVSRegistrarBase.Contract.SupportsAVS(&_ITaskAVSRegistrarBase.CallOpts, avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseCallerSession) SupportsAVS(avs common.Address) (bool, error) {
	return _ITaskAVSRegistrarBase.Contract.SupportsAVS(&_ITaskAVSRegistrarBase.CallOpts, avs)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactor) DeregisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.contract.Transact(opts, "deregisterOperator", operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.DeregisterOperator(&_ITaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactorSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.DeregisterOperator(&_ITaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.contract.Transact(opts, "registerOperator", operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.RegisterOperator(&_ITaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactorSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.RegisterOperator(&_ITaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds, data)
}

// SetAvsConfig is a paid mutator transaction binding the contract method 0xd1f2e81d.
//
// Solidity: function setAvsConfig((uint32,uint32[]) config) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactor) SetAvsConfig(opts *bind.TransactOpts, config ITaskAVSRegistrarBaseTypesAvsConfig) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.contract.Transact(opts, "setAvsConfig", config)
}

// SetAvsConfig is a paid mutator transaction binding the contract method 0xd1f2e81d.
//
// Solidity: function setAvsConfig((uint32,uint32[]) config) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) SetAvsConfig(config ITaskAVSRegistrarBaseTypesAvsConfig) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.SetAvsConfig(&_ITaskAVSRegistrarBase.TransactOpts, config)
}

// SetAvsConfig is a paid mutator transaction binding the contract method 0xd1f2e81d.
//
// Solidity: function setAvsConfig((uint32,uint32[]) config) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactorSession) SetAvsConfig(config ITaskAVSRegistrarBaseTypesAvsConfig) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.SetAvsConfig(&_ITaskAVSRegistrarBase.TransactOpts, config)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactor) UpdateSocket(opts *bind.TransactOpts, operator common.Address, socket string) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.contract.Transact(opts, "updateSocket", operator, socket)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseSession) UpdateSocket(operator common.Address, socket string) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.UpdateSocket(&_ITaskAVSRegistrarBase.TransactOpts, operator, socket)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseTransactorSession) UpdateSocket(operator common.Address, socket string) (*types.Transaction, error) {
	return _ITaskAVSRegistrarBase.Contract.UpdateSocket(&_ITaskAVSRegistrarBase.TransactOpts, operator, socket)
}

// ITaskAVSRegistrarBaseAvsConfigSetIterator is returned from FilterAvsConfigSet and is used to iterate over the raw logs and unpacked data for AvsConfigSet events raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseAvsConfigSetIterator struct {
	Event *ITaskAVSRegistrarBaseAvsConfigSet // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarBaseAvsConfigSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarBaseAvsConfigSet)
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
		it.Event = new(ITaskAVSRegistrarBaseAvsConfigSet)
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
func (it *ITaskAVSRegistrarBaseAvsConfigSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarBaseAvsConfigSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarBaseAvsConfigSet represents a AvsConfigSet event raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseAvsConfigSet struct {
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterAvsConfigSet is a free log retrieval operation binding the contract event 0x836f1d33f6d85cfc7b24565d309c6e1486cf56dd3d8267a9651e05b88342ef51.
//
// Solidity: event AvsConfigSet(uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) FilterAvsConfigSet(opts *bind.FilterOpts) (*ITaskAVSRegistrarBaseAvsConfigSetIterator, error) {

	logs, sub, err := _ITaskAVSRegistrarBase.contract.FilterLogs(opts, "AvsConfigSet")
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseAvsConfigSetIterator{contract: _ITaskAVSRegistrarBase.contract, event: "AvsConfigSet", logs: logs, sub: sub}, nil
}

// WatchAvsConfigSet is a free log subscription operation binding the contract event 0x836f1d33f6d85cfc7b24565d309c6e1486cf56dd3d8267a9651e05b88342ef51.
//
// Solidity: event AvsConfigSet(uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) WatchAvsConfigSet(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarBaseAvsConfigSet) (event.Subscription, error) {

	logs, sub, err := _ITaskAVSRegistrarBase.contract.WatchLogs(opts, "AvsConfigSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarBaseAvsConfigSet)
				if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "AvsConfigSet", log); err != nil {
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

// ParseAvsConfigSet is a log parse operation binding the contract event 0x836f1d33f6d85cfc7b24565d309c6e1486cf56dd3d8267a9651e05b88342ef51.
//
// Solidity: event AvsConfigSet(uint32 aggregatorOperatorSetId, uint32[] executorOperatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) ParseAvsConfigSet(log types.Log) (*ITaskAVSRegistrarBaseAvsConfigSet, error) {
	event := new(ITaskAVSRegistrarBaseAvsConfigSet)
	if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "AvsConfigSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskAVSRegistrarBaseOperatorDeregisteredIterator is returned from FilterOperatorDeregistered and is used to iterate over the raw logs and unpacked data for OperatorDeregistered events raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseOperatorDeregisteredIterator struct {
	Event *ITaskAVSRegistrarBaseOperatorDeregistered // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarBaseOperatorDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarBaseOperatorDeregistered)
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
		it.Event = new(ITaskAVSRegistrarBaseOperatorDeregistered)
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
func (it *ITaskAVSRegistrarBaseOperatorDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarBaseOperatorDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarBaseOperatorDeregistered represents a OperatorDeregistered event raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseOperatorDeregistered struct {
	Operator       common.Address
	OperatorSetIds []uint32
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOperatorDeregistered is a free log retrieval operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) FilterOperatorDeregistered(opts *bind.FilterOpts, operator []common.Address) (*ITaskAVSRegistrarBaseOperatorDeregisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ITaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorDeregistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseOperatorDeregisteredIterator{contract: _ITaskAVSRegistrarBase.contract, event: "OperatorDeregistered", logs: logs, sub: sub}, nil
}

// WatchOperatorDeregistered is a free log subscription operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) WatchOperatorDeregistered(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarBaseOperatorDeregistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ITaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorDeregistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarBaseOperatorDeregistered)
				if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
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

// ParseOperatorDeregistered is a log parse operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) ParseOperatorDeregistered(log types.Log) (*ITaskAVSRegistrarBaseOperatorDeregistered, error) {
	event := new(ITaskAVSRegistrarBaseOperatorDeregistered)
	if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskAVSRegistrarBaseOperatorRegisteredIterator is returned from FilterOperatorRegistered and is used to iterate over the raw logs and unpacked data for OperatorRegistered events raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseOperatorRegisteredIterator struct {
	Event *ITaskAVSRegistrarBaseOperatorRegistered // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarBaseOperatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarBaseOperatorRegistered)
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
		it.Event = new(ITaskAVSRegistrarBaseOperatorRegistered)
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
func (it *ITaskAVSRegistrarBaseOperatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarBaseOperatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarBaseOperatorRegistered represents a OperatorRegistered event raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseOperatorRegistered struct {
	Operator       common.Address
	OperatorSetIds []uint32
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOperatorRegistered is a free log retrieval operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) FilterOperatorRegistered(opts *bind.FilterOpts, operator []common.Address) (*ITaskAVSRegistrarBaseOperatorRegisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ITaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseOperatorRegisteredIterator{contract: _ITaskAVSRegistrarBase.contract, event: "OperatorRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorRegistered is a free log subscription operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) WatchOperatorRegistered(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarBaseOperatorRegistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ITaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarBaseOperatorRegistered)
				if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
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

// ParseOperatorRegistered is a log parse operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) ParseOperatorRegistered(log types.Log) (*ITaskAVSRegistrarBaseOperatorRegistered, error) {
	event := new(ITaskAVSRegistrarBaseOperatorRegistered)
	if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskAVSRegistrarBaseOperatorSocketSetIterator is returned from FilterOperatorSocketSet and is used to iterate over the raw logs and unpacked data for OperatorSocketSet events raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseOperatorSocketSetIterator struct {
	Event *ITaskAVSRegistrarBaseOperatorSocketSet // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarBaseOperatorSocketSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarBaseOperatorSocketSet)
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
		it.Event = new(ITaskAVSRegistrarBaseOperatorSocketSet)
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
func (it *ITaskAVSRegistrarBaseOperatorSocketSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarBaseOperatorSocketSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarBaseOperatorSocketSet represents a OperatorSocketSet event raised by the ITaskAVSRegistrarBase contract.
type ITaskAVSRegistrarBaseOperatorSocketSet struct {
	Operator common.Address
	Socket   string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketSet is a free log retrieval operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) FilterOperatorSocketSet(opts *bind.FilterOpts, operator []common.Address) (*ITaskAVSRegistrarBaseOperatorSocketSetIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ITaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorSocketSet", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarBaseOperatorSocketSetIterator{contract: _ITaskAVSRegistrarBase.contract, event: "OperatorSocketSet", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketSet is a free log subscription operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) WatchOperatorSocketSet(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarBaseOperatorSocketSet, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ITaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorSocketSet", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarBaseOperatorSocketSet)
				if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSocketSet", log); err != nil {
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

// ParseOperatorSocketSet is a log parse operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_ITaskAVSRegistrarBase *ITaskAVSRegistrarBaseFilterer) ParseOperatorSocketSet(log types.Log) (*ITaskAVSRegistrarBaseOperatorSocketSet, error) {
	event := new(ITaskAVSRegistrarBaseOperatorSocketSet)
	if err := _ITaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSocketSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
