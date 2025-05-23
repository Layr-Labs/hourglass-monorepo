// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ArtifactRegistryStorage

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

// ArtifactRegistryStorageMetaData contains all meta data concerning the ArtifactRegistryStorage contract.
var ArtifactRegistryStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"avsAddresses\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorAvs\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registries\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"avsId\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"}]",
}

// ArtifactRegistryStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use ArtifactRegistryStorageMetaData.ABI instead.
var ArtifactRegistryStorageABI = ArtifactRegistryStorageMetaData.ABI

// ArtifactRegistryStorage is an auto generated Go binding around an Ethereum contract.
type ArtifactRegistryStorage struct {
	ArtifactRegistryStorageCaller     // Read-only binding to the contract
	ArtifactRegistryStorageTransactor // Write-only binding to the contract
	ArtifactRegistryStorageFilterer   // Log filterer for contract events
}

// ArtifactRegistryStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type ArtifactRegistryStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArtifactRegistryStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ArtifactRegistryStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArtifactRegistryStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ArtifactRegistryStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArtifactRegistryStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ArtifactRegistryStorageSession struct {
	Contract     *ArtifactRegistryStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ArtifactRegistryStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ArtifactRegistryStorageCallerSession struct {
	Contract *ArtifactRegistryStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// ArtifactRegistryStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ArtifactRegistryStorageTransactorSession struct {
	Contract     *ArtifactRegistryStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// ArtifactRegistryStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type ArtifactRegistryStorageRaw struct {
	Contract *ArtifactRegistryStorage // Generic contract binding to access the raw methods on
}

// ArtifactRegistryStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ArtifactRegistryStorageCallerRaw struct {
	Contract *ArtifactRegistryStorageCaller // Generic read-only contract binding to access the raw methods on
}

// ArtifactRegistryStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ArtifactRegistryStorageTransactorRaw struct {
	Contract *ArtifactRegistryStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewArtifactRegistryStorage creates a new instance of ArtifactRegistryStorage, bound to a specific deployed contract.
func NewArtifactRegistryStorage(address common.Address, backend bind.ContractBackend) (*ArtifactRegistryStorage, error) {
	contract, err := bindArtifactRegistryStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryStorage{ArtifactRegistryStorageCaller: ArtifactRegistryStorageCaller{contract: contract}, ArtifactRegistryStorageTransactor: ArtifactRegistryStorageTransactor{contract: contract}, ArtifactRegistryStorageFilterer: ArtifactRegistryStorageFilterer{contract: contract}}, nil
}

// NewArtifactRegistryStorageCaller creates a new read-only instance of ArtifactRegistryStorage, bound to a specific deployed contract.
func NewArtifactRegistryStorageCaller(address common.Address, caller bind.ContractCaller) (*ArtifactRegistryStorageCaller, error) {
	contract, err := bindArtifactRegistryStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryStorageCaller{contract: contract}, nil
}

// NewArtifactRegistryStorageTransactor creates a new write-only instance of ArtifactRegistryStorage, bound to a specific deployed contract.
func NewArtifactRegistryStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*ArtifactRegistryStorageTransactor, error) {
	contract, err := bindArtifactRegistryStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryStorageTransactor{contract: contract}, nil
}

// NewArtifactRegistryStorageFilterer creates a new log filterer instance of ArtifactRegistryStorage, bound to a specific deployed contract.
func NewArtifactRegistryStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*ArtifactRegistryStorageFilterer, error) {
	contract, err := bindArtifactRegistryStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryStorageFilterer{contract: contract}, nil
}

// bindArtifactRegistryStorage binds a generic wrapper to an already deployed contract.
func bindArtifactRegistryStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ArtifactRegistryStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ArtifactRegistryStorage *ArtifactRegistryStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ArtifactRegistryStorage.Contract.ArtifactRegistryStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ArtifactRegistryStorage *ArtifactRegistryStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ArtifactRegistryStorage.Contract.ArtifactRegistryStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ArtifactRegistryStorage *ArtifactRegistryStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ArtifactRegistryStorage.Contract.ArtifactRegistryStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ArtifactRegistryStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ArtifactRegistryStorage *ArtifactRegistryStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ArtifactRegistryStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ArtifactRegistryStorage *ArtifactRegistryStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ArtifactRegistryStorage.Contract.contract.Transact(opts, method, params...)
}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCaller) AvsAddresses(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ArtifactRegistryStorage.contract.Call(opts, &out, "avsAddresses", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageSession) AvsAddresses(arg0 *big.Int) (common.Address, error) {
	return _ArtifactRegistryStorage.Contract.AvsAddresses(&_ArtifactRegistryStorage.CallOpts, arg0)
}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCallerSession) AvsAddresses(arg0 *big.Int) (common.Address, error) {
	return _ArtifactRegistryStorage.Contract.AvsAddresses(&_ArtifactRegistryStorage.CallOpts, arg0)
}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCaller) OperatorAvs(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ArtifactRegistryStorage.contract.Call(opts, &out, "operatorAvs", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageSession) OperatorAvs(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _ArtifactRegistryStorage.Contract.OperatorAvs(&_ArtifactRegistryStorage.CallOpts, arg0, arg1)
}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCallerSession) OperatorAvs(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _ArtifactRegistryStorage.Contract.OperatorAvs(&_ArtifactRegistryStorage.CallOpts, arg0, arg1)
}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCaller) Registries(opts *bind.CallOpts, arg0 common.Address) ([]byte, error) {
	var out []interface{}
	err := _ArtifactRegistryStorage.contract.Call(opts, &out, "registries", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageSession) Registries(arg0 common.Address) ([]byte, error) {
	return _ArtifactRegistryStorage.Contract.Registries(&_ArtifactRegistryStorage.CallOpts, arg0)
}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_ArtifactRegistryStorage *ArtifactRegistryStorageCallerSession) Registries(arg0 common.Address) ([]byte, error) {
	return _ArtifactRegistryStorage.Contract.Registries(&_ArtifactRegistryStorage.CallOpts, arg0)
}
