// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package AVSArtifactRegistryStorage

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

// AVSArtifactRegistryStorageMetaData contains all meta data concerning the AVSArtifactRegistryStorage contract.
var AVSArtifactRegistryStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"avsAddresses\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorAvs\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registries\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"avsId\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"}]",
}

// AVSArtifactRegistryStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use AVSArtifactRegistryStorageMetaData.ABI instead.
var AVSArtifactRegistryStorageABI = AVSArtifactRegistryStorageMetaData.ABI

// AVSArtifactRegistryStorage is an auto generated Go binding around an Ethereum contract.
type AVSArtifactRegistryStorage struct {
	AVSArtifactRegistryStorageCaller     // Read-only binding to the contract
	AVSArtifactRegistryStorageTransactor // Write-only binding to the contract
	AVSArtifactRegistryStorageFilterer   // Log filterer for contract events
}

// AVSArtifactRegistryStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type AVSArtifactRegistryStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AVSArtifactRegistryStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AVSArtifactRegistryStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AVSArtifactRegistryStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AVSArtifactRegistryStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AVSArtifactRegistryStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AVSArtifactRegistryStorageSession struct {
	Contract     *AVSArtifactRegistryStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts               // Call options to use throughout this session
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// AVSArtifactRegistryStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AVSArtifactRegistryStorageCallerSession struct {
	Contract *AVSArtifactRegistryStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                     // Call options to use throughout this session
}

// AVSArtifactRegistryStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AVSArtifactRegistryStorageTransactorSession struct {
	Contract     *AVSArtifactRegistryStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                     // Transaction auth options to use throughout this session
}

// AVSArtifactRegistryStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type AVSArtifactRegistryStorageRaw struct {
	Contract *AVSArtifactRegistryStorage // Generic contract binding to access the raw methods on
}

// AVSArtifactRegistryStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AVSArtifactRegistryStorageCallerRaw struct {
	Contract *AVSArtifactRegistryStorageCaller // Generic read-only contract binding to access the raw methods on
}

// AVSArtifactRegistryStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AVSArtifactRegistryStorageTransactorRaw struct {
	Contract *AVSArtifactRegistryStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAVSArtifactRegistryStorage creates a new instance of AVSArtifactRegistryStorage, bound to a specific deployed contract.
func NewAVSArtifactRegistryStorage(address common.Address, backend bind.ContractBackend) (*AVSArtifactRegistryStorage, error) {
	contract, err := bindAVSArtifactRegistryStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AVSArtifactRegistryStorage{AVSArtifactRegistryStorageCaller: AVSArtifactRegistryStorageCaller{contract: contract}, AVSArtifactRegistryStorageTransactor: AVSArtifactRegistryStorageTransactor{contract: contract}, AVSArtifactRegistryStorageFilterer: AVSArtifactRegistryStorageFilterer{contract: contract}}, nil
}

// NewAVSArtifactRegistryStorageCaller creates a new read-only instance of AVSArtifactRegistryStorage, bound to a specific deployed contract.
func NewAVSArtifactRegistryStorageCaller(address common.Address, caller bind.ContractCaller) (*AVSArtifactRegistryStorageCaller, error) {
	contract, err := bindAVSArtifactRegistryStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AVSArtifactRegistryStorageCaller{contract: contract}, nil
}

// NewAVSArtifactRegistryStorageTransactor creates a new write-only instance of AVSArtifactRegistryStorage, bound to a specific deployed contract.
func NewAVSArtifactRegistryStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*AVSArtifactRegistryStorageTransactor, error) {
	contract, err := bindAVSArtifactRegistryStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AVSArtifactRegistryStorageTransactor{contract: contract}, nil
}

// NewAVSArtifactRegistryStorageFilterer creates a new log filterer instance of AVSArtifactRegistryStorage, bound to a specific deployed contract.
func NewAVSArtifactRegistryStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*AVSArtifactRegistryStorageFilterer, error) {
	contract, err := bindAVSArtifactRegistryStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AVSArtifactRegistryStorageFilterer{contract: contract}, nil
}

// bindAVSArtifactRegistryStorage binds a generic wrapper to an already deployed contract.
func bindAVSArtifactRegistryStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AVSArtifactRegistryStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AVSArtifactRegistryStorage.Contract.AVSArtifactRegistryStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AVSArtifactRegistryStorage.Contract.AVSArtifactRegistryStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AVSArtifactRegistryStorage.Contract.AVSArtifactRegistryStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AVSArtifactRegistryStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AVSArtifactRegistryStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AVSArtifactRegistryStorage.Contract.contract.Transact(opts, method, params...)
}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCaller) AvsAddresses(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AVSArtifactRegistryStorage.contract.Call(opts, &out, "avsAddresses", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageSession) AvsAddresses(arg0 *big.Int) (common.Address, error) {
	return _AVSArtifactRegistryStorage.Contract.AvsAddresses(&_AVSArtifactRegistryStorage.CallOpts, arg0)
}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCallerSession) AvsAddresses(arg0 *big.Int) (common.Address, error) {
	return _AVSArtifactRegistryStorage.Contract.AvsAddresses(&_AVSArtifactRegistryStorage.CallOpts, arg0)
}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCaller) OperatorAvs(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AVSArtifactRegistryStorage.contract.Call(opts, &out, "operatorAvs", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageSession) OperatorAvs(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _AVSArtifactRegistryStorage.Contract.OperatorAvs(&_AVSArtifactRegistryStorage.CallOpts, arg0, arg1)
}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCallerSession) OperatorAvs(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _AVSArtifactRegistryStorage.Contract.OperatorAvs(&_AVSArtifactRegistryStorage.CallOpts, arg0, arg1)
}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCaller) Registries(opts *bind.CallOpts, arg0 common.Address) ([]byte, error) {
	var out []interface{}
	err := _AVSArtifactRegistryStorage.contract.Call(opts, &out, "registries", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageSession) Registries(arg0 common.Address) ([]byte, error) {
	return _AVSArtifactRegistryStorage.Contract.Registries(&_AVSArtifactRegistryStorage.CallOpts, arg0)
}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_AVSArtifactRegistryStorage *AVSArtifactRegistryStorageCallerSession) Registries(arg0 common.Address) ([]byte, error) {
	return _AVSArtifactRegistryStorage.Contract.Registries(&_AVSArtifactRegistryStorage.CallOpts, arg0)
}
