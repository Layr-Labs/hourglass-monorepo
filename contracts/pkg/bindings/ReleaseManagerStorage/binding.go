// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ReleaseManagerStorage

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

// IReleaseManagerArtifact is an auto generated low-level Go binding around an user-defined struct.
type IReleaseManagerArtifact struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}

// IReleaseManagerArtifactPromotion is an auto generated low-level Go binding around an user-defined struct.
type IReleaseManagerArtifactPromotion struct {
	PromotionStatus uint8
	Digest          [32]byte
	RegistryUrl     string
	OperatorSetIds  [][32]byte
}

// IReleaseManagerPromotedArtifact is an auto generated low-level Go binding around an user-defined struct.
type IReleaseManagerPromotedArtifact struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}

// ReleaseManagerStorageMetaData contains all meta data concerning the ReleaseManagerStorage contract.
var ReleaseManagerStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"allPromotedArtifacts\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"artifactExists\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"artifacts\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregister\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLatestPromotedArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.PromotedArtifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotedArtifactAtBlock\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.PromotedArtifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotedArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.PromotedArtifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionCheckpointAt\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"pos\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"artifactIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionCheckpointCount\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionHistory\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.PromotedArtifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionStatusAtBlock\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"permissionController\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPermissionController\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"promoteArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"promotions\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.ArtifactPromotion[]\",\"components\":[{\"name\":\"promotionStatus\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"operatorSetIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}]},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"publishArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"artifacts\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.Artifact[]\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"register\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registeredAVS\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updatePromotionStatus\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"newStatus\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AVSDeregistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AVSRegistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ArtifactPublished\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOperatingSystem\"},{\"name\":\"artifactType\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumArtifactType\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ArtifactsPromoted\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"version\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"digests\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"statuses\",\"type\":\"uint8[]\",\"indexed\":false,\"internalType\":\"enumPromotionStatus[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PromotionStatusUpdated\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"oldStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumPromotionStatus\"},{\"name\":\"newStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumPromotionStatus\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AVSAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AVSNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArrayLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArtifactNotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDeadline\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Unauthorized\",\"inputs\":[]}]",
}

// ReleaseManagerStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use ReleaseManagerStorageMetaData.ABI instead.
var ReleaseManagerStorageABI = ReleaseManagerStorageMetaData.ABI

// ReleaseManagerStorage is an auto generated Go binding around an Ethereum contract.
type ReleaseManagerStorage struct {
	ReleaseManagerStorageCaller     // Read-only binding to the contract
	ReleaseManagerStorageTransactor // Write-only binding to the contract
	ReleaseManagerStorageFilterer   // Log filterer for contract events
}

// ReleaseManagerStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReleaseManagerStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReleaseManagerStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReleaseManagerStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReleaseManagerStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ReleaseManagerStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReleaseManagerStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReleaseManagerStorageSession struct {
	Contract     *ReleaseManagerStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// ReleaseManagerStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReleaseManagerStorageCallerSession struct {
	Contract *ReleaseManagerStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// ReleaseManagerStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReleaseManagerStorageTransactorSession struct {
	Contract     *ReleaseManagerStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// ReleaseManagerStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReleaseManagerStorageRaw struct {
	Contract *ReleaseManagerStorage // Generic contract binding to access the raw methods on
}

// ReleaseManagerStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReleaseManagerStorageCallerRaw struct {
	Contract *ReleaseManagerStorageCaller // Generic read-only contract binding to access the raw methods on
}

// ReleaseManagerStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReleaseManagerStorageTransactorRaw struct {
	Contract *ReleaseManagerStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReleaseManagerStorage creates a new instance of ReleaseManagerStorage, bound to a specific deployed contract.
func NewReleaseManagerStorage(address common.Address, backend bind.ContractBackend) (*ReleaseManagerStorage, error) {
	contract, err := bindReleaseManagerStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorage{ReleaseManagerStorageCaller: ReleaseManagerStorageCaller{contract: contract}, ReleaseManagerStorageTransactor: ReleaseManagerStorageTransactor{contract: contract}, ReleaseManagerStorageFilterer: ReleaseManagerStorageFilterer{contract: contract}}, nil
}

// NewReleaseManagerStorageCaller creates a new read-only instance of ReleaseManagerStorage, bound to a specific deployed contract.
func NewReleaseManagerStorageCaller(address common.Address, caller bind.ContractCaller) (*ReleaseManagerStorageCaller, error) {
	contract, err := bindReleaseManagerStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageCaller{contract: contract}, nil
}

// NewReleaseManagerStorageTransactor creates a new write-only instance of ReleaseManagerStorage, bound to a specific deployed contract.
func NewReleaseManagerStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*ReleaseManagerStorageTransactor, error) {
	contract, err := bindReleaseManagerStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageTransactor{contract: contract}, nil
}

// NewReleaseManagerStorageFilterer creates a new log filterer instance of ReleaseManagerStorage, bound to a specific deployed contract.
func NewReleaseManagerStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*ReleaseManagerStorageFilterer, error) {
	contract, err := bindReleaseManagerStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageFilterer{contract: contract}, nil
}

// bindReleaseManagerStorage binds a generic wrapper to an already deployed contract.
func bindReleaseManagerStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ReleaseManagerStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReleaseManagerStorage *ReleaseManagerStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ReleaseManagerStorage.Contract.ReleaseManagerStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReleaseManagerStorage *ReleaseManagerStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.ReleaseManagerStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReleaseManagerStorage *ReleaseManagerStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.ReleaseManagerStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ReleaseManagerStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.contract.Transact(opts, method, params...)
}

// AllPromotedArtifacts is a free data retrieval call binding the contract method 0xae738920.
//
// Solidity: function allPromotedArtifacts(address , bytes32 , uint256 ) view returns(bytes32 digest, string registryUrl, uint8 status, string version, uint256 deploymentDeadline, uint256 promotedAt)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) AllPromotedArtifacts(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte, arg2 *big.Int) (struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "allPromotedArtifacts", arg0, arg1, arg2)

	outstruct := new(struct {
		Digest             [32]byte
		RegistryUrl        string
		Status             uint8
		Version            string
		DeploymentDeadline *big.Int
		PromotedAt         *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Digest = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.RegistryUrl = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Status = *abi.ConvertType(out[2], new(uint8)).(*uint8)
	outstruct.Version = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.DeploymentDeadline = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.PromotedAt = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// AllPromotedArtifacts is a free data retrieval call binding the contract method 0xae738920.
//
// Solidity: function allPromotedArtifacts(address , bytes32 , uint256 ) view returns(bytes32 digest, string registryUrl, uint8 status, string version, uint256 deploymentDeadline, uint256 promotedAt)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) AllPromotedArtifacts(arg0 common.Address, arg1 [32]byte, arg2 *big.Int) (struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}, error) {
	return _ReleaseManagerStorage.Contract.AllPromotedArtifacts(&_ReleaseManagerStorage.CallOpts, arg0, arg1, arg2)
}

// AllPromotedArtifacts is a free data retrieval call binding the contract method 0xae738920.
//
// Solidity: function allPromotedArtifacts(address , bytes32 , uint256 ) view returns(bytes32 digest, string registryUrl, uint8 status, string version, uint256 deploymentDeadline, uint256 promotedAt)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) AllPromotedArtifacts(arg0 common.Address, arg1 [32]byte, arg2 *big.Int) (struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}, error) {
	return _ReleaseManagerStorage.Contract.AllPromotedArtifacts(&_ReleaseManagerStorage.CallOpts, arg0, arg1, arg2)
}

// ArtifactExists is a free data retrieval call binding the contract method 0xb2daca5c.
//
// Solidity: function artifactExists(bytes32 ) view returns(bool)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) ArtifactExists(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "artifactExists", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ArtifactExists is a free data retrieval call binding the contract method 0xb2daca5c.
//
// Solidity: function artifactExists(bytes32 ) view returns(bool)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) ArtifactExists(arg0 [32]byte) (bool, error) {
	return _ReleaseManagerStorage.Contract.ArtifactExists(&_ReleaseManagerStorage.CallOpts, arg0)
}

// ArtifactExists is a free data retrieval call binding the contract method 0xb2daca5c.
//
// Solidity: function artifactExists(bytes32 ) view returns(bool)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) ArtifactExists(arg0 [32]byte) (bool, error) {
	return _ReleaseManagerStorage.Contract.ArtifactExists(&_ReleaseManagerStorage.CallOpts, arg0)
}

// Artifacts is a free data retrieval call binding the contract method 0xa63e3a37.
//
// Solidity: function artifacts(bytes32 ) view returns(uint8 artifactType, uint8 architecture, uint8 os, bytes32 digest, string registryUrl, uint256 publishedAt)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) Artifacts(opts *bind.CallOpts, arg0 [32]byte) (struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "artifacts", arg0)

	outstruct := new(struct {
		ArtifactType uint8
		Architecture uint8
		Os           uint8
		Digest       [32]byte
		RegistryUrl  string
		PublishedAt  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ArtifactType = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.Architecture = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.Os = *abi.ConvertType(out[2], new(uint8)).(*uint8)
	outstruct.Digest = *abi.ConvertType(out[3], new([32]byte)).(*[32]byte)
	outstruct.RegistryUrl = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.PublishedAt = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Artifacts is a free data retrieval call binding the contract method 0xa63e3a37.
//
// Solidity: function artifacts(bytes32 ) view returns(uint8 artifactType, uint8 architecture, uint8 os, bytes32 digest, string registryUrl, uint256 publishedAt)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) Artifacts(arg0 [32]byte) (struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}, error) {
	return _ReleaseManagerStorage.Contract.Artifacts(&_ReleaseManagerStorage.CallOpts, arg0)
}

// Artifacts is a free data retrieval call binding the contract method 0xa63e3a37.
//
// Solidity: function artifacts(bytes32 ) view returns(uint8 artifactType, uint8 architecture, uint8 os, bytes32 digest, string registryUrl, uint256 publishedAt)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) Artifacts(arg0 [32]byte) (struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}, error) {
	return _ReleaseManagerStorage.Contract.Artifacts(&_ReleaseManagerStorage.CallOpts, arg0)
}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetArtifact(opts *bind.CallOpts, avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getArtifact", avs, digest)

	if err != nil {
		return *new(IReleaseManagerArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerArtifact)).(*IReleaseManagerArtifact)

	return out0, err

}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetArtifact(avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetArtifact(&_ReleaseManagerStorage.CallOpts, avs, digest)
}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetArtifact(avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetArtifact(&_ReleaseManagerStorage.CallOpts, avs, digest)
}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetLatestPromotedArtifact(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getLatestPromotedArtifact", avs, operatorSetId)

	if err != nil {
		return *new(IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerPromotedArtifact)).(*IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetLatestPromotedArtifact(avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetLatestPromotedArtifact(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetLatestPromotedArtifact(avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetLatestPromotedArtifact(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetPromotedArtifactAtBlock(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getPromotedArtifactAtBlock", avs, operatorSetId, blockNumber)

	if err != nil {
		return *new(IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerPromotedArtifact)).(*IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetPromotedArtifactAtBlock(avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetPromotedArtifactAtBlock(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId, blockNumber)
}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetPromotedArtifactAtBlock(avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetPromotedArtifactAtBlock(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId, blockNumber)
}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetPromotedArtifacts(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getPromotedArtifacts", avs, operatorSetId)

	if err != nil {
		return *new([]IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]IReleaseManagerPromotedArtifact)).(*[]IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetPromotedArtifacts(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetPromotedArtifacts(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetPromotedArtifacts(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetPromotedArtifacts(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetPromotionCheckpointAt(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getPromotionCheckpointAt", avs, operatorSetId, pos)

	outstruct := new(struct {
		BlockNumber   *big.Int
		ArtifactIndex *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.BlockNumber = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ArtifactIndex = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetPromotionCheckpointAt(avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionCheckpointAt(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId, pos)
}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetPromotionCheckpointAt(avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionCheckpointAt(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId, pos)
}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetPromotionCheckpointCount(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getPromotionCheckpointCount", avs, operatorSetId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetPromotionCheckpointCount(avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionCheckpointCount(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetPromotionCheckpointCount(avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionCheckpointCount(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetPromotionHistory(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getPromotionHistory", avs, operatorSetId)

	if err != nil {
		return *new([]IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]IReleaseManagerPromotedArtifact)).(*[]IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetPromotionHistory(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionHistory(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetPromotionHistory(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionHistory(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId)
}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) GetPromotionStatusAtBlock(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "getPromotionStatusAtBlock", avs, operatorSetId, digest, blockNumber)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) GetPromotionStatusAtBlock(avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionStatusAtBlock(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId, digest, blockNumber)
}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) GetPromotionStatusAtBlock(avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	return _ReleaseManagerStorage.Contract.GetPromotionStatusAtBlock(&_ReleaseManagerStorage.CallOpts, avs, operatorSetId, digest, blockNumber)
}

// PermissionController is a free data retrieval call binding the contract method 0x4657e26a.
//
// Solidity: function permissionController() view returns(address)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) PermissionController(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "permissionController")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PermissionController is a free data retrieval call binding the contract method 0x4657e26a.
//
// Solidity: function permissionController() view returns(address)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) PermissionController() (common.Address, error) {
	return _ReleaseManagerStorage.Contract.PermissionController(&_ReleaseManagerStorage.CallOpts)
}

// PermissionController is a free data retrieval call binding the contract method 0x4657e26a.
//
// Solidity: function permissionController() view returns(address)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) PermissionController() (common.Address, error) {
	return _ReleaseManagerStorage.Contract.PermissionController(&_ReleaseManagerStorage.CallOpts)
}

// RegisteredAVS is a free data retrieval call binding the contract method 0xbf2d8e07.
//
// Solidity: function registeredAVS(address ) view returns(bool)
func (_ReleaseManagerStorage *ReleaseManagerStorageCaller) RegisteredAVS(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ReleaseManagerStorage.contract.Call(opts, &out, "registeredAVS", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RegisteredAVS is a free data retrieval call binding the contract method 0xbf2d8e07.
//
// Solidity: function registeredAVS(address ) view returns(bool)
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) RegisteredAVS(arg0 common.Address) (bool, error) {
	return _ReleaseManagerStorage.Contract.RegisteredAVS(&_ReleaseManagerStorage.CallOpts, arg0)
}

// RegisteredAVS is a free data retrieval call binding the contract method 0xbf2d8e07.
//
// Solidity: function registeredAVS(address ) view returns(bool)
func (_ReleaseManagerStorage *ReleaseManagerStorageCallerSession) RegisteredAVS(arg0 common.Address) (bool, error) {
	return _ReleaseManagerStorage.Contract.RegisteredAVS(&_ReleaseManagerStorage.CallOpts, arg0)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactor) Deregister(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _ReleaseManagerStorage.contract.Transact(opts, "deregister", avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.Deregister(&_ReleaseManagerStorage.TransactOpts, avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.Deregister(&_ReleaseManagerStorage.TransactOpts, avs)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactor) PromoteArtifacts(opts *bind.TransactOpts, avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _ReleaseManagerStorage.contract.Transact(opts, "promoteArtifacts", avs, promotions, version, deploymentDeadline)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) PromoteArtifacts(avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.PromoteArtifacts(&_ReleaseManagerStorage.TransactOpts, avs, promotions, version, deploymentDeadline)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorSession) PromoteArtifacts(avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.PromoteArtifacts(&_ReleaseManagerStorage.TransactOpts, avs, promotions, version, deploymentDeadline)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] artifacts) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactor) PublishArtifacts(opts *bind.TransactOpts, avs common.Address, artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _ReleaseManagerStorage.contract.Transact(opts, "publishArtifacts", avs, artifacts)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] artifacts) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) PublishArtifacts(avs common.Address, artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.PublishArtifacts(&_ReleaseManagerStorage.TransactOpts, avs, artifacts)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] artifacts) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorSession) PublishArtifacts(avs common.Address, artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.PublishArtifacts(&_ReleaseManagerStorage.TransactOpts, avs, artifacts)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactor) Register(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _ReleaseManagerStorage.contract.Transact(opts, "register", avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) Register(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.Register(&_ReleaseManagerStorage.TransactOpts, avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorSession) Register(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.Register(&_ReleaseManagerStorage.TransactOpts, avs)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactor) UpdatePromotionStatus(opts *bind.TransactOpts, avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _ReleaseManagerStorage.contract.Transact(opts, "updatePromotionStatus", avs, digest, operatorSetId, newStatus)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageSession) UpdatePromotionStatus(avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.UpdatePromotionStatus(&_ReleaseManagerStorage.TransactOpts, avs, digest, operatorSetId, newStatus)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_ReleaseManagerStorage *ReleaseManagerStorageTransactorSession) UpdatePromotionStatus(avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _ReleaseManagerStorage.Contract.UpdatePromotionStatus(&_ReleaseManagerStorage.TransactOpts, avs, digest, operatorSetId, newStatus)
}

// ReleaseManagerStorageAVSDeregisteredIterator is returned from FilterAVSDeregistered and is used to iterate over the raw logs and unpacked data for AVSDeregistered events raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageAVSDeregisteredIterator struct {
	Event *ReleaseManagerStorageAVSDeregistered // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerStorageAVSDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerStorageAVSDeregistered)
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
		it.Event = new(ReleaseManagerStorageAVSDeregistered)
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
func (it *ReleaseManagerStorageAVSDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerStorageAVSDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerStorageAVSDeregistered represents a AVSDeregistered event raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageAVSDeregistered struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAVSDeregistered is a free log retrieval operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) FilterAVSDeregistered(opts *bind.FilterOpts, avs []common.Address) (*ReleaseManagerStorageAVSDeregisteredIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.FilterLogs(opts, "AVSDeregistered", avsRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageAVSDeregisteredIterator{contract: _ReleaseManagerStorage.contract, event: "AVSDeregistered", logs: logs, sub: sub}, nil
}

// WatchAVSDeregistered is a free log subscription operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) WatchAVSDeregistered(opts *bind.WatchOpts, sink chan<- *ReleaseManagerStorageAVSDeregistered, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.WatchLogs(opts, "AVSDeregistered", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerStorageAVSDeregistered)
				if err := _ReleaseManagerStorage.contract.UnpackLog(event, "AVSDeregistered", log); err != nil {
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

// ParseAVSDeregistered is a log parse operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) ParseAVSDeregistered(log types.Log) (*ReleaseManagerStorageAVSDeregistered, error) {
	event := new(ReleaseManagerStorageAVSDeregistered)
	if err := _ReleaseManagerStorage.contract.UnpackLog(event, "AVSDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerStorageAVSRegisteredIterator is returned from FilterAVSRegistered and is used to iterate over the raw logs and unpacked data for AVSRegistered events raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageAVSRegisteredIterator struct {
	Event *ReleaseManagerStorageAVSRegistered // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerStorageAVSRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerStorageAVSRegistered)
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
		it.Event = new(ReleaseManagerStorageAVSRegistered)
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
func (it *ReleaseManagerStorageAVSRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerStorageAVSRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerStorageAVSRegistered represents a AVSRegistered event raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageAVSRegistered struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAVSRegistered is a free log retrieval operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) FilterAVSRegistered(opts *bind.FilterOpts, avs []common.Address) (*ReleaseManagerStorageAVSRegisteredIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.FilterLogs(opts, "AVSRegistered", avsRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageAVSRegisteredIterator{contract: _ReleaseManagerStorage.contract, event: "AVSRegistered", logs: logs, sub: sub}, nil
}

// WatchAVSRegistered is a free log subscription operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) WatchAVSRegistered(opts *bind.WatchOpts, sink chan<- *ReleaseManagerStorageAVSRegistered, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.WatchLogs(opts, "AVSRegistered", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerStorageAVSRegistered)
				if err := _ReleaseManagerStorage.contract.UnpackLog(event, "AVSRegistered", log); err != nil {
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

// ParseAVSRegistered is a log parse operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) ParseAVSRegistered(log types.Log) (*ReleaseManagerStorageAVSRegistered, error) {
	event := new(ReleaseManagerStorageAVSRegistered)
	if err := _ReleaseManagerStorage.contract.UnpackLog(event, "AVSRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerStorageArtifactPublishedIterator is returned from FilterArtifactPublished and is used to iterate over the raw logs and unpacked data for ArtifactPublished events raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageArtifactPublishedIterator struct {
	Event *ReleaseManagerStorageArtifactPublished // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerStorageArtifactPublishedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerStorageArtifactPublished)
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
		it.Event = new(ReleaseManagerStorageArtifactPublished)
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
func (it *ReleaseManagerStorageArtifactPublishedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerStorageArtifactPublishedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerStorageArtifactPublished represents a ArtifactPublished event raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageArtifactPublished struct {
	Avs          common.Address
	Digest       [32]byte
	RegistryUrl  string
	Architecture uint8
	Os           uint8
	ArtifactType uint8
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterArtifactPublished is a free log retrieval operation binding the contract event 0x622e1c25f0e4dcedadb24c8f0bbe6ceaa3776cbeb58b17a2d5d8ac8ae31a7502.
//
// Solidity: event ArtifactPublished(address indexed avs, bytes32 indexed digest, string registryUrl, uint8 architecture, uint8 os, uint8 artifactType)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) FilterArtifactPublished(opts *bind.FilterOpts, avs []common.Address, digest [][32]byte) (*ReleaseManagerStorageArtifactPublishedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.FilterLogs(opts, "ArtifactPublished", avsRule, digestRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageArtifactPublishedIterator{contract: _ReleaseManagerStorage.contract, event: "ArtifactPublished", logs: logs, sub: sub}, nil
}

// WatchArtifactPublished is a free log subscription operation binding the contract event 0x622e1c25f0e4dcedadb24c8f0bbe6ceaa3776cbeb58b17a2d5d8ac8ae31a7502.
//
// Solidity: event ArtifactPublished(address indexed avs, bytes32 indexed digest, string registryUrl, uint8 architecture, uint8 os, uint8 artifactType)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) WatchArtifactPublished(opts *bind.WatchOpts, sink chan<- *ReleaseManagerStorageArtifactPublished, avs []common.Address, digest [][32]byte) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.WatchLogs(opts, "ArtifactPublished", avsRule, digestRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerStorageArtifactPublished)
				if err := _ReleaseManagerStorage.contract.UnpackLog(event, "ArtifactPublished", log); err != nil {
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

// ParseArtifactPublished is a log parse operation binding the contract event 0x622e1c25f0e4dcedadb24c8f0bbe6ceaa3776cbeb58b17a2d5d8ac8ae31a7502.
//
// Solidity: event ArtifactPublished(address indexed avs, bytes32 indexed digest, string registryUrl, uint8 architecture, uint8 os, uint8 artifactType)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) ParseArtifactPublished(log types.Log) (*ReleaseManagerStorageArtifactPublished, error) {
	event := new(ReleaseManagerStorageArtifactPublished)
	if err := _ReleaseManagerStorage.contract.UnpackLog(event, "ArtifactPublished", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerStorageArtifactsPromotedIterator is returned from FilterArtifactsPromoted and is used to iterate over the raw logs and unpacked data for ArtifactsPromoted events raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageArtifactsPromotedIterator struct {
	Event *ReleaseManagerStorageArtifactsPromoted // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerStorageArtifactsPromotedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerStorageArtifactsPromoted)
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
		it.Event = new(ReleaseManagerStorageArtifactsPromoted)
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
func (it *ReleaseManagerStorageArtifactsPromotedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerStorageArtifactsPromotedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerStorageArtifactsPromoted represents a ArtifactsPromoted event raised by the ReleaseManagerStorage contract.
type ReleaseManagerStorageArtifactsPromoted struct {
	Avs                common.Address
	Version            common.Hash
	DeploymentDeadline *big.Int
	Digests            [][32]byte
	Statuses           []uint8
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterArtifactsPromoted is a free log retrieval operation binding the contract event 0x7fd55cb6307da041fb4711b42eb59c940bb12e76dc208c9be19be9abde37815d.
//
// Solidity: event ArtifactsPromoted(address indexed avs, string indexed version, uint256 deploymentDeadline, bytes32[] digests, uint8[] statuses)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) FilterArtifactsPromoted(opts *bind.FilterOpts, avs []common.Address, version []string) (*ReleaseManagerStorageArtifactsPromotedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.FilterLogs(opts, "ArtifactsPromoted", avsRule, versionRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStorageArtifactsPromotedIterator{contract: _ReleaseManagerStorage.contract, event: "ArtifactsPromoted", logs: logs, sub: sub}, nil
}

// WatchArtifactsPromoted is a free log subscription operation binding the contract event 0x7fd55cb6307da041fb4711b42eb59c940bb12e76dc208c9be19be9abde37815d.
//
// Solidity: event ArtifactsPromoted(address indexed avs, string indexed version, uint256 deploymentDeadline, bytes32[] digests, uint8[] statuses)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) WatchArtifactsPromoted(opts *bind.WatchOpts, sink chan<- *ReleaseManagerStorageArtifactsPromoted, avs []common.Address, version []string) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.WatchLogs(opts, "ArtifactsPromoted", avsRule, versionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerStorageArtifactsPromoted)
				if err := _ReleaseManagerStorage.contract.UnpackLog(event, "ArtifactsPromoted", log); err != nil {
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

// ParseArtifactsPromoted is a log parse operation binding the contract event 0x7fd55cb6307da041fb4711b42eb59c940bb12e76dc208c9be19be9abde37815d.
//
// Solidity: event ArtifactsPromoted(address indexed avs, string indexed version, uint256 deploymentDeadline, bytes32[] digests, uint8[] statuses)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) ParseArtifactsPromoted(log types.Log) (*ReleaseManagerStorageArtifactsPromoted, error) {
	event := new(ReleaseManagerStorageArtifactsPromoted)
	if err := _ReleaseManagerStorage.contract.UnpackLog(event, "ArtifactsPromoted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerStoragePromotionStatusUpdatedIterator is returned from FilterPromotionStatusUpdated and is used to iterate over the raw logs and unpacked data for PromotionStatusUpdated events raised by the ReleaseManagerStorage contract.
type ReleaseManagerStoragePromotionStatusUpdatedIterator struct {
	Event *ReleaseManagerStoragePromotionStatusUpdated // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerStoragePromotionStatusUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerStoragePromotionStatusUpdated)
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
		it.Event = new(ReleaseManagerStoragePromotionStatusUpdated)
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
func (it *ReleaseManagerStoragePromotionStatusUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerStoragePromotionStatusUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerStoragePromotionStatusUpdated represents a PromotionStatusUpdated event raised by the ReleaseManagerStorage contract.
type ReleaseManagerStoragePromotionStatusUpdated struct {
	Avs           common.Address
	Digest        [32]byte
	OperatorSetId [32]byte
	OldStatus     uint8
	NewStatus     uint8
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterPromotionStatusUpdated is a free log retrieval operation binding the contract event 0x1341ac2513ee1dd81316f76f3a7840608cd4dc5a3302d2fb45ea1ad24d355671.
//
// Solidity: event PromotionStatusUpdated(address indexed avs, bytes32 indexed digest, bytes32 indexed operatorSetId, uint8 oldStatus, uint8 newStatus)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) FilterPromotionStatusUpdated(opts *bind.FilterOpts, avs []common.Address, digest [][32]byte, operatorSetId [][32]byte) (*ReleaseManagerStoragePromotionStatusUpdatedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.FilterLogs(opts, "PromotionStatusUpdated", avsRule, digestRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerStoragePromotionStatusUpdatedIterator{contract: _ReleaseManagerStorage.contract, event: "PromotionStatusUpdated", logs: logs, sub: sub}, nil
}

// WatchPromotionStatusUpdated is a free log subscription operation binding the contract event 0x1341ac2513ee1dd81316f76f3a7840608cd4dc5a3302d2fb45ea1ad24d355671.
//
// Solidity: event PromotionStatusUpdated(address indexed avs, bytes32 indexed digest, bytes32 indexed operatorSetId, uint8 oldStatus, uint8 newStatus)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) WatchPromotionStatusUpdated(opts *bind.WatchOpts, sink chan<- *ReleaseManagerStoragePromotionStatusUpdated, avs []common.Address, digest [][32]byte, operatorSetId [][32]byte) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _ReleaseManagerStorage.contract.WatchLogs(opts, "PromotionStatusUpdated", avsRule, digestRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerStoragePromotionStatusUpdated)
				if err := _ReleaseManagerStorage.contract.UnpackLog(event, "PromotionStatusUpdated", log); err != nil {
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

// ParsePromotionStatusUpdated is a log parse operation binding the contract event 0x1341ac2513ee1dd81316f76f3a7840608cd4dc5a3302d2fb45ea1ad24d355671.
//
// Solidity: event PromotionStatusUpdated(address indexed avs, bytes32 indexed digest, bytes32 indexed operatorSetId, uint8 oldStatus, uint8 newStatus)
func (_ReleaseManagerStorage *ReleaseManagerStorageFilterer) ParsePromotionStatusUpdated(log types.Log) (*ReleaseManagerStoragePromotionStatusUpdated, error) {
	event := new(ReleaseManagerStoragePromotionStatusUpdated)
	if err := _ReleaseManagerStorage.contract.UnpackLog(event, "PromotionStatusUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
