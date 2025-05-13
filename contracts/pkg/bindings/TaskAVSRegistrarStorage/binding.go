// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package TaskAVSRegistrarStorage

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

// ITaskAVSRegistrarTypesPubkeyInfo is an auto generated low-level Go binding around an user-defined struct.
type ITaskAVSRegistrarTypesPubkeyInfo struct {
	PubkeyG1   BN254G1Point
	PubkeyG2   BN254G2Point
	PubkeyHash [32]byte
}

// ITaskAVSRegistrarTypesPubkeyInfoAndSocket is an auto generated low-level Go binding around an user-defined struct.
type ITaskAVSRegistrarTypesPubkeyInfoAndSocket struct {
	PubkeyInfo ITaskAVSRegistrarTypesPubkeyInfo
	Socket     string
}

// TaskAVSRegistrarStorageMetaData contains all meta data concerning the TaskAVSRegistrarStorage contract.
var TaskAVSRegistrarStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"ALLOCATION_MANAGER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"AVS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PUBKEY_REGISTRATION_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculatePubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"currentApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getBatchOperatorPubkeyInfoAndSocket\",\"inputs\":[{\"name\":\"operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfoAndSocket[]\",\"components\":[{\"name\":\"pubkeyInfo\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFromPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyG2\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkeyInfo\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyHashToOperator\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyHashToSocket\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateOperatorSocket\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"NewPubkeyRegistration\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSetApkUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"apk\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"BLSPubkeyAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAVS\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBLSSignatureOrPrivateKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyAllocationManager\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroPubKey\",\"inputs\":[]}]",
}

// TaskAVSRegistrarStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use TaskAVSRegistrarStorageMetaData.ABI instead.
var TaskAVSRegistrarStorageABI = TaskAVSRegistrarStorageMetaData.ABI

// TaskAVSRegistrarStorage is an auto generated Go binding around an Ethereum contract.
type TaskAVSRegistrarStorage struct {
	TaskAVSRegistrarStorageCaller     // Read-only binding to the contract
	TaskAVSRegistrarStorageTransactor // Write-only binding to the contract
	TaskAVSRegistrarStorageFilterer   // Log filterer for contract events
}

// TaskAVSRegistrarStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TaskAVSRegistrarStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TaskAVSRegistrarStorageSession struct {
	Contract     *TaskAVSRegistrarStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TaskAVSRegistrarStorageCallerSession struct {
	Contract *TaskAVSRegistrarStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// TaskAVSRegistrarStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TaskAVSRegistrarStorageTransactorSession struct {
	Contract     *TaskAVSRegistrarStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type TaskAVSRegistrarStorageRaw struct {
	Contract *TaskAVSRegistrarStorage // Generic contract binding to access the raw methods on
}

// TaskAVSRegistrarStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarStorageCallerRaw struct {
	Contract *TaskAVSRegistrarStorageCaller // Generic read-only contract binding to access the raw methods on
}

// TaskAVSRegistrarStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarStorageTransactorRaw struct {
	Contract *TaskAVSRegistrarStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTaskAVSRegistrarStorage creates a new instance of TaskAVSRegistrarStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarStorage(address common.Address, backend bind.ContractBackend) (*TaskAVSRegistrarStorage, error) {
	contract, err := bindTaskAVSRegistrarStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorage{TaskAVSRegistrarStorageCaller: TaskAVSRegistrarStorageCaller{contract: contract}, TaskAVSRegistrarStorageTransactor: TaskAVSRegistrarStorageTransactor{contract: contract}, TaskAVSRegistrarStorageFilterer: TaskAVSRegistrarStorageFilterer{contract: contract}}, nil
}

// NewTaskAVSRegistrarStorageCaller creates a new read-only instance of TaskAVSRegistrarStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarStorageCaller(address common.Address, caller bind.ContractCaller) (*TaskAVSRegistrarStorageCaller, error) {
	contract, err := bindTaskAVSRegistrarStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorageCaller{contract: contract}, nil
}

// NewTaskAVSRegistrarStorageTransactor creates a new write-only instance of TaskAVSRegistrarStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*TaskAVSRegistrarStorageTransactor, error) {
	contract, err := bindTaskAVSRegistrarStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorageTransactor{contract: contract}, nil
}

// NewTaskAVSRegistrarStorageFilterer creates a new log filterer instance of TaskAVSRegistrarStorage, bound to a specific deployed contract.
func NewTaskAVSRegistrarStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*TaskAVSRegistrarStorageFilterer, error) {
	contract, err := bindTaskAVSRegistrarStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorageFilterer{contract: contract}, nil
}

// bindTaskAVSRegistrarStorage binds a generic wrapper to an already deployed contract.
func bindTaskAVSRegistrarStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TaskAVSRegistrarStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrarStorage.Contract.TaskAVSRegistrarStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.TaskAVSRegistrarStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.TaskAVSRegistrarStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrarStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.contract.Transact(opts, method, params...)
}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) ALLOCATIONMANAGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "ALLOCATION_MANAGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrarStorage.CallOpts)
}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrarStorage.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) AVS(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "AVS")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.AVS(&_TaskAVSRegistrarStorage.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.AVS(&_TaskAVSRegistrarStorage.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) PUBKEYREGISTRATIONTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "PUBKEY_REGISTRATION_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrarStorage.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrarStorage.CallOpts)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) CalculatePubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "calculatePubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) CurrentApk(opts *bind.CallOpts, operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "currentApk", operatorSetId)

	outstruct := new(struct {
		X *big.Int
		Y *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.X = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Y = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) CurrentApk(operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarStorage.Contract.CurrentApk(&_TaskAVSRegistrarStorage.CallOpts, operatorSetId)
}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) CurrentApk(operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarStorage.Contract.CurrentApk(&_TaskAVSRegistrarStorage.CallOpts, operatorSetId)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetApk(opts *bind.CallOpts, operatorSetId uint8) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getApk", operatorSetId)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _TaskAVSRegistrarStorage.Contract.GetApk(&_TaskAVSRegistrarStorage.CallOpts, operatorSetId)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _TaskAVSRegistrarStorage.Contract.GetApk(&_TaskAVSRegistrarStorage.CallOpts, operatorSetId)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetBatchOperatorPubkeyInfoAndSocket(opts *bind.CallOpts, operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getBatchOperatorPubkeyInfoAndSocket", operators)

	if err != nil {
		return *new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket), err
	}

	out0 := *abi.ConvertType(out[0], new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)).(*[]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)

	return out0, err

}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _TaskAVSRegistrarStorage.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_TaskAVSRegistrarStorage.CallOpts, operators)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _TaskAVSRegistrarStorage.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_TaskAVSRegistrarStorage.CallOpts, operators)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getOperatorFromPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetOperatorPubkeyG2(opts *bind.CallOpts, operator common.Address) (BN254G2Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getOperatorPubkeyG2", operator)

	if err != nil {
		return *new(BN254G2Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G2Point)).(*BN254G2Point)

	return out0, err

}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorPubkeyG2(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorPubkeyG2(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetOperatorPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getOperatorPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetOperatorSocketByOperator(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getOperatorSocketByOperator", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetOperatorSocketByPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getOperatorSocketByPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, [32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getRegisteredPubkey", operator)

	if err != nil {
		return *new(BN254G1Point), *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)
	out1 := *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)

	return out0, out1, err

}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.GetRegisteredPubkey(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.GetRegisteredPubkey(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) GetRegisteredPubkeyInfo(opts *bind.CallOpts, operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "getRegisteredPubkeyInfo", operator)

	if err != nil {
		return *new(ITaskAVSRegistrarTypesPubkeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskAVSRegistrarTypesPubkeyInfo)).(*ITaskAVSRegistrarTypesPubkeyInfo)

	return out0, err

}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _TaskAVSRegistrarStorage.Contract.GetRegisteredPubkeyInfo(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _TaskAVSRegistrarStorage.Contract.GetRegisteredPubkeyInfo(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) OperatorToPubkey(opts *bind.CallOpts, operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "operatorToPubkey", operator)

	outstruct := new(struct {
		X *big.Int
		Y *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.X = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Y = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarStorage.Contract.OperatorToPubkey(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarStorage.Contract.OperatorToPubkey(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) OperatorToPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "operatorToPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarStorage.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) OperatorToSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "operatorToSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.OperatorToSocket(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.OperatorToSocket(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) PubkeyHashToOperator(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "pubkeyHashToOperator", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.PubkeyHashToOperator(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarStorage.Contract.PubkeyHashToOperator(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) PubkeyHashToSocket(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "pubkeyHashToSocket", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.PubkeyHashToSocket(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarStorage.Contract.PubkeyHashToSocket(&_TaskAVSRegistrarStorage.CallOpts, pubkeyHash)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) PubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "pubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrarStorage.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrarStorage.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrarStorage.CallOpts, operator)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCaller) SupportsAVS(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _TaskAVSRegistrarStorage.contract.Call(opts, &out, "supportsAVS", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrarStorage.Contract.SupportsAVS(&_TaskAVSRegistrarStorage.CallOpts, avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageCallerSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrarStorage.Contract.SupportsAVS(&_TaskAVSRegistrarStorage.CallOpts, avs)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactor) DeregisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.contract.Transact(opts, "deregisterOperator", operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.DeregisterOperator(&_TaskAVSRegistrarStorage.TransactOpts, operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactorSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.DeregisterOperator(&_TaskAVSRegistrarStorage.TransactOpts, operator, avs, operatorSetIds)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.contract.Transact(opts, "registerOperator", operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.RegisterOperator(&_TaskAVSRegistrarStorage.TransactOpts, operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactorSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.RegisterOperator(&_TaskAVSRegistrarStorage.TransactOpts, operator, avs, operatorSetIds, data)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactor) UpdateOperatorSocket(opts *bind.TransactOpts, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.contract.Transact(opts, "updateOperatorSocket", socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.UpdateOperatorSocket(&_TaskAVSRegistrarStorage.TransactOpts, socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageTransactorSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarStorage.Contract.UpdateOperatorSocket(&_TaskAVSRegistrarStorage.TransactOpts, socket)
}

// TaskAVSRegistrarStorageNewPubkeyRegistrationIterator is returned from FilterNewPubkeyRegistration and is used to iterate over the raw logs and unpacked data for NewPubkeyRegistration events raised by the TaskAVSRegistrarStorage contract.
type TaskAVSRegistrarStorageNewPubkeyRegistrationIterator struct {
	Event *TaskAVSRegistrarStorageNewPubkeyRegistration // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarStorageNewPubkeyRegistrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarStorageNewPubkeyRegistration)
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
		it.Event = new(TaskAVSRegistrarStorageNewPubkeyRegistration)
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
func (it *TaskAVSRegistrarStorageNewPubkeyRegistrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarStorageNewPubkeyRegistrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarStorageNewPubkeyRegistration represents a NewPubkeyRegistration event raised by the TaskAVSRegistrarStorage contract.
type TaskAVSRegistrarStorageNewPubkeyRegistration struct {
	Operator   common.Address
	PubkeyHash [32]byte
	PubkeyG1   BN254G1Point
	PubkeyG2   BN254G2Point
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewPubkeyRegistration is a free log retrieval operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarStorageNewPubkeyRegistrationIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarStorage.contract.FilterLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorageNewPubkeyRegistrationIterator{contract: _TaskAVSRegistrarStorage.contract, event: "NewPubkeyRegistration", logs: logs, sub: sub}, nil
}

// WatchNewPubkeyRegistration is a free log subscription operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarStorageNewPubkeyRegistration, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarStorage.contract.WatchLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarStorageNewPubkeyRegistration)
				if err := _TaskAVSRegistrarStorage.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
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

// ParseNewPubkeyRegistration is a log parse operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) ParseNewPubkeyRegistration(log types.Log) (*TaskAVSRegistrarStorageNewPubkeyRegistration, error) {
	event := new(TaskAVSRegistrarStorageNewPubkeyRegistration)
	if err := _TaskAVSRegistrarStorage.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator is returned from FilterOperatorSetApkUpdated and is used to iterate over the raw logs and unpacked data for OperatorSetApkUpdated events raised by the TaskAVSRegistrarStorage contract.
type TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator struct {
	Event *TaskAVSRegistrarStorageOperatorSetApkUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarStorageOperatorSetApkUpdated)
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
		it.Event = new(TaskAVSRegistrarStorageOperatorSetApkUpdated)
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
func (it *TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarStorageOperatorSetApkUpdated represents a OperatorSetApkUpdated event raised by the TaskAVSRegistrarStorage contract.
type TaskAVSRegistrarStorageOperatorSetApkUpdated struct {
	Operator      common.Address
	PubkeyHash    [32]byte
	OperatorSetId uint32
	Apk           BN254G1Point
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetApkUpdated is a free log retrieval operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) FilterOperatorSetApkUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (*TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _TaskAVSRegistrarStorage.contract.FilterLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorageOperatorSetApkUpdatedIterator{contract: _TaskAVSRegistrarStorage.contract, event: "OperatorSetApkUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSetApkUpdated is a free log subscription operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) WatchOperatorSetApkUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarStorageOperatorSetApkUpdated, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _TaskAVSRegistrarStorage.contract.WatchLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarStorageOperatorSetApkUpdated)
				if err := _TaskAVSRegistrarStorage.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
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

// ParseOperatorSetApkUpdated is a log parse operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) ParseOperatorSetApkUpdated(log types.Log) (*TaskAVSRegistrarStorageOperatorSetApkUpdated, error) {
	event := new(TaskAVSRegistrarStorageOperatorSetApkUpdated)
	if err := _TaskAVSRegistrarStorage.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarStorageOperatorSocketUpdatedIterator is returned from FilterOperatorSocketUpdated and is used to iterate over the raw logs and unpacked data for OperatorSocketUpdated events raised by the TaskAVSRegistrarStorage contract.
type TaskAVSRegistrarStorageOperatorSocketUpdatedIterator struct {
	Event *TaskAVSRegistrarStorageOperatorSocketUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarStorageOperatorSocketUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarStorageOperatorSocketUpdated)
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
		it.Event = new(TaskAVSRegistrarStorageOperatorSocketUpdated)
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
func (it *TaskAVSRegistrarStorageOperatorSocketUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarStorageOperatorSocketUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarStorageOperatorSocketUpdated represents a OperatorSocketUpdated event raised by the TaskAVSRegistrarStorage contract.
type TaskAVSRegistrarStorageOperatorSocketUpdated struct {
	Operator   common.Address
	PubkeyHash [32]byte
	Socket     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketUpdated is a free log retrieval operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) FilterOperatorSocketUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarStorageOperatorSocketUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarStorage.contract.FilterLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarStorageOperatorSocketUpdatedIterator{contract: _TaskAVSRegistrarStorage.contract, event: "OperatorSocketUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketUpdated is a free log subscription operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) WatchOperatorSocketUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarStorageOperatorSocketUpdated, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarStorage.contract.WatchLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarStorageOperatorSocketUpdated)
				if err := _TaskAVSRegistrarStorage.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
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

// ParseOperatorSocketUpdated is a log parse operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarStorage *TaskAVSRegistrarStorageFilterer) ParseOperatorSocketUpdated(log types.Log) (*TaskAVSRegistrarStorageOperatorSocketUpdated, error) {
	event := new(TaskAVSRegistrarStorageOperatorSocketUpdated)
	if err := _TaskAVSRegistrarStorage.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
