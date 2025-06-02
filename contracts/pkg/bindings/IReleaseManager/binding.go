// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package IReleaseManager

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

// IReleaseManagerMetaData contains all meta data concerning the IReleaseManager contract.
var IReleaseManagerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"deregister\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLatestPromotedArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.PromotedArtifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotedArtifactAtBlock\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.PromotedArtifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotedArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.PromotedArtifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionCheckpointAt\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"pos\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"artifactIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionCheckpointCount\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionHistory\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.PromotedArtifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionStatusAtBlock\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"promoteArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"promotions\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.ArtifactPromotion[]\",\"components\":[{\"name\":\"promotionStatus\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"operatorSetIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}]},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"publishArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"artifacts\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.Artifact[]\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"register\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updatePromotionStatus\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"newStatus\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AVSDeregistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AVSRegistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ArtifactPublished\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOperatingSystem\"},{\"name\":\"artifactType\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumArtifactType\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ArtifactsPromoted\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"version\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"digests\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"statuses\",\"type\":\"uint8[]\",\"indexed\":false,\"internalType\":\"enumPromotionStatus[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PromotionStatusUpdated\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"oldStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumPromotionStatus\"},{\"name\":\"newStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumPromotionStatus\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AVSAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AVSNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArrayLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArtifactNotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDeadline\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Unauthorized\",\"inputs\":[]}]",
}

// IReleaseManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use IReleaseManagerMetaData.ABI instead.
var IReleaseManagerABI = IReleaseManagerMetaData.ABI

// IReleaseManager is an auto generated Go binding around an Ethereum contract.
type IReleaseManager struct {
	IReleaseManagerCaller     // Read-only binding to the contract
	IReleaseManagerTransactor // Write-only binding to the contract
	IReleaseManagerFilterer   // Log filterer for contract events
}

// IReleaseManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type IReleaseManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IReleaseManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IReleaseManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IReleaseManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IReleaseManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IReleaseManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IReleaseManagerSession struct {
	Contract     *IReleaseManager  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IReleaseManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IReleaseManagerCallerSession struct {
	Contract *IReleaseManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// IReleaseManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IReleaseManagerTransactorSession struct {
	Contract     *IReleaseManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// IReleaseManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type IReleaseManagerRaw struct {
	Contract *IReleaseManager // Generic contract binding to access the raw methods on
}

// IReleaseManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IReleaseManagerCallerRaw struct {
	Contract *IReleaseManagerCaller // Generic read-only contract binding to access the raw methods on
}

// IReleaseManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IReleaseManagerTransactorRaw struct {
	Contract *IReleaseManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIReleaseManager creates a new instance of IReleaseManager, bound to a specific deployed contract.
func NewIReleaseManager(address common.Address, backend bind.ContractBackend) (*IReleaseManager, error) {
	contract, err := bindIReleaseManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IReleaseManager{IReleaseManagerCaller: IReleaseManagerCaller{contract: contract}, IReleaseManagerTransactor: IReleaseManagerTransactor{contract: contract}, IReleaseManagerFilterer: IReleaseManagerFilterer{contract: contract}}, nil
}

// NewIReleaseManagerCaller creates a new read-only instance of IReleaseManager, bound to a specific deployed contract.
func NewIReleaseManagerCaller(address common.Address, caller bind.ContractCaller) (*IReleaseManagerCaller, error) {
	contract, err := bindIReleaseManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerCaller{contract: contract}, nil
}

// NewIReleaseManagerTransactor creates a new write-only instance of IReleaseManager, bound to a specific deployed contract.
func NewIReleaseManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*IReleaseManagerTransactor, error) {
	contract, err := bindIReleaseManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerTransactor{contract: contract}, nil
}

// NewIReleaseManagerFilterer creates a new log filterer instance of IReleaseManager, bound to a specific deployed contract.
func NewIReleaseManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*IReleaseManagerFilterer, error) {
	contract, err := bindIReleaseManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerFilterer{contract: contract}, nil
}

// bindIReleaseManager binds a generic wrapper to an already deployed contract.
func bindIReleaseManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IReleaseManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IReleaseManager *IReleaseManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IReleaseManager.Contract.IReleaseManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IReleaseManager *IReleaseManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IReleaseManager.Contract.IReleaseManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IReleaseManager *IReleaseManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IReleaseManager.Contract.IReleaseManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IReleaseManager *IReleaseManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IReleaseManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IReleaseManager *IReleaseManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IReleaseManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IReleaseManager *IReleaseManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IReleaseManager.Contract.contract.Transact(opts, method, params...)
}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_IReleaseManager *IReleaseManagerCaller) GetArtifact(opts *bind.CallOpts, avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getArtifact", avs, digest)

	if err != nil {
		return *new(IReleaseManagerArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerArtifact)).(*IReleaseManagerArtifact)

	return out0, err

}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_IReleaseManager *IReleaseManagerSession) GetArtifact(avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	return _IReleaseManager.Contract.GetArtifact(&_IReleaseManager.CallOpts, avs, digest)
}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_IReleaseManager *IReleaseManagerCallerSession) GetArtifact(avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	return _IReleaseManager.Contract.GetArtifact(&_IReleaseManager.CallOpts, avs, digest)
}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_IReleaseManager *IReleaseManagerCaller) GetLatestPromotedArtifact(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getLatestPromotedArtifact", avs, operatorSetId)

	if err != nil {
		return *new(IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerPromotedArtifact)).(*IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_IReleaseManager *IReleaseManagerSession) GetLatestPromotedArtifact(avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetLatestPromotedArtifact(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_IReleaseManager *IReleaseManagerCallerSession) GetLatestPromotedArtifact(avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetLatestPromotedArtifact(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_IReleaseManager *IReleaseManagerCaller) GetPromotedArtifactAtBlock(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getPromotedArtifactAtBlock", avs, operatorSetId, blockNumber)

	if err != nil {
		return *new(IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerPromotedArtifact)).(*IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_IReleaseManager *IReleaseManagerSession) GetPromotedArtifactAtBlock(avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetPromotedArtifactAtBlock(&_IReleaseManager.CallOpts, avs, operatorSetId, blockNumber)
}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_IReleaseManager *IReleaseManagerCallerSession) GetPromotedArtifactAtBlock(avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetPromotedArtifactAtBlock(&_IReleaseManager.CallOpts, avs, operatorSetId, blockNumber)
}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_IReleaseManager *IReleaseManagerCaller) GetPromotedArtifacts(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getPromotedArtifacts", avs, operatorSetId)

	if err != nil {
		return *new([]IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]IReleaseManagerPromotedArtifact)).(*[]IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_IReleaseManager *IReleaseManagerSession) GetPromotedArtifacts(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetPromotedArtifacts(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_IReleaseManager *IReleaseManagerCallerSession) GetPromotedArtifacts(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetPromotedArtifacts(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_IReleaseManager *IReleaseManagerCaller) GetPromotionCheckpointAt(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getPromotionCheckpointAt", avs, operatorSetId, pos)

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
func (_IReleaseManager *IReleaseManagerSession) GetPromotionCheckpointAt(avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	return _IReleaseManager.Contract.GetPromotionCheckpointAt(&_IReleaseManager.CallOpts, avs, operatorSetId, pos)
}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_IReleaseManager *IReleaseManagerCallerSession) GetPromotionCheckpointAt(avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	return _IReleaseManager.Contract.GetPromotionCheckpointAt(&_IReleaseManager.CallOpts, avs, operatorSetId, pos)
}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_IReleaseManager *IReleaseManagerCaller) GetPromotionCheckpointCount(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getPromotionCheckpointCount", avs, operatorSetId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_IReleaseManager *IReleaseManagerSession) GetPromotionCheckpointCount(avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	return _IReleaseManager.Contract.GetPromotionCheckpointCount(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_IReleaseManager *IReleaseManagerCallerSession) GetPromotionCheckpointCount(avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	return _IReleaseManager.Contract.GetPromotionCheckpointCount(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_IReleaseManager *IReleaseManagerCaller) GetPromotionHistory(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getPromotionHistory", avs, operatorSetId)

	if err != nil {
		return *new([]IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]IReleaseManagerPromotedArtifact)).(*[]IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_IReleaseManager *IReleaseManagerSession) GetPromotionHistory(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetPromotionHistory(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_IReleaseManager *IReleaseManagerCallerSession) GetPromotionHistory(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _IReleaseManager.Contract.GetPromotionHistory(&_IReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_IReleaseManager *IReleaseManagerCaller) GetPromotionStatusAtBlock(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	var out []interface{}
	err := _IReleaseManager.contract.Call(opts, &out, "getPromotionStatusAtBlock", avs, operatorSetId, digest, blockNumber)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_IReleaseManager *IReleaseManagerSession) GetPromotionStatusAtBlock(avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	return _IReleaseManager.Contract.GetPromotionStatusAtBlock(&_IReleaseManager.CallOpts, avs, operatorSetId, digest, blockNumber)
}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_IReleaseManager *IReleaseManagerCallerSession) GetPromotionStatusAtBlock(avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	return _IReleaseManager.Contract.GetPromotionStatusAtBlock(&_IReleaseManager.CallOpts, avs, operatorSetId, digest, blockNumber)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_IReleaseManager *IReleaseManagerTransactor) Deregister(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _IReleaseManager.contract.Transact(opts, "deregister", avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_IReleaseManager *IReleaseManagerSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _IReleaseManager.Contract.Deregister(&_IReleaseManager.TransactOpts, avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_IReleaseManager *IReleaseManagerTransactorSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _IReleaseManager.Contract.Deregister(&_IReleaseManager.TransactOpts, avs)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_IReleaseManager *IReleaseManagerTransactor) PromoteArtifacts(opts *bind.TransactOpts, avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _IReleaseManager.contract.Transact(opts, "promoteArtifacts", avs, promotions, version, deploymentDeadline)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_IReleaseManager *IReleaseManagerSession) PromoteArtifacts(avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _IReleaseManager.Contract.PromoteArtifacts(&_IReleaseManager.TransactOpts, avs, promotions, version, deploymentDeadline)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_IReleaseManager *IReleaseManagerTransactorSession) PromoteArtifacts(avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _IReleaseManager.Contract.PromoteArtifacts(&_IReleaseManager.TransactOpts, avs, promotions, version, deploymentDeadline)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] artifacts) returns()
func (_IReleaseManager *IReleaseManagerTransactor) PublishArtifacts(opts *bind.TransactOpts, avs common.Address, artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _IReleaseManager.contract.Transact(opts, "publishArtifacts", avs, artifacts)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] artifacts) returns()
func (_IReleaseManager *IReleaseManagerSession) PublishArtifacts(avs common.Address, artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _IReleaseManager.Contract.PublishArtifacts(&_IReleaseManager.TransactOpts, avs, artifacts)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] artifacts) returns()
func (_IReleaseManager *IReleaseManagerTransactorSession) PublishArtifacts(avs common.Address, artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _IReleaseManager.Contract.PublishArtifacts(&_IReleaseManager.TransactOpts, avs, artifacts)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_IReleaseManager *IReleaseManagerTransactor) Register(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _IReleaseManager.contract.Transact(opts, "register", avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_IReleaseManager *IReleaseManagerSession) Register(avs common.Address) (*types.Transaction, error) {
	return _IReleaseManager.Contract.Register(&_IReleaseManager.TransactOpts, avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_IReleaseManager *IReleaseManagerTransactorSession) Register(avs common.Address) (*types.Transaction, error) {
	return _IReleaseManager.Contract.Register(&_IReleaseManager.TransactOpts, avs)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_IReleaseManager *IReleaseManagerTransactor) UpdatePromotionStatus(opts *bind.TransactOpts, avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _IReleaseManager.contract.Transact(opts, "updatePromotionStatus", avs, digest, operatorSetId, newStatus)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_IReleaseManager *IReleaseManagerSession) UpdatePromotionStatus(avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _IReleaseManager.Contract.UpdatePromotionStatus(&_IReleaseManager.TransactOpts, avs, digest, operatorSetId, newStatus)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_IReleaseManager *IReleaseManagerTransactorSession) UpdatePromotionStatus(avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _IReleaseManager.Contract.UpdatePromotionStatus(&_IReleaseManager.TransactOpts, avs, digest, operatorSetId, newStatus)
}

// IReleaseManagerAVSDeregisteredIterator is returned from FilterAVSDeregistered and is used to iterate over the raw logs and unpacked data for AVSDeregistered events raised by the IReleaseManager contract.
type IReleaseManagerAVSDeregisteredIterator struct {
	Event *IReleaseManagerAVSDeregistered // Event containing the contract specifics and raw log

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
func (it *IReleaseManagerAVSDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IReleaseManagerAVSDeregistered)
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
		it.Event = new(IReleaseManagerAVSDeregistered)
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
func (it *IReleaseManagerAVSDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IReleaseManagerAVSDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IReleaseManagerAVSDeregistered represents a AVSDeregistered event raised by the IReleaseManager contract.
type IReleaseManagerAVSDeregistered struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAVSDeregistered is a free log retrieval operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_IReleaseManager *IReleaseManagerFilterer) FilterAVSDeregistered(opts *bind.FilterOpts, avs []common.Address) (*IReleaseManagerAVSDeregisteredIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IReleaseManager.contract.FilterLogs(opts, "AVSDeregistered", avsRule)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerAVSDeregisteredIterator{contract: _IReleaseManager.contract, event: "AVSDeregistered", logs: logs, sub: sub}, nil
}

// WatchAVSDeregistered is a free log subscription operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_IReleaseManager *IReleaseManagerFilterer) WatchAVSDeregistered(opts *bind.WatchOpts, sink chan<- *IReleaseManagerAVSDeregistered, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IReleaseManager.contract.WatchLogs(opts, "AVSDeregistered", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IReleaseManagerAVSDeregistered)
				if err := _IReleaseManager.contract.UnpackLog(event, "AVSDeregistered", log); err != nil {
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
func (_IReleaseManager *IReleaseManagerFilterer) ParseAVSDeregistered(log types.Log) (*IReleaseManagerAVSDeregistered, error) {
	event := new(IReleaseManagerAVSDeregistered)
	if err := _IReleaseManager.contract.UnpackLog(event, "AVSDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IReleaseManagerAVSRegisteredIterator is returned from FilterAVSRegistered and is used to iterate over the raw logs and unpacked data for AVSRegistered events raised by the IReleaseManager contract.
type IReleaseManagerAVSRegisteredIterator struct {
	Event *IReleaseManagerAVSRegistered // Event containing the contract specifics and raw log

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
func (it *IReleaseManagerAVSRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IReleaseManagerAVSRegistered)
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
		it.Event = new(IReleaseManagerAVSRegistered)
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
func (it *IReleaseManagerAVSRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IReleaseManagerAVSRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IReleaseManagerAVSRegistered represents a AVSRegistered event raised by the IReleaseManager contract.
type IReleaseManagerAVSRegistered struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAVSRegistered is a free log retrieval operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_IReleaseManager *IReleaseManagerFilterer) FilterAVSRegistered(opts *bind.FilterOpts, avs []common.Address) (*IReleaseManagerAVSRegisteredIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IReleaseManager.contract.FilterLogs(opts, "AVSRegistered", avsRule)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerAVSRegisteredIterator{contract: _IReleaseManager.contract, event: "AVSRegistered", logs: logs, sub: sub}, nil
}

// WatchAVSRegistered is a free log subscription operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_IReleaseManager *IReleaseManagerFilterer) WatchAVSRegistered(opts *bind.WatchOpts, sink chan<- *IReleaseManagerAVSRegistered, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _IReleaseManager.contract.WatchLogs(opts, "AVSRegistered", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IReleaseManagerAVSRegistered)
				if err := _IReleaseManager.contract.UnpackLog(event, "AVSRegistered", log); err != nil {
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
func (_IReleaseManager *IReleaseManagerFilterer) ParseAVSRegistered(log types.Log) (*IReleaseManagerAVSRegistered, error) {
	event := new(IReleaseManagerAVSRegistered)
	if err := _IReleaseManager.contract.UnpackLog(event, "AVSRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IReleaseManagerArtifactPublishedIterator is returned from FilterArtifactPublished and is used to iterate over the raw logs and unpacked data for ArtifactPublished events raised by the IReleaseManager contract.
type IReleaseManagerArtifactPublishedIterator struct {
	Event *IReleaseManagerArtifactPublished // Event containing the contract specifics and raw log

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
func (it *IReleaseManagerArtifactPublishedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IReleaseManagerArtifactPublished)
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
		it.Event = new(IReleaseManagerArtifactPublished)
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
func (it *IReleaseManagerArtifactPublishedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IReleaseManagerArtifactPublishedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IReleaseManagerArtifactPublished represents a ArtifactPublished event raised by the IReleaseManager contract.
type IReleaseManagerArtifactPublished struct {
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
func (_IReleaseManager *IReleaseManagerFilterer) FilterArtifactPublished(opts *bind.FilterOpts, avs []common.Address, digest [][32]byte) (*IReleaseManagerArtifactPublishedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _IReleaseManager.contract.FilterLogs(opts, "ArtifactPublished", avsRule, digestRule)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerArtifactPublishedIterator{contract: _IReleaseManager.contract, event: "ArtifactPublished", logs: logs, sub: sub}, nil
}

// WatchArtifactPublished is a free log subscription operation binding the contract event 0x622e1c25f0e4dcedadb24c8f0bbe6ceaa3776cbeb58b17a2d5d8ac8ae31a7502.
//
// Solidity: event ArtifactPublished(address indexed avs, bytes32 indexed digest, string registryUrl, uint8 architecture, uint8 os, uint8 artifactType)
func (_IReleaseManager *IReleaseManagerFilterer) WatchArtifactPublished(opts *bind.WatchOpts, sink chan<- *IReleaseManagerArtifactPublished, avs []common.Address, digest [][32]byte) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _IReleaseManager.contract.WatchLogs(opts, "ArtifactPublished", avsRule, digestRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IReleaseManagerArtifactPublished)
				if err := _IReleaseManager.contract.UnpackLog(event, "ArtifactPublished", log); err != nil {
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
func (_IReleaseManager *IReleaseManagerFilterer) ParseArtifactPublished(log types.Log) (*IReleaseManagerArtifactPublished, error) {
	event := new(IReleaseManagerArtifactPublished)
	if err := _IReleaseManager.contract.UnpackLog(event, "ArtifactPublished", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IReleaseManagerArtifactsPromotedIterator is returned from FilterArtifactsPromoted and is used to iterate over the raw logs and unpacked data for ArtifactsPromoted events raised by the IReleaseManager contract.
type IReleaseManagerArtifactsPromotedIterator struct {
	Event *IReleaseManagerArtifactsPromoted // Event containing the contract specifics and raw log

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
func (it *IReleaseManagerArtifactsPromotedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IReleaseManagerArtifactsPromoted)
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
		it.Event = new(IReleaseManagerArtifactsPromoted)
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
func (it *IReleaseManagerArtifactsPromotedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IReleaseManagerArtifactsPromotedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IReleaseManagerArtifactsPromoted represents a ArtifactsPromoted event raised by the IReleaseManager contract.
type IReleaseManagerArtifactsPromoted struct {
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
func (_IReleaseManager *IReleaseManagerFilterer) FilterArtifactsPromoted(opts *bind.FilterOpts, avs []common.Address, version []string) (*IReleaseManagerArtifactsPromotedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}

	logs, sub, err := _IReleaseManager.contract.FilterLogs(opts, "ArtifactsPromoted", avsRule, versionRule)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerArtifactsPromotedIterator{contract: _IReleaseManager.contract, event: "ArtifactsPromoted", logs: logs, sub: sub}, nil
}

// WatchArtifactsPromoted is a free log subscription operation binding the contract event 0x7fd55cb6307da041fb4711b42eb59c940bb12e76dc208c9be19be9abde37815d.
//
// Solidity: event ArtifactsPromoted(address indexed avs, string indexed version, uint256 deploymentDeadline, bytes32[] digests, uint8[] statuses)
func (_IReleaseManager *IReleaseManagerFilterer) WatchArtifactsPromoted(opts *bind.WatchOpts, sink chan<- *IReleaseManagerArtifactsPromoted, avs []common.Address, version []string) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}

	logs, sub, err := _IReleaseManager.contract.WatchLogs(opts, "ArtifactsPromoted", avsRule, versionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IReleaseManagerArtifactsPromoted)
				if err := _IReleaseManager.contract.UnpackLog(event, "ArtifactsPromoted", log); err != nil {
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
func (_IReleaseManager *IReleaseManagerFilterer) ParseArtifactsPromoted(log types.Log) (*IReleaseManagerArtifactsPromoted, error) {
	event := new(IReleaseManagerArtifactsPromoted)
	if err := _IReleaseManager.contract.UnpackLog(event, "ArtifactsPromoted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// IReleaseManagerPromotionStatusUpdatedIterator is returned from FilterPromotionStatusUpdated and is used to iterate over the raw logs and unpacked data for PromotionStatusUpdated events raised by the IReleaseManager contract.
type IReleaseManagerPromotionStatusUpdatedIterator struct {
	Event *IReleaseManagerPromotionStatusUpdated // Event containing the contract specifics and raw log

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
func (it *IReleaseManagerPromotionStatusUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IReleaseManagerPromotionStatusUpdated)
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
		it.Event = new(IReleaseManagerPromotionStatusUpdated)
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
func (it *IReleaseManagerPromotionStatusUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IReleaseManagerPromotionStatusUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IReleaseManagerPromotionStatusUpdated represents a PromotionStatusUpdated event raised by the IReleaseManager contract.
type IReleaseManagerPromotionStatusUpdated struct {
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
func (_IReleaseManager *IReleaseManagerFilterer) FilterPromotionStatusUpdated(opts *bind.FilterOpts, avs []common.Address, digest [][32]byte, operatorSetId [][32]byte) (*IReleaseManagerPromotionStatusUpdatedIterator, error) {

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

	logs, sub, err := _IReleaseManager.contract.FilterLogs(opts, "PromotionStatusUpdated", avsRule, digestRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &IReleaseManagerPromotionStatusUpdatedIterator{contract: _IReleaseManager.contract, event: "PromotionStatusUpdated", logs: logs, sub: sub}, nil
}

// WatchPromotionStatusUpdated is a free log subscription operation binding the contract event 0x1341ac2513ee1dd81316f76f3a7840608cd4dc5a3302d2fb45ea1ad24d355671.
//
// Solidity: event PromotionStatusUpdated(address indexed avs, bytes32 indexed digest, bytes32 indexed operatorSetId, uint8 oldStatus, uint8 newStatus)
func (_IReleaseManager *IReleaseManagerFilterer) WatchPromotionStatusUpdated(opts *bind.WatchOpts, sink chan<- *IReleaseManagerPromotionStatusUpdated, avs []common.Address, digest [][32]byte, operatorSetId [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _IReleaseManager.contract.WatchLogs(opts, "PromotionStatusUpdated", avsRule, digestRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IReleaseManagerPromotionStatusUpdated)
				if err := _IReleaseManager.contract.UnpackLog(event, "PromotionStatusUpdated", log); err != nil {
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
func (_IReleaseManager *IReleaseManagerFilterer) ParsePromotionStatusUpdated(log types.Log) (*IReleaseManagerPromotionStatusUpdated, error) {
	event := new(IReleaseManagerPromotionStatusUpdated)
	if err := _IReleaseManager.contract.UnpackLog(event, "PromotionStatusUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
