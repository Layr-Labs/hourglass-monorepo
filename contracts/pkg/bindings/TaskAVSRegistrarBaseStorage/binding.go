// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package TaskAVSRegistrarBaseStorage

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

// TaskAVSRegistrarBaseStorageMetaData contains all meta data concerning the TaskAVSRegistrarBaseStorage contract.
var TaskAVSRegistrarBaseStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"allocationManager\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"avs\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOperatorSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"keyRegistrar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIKeyRegistrar\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"_avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorDeregistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketSet\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CallerNotOperator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DataLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"KeyNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotAllocationManager\",\"inputs\":[]}]",
}

// TaskAVSRegistrarBaseStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use TaskAVSRegistrarBaseStorageMetaData.ABI instead.
var TaskAVSRegistrarBaseStorageABI = TaskAVSRegistrarBaseStorageMetaData.ABI

// TaskAVSRegistrarBaseStorage is an auto generated Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseStorage struct {
	TaskAVSRegistrarBaseStorageCaller     // Read-only binding to the contract
	TaskAVSRegistrarBaseStorageTransactor // Write-only binding to the contract
	TaskAVSRegistrarBaseStorageFilterer   // Log filterer for contract events
}

// TaskAVSRegistrarBaseStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarBaseStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarBaseStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TaskAVSRegistrarBaseStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarBaseStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TaskAVSRegistrarBaseStorageSession struct {
	Contract     *TaskAVSRegistrarBaseStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                // Call options to use throughout this session
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarBaseStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TaskAVSRegistrarBaseStorageCallerSession struct {
	Contract *TaskAVSRegistrarBaseStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                      // Call options to use throughout this session
}

// TaskAVSRegistrarBaseStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TaskAVSRegistrarBaseStorageTransactorSession struct {
	Contract     *TaskAVSRegistrarBaseStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                      // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarBaseStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseStorageRaw struct {
	Contract *TaskAVSRegistrarBaseStorage // Generic contract binding to access the raw methods on
}

// TaskAVSRegistrarBaseStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseStorageCallerRaw struct {
	Contract *TaskAVSRegistrarBaseStorageCaller // Generic read-only contract binding to access the raw methods on
}

// TaskAVSRegistrarBaseStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarBaseStorageTransactorRaw struct {
	Contract *TaskAVSRegistrarBaseStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTaskAVSRegistrarBaseStorage creates a new instance of TaskAVSRegistrarBaseStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseStorage(address common.Address, backend bind.ContractBackend) (*TaskAVSRegistrarBaseStorage, error) {
	contract, err := bindTaskAVSRegistrarBaseStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorage{TaskAVSRegistrarBaseStorageCaller: TaskAVSRegistrarBaseStorageCaller{contract: contract}, TaskAVSRegistrarBaseStorageTransactor: TaskAVSRegistrarBaseStorageTransactor{contract: contract}, TaskAVSRegistrarBaseStorageFilterer: TaskAVSRegistrarBaseStorageFilterer{contract: contract}}, nil
}

// NewTaskAVSRegistrarBaseStorageCaller creates a new read-only instance of TaskAVSRegistrarBaseStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseStorageCaller(address common.Address, caller bind.ContractCaller) (*TaskAVSRegistrarBaseStorageCaller, error) {
	contract, err := bindTaskAVSRegistrarBaseStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageCaller{contract: contract}, nil
}

// NewTaskAVSRegistrarBaseStorageTransactor creates a new write-only instance of TaskAVSRegistrarBaseStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*TaskAVSRegistrarBaseStorageTransactor, error) {
	contract, err := bindTaskAVSRegistrarBaseStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageTransactor{contract: contract}, nil
}

// NewTaskAVSRegistrarBaseStorageFilterer creates a new log filterer instance of TaskAVSRegistrarBaseStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarBaseStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*TaskAVSRegistrarBaseStorageFilterer, error) {
	contract, err := bindTaskAVSRegistrarBaseStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageFilterer{contract: contract}, nil
}

// bindTaskAVSRegistrarBaseStorage binds a generic wrapper to an already deployed contract.
func bindTaskAVSRegistrarBaseStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TaskAVSRegistrarBaseStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrarBaseStorage.Contract.TaskAVSRegistrarBaseStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.TaskAVSRegistrarBaseStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.TaskAVSRegistrarBaseStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrarBaseStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.contract.Transact(opts, method, params...)
}

// AllocationManager is a free data retrieval call binding the contract method 0xca8aa7c7.
//
// Solidity: function allocationManager() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) AllocationManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "allocationManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AllocationManager is a free data retrieval call binding the contract method 0xca8aa7c7.
//
// Solidity: function allocationManager() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) AllocationManager() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.AllocationManager(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// AllocationManager is a free data retrieval call binding the contract method 0xca8aa7c7.
//
// Solidity: function allocationManager() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) AllocationManager() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.AllocationManager(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// Avs is a free data retrieval call binding the contract method 0xde1164bb.
//
// Solidity: function avs() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) Avs(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "avs")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Avs is a free data retrieval call binding the contract method 0xde1164bb.
//
// Solidity: function avs() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) Avs() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.Avs(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// Avs is a free data retrieval call binding the contract method 0xde1164bb.
//
// Solidity: function avs() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) Avs() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.Avs(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetOperatorSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getOperatorSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetOperatorSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x8481931d.
//
// Solidity: function getOperatorSocket(address operator) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetOperatorSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// KeyRegistrar is a free data retrieval call binding the contract method 0x3ec45c7e.
//
// Solidity: function keyRegistrar() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) KeyRegistrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "keyRegistrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KeyRegistrar is a free data retrieval call binding the contract method 0x3ec45c7e.
//
// Solidity: function keyRegistrar() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) KeyRegistrar() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.KeyRegistrar(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// KeyRegistrar is a free data retrieval call binding the contract method 0x3ec45c7e.
//
// Solidity: function keyRegistrar() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) KeyRegistrar() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.KeyRegistrar(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address _avs) view returns(bool)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) SupportsAVS(opts *bind.CallOpts, _avs common.Address) (bool, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "supportsAVS", _avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address _avs) view returns(bool)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) SupportsAVS(_avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.SupportsAVS(&_TaskAVSRegistrarBaseStorage.CallOpts, _avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address _avs) view returns(bool)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) SupportsAVS(_avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.SupportsAVS(&_TaskAVSRegistrarBaseStorage.CallOpts, _avs)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactor) DeregisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.contract.Transact(opts, "deregisterOperator", operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.DeregisterOperator(&_TaskAVSRegistrarBaseStorage.TransactOpts, operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactorSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.DeregisterOperator(&_TaskAVSRegistrarBaseStorage.TransactOpts, operator, avs, operatorSetIds)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.contract.Transact(opts, "registerOperator", operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.RegisterOperator(&_TaskAVSRegistrarBaseStorage.TransactOpts, operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactorSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.RegisterOperator(&_TaskAVSRegistrarBaseStorage.TransactOpts, operator, avs, operatorSetIds, data)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactor) UpdateSocket(opts *bind.TransactOpts, operator common.Address, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.contract.Transact(opts, "updateSocket", operator, socket)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) UpdateSocket(operator common.Address, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.UpdateSocket(&_TaskAVSRegistrarBaseStorage.TransactOpts, operator, socket)
}

// UpdateSocket is a paid mutator transaction binding the contract method 0x6591666a.
//
// Solidity: function updateSocket(address operator, string socket) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactorSession) UpdateSocket(operator common.Address, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.UpdateSocket(&_TaskAVSRegistrarBaseStorage.TransactOpts, operator, socket)
}

// TaskAVSRegistrarBaseStorageInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageInitializedIterator struct {
	Event *TaskAVSRegistrarBaseStorageInitialized // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageInitialized)
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
		it.Event = new(TaskAVSRegistrarBaseStorageInitialized)
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
func (it *TaskAVSRegistrarBaseStorageInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageInitialized represents a Initialized event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterInitialized(opts *bind.FilterOpts) (*TaskAVSRegistrarBaseStorageInitializedIterator, error) {

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageInitializedIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageInitialized) (event.Subscription, error) {

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageInitialized)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseInitialized(log types.Log) (*TaskAVSRegistrarBaseStorageInitialized, error) {
	event := new(TaskAVSRegistrarBaseStorageInitialized)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator is returned from FilterOperatorDeregistered and is used to iterate over the raw logs and unpacked data for OperatorDeregistered events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator struct {
	Event *TaskAVSRegistrarBaseStorageOperatorDeregistered // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageOperatorDeregistered)
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
		it.Event = new(TaskAVSRegistrarBaseStorageOperatorDeregistered)
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
func (it *TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageOperatorDeregistered represents a OperatorDeregistered event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorDeregistered struct {
	Operator       common.Address
	OperatorSetIds []uint32
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOperatorDeregistered is a free log retrieval operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterOperatorDeregistered(opts *bind.FilterOpts, operator []common.Address) (*TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "OperatorDeregistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageOperatorDeregisteredIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "OperatorDeregistered", logs: logs, sub: sub}, nil
}

// WatchOperatorDeregistered is a free log subscription operation binding the contract event 0xf8aaad08ee23b49c9bb44e3bca6c7efa43442fc4281245a7f2475aa2632718d1.
//
// Solidity: event OperatorDeregistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchOperatorDeregistered(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageOperatorDeregistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "OperatorDeregistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageOperatorDeregistered)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseOperatorDeregistered(log types.Log) (*TaskAVSRegistrarBaseStorageOperatorDeregistered, error) {
	event := new(TaskAVSRegistrarBaseStorageOperatorDeregistered)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseStorageOperatorRegisteredIterator is returned from FilterOperatorRegistered and is used to iterate over the raw logs and unpacked data for OperatorRegistered events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorRegisteredIterator struct {
	Event *TaskAVSRegistrarBaseStorageOperatorRegistered // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageOperatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageOperatorRegistered)
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
		it.Event = new(TaskAVSRegistrarBaseStorageOperatorRegistered)
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
func (it *TaskAVSRegistrarBaseStorageOperatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageOperatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageOperatorRegistered represents a OperatorRegistered event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorRegistered struct {
	Operator       common.Address
	OperatorSetIds []uint32
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOperatorRegistered is a free log retrieval operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterOperatorRegistered(opts *bind.FilterOpts, operator []common.Address) (*TaskAVSRegistrarBaseStorageOperatorRegisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageOperatorRegisteredIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "OperatorRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorRegistered is a free log subscription operation binding the contract event 0x9efdc3d07eb312e06bf36ea85db02aec96817d7c7421f919027b240eaf34035d.
//
// Solidity: event OperatorRegistered(address indexed operator, uint32[] operatorSetIds)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchOperatorRegistered(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageOperatorRegistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageOperatorRegistered)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseOperatorRegistered(log types.Log) (*TaskAVSRegistrarBaseStorageOperatorRegistered, error) {
	event := new(TaskAVSRegistrarBaseStorageOperatorRegistered)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseStorageOperatorSocketSetIterator is returned from FilterOperatorSocketSet and is used to iterate over the raw logs and unpacked data for OperatorSocketSet events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorSocketSetIterator struct {
	Event *TaskAVSRegistrarBaseStorageOperatorSocketSet // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageOperatorSocketSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageOperatorSocketSet)
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
		it.Event = new(TaskAVSRegistrarBaseStorageOperatorSocketSet)
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
func (it *TaskAVSRegistrarBaseStorageOperatorSocketSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageOperatorSocketSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageOperatorSocketSet represents a OperatorSocketSet event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorSocketSet struct {
	Operator common.Address
	Socket   string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketSet is a free log retrieval operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterOperatorSocketSet(opts *bind.FilterOpts, operator []common.Address) (*TaskAVSRegistrarBaseStorageOperatorSocketSetIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "OperatorSocketSet", operatorRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageOperatorSocketSetIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "OperatorSocketSet", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketSet is a free log subscription operation binding the contract event 0x0728b43b8c8244bf835bc60bb800c6834d28d6b696427683617f8d4b0878054b.
//
// Solidity: event OperatorSocketSet(address indexed operator, string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchOperatorSocketSet(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageOperatorSocketSet, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "OperatorSocketSet", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageOperatorSocketSet)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorSocketSet", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseOperatorSocketSet(log types.Log) (*TaskAVSRegistrarBaseStorageOperatorSocketSet, error) {
	event := new(TaskAVSRegistrarBaseStorageOperatorSocketSet)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorSocketSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
