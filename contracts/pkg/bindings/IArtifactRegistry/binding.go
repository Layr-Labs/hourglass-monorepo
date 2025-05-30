// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package IArtifactRegistry

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

// IArtifactRegistryArtifact is an auto generated low-level Go binding around an user-defined struct.
type IArtifactRegistryArtifact struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Distro       uint8
	Digest       []byte
	RegistryUrl  []byte
}

// IArtifactRegistryArtifactReleases is an auto generated low-level Go binding around an user-defined struct.
type IArtifactRegistryArtifactReleases struct {
	Artifacts []IArtifactRegistryArtifact
}

// IArtifactRegistryMetaData contains all meta data concerning the IArtifactRegistry contract.
var IArtifactRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"deregister\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIArtifactRegistry.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.ArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Architecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.OperatingSystem\"},{\"name\":\"distro\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Distribution\"},{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"listArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIArtifactRegistry.ArtifactReleases[]\",\"components\":[{\"name\":\"artifacts\",\"type\":\"tuple[]\",\"internalType\":\"structIArtifactRegistry.Artifact[]\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.ArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Architecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.OperatingSystem\"},{\"name\":\"distro\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Distribution\"},{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"publishArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"newArtifact\",\"type\":\"tuple\",\"internalType\":\"structIArtifactRegistry.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.ArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Architecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.OperatingSystem\"},{\"name\":\"distro\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Distribution\"},{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"register\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"DeregisteredAvs\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PublishedArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes\",\"indexed\":true,\"internalType\":\"bytes\"},{\"name\":\"newArtifact\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIArtifactRegistry.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.ArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Architecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.OperatingSystem\"},{\"name\":\"distro\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Distribution\"},{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"previousArtifact\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structIArtifactRegistry.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.ArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Architecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.OperatingSystem\"},{\"name\":\"distro\",\"type\":\"uint8\",\"internalType\":\"enumIArtifactRegistry.Distribution\"},{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegisteredAvs\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false}]",
}

// IArtifactRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use IArtifactRegistryMetaData.ABI instead.
var IArtifactRegistryABI = IArtifactRegistryMetaData.ABI

// IArtifactRegistry is an auto generated Go binding around an Ethereum contract.
type IArtifactRegistry struct {
	IArtifactRegistryCaller     // Read-only binding to the contract
	IArtifactRegistryTransactor // Write-only binding to the contract
	IArtifactRegistryFilterer   // Log filterer for contract events
}

// IArtifactRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type IArtifactRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IArtifactRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IArtifactRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IArtifactRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IArtifactRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IArtifactRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IArtifactRegistrySession struct {
	Contract     *IArtifactRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IArtifactRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IArtifactRegistryCallerSession struct {
	Contract *IArtifactRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// IArtifactRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IArtifactRegistryTransactorSession struct {
	Contract     *IArtifactRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// IArtifactRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type IArtifactRegistryRaw struct {
	Contract *IArtifactRegistry // Generic contract binding to access the raw methods on
}

// IArtifactRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IArtifactRegistryCallerRaw struct {
	Contract *IArtifactRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// IArtifactRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IArtifactRegistryTransactorRaw struct {
	Contract *IArtifactRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIArtifactRegistry creates a new instance of IArtifactRegistry, bound to a specific deployed contract.
func NewIArtifactRegistry(address common.Address, backend bind.ContractBackend) (*IArtifactRegistry, error) {
	contract, err := bindIArtifactRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistry{IArtifactRegistryCaller: IArtifactRegistryCaller{contract: contract}, IArtifactRegistryTransactor: IArtifactRegistryTransactor{contract: contract}, IArtifactRegistryFilterer: IArtifactRegistryFilterer{contract: contract}}, nil
}

// NewIArtifactRegistryCaller creates a new read-only instance of IArtifactRegistry, bound to a specific deployed contract.
func NewIArtifactRegistryCaller(address common.Address, caller bind.ContractCaller) (*IArtifactRegistryCaller, error) {
	contract, err := bindIArtifactRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistryCaller{contract: contract}, nil
}

// NewIArtifactRegistryTransactor creates a new write-only instance of IArtifactRegistry, bound to a specific deployed contract.
func NewIArtifactRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*IArtifactRegistryTransactor, error) {
	contract, err := bindIArtifactRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistryTransactor{contract: contract}, nil
}

// NewIArtifactRegistryFilterer creates a new log filterer instance of IArtifactRegistry, bound to a specific deployed contract.
func NewIArtifactRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*IArtifactRegistryFilterer, error) {
	contract, err := bindIArtifactRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistryFilterer{contract: contract}, nil
}

// bindIArtifactRegistry binds a generic wrapper to an already deployed contract.
func bindIArtifactRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IArtifactRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IArtifactRegistry *IArtifactRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IArtifactRegistry.Contract.IArtifactRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IArtifactRegistry *IArtifactRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.IArtifactRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IArtifactRegistry *IArtifactRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.IArtifactRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IArtifactRegistry *IArtifactRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IArtifactRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IArtifactRegistry *IArtifactRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IArtifactRegistry *IArtifactRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetArtifact is a free data retrieval call binding the contract method 0x2e9ef963.
//
// Solidity: function getArtifact(address avs, bytes operatorSetId) view returns((uint8,uint8,uint8,uint8,bytes,bytes))
func (_IArtifactRegistry *IArtifactRegistryCaller) GetArtifact(opts *bind.CallOpts, avs common.Address, operatorSetId []byte) (IArtifactRegistryArtifact, error) {
	var out []interface{}
	err := _IArtifactRegistry.contract.Call(opts, &out, "getArtifact", avs, operatorSetId)

	if err != nil {
		return *new(IArtifactRegistryArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IArtifactRegistryArtifact)).(*IArtifactRegistryArtifact)

	return out0, err

}

// GetArtifact is a free data retrieval call binding the contract method 0x2e9ef963.
//
// Solidity: function getArtifact(address avs, bytes operatorSetId) view returns((uint8,uint8,uint8,uint8,bytes,bytes))
func (_IArtifactRegistry *IArtifactRegistrySession) GetArtifact(avs common.Address, operatorSetId []byte) (IArtifactRegistryArtifact, error) {
	return _IArtifactRegistry.Contract.GetArtifact(&_IArtifactRegistry.CallOpts, avs, operatorSetId)
}

// GetArtifact is a free data retrieval call binding the contract method 0x2e9ef963.
//
// Solidity: function getArtifact(address avs, bytes operatorSetId) view returns((uint8,uint8,uint8,uint8,bytes,bytes))
func (_IArtifactRegistry *IArtifactRegistryCallerSession) GetArtifact(avs common.Address, operatorSetId []byte) (IArtifactRegistryArtifact, error) {
	return _IArtifactRegistry.Contract.GetArtifact(&_IArtifactRegistry.CallOpts, avs, operatorSetId)
}

// ListArtifacts is a free data retrieval call binding the contract method 0xf18d1677.
//
// Solidity: function listArtifacts(address avs) view returns(((uint8,uint8,uint8,uint8,bytes,bytes)[])[])
func (_IArtifactRegistry *IArtifactRegistryCaller) ListArtifacts(opts *bind.CallOpts, avs common.Address) ([]IArtifactRegistryArtifactReleases, error) {
	var out []interface{}
	err := _IArtifactRegistry.contract.Call(opts, &out, "listArtifacts", avs)

	if err != nil {
		return *new([]IArtifactRegistryArtifactReleases), err
	}

	out0 := *abi.ConvertType(out[0], new([]IArtifactRegistryArtifactReleases)).(*[]IArtifactRegistryArtifactReleases)

	return out0, err

}

// ListArtifacts is a free data retrieval call binding the contract method 0xf18d1677.
//
// Solidity: function listArtifacts(address avs) view returns(((uint8,uint8,uint8,uint8,bytes,bytes)[])[])
func (_IArtifactRegistry *IArtifactRegistrySession) ListArtifacts(avs common.Address) ([]IArtifactRegistryArtifactReleases, error) {
	return _IArtifactRegistry.Contract.ListArtifacts(&_IArtifactRegistry.CallOpts, avs)
}

// ListArtifacts is a free data retrieval call binding the contract method 0xf18d1677.
//
// Solidity: function listArtifacts(address avs) view returns(((uint8,uint8,uint8,uint8,bytes,bytes)[])[])
func (_IArtifactRegistry *IArtifactRegistryCallerSession) ListArtifacts(avs common.Address) ([]IArtifactRegistryArtifactReleases, error) {
	return _IArtifactRegistry.Contract.ListArtifacts(&_IArtifactRegistry.CallOpts, avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_IArtifactRegistry *IArtifactRegistryTransactor) Deregister(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _IArtifactRegistry.contract.Transact(opts, "deregister", avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_IArtifactRegistry *IArtifactRegistrySession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.Deregister(&_IArtifactRegistry.TransactOpts, avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_IArtifactRegistry *IArtifactRegistryTransactorSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.Deregister(&_IArtifactRegistry.TransactOpts, avs)
}

// PublishArtifact is a paid mutator transaction binding the contract method 0xbcc6efd0.
//
// Solidity: function publishArtifact(address avs, bytes operatorSetId, (uint8,uint8,uint8,uint8,bytes,bytes) newArtifact) returns()
func (_IArtifactRegistry *IArtifactRegistryTransactor) PublishArtifact(opts *bind.TransactOpts, avs common.Address, operatorSetId []byte, newArtifact IArtifactRegistryArtifact) (*types.Transaction, error) {
	return _IArtifactRegistry.contract.Transact(opts, "publishArtifact", avs, operatorSetId, newArtifact)
}

// PublishArtifact is a paid mutator transaction binding the contract method 0xbcc6efd0.
//
// Solidity: function publishArtifact(address avs, bytes operatorSetId, (uint8,uint8,uint8,uint8,bytes,bytes) newArtifact) returns()
func (_IArtifactRegistry *IArtifactRegistrySession) PublishArtifact(avs common.Address, operatorSetId []byte, newArtifact IArtifactRegistryArtifact) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.PublishArtifact(&_IArtifactRegistry.TransactOpts, avs, operatorSetId, newArtifact)
}

// PublishArtifact is a paid mutator transaction binding the contract method 0xbcc6efd0.
//
// Solidity: function publishArtifact(address avs, bytes operatorSetId, (uint8,uint8,uint8,uint8,bytes,bytes) newArtifact) returns()
func (_IArtifactRegistry *IArtifactRegistryTransactorSession) PublishArtifact(avs common.Address, operatorSetId []byte, newArtifact IArtifactRegistryArtifact) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.PublishArtifact(&_IArtifactRegistry.TransactOpts, avs, operatorSetId, newArtifact)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_IArtifactRegistry *IArtifactRegistryTransactor) Register(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _IArtifactRegistry.contract.Transact(opts, "register", avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_IArtifactRegistry *IArtifactRegistrySession) Register(avs common.Address) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.Register(&_IArtifactRegistry.TransactOpts, avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_IArtifactRegistry *IArtifactRegistryTransactorSession) Register(avs common.Address) (*types.Transaction, error) {
	return _IArtifactRegistry.Contract.Register(&_IArtifactRegistry.TransactOpts, avs)
}

// IArtifactRegistryDeregisteredAvsIterator is returned from FilterDeregisteredAvs and is used to iterate over the raw logs and unpacked data for DeregisteredAvs events raised by the IArtifactRegistry contract.
type IArtifactRegistryDeregisteredAvsIterator struct {
	Event *IArtifactRegistryDeregisteredAvs // Event containing the contract specifics and raw log

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
func (it *IArtifactRegistryDeregisteredAvsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IArtifactRegistryDeregisteredAvs)
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
		it.Event = new(IArtifactRegistryDeregisteredAvs)
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
func (it *IArtifactRegistryDeregisteredAvsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IArtifactRegistryDeregisteredAvsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IArtifactRegistryDeregisteredAvs represents a DeregisteredAvs event raised by the IArtifactRegistry contract.
type IArtifactRegistryDeregisteredAvs struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDeregisteredAvs is a free log retrieval operation binding the contract event 0x7945297343210cdfa6011428609630f57248645b31ef38facb55ab2f28753c23.
//
// Solidity: event DeregisteredAvs(address indexed avs)
func (_IArtifactRegistry *IArtifactRegistryFilterer) FilterDeregisteredAvs(opts *bind.FilterOpts, avs []common.Address) (*IArtifactRegistryDeregisteredAvsIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IArtifactRegistry.contract.FilterLogs(opts, "DeregisteredAvs", avsRule)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistryDeregisteredAvsIterator{contract: _IArtifactRegistry.contract, event: "DeregisteredAvs", logs: logs, sub: sub}, nil
}

// WatchDeregisteredAvs is a free log subscription operation binding the contract event 0x7945297343210cdfa6011428609630f57248645b31ef38facb55ab2f28753c23.
//
// Solidity: event DeregisteredAvs(address indexed avs)
func (_IArtifactRegistry *IArtifactRegistryFilterer) WatchDeregisteredAvs(opts *bind.WatchOpts, sink chan<- *IArtifactRegistryDeregisteredAvs, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IArtifactRegistry.contract.WatchLogs(opts, "DeregisteredAvs", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IArtifactRegistryDeregisteredAvs)
				if err := _IArtifactRegistry.contract.UnpackLog(event, "DeregisteredAvs", log); err != nil {
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

// ParseDeregisteredAvs is a log parse operation binding the contract event 0x7945297343210cdfa6011428609630f57248645b31ef38facb55ab2f28753c23.
//
// Solidity: event DeregisteredAvs(address indexed avs)
func (_IArtifactRegistry *IArtifactRegistryFilterer) ParseDeregisteredAvs(log types.Log) (*IArtifactRegistryDeregisteredAvs, error) {
	event := new(IArtifactRegistryDeregisteredAvs)
	if err := _IArtifactRegistry.contract.UnpackLog(event, "DeregisteredAvs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IArtifactRegistryPublishedArtifactIterator is returned from FilterPublishedArtifact and is used to iterate over the raw logs and unpacked data for PublishedArtifact events raised by the IArtifactRegistry contract.
type IArtifactRegistryPublishedArtifactIterator struct {
	Event *IArtifactRegistryPublishedArtifact // Event containing the contract specifics and raw log

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
func (it *IArtifactRegistryPublishedArtifactIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IArtifactRegistryPublishedArtifact)
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
		it.Event = new(IArtifactRegistryPublishedArtifact)
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
func (it *IArtifactRegistryPublishedArtifactIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IArtifactRegistryPublishedArtifactIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IArtifactRegistryPublishedArtifact represents a PublishedArtifact event raised by the IArtifactRegistry contract.
type IArtifactRegistryPublishedArtifact struct {
	Avs              common.Address
	OperatorSetId    common.Hash
	NewArtifact      IArtifactRegistryArtifact
	PreviousArtifact IArtifactRegistryArtifact
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterPublishedArtifact is a free log retrieval operation binding the contract event 0xf6cd831602d33123519e288ab94e4b0790f5da0c056eab241e7526e45fbe88b5.
//
// Solidity: event PublishedArtifact(address indexed avs, bytes indexed operatorSetId, (uint8,uint8,uint8,uint8,bytes,bytes) newArtifact, (uint8,uint8,uint8,uint8,bytes,bytes) previousArtifact)
func (_IArtifactRegistry *IArtifactRegistryFilterer) FilterPublishedArtifact(opts *bind.FilterOpts, avs []common.Address, operatorSetId [][]byte) (*IArtifactRegistryPublishedArtifactIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _IArtifactRegistry.contract.FilterLogs(opts, "PublishedArtifact", avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistryPublishedArtifactIterator{contract: _IArtifactRegistry.contract, event: "PublishedArtifact", logs: logs, sub: sub}, nil
}

// WatchPublishedArtifact is a free log subscription operation binding the contract event 0xf6cd831602d33123519e288ab94e4b0790f5da0c056eab241e7526e45fbe88b5.
//
// Solidity: event PublishedArtifact(address indexed avs, bytes indexed operatorSetId, (uint8,uint8,uint8,uint8,bytes,bytes) newArtifact, (uint8,uint8,uint8,uint8,bytes,bytes) previousArtifact)
func (_IArtifactRegistry *IArtifactRegistryFilterer) WatchPublishedArtifact(opts *bind.WatchOpts, sink chan<- *IArtifactRegistryPublishedArtifact, avs []common.Address, operatorSetId [][]byte) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _IArtifactRegistry.contract.WatchLogs(opts, "PublishedArtifact", avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IArtifactRegistryPublishedArtifact)
				if err := _IArtifactRegistry.contract.UnpackLog(event, "PublishedArtifact", log); err != nil {
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

// ParsePublishedArtifact is a log parse operation binding the contract event 0xf6cd831602d33123519e288ab94e4b0790f5da0c056eab241e7526e45fbe88b5.
//
// Solidity: event PublishedArtifact(address indexed avs, bytes indexed operatorSetId, (uint8,uint8,uint8,uint8,bytes,bytes) newArtifact, (uint8,uint8,uint8,uint8,bytes,bytes) previousArtifact)
func (_IArtifactRegistry *IArtifactRegistryFilterer) ParsePublishedArtifact(log types.Log) (*IArtifactRegistryPublishedArtifact, error) {
	event := new(IArtifactRegistryPublishedArtifact)
	if err := _IArtifactRegistry.contract.UnpackLog(event, "PublishedArtifact", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IArtifactRegistryRegisteredAvsIterator is returned from FilterRegisteredAvs and is used to iterate over the raw logs and unpacked data for RegisteredAvs events raised by the IArtifactRegistry contract.
type IArtifactRegistryRegisteredAvsIterator struct {
	Event *IArtifactRegistryRegisteredAvs // Event containing the contract specifics and raw log

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
func (it *IArtifactRegistryRegisteredAvsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IArtifactRegistryRegisteredAvs)
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
		it.Event = new(IArtifactRegistryRegisteredAvs)
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
func (it *IArtifactRegistryRegisteredAvsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IArtifactRegistryRegisteredAvsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IArtifactRegistryRegisteredAvs represents a RegisteredAvs event raised by the IArtifactRegistry contract.
type IArtifactRegistryRegisteredAvs struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRegisteredAvs is a free log retrieval operation binding the contract event 0x24976ce5cfa1b02b826fd06220c30094385eb5fc48e73b3d55e92653463e9255.
//
// Solidity: event RegisteredAvs(address indexed avs)
func (_IArtifactRegistry *IArtifactRegistryFilterer) FilterRegisteredAvs(opts *bind.FilterOpts, avs []common.Address) (*IArtifactRegistryRegisteredAvsIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IArtifactRegistry.contract.FilterLogs(opts, "RegisteredAvs", avsRule)
	if err != nil {
		return nil, err
	}
	return &IArtifactRegistryRegisteredAvsIterator{contract: _IArtifactRegistry.contract, event: "RegisteredAvs", logs: logs, sub: sub}, nil
}

// WatchRegisteredAvs is a free log subscription operation binding the contract event 0x24976ce5cfa1b02b826fd06220c30094385eb5fc48e73b3d55e92653463e9255.
//
// Solidity: event RegisteredAvs(address indexed avs)
func (_IArtifactRegistry *IArtifactRegistryFilterer) WatchRegisteredAvs(opts *bind.WatchOpts, sink chan<- *IArtifactRegistryRegisteredAvs, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IArtifactRegistry.contract.WatchLogs(opts, "RegisteredAvs", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IArtifactRegistryRegisteredAvs)
				if err := _IArtifactRegistry.contract.UnpackLog(event, "RegisteredAvs", log); err != nil {
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

// ParseRegisteredAvs is a log parse operation binding the contract event 0x24976ce5cfa1b02b826fd06220c30094385eb5fc48e73b3d55e92653463e9255.
//
// Solidity: event RegisteredAvs(address indexed avs)
func (_IArtifactRegistry *IArtifactRegistryFilterer) ParseRegisteredAvs(log types.Log) (*IArtifactRegistryRegisteredAvs, error) {
	event := new(IArtifactRegistryRegisteredAvs)
	if err := _IArtifactRegistry.contract.UnpackLog(event, "RegisteredAvs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
