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

// TaskAVSRegistrarBaseMetaData contains all meta data concerning the TaskAVSRegistrarBase contract.
var TaskAVSRegistrarBaseMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"ALLOCATION_MANAGER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"AVS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PUBKEY_REGISTRATION_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculatePubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"currentApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getBatchOperatorPubkeyInfoAndSocket\",\"inputs\":[{\"name\":\"operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfoAndSocket[]\",\"components\":[{\"name\":\"pubkeyInfo\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFromPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyG2\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkeyInfo\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"packRegisterPayload\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"pubkeyRegistrationParams\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyRegistrationParams\",\"components\":[{\"name\":\"pubkeyRegistrationSignature\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"pubkeyHashToOperator\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyHashToSocket\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateOperatorSocket\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NewPubkeyRegistration\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSetApkUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"apk\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"BLSPubkeyAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECAddFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECMulFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECPairingFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpModFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAVS\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBLSSignatureOrPrivateKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyAllocationManager\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"ZeroPubKey\",\"inputs\":[]}]",
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

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) ALLOCATIONMANAGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "ALLOCATION_MANAGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrarBase.CallOpts)
}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrarBase.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) AVS(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "AVS")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.AVS(&_TaskAVSRegistrarBase.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.AVS(&_TaskAVSRegistrarBase.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) PUBKEYREGISTRATIONTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "PUBKEY_REGISTRATION_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrarBase.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrarBase.CallOpts)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) CalculatePubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "calculatePubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) CurrentApk(opts *bind.CallOpts, operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "currentApk", operatorSetId)

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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) CurrentApk(operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBase.Contract.CurrentApk(&_TaskAVSRegistrarBase.CallOpts, operatorSetId)
}

// CurrentApk is a free data retrieval call binding the contract method 0x7d04529a.
//
// Solidity: function currentApk(uint32 operatorSetId) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) CurrentApk(operatorSetId uint32) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBase.Contract.CurrentApk(&_TaskAVSRegistrarBase.CallOpts, operatorSetId)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _TaskAVSRegistrarBase.Contract.Eip712Domain(&_TaskAVSRegistrarBase.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _TaskAVSRegistrarBase.Contract.Eip712Domain(&_TaskAVSRegistrarBase.CallOpts)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetApk(opts *bind.CallOpts, operatorSetId uint8) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getApk", operatorSetId)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _TaskAVSRegistrarBase.Contract.GetApk(&_TaskAVSRegistrarBase.CallOpts, operatorSetId)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _TaskAVSRegistrarBase.Contract.GetApk(&_TaskAVSRegistrarBase.CallOpts, operatorSetId)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetBatchOperatorPubkeyInfoAndSocket(opts *bind.CallOpts, operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getBatchOperatorPubkeyInfoAndSocket", operators)

	if err != nil {
		return *new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket), err
	}

	out0 := *abi.ConvertType(out[0], new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)).(*[]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)

	return out0, err

}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _TaskAVSRegistrarBase.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_TaskAVSRegistrarBase.CallOpts, operators)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _TaskAVSRegistrarBase.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_TaskAVSRegistrarBase.CallOpts, operators)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorFromPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetOperatorPubkeyG2(opts *bind.CallOpts, operator common.Address) (BN254G2Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorPubkeyG2", operator)

	if err != nil {
		return *new(BN254G2Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G2Point)).(*BN254G2Point)

	return out0, err

}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorPubkeyG2(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorPubkeyG2(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetOperatorPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetOperatorSocketByOperator(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorSocketByOperator", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetOperatorSocketByPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getOperatorSocketByPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBase.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, [32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getRegisteredPubkey", operator)

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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.GetRegisteredPubkey(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.GetRegisteredPubkey(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) GetRegisteredPubkeyInfo(opts *bind.CallOpts, operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "getRegisteredPubkeyInfo", operator)

	if err != nil {
		return *new(ITaskAVSRegistrarTypesPubkeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskAVSRegistrarTypesPubkeyInfo)).(*ITaskAVSRegistrarTypesPubkeyInfo)

	return out0, err

}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _TaskAVSRegistrarBase.Contract.GetRegisteredPubkeyInfo(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _TaskAVSRegistrarBase.Contract.GetRegisteredPubkeyInfo(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) OperatorToPubkey(opts *bind.CallOpts, operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "operatorToPubkey", operator)

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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBase.Contract.OperatorToPubkey(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrarBase.Contract.OperatorToPubkey(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) OperatorToPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "operatorToPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrarBase.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) OperatorToSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "operatorToSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBase.Contract.OperatorToSocket(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrarBase.Contract.OperatorToSocket(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) PackRegisterPayload(opts *bind.CallOpts, socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "packRegisterPayload", socket, pubkeyRegistrationParams)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) PackRegisterPayload(socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	return _TaskAVSRegistrarBase.Contract.PackRegisterPayload(&_TaskAVSRegistrarBase.CallOpts, socket, pubkeyRegistrationParams)
}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) PackRegisterPayload(socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	return _TaskAVSRegistrarBase.Contract.PackRegisterPayload(&_TaskAVSRegistrarBase.CallOpts, socket, pubkeyRegistrationParams)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) PubkeyHashToOperator(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "pubkeyHashToOperator", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.PubkeyHashToOperator(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrarBase.Contract.PubkeyHashToOperator(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) PubkeyHashToSocket(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "pubkeyHashToSocket", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBase.Contract.PubkeyHashToSocket(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrarBase.Contract.PubkeyHashToSocket(&_TaskAVSRegistrarBase.CallOpts, pubkeyHash)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) PubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "pubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrarBase.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrarBase.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrarBase.CallOpts, operator)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCaller) SupportsAVS(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _TaskAVSRegistrarBase.contract.Call(opts, &out, "supportsAVS", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBase.Contract.SupportsAVS(&_TaskAVSRegistrarBase.CallOpts, avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseCallerSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrarBase.Contract.SupportsAVS(&_TaskAVSRegistrarBase.CallOpts, avs)
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

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactor) UpdateOperatorSocket(opts *bind.TransactOpts, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.contract.Transact(opts, "updateOperatorSocket", socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.UpdateOperatorSocket(&_TaskAVSRegistrarBase.TransactOpts, socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseTransactorSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrarBase.Contract.UpdateOperatorSocket(&_TaskAVSRegistrarBase.TransactOpts, socket)
}

// TaskAVSRegistrarBaseEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseEIP712DomainChangedIterator struct {
	Event *TaskAVSRegistrarBaseEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseEIP712DomainChanged)
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
		it.Event = new(TaskAVSRegistrarBaseEIP712DomainChanged)
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
func (it *TaskAVSRegistrarBaseEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseEIP712DomainChanged represents a EIP712DomainChanged event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*TaskAVSRegistrarBaseEIP712DomainChangedIterator, error) {

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseEIP712DomainChangedIterator{contract: _TaskAVSRegistrarBase.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseEIP712DomainChanged)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseEIP712DomainChanged(log types.Log) (*TaskAVSRegistrarBaseEIP712DomainChanged, error) {
	event := new(TaskAVSRegistrarBaseEIP712DomainChanged)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseNewPubkeyRegistrationIterator is returned from FilterNewPubkeyRegistration and is used to iterate over the raw logs and unpacked data for NewPubkeyRegistration events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseNewPubkeyRegistrationIterator struct {
	Event *TaskAVSRegistrarBaseNewPubkeyRegistration // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseNewPubkeyRegistrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseNewPubkeyRegistration)
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
		it.Event = new(TaskAVSRegistrarBaseNewPubkeyRegistration)
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
func (it *TaskAVSRegistrarBaseNewPubkeyRegistrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseNewPubkeyRegistrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseNewPubkeyRegistration represents a NewPubkeyRegistration event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseNewPubkeyRegistration struct {
	Operator   common.Address
	PubkeyHash [32]byte
	PubkeyG1   BN254G1Point
	PubkeyG2   BN254G2Point
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewPubkeyRegistration is a free log retrieval operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarBaseNewPubkeyRegistrationIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseNewPubkeyRegistrationIterator{contract: _TaskAVSRegistrarBase.contract, event: "NewPubkeyRegistration", logs: logs, sub: sub}, nil
}

// WatchNewPubkeyRegistration is a free log subscription operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseNewPubkeyRegistration, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseNewPubkeyRegistration)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseNewPubkeyRegistration(log types.Log) (*TaskAVSRegistrarBaseNewPubkeyRegistration, error) {
	event := new(TaskAVSRegistrarBaseNewPubkeyRegistration)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator is returned from FilterOperatorSetApkUpdated and is used to iterate over the raw logs and unpacked data for OperatorSetApkUpdated events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator struct {
	Event *TaskAVSRegistrarBaseOperatorSetApkUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseOperatorSetApkUpdated)
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
		it.Event = new(TaskAVSRegistrarBaseOperatorSetApkUpdated)
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
func (it *TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseOperatorSetApkUpdated represents a OperatorSetApkUpdated event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorSetApkUpdated struct {
	Operator      common.Address
	PubkeyHash    [32]byte
	OperatorSetId uint32
	Apk           BN254G1Point
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetApkUpdated is a free log retrieval operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterOperatorSetApkUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (*TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator, error) {

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

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseOperatorSetApkUpdatedIterator{contract: _TaskAVSRegistrarBase.contract, event: "OperatorSetApkUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSetApkUpdated is a free log subscription operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchOperatorSetApkUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseOperatorSetApkUpdated, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (event.Subscription, error) {

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

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseOperatorSetApkUpdated)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseOperatorSetApkUpdated(log types.Log) (*TaskAVSRegistrarBaseOperatorSetApkUpdated, error) {
	event := new(TaskAVSRegistrarBaseOperatorSetApkUpdated)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarBaseOperatorSocketUpdatedIterator is returned from FilterOperatorSocketUpdated and is used to iterate over the raw logs and unpacked data for OperatorSocketUpdated events raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorSocketUpdatedIterator struct {
	Event *TaskAVSRegistrarBaseOperatorSocketUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarBaseOperatorSocketUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarBaseOperatorSocketUpdated)
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
		it.Event = new(TaskAVSRegistrarBaseOperatorSocketUpdated)
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
func (it *TaskAVSRegistrarBaseOperatorSocketUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarBaseOperatorSocketUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarBaseOperatorSocketUpdated represents a OperatorSocketUpdated event raised by the TaskAVSRegistrarBase contract.
type TaskAVSRegistrarBaseOperatorSocketUpdated struct {
	Operator   common.Address
	PubkeyHash [32]byte
	Socket     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketUpdated is a free log retrieval operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) FilterOperatorSocketUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarBaseOperatorSocketUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.FilterLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarBaseOperatorSocketUpdatedIterator{contract: _TaskAVSRegistrarBase.contract, event: "OperatorSocketUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketUpdated is a free log subscription operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) WatchOperatorSocketUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarBaseOperatorSocketUpdated, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrarBase.contract.WatchLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarBaseOperatorSocketUpdated)
				if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
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
func (_TaskAVSRegistrarBase *TaskAVSRegistrarBaseFilterer) ParseOperatorSocketUpdated(log types.Log) (*TaskAVSRegistrarBaseOperatorSocketUpdated, error) {
	event := new(TaskAVSRegistrarBaseOperatorSocketUpdated)
	if err := _TaskAVSRegistrarBase.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
