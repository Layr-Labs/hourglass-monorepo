// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ITaskAVSRegistrar

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

// ITaskAVSRegistrarMetaData contains all meta data concerning the ITaskAVSRegistrar contract.
var ITaskAVSRegistrarMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"calculatePubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getApk\",\"inputs\":[{\"name\":\"operatorSetId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getBatchOperatorPubkeyInfoAndSocket\",\"inputs\":[{\"name\":\"operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfoAndSocket[]\",\"components\":[{\"name\":\"pubkeyInfo\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFromPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyG2\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkeyInfo\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyInfo\",\"components\":[{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"packRegisterPayload\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"pubkeyRegistrationParams\",\"type\":\"tuple\",\"internalType\":\"structITaskAVSRegistrarTypes.PubkeyRegistrationParams\",\"components\":[{\"name\":\"pubkeyRegistrationSignature\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"pubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateOperatorSocket\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"NewPubkeyRegistration\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSetApkUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"apk\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"BLSPubkeyAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAVS\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBLSSignatureOrPrivateKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyAllocationManager\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroPubKey\",\"inputs\":[]}]",
}

// ITaskAVSRegistrarABI is the input ABI used to generate the binding from.
// Deprecated: Use ITaskAVSRegistrarMetaData.ABI instead.
var ITaskAVSRegistrarABI = ITaskAVSRegistrarMetaData.ABI

// ITaskAVSRegistrar is an auto generated Go binding around an Ethereum contract.
type ITaskAVSRegistrar struct {
	ITaskAVSRegistrarCaller     // Read-only binding to the contract
	ITaskAVSRegistrarTransactor // Write-only binding to the contract
	ITaskAVSRegistrarFilterer   // Log filterer for contract events
}

// ITaskAVSRegistrarCaller is an auto generated read-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskAVSRegistrarTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskAVSRegistrarFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ITaskAVSRegistrarFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITaskAVSRegistrarSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ITaskAVSRegistrarSession struct {
	Contract     *ITaskAVSRegistrar // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ITaskAVSRegistrarCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ITaskAVSRegistrarCallerSession struct {
	Contract *ITaskAVSRegistrarCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// ITaskAVSRegistrarTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ITaskAVSRegistrarTransactorSession struct {
	Contract     *ITaskAVSRegistrarTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// ITaskAVSRegistrarRaw is an auto generated low-level Go binding around an Ethereum contract.
type ITaskAVSRegistrarRaw struct {
	Contract *ITaskAVSRegistrar // Generic contract binding to access the raw methods on
}

// ITaskAVSRegistrarCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarCallerRaw struct {
	Contract *ITaskAVSRegistrarCaller // Generic read-only contract binding to access the raw methods on
}

// ITaskAVSRegistrarTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ITaskAVSRegistrarTransactorRaw struct {
	Contract *ITaskAVSRegistrarTransactor // Generic write-only contract binding to access the raw methods on
}

// NewITaskAVSRegistrar creates a new instance of ITaskAVSRegistrar, bound to a specific deployed contract.
func NewITaskAVSRegistrar(address common.Address, backend bind.ContractBackend) (*ITaskAVSRegistrar, error) {
	contract, err := bindITaskAVSRegistrar(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrar{ITaskAVSRegistrarCaller: ITaskAVSRegistrarCaller{contract: contract}, ITaskAVSRegistrarTransactor: ITaskAVSRegistrarTransactor{contract: contract}, ITaskAVSRegistrarFilterer: ITaskAVSRegistrarFilterer{contract: contract}}, nil
}

// NewITaskAVSRegistrarCaller creates a new read-only instance of ITaskAVSRegistrar, bound to a specific deployed contract.
func NewITaskAVSRegistrarCaller(address common.Address, caller bind.ContractCaller) (*ITaskAVSRegistrarCaller, error) {
	contract, err := bindITaskAVSRegistrar(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarCaller{contract: contract}, nil
}

// NewITaskAVSRegistrarTransactor creates a new write-only instance of ITaskAVSRegistrar, bound to a specific deployed contract.
func NewITaskAVSRegistrarTransactor(address common.Address, transactor bind.ContractTransactor) (*ITaskAVSRegistrarTransactor, error) {
	contract, err := bindITaskAVSRegistrar(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarTransactor{contract: contract}, nil
}

// NewITaskAVSRegistrarFilterer creates a new log filterer instance of ITaskAVSRegistrar, bound to a specific deployed contract.
func NewITaskAVSRegistrarFilterer(address common.Address, filterer bind.ContractFilterer) (*ITaskAVSRegistrarFilterer, error) {
	contract, err := bindITaskAVSRegistrar(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarFilterer{contract: contract}, nil
}

// bindITaskAVSRegistrar binds a generic wrapper to an already deployed contract.
func bindITaskAVSRegistrar(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ITaskAVSRegistrarMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITaskAVSRegistrar *ITaskAVSRegistrarRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ITaskAVSRegistrar.Contract.ITaskAVSRegistrarCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITaskAVSRegistrar *ITaskAVSRegistrarRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.ITaskAVSRegistrarTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITaskAVSRegistrar *ITaskAVSRegistrarRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.ITaskAVSRegistrarTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ITaskAVSRegistrar.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.contract.Transact(opts, method, params...)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) CalculatePubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "calculatePubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _ITaskAVSRegistrar.Contract.CalculatePubkeyRegistrationMessageHash(&_ITaskAVSRegistrar.CallOpts, operator)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _ITaskAVSRegistrar.Contract.CalculatePubkeyRegistrationMessageHash(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetApk(opts *bind.CallOpts, operatorSetId uint8) (BN254G1Point, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getApk", operatorSetId)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _ITaskAVSRegistrar.Contract.GetApk(&_ITaskAVSRegistrar.CallOpts, operatorSetId)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 operatorSetId) view returns((uint256,uint256))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetApk(operatorSetId uint8) (BN254G1Point, error) {
	return _ITaskAVSRegistrar.Contract.GetApk(&_ITaskAVSRegistrar.CallOpts, operatorSetId)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetBatchOperatorPubkeyInfoAndSocket(opts *bind.CallOpts, operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getBatchOperatorPubkeyInfoAndSocket", operators)

	if err != nil {
		return *new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket), err
	}

	out0 := *abi.ConvertType(out[0], new([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)).(*[]ITaskAVSRegistrarTypesPubkeyInfoAndSocket)

	return out0, err

}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _ITaskAVSRegistrar.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_ITaskAVSRegistrar.CallOpts, operators)
}

// GetBatchOperatorPubkeyInfoAndSocket is a free data retrieval call binding the contract method 0x3da35ac8.
//
// Solidity: function getBatchOperatorPubkeyInfoAndSocket(address[] operators) view returns((((uint256,uint256),(uint256[2],uint256[2]),bytes32),string)[])
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetBatchOperatorPubkeyInfoAndSocket(operators []common.Address) ([]ITaskAVSRegistrarTypesPubkeyInfoAndSocket, error) {
	return _ITaskAVSRegistrar.Contract.GetBatchOperatorPubkeyInfoAndSocket(&_ITaskAVSRegistrar.CallOpts, operators)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getOperatorFromPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorFromPubkeyHash(&_ITaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorFromPubkeyHash(&_ITaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetOperatorPubkeyG2(opts *bind.CallOpts, operator common.Address) (BN254G2Point, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getOperatorPubkeyG2", operator)

	if err != nil {
		return *new(BN254G2Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G2Point)).(*BN254G2Point)

	return out0, err

}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorPubkeyG2(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorPubkeyG2 is a free data retrieval call binding the contract method 0x67169911.
//
// Solidity: function getOperatorPubkeyG2(address operator) view returns((uint256[2],uint256[2]))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetOperatorPubkeyG2(operator common.Address) (BN254G2Point, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorPubkeyG2(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetOperatorPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getOperatorPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorPubkeyHash(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorPubkeyHash(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetOperatorSocketByOperator(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getOperatorSocketByOperator", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorSocketByOperator(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorSocketByOperator(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetOperatorSocketByPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getOperatorSocketByPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorSocketByPubkeyHash(&_ITaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _ITaskAVSRegistrar.Contract.GetOperatorSocketByPubkeyHash(&_ITaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, [32]byte, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getRegisteredPubkey", operator)

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
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _ITaskAVSRegistrar.Contract.GetRegisteredPubkey(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _ITaskAVSRegistrar.Contract.GetRegisteredPubkey(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) GetRegisteredPubkeyInfo(opts *bind.CallOpts, operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "getRegisteredPubkeyInfo", operator)

	if err != nil {
		return *new(ITaskAVSRegistrarTypesPubkeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ITaskAVSRegistrarTypesPubkeyInfo)).(*ITaskAVSRegistrarTypesPubkeyInfo)

	return out0, err

}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _ITaskAVSRegistrar.Contract.GetRegisteredPubkeyInfo(&_ITaskAVSRegistrar.CallOpts, operator)
}

// GetRegisteredPubkeyInfo is a free data retrieval call binding the contract method 0xc7e66dc4.
//
// Solidity: function getRegisteredPubkeyInfo(address operator) view returns(((uint256,uint256),(uint256[2],uint256[2]),bytes32))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) GetRegisteredPubkeyInfo(operator common.Address) (ITaskAVSRegistrarTypesPubkeyInfo, error) {
	return _ITaskAVSRegistrar.Contract.GetRegisteredPubkeyInfo(&_ITaskAVSRegistrar.CallOpts, operator)
}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) PackRegisterPayload(opts *bind.CallOpts, socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "packRegisterPayload", socket, pubkeyRegistrationParams)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) PackRegisterPayload(socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	return _ITaskAVSRegistrar.Contract.PackRegisterPayload(&_ITaskAVSRegistrar.CallOpts, socket, pubkeyRegistrationParams)
}

// PackRegisterPayload is a free data retrieval call binding the contract method 0x3d3e91b0.
//
// Solidity: function packRegisterPayload(string socket, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) pubkeyRegistrationParams) pure returns(bytes)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) PackRegisterPayload(socket string, pubkeyRegistrationParams ITaskAVSRegistrarTypesPubkeyRegistrationParams) ([]byte, error) {
	return _ITaskAVSRegistrar.Contract.PackRegisterPayload(&_ITaskAVSRegistrar.CallOpts, socket, pubkeyRegistrationParams)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) PubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) (BN254G1Point, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "pubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _ITaskAVSRegistrar.Contract.PubkeyRegistrationMessageHash(&_ITaskAVSRegistrar.CallOpts, operator)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _ITaskAVSRegistrar.Contract.PubkeyRegistrationMessageHash(&_ITaskAVSRegistrar.CallOpts, operator)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCaller) SupportsAVS(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _ITaskAVSRegistrar.contract.Call(opts, &out, "supportsAVS", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) SupportsAVS(avs common.Address) (bool, error) {
	return _ITaskAVSRegistrar.Contract.SupportsAVS(&_ITaskAVSRegistrar.CallOpts, avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarCallerSession) SupportsAVS(avs common.Address) (bool, error) {
	return _ITaskAVSRegistrar.Contract.SupportsAVS(&_ITaskAVSRegistrar.CallOpts, avs)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactor) DeregisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.contract.Transact(opts, "deregisterOperator", operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.DeregisterOperator(&_ITaskAVSRegistrar.TransactOpts, operator, avs, operatorSetIds)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address operator, address avs, uint32[] operatorSetIds) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactorSession) DeregisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.DeregisterOperator(&_ITaskAVSRegistrar.TransactOpts, operator, avs, operatorSetIds)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.contract.Transact(opts, "registerOperator", operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.RegisterOperator(&_ITaskAVSRegistrar.TransactOpts, operator, avs, operatorSetIds, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] operatorSetIds, bytes data) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactorSession) RegisterOperator(operator common.Address, avs common.Address, operatorSetIds []uint32, data []byte) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.RegisterOperator(&_ITaskAVSRegistrar.TransactOpts, operator, avs, operatorSetIds, data)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactor) UpdateOperatorSocket(opts *bind.TransactOpts, socket string) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.contract.Transact(opts, "updateOperatorSocket", socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.UpdateOperatorSocket(&_ITaskAVSRegistrar.TransactOpts, socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_ITaskAVSRegistrar *ITaskAVSRegistrarTransactorSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _ITaskAVSRegistrar.Contract.UpdateOperatorSocket(&_ITaskAVSRegistrar.TransactOpts, socket)
}

// ITaskAVSRegistrarNewPubkeyRegistrationIterator is returned from FilterNewPubkeyRegistration and is used to iterate over the raw logs and unpacked data for NewPubkeyRegistration events raised by the ITaskAVSRegistrar contract.
type ITaskAVSRegistrarNewPubkeyRegistrationIterator struct {
	Event *ITaskAVSRegistrarNewPubkeyRegistration // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarNewPubkeyRegistrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarNewPubkeyRegistration)
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
		it.Event = new(ITaskAVSRegistrarNewPubkeyRegistration)
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
func (it *ITaskAVSRegistrarNewPubkeyRegistrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarNewPubkeyRegistrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarNewPubkeyRegistration represents a NewPubkeyRegistration event raised by the ITaskAVSRegistrar contract.
type ITaskAVSRegistrarNewPubkeyRegistration struct {
	Operator   common.Address
	PubkeyHash [32]byte
	PubkeyG1   BN254G1Point
	PubkeyG2   BN254G2Point
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewPubkeyRegistration is a free log retrieval operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*ITaskAVSRegistrarNewPubkeyRegistrationIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _ITaskAVSRegistrar.contract.FilterLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarNewPubkeyRegistrationIterator{contract: _ITaskAVSRegistrar.contract, event: "NewPubkeyRegistration", logs: logs, sub: sub}, nil
}

// WatchNewPubkeyRegistration is a free log subscription operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarNewPubkeyRegistration, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _ITaskAVSRegistrar.contract.WatchLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarNewPubkeyRegistration)
				if err := _ITaskAVSRegistrar.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
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
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) ParseNewPubkeyRegistration(log types.Log) (*ITaskAVSRegistrarNewPubkeyRegistration, error) {
	event := new(ITaskAVSRegistrarNewPubkeyRegistration)
	if err := _ITaskAVSRegistrar.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskAVSRegistrarOperatorSetApkUpdatedIterator is returned from FilterOperatorSetApkUpdated and is used to iterate over the raw logs and unpacked data for OperatorSetApkUpdated events raised by the ITaskAVSRegistrar contract.
type ITaskAVSRegistrarOperatorSetApkUpdatedIterator struct {
	Event *ITaskAVSRegistrarOperatorSetApkUpdated // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarOperatorSetApkUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarOperatorSetApkUpdated)
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
		it.Event = new(ITaskAVSRegistrarOperatorSetApkUpdated)
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
func (it *ITaskAVSRegistrarOperatorSetApkUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarOperatorSetApkUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarOperatorSetApkUpdated represents a OperatorSetApkUpdated event raised by the ITaskAVSRegistrar contract.
type ITaskAVSRegistrarOperatorSetApkUpdated struct {
	Operator      common.Address
	PubkeyHash    [32]byte
	OperatorSetId uint32
	Apk           BN254G1Point
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorSetApkUpdated is a free log retrieval operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) FilterOperatorSetApkUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (*ITaskAVSRegistrarOperatorSetApkUpdatedIterator, error) {

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

	logs, sub, err := _ITaskAVSRegistrar.contract.FilterLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarOperatorSetApkUpdatedIterator{contract: _ITaskAVSRegistrar.contract, event: "OperatorSetApkUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSetApkUpdated is a free log subscription operation binding the contract event 0x4578729147fad323579cca24cee225babfdedb43e063c28e1505b179fc8a2fe1.
//
// Solidity: event OperatorSetApkUpdated(address indexed operator, bytes32 indexed pubkeyHash, uint32 indexed operatorSetId, (uint256,uint256) apk)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) WatchOperatorSetApkUpdated(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarOperatorSetApkUpdated, operator []common.Address, pubkeyHash [][32]byte, operatorSetId []uint32) (event.Subscription, error) {

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

	logs, sub, err := _ITaskAVSRegistrar.contract.WatchLogs(opts, "OperatorSetApkUpdated", operatorRule, pubkeyHashRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarOperatorSetApkUpdated)
				if err := _ITaskAVSRegistrar.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
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
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) ParseOperatorSetApkUpdated(log types.Log) (*ITaskAVSRegistrarOperatorSetApkUpdated, error) {
	event := new(ITaskAVSRegistrarOperatorSetApkUpdated)
	if err := _ITaskAVSRegistrar.contract.UnpackLog(event, "OperatorSetApkUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ITaskAVSRegistrarOperatorSocketUpdatedIterator is returned from FilterOperatorSocketUpdated and is used to iterate over the raw logs and unpacked data for OperatorSocketUpdated events raised by the ITaskAVSRegistrar contract.
type ITaskAVSRegistrarOperatorSocketUpdatedIterator struct {
	Event *ITaskAVSRegistrarOperatorSocketUpdated // Event containing the contract specifics and raw log

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
func (it *ITaskAVSRegistrarOperatorSocketUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITaskAVSRegistrarOperatorSocketUpdated)
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
		it.Event = new(ITaskAVSRegistrarOperatorSocketUpdated)
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
func (it *ITaskAVSRegistrarOperatorSocketUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITaskAVSRegistrarOperatorSocketUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITaskAVSRegistrarOperatorSocketUpdated represents a OperatorSocketUpdated event raised by the ITaskAVSRegistrar contract.
type ITaskAVSRegistrarOperatorSocketUpdated struct {
	Operator   common.Address
	PubkeyHash [32]byte
	Socket     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketUpdated is a free log retrieval operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) FilterOperatorSocketUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*ITaskAVSRegistrarOperatorSocketUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _ITaskAVSRegistrar.contract.FilterLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &ITaskAVSRegistrarOperatorSocketUpdatedIterator{contract: _ITaskAVSRegistrar.contract, event: "OperatorSocketUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketUpdated is a free log subscription operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) WatchOperatorSocketUpdated(opts *bind.WatchOpts, sink chan<- *ITaskAVSRegistrarOperatorSocketUpdated, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _ITaskAVSRegistrar.contract.WatchLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITaskAVSRegistrarOperatorSocketUpdated)
				if err := _ITaskAVSRegistrar.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
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
func (_ITaskAVSRegistrar *ITaskAVSRegistrarFilterer) ParseOperatorSocketUpdated(log types.Log) (*ITaskAVSRegistrarOperatorSocketUpdated, error) {
	event := new(ITaskAVSRegistrarOperatorSocketUpdated)
	if err := _ITaskAVSRegistrar.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
