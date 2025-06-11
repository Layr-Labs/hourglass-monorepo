// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package TaskAVSRegistrarBase

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

// TaskAVSRegistrarBaseMetaData contains all meta data concerning the TaskAVSRegistrarBase contract.
var TaskAVSRegistrarBaseMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"allocationManager\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"avs\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOperatorSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"keyRegistrar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIKeyRegistrar\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"_avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorDeregistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketSet\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CallerNotOperator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DataLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"KeyNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAllocationManager\",\"inputs\":[]}]",
}

// TaskAVSRegistrarBaseABI is the input ABI used to generate the binding from.
// Deprecated: Use TaskAVSRegistrarBaseMetaData.ABI instead.
var TaskAVSRegistrarBaseABI = TaskAVSRegistrarBaseMetaData.ABI

// TaskAVSRegistrarBase is an auto generated Go binding around an Ethereum contract.
type TaskAVSRegistrarBase struct {
	TaskAVSRegistrarBaseCaller     // Read-only binding to the contract
	TaskAVSRegistrarBaseTransactor // Write-only binding to the contract
	TaskAVSRegistrarBaseFilterer   // Log filterer for contract events
}

// TaskAVSRegistrarBaseCaller is an auto generated read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarBaseTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarBaseFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TaskAVSRegistrarBaseFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarBaseSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TaskAVSRegistrarBaseSession struct {
	Contract     *TaskAVSRegistrarBase // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarBaseCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TaskAVSRegistrarBaseCallerSession struct {
	Contract *TaskAVSRegistrarBaseCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// TaskAVSRegistrarBaseTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TaskAVSRegistrarBaseTransactorSession struct {
	Contract     *TaskAVSRegistrarBaseTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarBaseRaw is an auto generated low-level Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseRaw struct {
	Contract *TaskAVSRegistrarBase // Generic contract binding to access the raw methods on
}

// TaskAVSRegistrarBaseCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseCallerRaw struct {
	Contract *TaskAVSRegistrarBaseCaller // Generic read-only contract binding to access the raw methods on
}

// TaskAVSRegistrarBaseTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseTransactorRaw struct {
	Contract *TaskAVSRegistrarBaseTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTaskAVSRegistrarBase creates a new instance of TaskAVSRegistrarBase, bound to a specific deployed contract.
func NewTaskAVSRegistrarBase(address common.Address, backend bind.ContractBackend) (*TaskAVSRegistrarBase, error) {
	contract, err := bindTaskAVSRegistrarBase(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBase{TaskAVSRegistrarBaseCaller: TaskAVSRegistrarBaseCaller{contract: contract}, TaskAVSRegistrarBaseTransactor: TaskAVSRegistrarBaseTransactor{contract: contract}, TaskAVSRegistrarBaseFilterer: TaskAVSRegistrarBaseFilterer{contract: contract}}, nil
}

// NewTaskAVSRegistrarBaseCaller creates a new read-only instance of TaskAVSRegistrarBase, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseCaller(address common.Address, caller bind.ContractCaller) (*TaskAVSRegistrarBaseCaller, error) {
	contract, err := bindTaskAVSRegistrarBase(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseCaller{contract: contract}, nil
}

// NewTaskAVSRegistrarBaseTransactor creates a new write-only instance of TaskAVSRegistrarBase, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseTransactor(address common.Address, transactor bind.ContractTransactor) (*TaskAVSRegistrarBaseTransactor, error) {
	contract, err := bindTaskAVSRegistrarBase(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseTransactor{contract: contract}, nil
}

// NewTaskAVSRegistrarBaseFilterer creates a new log filterer instance of TaskAVSRegistrarBase, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseFilterer(address common.Address, filterer bind.ContractFilterer) (*TaskAVSRegistrarBaseFilterer, error) {
	contract, err := bindTaskAVSRegistrarBase(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseFilterer{contract: contract}, nil
}

// bindTaskAVSRegistrarBase binds a generic wrapper to an already deployed contract.
func bindTaskAVSRegistrarBase(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TaskAVSRegistrarBaseMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrarBase.Contract.TaskAVSRegistrarBaseCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.TaskAVSRegistrarBaseTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.TaskAVSRegistrarBaseTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrarBase.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.contract.Transact(opts, method, params...)
}

// AllocationManager is a free data retrieval call binding the contract method 0xca8aa7c7.
//
// Solidity: function allocationManager() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) AllocationManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "allocationManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AllocationManager is a free data retrieval call binding the contract method 0xca8aa7c7.
//
// Solidity: function allocationManager() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) AllocationManager() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.AllocationManager(&_TaskAVSRegistrarBase.CallOpts)
}

// AllocationManager is a free data retrieval call binding the contract method 0xca8aa7c7.
//
// Solidity: function allocationManager() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) AllocationManager() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.AllocationManager(&_TaskAVSRegistrarBase.CallOpts)
}

// Avs is a free data retrieval call binding the contract method 0xde1164bb.
//
// Solidity: function avs() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) Avs(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "avs")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Avs is a free data retrieval call binding the contract method 0xde1164bb.
//
// Solidity: function avs() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) Avs() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.Avs(&_TaskAVSRegistrarBase.CallOpts)
}

// Avs is a free data retrieval call binding the contract method 0xde1164bb.
//
// Solidity: function avs() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) Avs() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.Avs(&_TaskAVSRegistrarBase.CallOpts)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetOperatorSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetOperatorSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorSocket(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetOperatorSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorSocket(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// KeyRegistrar is a free data retrieval call binding the contract method 0x3ec45c7e.
//
// Solidity: function keyRegistrar() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) KeyRegistrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "keyRegistrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KeyRegistrar is a free data retrieval call binding the contract method 0x3ec45c7e.
//
// Solidity: function keyRegistrar() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) KeyRegistrar() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.KeyRegistrar(&_TaskAVSRegistrarBase.CallOpts)
}

// KeyRegistrar is a free data retrieval call binding the contract method 0x3ec45c7e.
//
// Solidity: function keyRegistrar() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) KeyRegistrar() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.KeyRegistrar(&_TaskAVSRegistrarBase.CallOpts)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address _avs) view returns(bool)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) SupportsAVS(opts *bind.CallOpts, _avs common.Address) (bool, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "supportsAVS", _avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address _avs) view returns(bool)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) SupportsAVS(_avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBase.Contract.SupportsAVS(&_TaskAVSRegistrarBase.CallOpts, _avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address _avs) view returns(bool)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) SupportsAVS(_avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBase.Contract.SupportsAVS(&_TaskAVSRegistrarBase.CallOpts, _avs)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactor) DeregisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.contract.Transact(opts, "deregisterOperator", operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.DeregisterOperator(&_TaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactorSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.DeregisterOperator(&_TaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.contract.Transact(opts, "registerOperator", operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.RegisterOperator(&_TaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactorSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.RegisterOperator(&_TaskAVSRegistrarBase.TransactOpts, operator, avs, operatorSetIds, data)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactor) UpdateSocket(opts *bind.TransactOpts, operator common.Address, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.contract.Transact(opts, "updateSocket", operator, socket)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) UpdateSocket(operator common.Address, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.UpdateSocket(&_TaskAVSRegistrarBase.TransactOpts, operator, socket)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactorSession) UpdateSocket(operator common.Address, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.UpdateSocket(&_TaskAVSRegistrarBase.TransactOpts, operator, socket)
}

// TaskAVSRegistrarBaseInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseInitializedIterator struct {
	Event *TaskAVSRegistrarBaseInitialized // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseInitialized)
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
		it.Event = new(TaskAVSRegistrarBaseInitialized)
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
func (it *TaskAVSRegistrarBaseInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseInitialized represents a Initialized event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterInitialized(opts *bind.FilterOpts) (*TaskAVSRegistrarBaseInitializedIterator, error) {

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseInitializedIterator{contract: _TaskAVSRegistrarBase.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseInitialized) (event.Subscription, error) {

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseInitialized)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseInitialized(log types.Log) (*TaskAVSRegistrarBaseInitialized, error) {
	event := new(TaskAVSRegistrarBaseInitialized)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseOperatorDeregisteredIterator is returned from FilterOperatorDeregistered and is used to iterate over the raw logs and unpacked data for OperatorDeregistered events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorDeregisteredIterator struct {
	Event *TaskAVSRegistrarBaseOperatorDeregistered // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseOperatorDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseOperatorDeregistered)
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
		it.Event = new(TaskAVSRegistrarBaseOperatorDeregistered)
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
func (it *TaskAVSRegistrarBaseOperatorDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseOperatorDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseOperatorDeregistered represents a OperatorDeregistered event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorDeregistered struct {
	Operator       common.Address
	OperatorSetIds []uint32
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOperatorDeregistered is a free log retrieval operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterOperatorDeregistered(opts *bind.FilterOpts, operator []common.Address) (*TaskAVSRegistrarBaseOperatorDeregisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorDeregistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseOperatorDeregisteredIterator{contract: _TaskAVSRegistrarBase.contract, event: "OperatorDeregistered", logs: logs, sub: sub}, nil
}

// WatchOperatorDeregistered is a free log subscription operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchOperatorDeregistered(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseOperatorDeregistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorDeregistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseOperatorDeregistered)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseOperatorDeregistered(log types.Log) (*TaskAVSRegistrarBaseOperatorDeregistered, error) {
	event := new(TaskAVSRegistrarBaseOperatorDeregistered)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseOperatorRegisteredIterator is returned from FilterOperatorRegistered and is used to iterate over the raw logs and unpacked data for OperatorRegistered events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorRegisteredIterator struct {
	Event *TaskAVSRegistrarBaseOperatorRegistered // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseOperatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseOperatorRegistered)
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
		it.Event = new(TaskAVSRegistrarBaseOperatorRegistered)
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
func (it *TaskAVSRegistrarBaseOperatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseOperatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseOperatorRegistered represents a OperatorRegistered event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorRegistered struct {
	Operator       common.Address
	OperatorSetIds []uint32
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOperatorRegistered is a free log retrieval operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterOperatorRegistered(opts *bind.FilterOpts, operator []common.Address) (*TaskAVSRegistrarBaseOperatorRegisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseOperatorRegisteredIterator{contract: _TaskAVSRegistrarBase.contract, event: "OperatorRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorRegistered is a free log subscription operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchOperatorRegistered(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseOperatorRegistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseOperatorRegistered)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseOperatorRegistered(log types.Log) (*TaskAVSRegistrarBaseOperatorRegistered, error) {
	event := new(TaskAVSRegistrarBaseOperatorRegistered)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseOperatorSocketSetIterator is returned from FilterOperatorSocketSet and is used to iterate over the raw logs and unpacked data for OperatorSocketSet events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorSocketSetIterator struct {
	Event *TaskAVSRegistrarBaseOperatorSocketSet // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseOperatorSocketSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseOperatorSocketSet)
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
		it.Event = new(TaskAVSRegistrarBaseOperatorSocketSet)
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
func (it *TaskAVSRegistrarBaseOperatorSocketSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseOperatorSocketSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseOperatorSocketSet represents a OperatorSocketSet event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorSocketSet struct {
	Operator common.Address
	Socket   string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketSet is a free log retrieval operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterOperatorSocketSet(opts *bind.FilterOpts, operator []common.Address) (*TaskAVSRegistrarBaseOperatorSocketSetIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorSocketSet", operatorRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseOperatorSocketSetIterator{contract: _TaskAVSRegistrarBase.contract, event: "OperatorSocketSet", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketSet is a free log subscription operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchOperatorSocketSet(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseOperatorSocketSet, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorSocketSet", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseOperatorSocketSet)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSocketSet", log); err != nil {
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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseOperatorSocketSet(log types.Log) (*TaskAVSRegistrarBaseOperatorSocketSet, error) {
	event := new(TaskAVSRegistrarBaseOperatorSocketSet)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSocketSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
