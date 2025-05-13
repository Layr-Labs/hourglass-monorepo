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

// ITaskAVSRegistrarTypesPubkeyRegistrationParams is an auto generated low-level Go binding around an user-defined struct.
type ITaskAVSRegistrarTypesPubkeyRegistrationParams struct {
	PubkeyRegistrationSignature BN254G1Point
	PubkeyG1                    BN254G1Point
	PubkeyG2                    BN254G2Point
}

// TaskAVSRegistrarBaseStorageMetaData contains all meta data concerning the TaskAVSRegistrarBaseStorage contract.
var TaskAVSRegistrarBaseStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"ALLOCATION_MANAGER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"AVS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PUBKEY_REGISTRATION_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculatePubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"currentApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getBatchOperatorPubkeyInfoAndSocket\",\"inputs\":[{\"name\":\"operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfoAndSocket[]\",\"components\":[{\"name\":\"pubkeyInfo\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFromPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyG2\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkeyInfo\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"packRegisterPayload\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"pubkeyRegistrationParams\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyRegistrationParams\",\"components\":[{\"name\":\"pubkeyRegistrationSignature\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"pubkeyHashToOperator\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyHashToSocket\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateOperatorSocket\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"NewPubkeyRegistration\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSetApkUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"apk\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"BLSPubkeyAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAVS\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBLSSignatureOrPrivateKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyAllocationManager\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroPubKey\",\"inputs\":[]}]",
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

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) ALLOCATIONMANAGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "ALLOCATION_MANAGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) AVS(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "AVS")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.AVS(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.AVS(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) PUBKEYREGISTRATIONTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "PUBKEY_REGISTRATION_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrarBaseStorage.CallOpts)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) CalculatePubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "calculatePubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) CurrentApk(opts *bind.CallOpts, operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "currentApk", operatorSetId)

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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) CurrentApk(operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.CurrentApk(&_TaskAVSRegistrarBaseStorage.CallOpts, operatorSetId)
}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) CurrentApk(operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.CurrentApk(&_TaskAVSRegistrarBaseStorage.CallOpts, operatorSetId)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetApk(opts *bind.CallOpts, operatorSetId uint8) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getApk", operatorSetId)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetApk(&_TaskAVSRegistrarBaseStorage.CallOpts, operatorSetId)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetApk(&_TaskAVSRegistrarBaseStorage.CallOpts, operatorSetId)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetBatchOperatorPubkeyInfoAndSocket(opts *bind.CallOpts, operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getBatchOperatorPubkeyInfoAndSocket", operators)

	if err != nil {
		return *new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket), err
	}

	out0 := *abi.ConvertType(out[0], new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)).(*[]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)

	return out0, err

}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, operators)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, operators)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getOperatorFromPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetOperatorPubkeyG2(opts *bind.CallOpts, operator common.Address) (BN254G2Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getOperatorPubkeyG2", operator)

	if err != nil {
		return *new(BN254G2Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G2Point)).(*BN254G2Point)

	return out0, err

}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorPubkeyG2(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorPubkeyG2(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetOperatorPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getOperatorPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetOperatorSocketByOperator(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getOperatorSocketByOperator", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetOperatorSocketByPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getOperatorSocketByPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, [32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getRegisteredPubkey", operator)

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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetRegisteredPubkey(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetRegisteredPubkey(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) GetRegisteredPubkeyInfo(opts *bind.CallOpts, operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "getRegisteredPubkeyInfo", operator)

	if err != nil {
		return *new(ITaskAVSRegistrarTypesPubkeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskAVSRegistrarTypesPubkeyInfo)).(*ITaskAVSRegistrarTypesPubkeyInfo)

	return out0, err

}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetRegisteredPubkeyInfo(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.GetRegisteredPubkeyInfo(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) OperatorToPubkey(opts *bind.CallOpts, operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "operatorToPubkey", operator)

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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.OperatorToPubkey(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.OperatorToPubkey(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) OperatorToPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "operatorToPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) OperatorToSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "operatorToSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.OperatorToSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.OperatorToSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) PackRegisterPayload(opts *bind.CallOpts, socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "packRegisterPayload", socket, pubkeyRegistrationParams)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) PackRegisterPayload(socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PackRegisterPayload(&_TaskAVSRegistrarBaseStorage.CallOpts, socket, pubkeyRegistrationParams)
}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) PackRegisterPayload(socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PackRegisterPayload(&_TaskAVSRegistrarBaseStorage.CallOpts, socket, pubkeyRegistrationParams)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) PubkeyHashToOperator(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "pubkeyHashToOperator", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PubkeyHashToOperator(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PubkeyHashToOperator(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) PubkeyHashToSocket(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "pubkeyHashToSocket", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PubkeyHashToSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PubkeyHashToSocket(&_TaskAVSRegistrarBaseStorage.CallOpts, pubkeyHash)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) PubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "pubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrarBaseStorage.CallOpts, operator)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCaller) SupportsAVS(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBaseStorage.contract.Call(opts, &out, "supportsAVS", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.SupportsAVS(&_TaskAVSRegistrarBaseStorage.CallOpts, avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageCallerSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.SupportsAVS(&_TaskAVSRegistrarBaseStorage.CallOpts, avs)
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

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactor) UpdateOperatorSocket(opts *bind.TransactOpts, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.contract.Transact(opts, "updateOperatorSocket", socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.UpdateOperatorSocket(&_TaskAVSRegistrarBaseStorage.TransactOpts, socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageTransactorSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBaseStorage.Contract.UpdateOperatorSocket(&_TaskAVSRegistrarBaseStorage.TransactOpts, socket)
}

// TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator is returned from FilterNewPubkeyRegistration and is used to iterate over the raw logs and unpacked data for NewPubkeyRegistration events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator struct {
	Event *TaskAVSRegistrarBaseStorageNewPubkeyRegistration // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageNewPubkeyRegistration)
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
		it.Event = new(TaskAVSRegistrarBaseStorageNewPubkeyRegistration)
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
func (it *TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageNewPubkeyRegistration represents a NewPubkeyRegistration event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageNewPubkeyRegistration struct {
	Operator   common.Address
	PubkeyHash [32]byte
	PubkeyG1   BN254G1Point
	PubkeyG2   BN254G2Point
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewPubkeyRegistration is a free log retrieval operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageNewPubkeyRegistrationIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "NewPubkeyRegistration", logs: logs, sub: sub}, nil
}

// WatchNewPubkeyRegistration is a free log subscription operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageNewPubkeyRegistration, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageNewPubkeyRegistration)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseNewPubkeyRegistration(log types.Log) (*TaskAVSRegistrarBaseStorageNewPubkeyRegistration, error) {
	event := new(TaskAVSRegistrarBaseStorageNewPubkeyRegistration)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator is returned from FilterOperatorSetApkUpdated and is used to iterate over the raw logs and unpacked data for OperatorSetApkUpdated events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator struct {
	Event *TaskAVSRegistrarBaseStorageOperatorSetApkUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageOperatorSetApkUpdated)
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
		it.Event = new(TaskAVSRegistrarBaseStorageOperatorSetApkUpdated)
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
func (it *TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageOperatorSetApkUpdated represents a OperatorSetApkUpdated event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorSetApkUpdated struct {
	Operator      common.Address
	PubkeyHash    [32]byte
	OperatorSetId uint32
	Apk           BN254G1Point
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetApkUpdated is a free log retrieval operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterOperatorSetApkUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (*TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator, error) {

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

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageOperatorSetApkUpdatedIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "OperatorSetApkUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSetApkUpdated is a free log subscription operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchOperatorSetApkUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageOperatorSetApkUpdated, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (event.Subscription, error) {

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

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageOperatorSetApkUpdated)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseOperatorSetApkUpdated(log types.Log) (*TaskAVSRegistrarBaseStorageOperatorSetApkUpdated, error) {
	event := new(TaskAVSRegistrarBaseStorageOperatorSetApkUpdated)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator is returned from FilterOperatorSocketUpdated and is used to iterate over the raw logs and unpacked data for OperatorSocketUpdated events raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator struct {
	Event *TaskAVSRegistrarBaseStorageOperatorSocketUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseStorageOperatorSocketUpdated)
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
		it.Event = new(TaskAVSRegistrarBaseStorageOperatorSocketUpdated)
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
func (it *TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseStorageOperatorSocketUpdated represents a OperatorSocketUpdated event raised by the TaskAVSRegistrarBaseStorage contract.
type TaskAVSRegistrarBaseStorageOperatorSocketUpdated struct {
	Operator   common.Address
	PubkeyHash [32]byte
	Socket     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketUpdated is a free log retrieval operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) FilterOperatorSocketUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.FilterLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseStorageOperatorSocketUpdatedIterator{contract: _TaskAVSRegistrarBaseStorage.contract, event: "OperatorSocketUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketUpdated is a free log subscription operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) WatchOperatorSocketUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseStorageOperatorSocketUpdated, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBaseStorage.contract.WatchLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseStorageOperatorSocketUpdated)
				if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
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
func (_TaskAVSRegistrarBaseStorage *TaskAVSRegistrarBaseStorageFilterer) ParseOperatorSocketUpdated(log types.Log) (*TaskAVSRegistrarBaseStorageOperatorSocketUpdated, error) {
	event := new(TaskAVSRegistrarBaseStorageOperatorSocketUpdated)
	if err := _TaskAVSRegistrarBaseStorage.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
