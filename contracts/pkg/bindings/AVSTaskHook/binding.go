// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package AVSTaskHook

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

// IBN254CertificateVerifierBN254Certificate is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierBN254Certificate struct {
	ReferenceTimestamp uint32
	MessageHash        [32]byte
	Sig                BN254G1Point
	Apk                BN254G2Point
	NonsignerIndices   []uint32
	NonSignerWitnesses []IBN254CertificateVerifierBN254OperatorInfoWitness
}

// IBN254CertificateVerifierBN254OperatorInfo is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierBN254OperatorInfo struct {
	Pubkey  BN254G1Point
	Weights []*big.Int
}

// IBN254CertificateVerifierBN254OperatorInfoWitness is an auto generated low-level Go binding around an user-defined struct.
type IBN254CertificateVerifierBN254OperatorInfoWitness struct {
	OperatorIndex      uint32
	OperatorInfoProofs []byte
	OperatorInfo       IBN254CertificateVerifierBN254OperatorInfo
}

// OperatorSet is an auto generated low-level Go binding around an user-defined struct.
type OperatorSet struct {
	Avs common.Address
	Id  uint32
}

// AVSTaskHookMetaData contains all meta data concerning the AVSTaskHook contract.
var AVSTaskHookMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"validatePostTaskCreation\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"validatePreTaskCreation\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structOperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"validateTaskResultSubmission\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.GetG1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b5061067f8061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c80631e5e8313146100435780638679c78114610057578063e507027a14610068575b5f5ffd5b6100556100513660046104b3565b5050565b005b610055610065366004610595565b50565b6100556100763660046105c2565b505050565b634e487b7160e01b5f52604160045260245ffd5b604080519081016001600160401b03811182821017156100b1576100b161007b565b60405290565b604051606081016001600160401b03811182821017156100b1576100b161007b565b60405160c081016001600160401b03811182821017156100b1576100b161007b565b604051601f8201601f191681016001600160401b03811182821017156101235761012361007b565b604052919050565b803563ffffffff8116811461013e575f5ffd5b919050565b5f60408284031215610153575f5ffd5b61015b61008f565b823581526020928301359281019290925250919050565b5f82601f830112610181575f5ffd5b61018961008f565b80604084018581111561019a575f5ffd5b845b818110156101b457803584526020938401930161019c565b509095945050505050565b5f608082840312156101cf575f5ffd5b6101d761008f565b90506101e38383610172565b81526101f28360408401610172565b602082015292915050565b5f6001600160401b038211156102155761021561007b565b5060051b60200190565b5f82601f83011261022e575f5ffd5b813561024161023c826101fd565b6100fb565b8082825260208201915060208360051b860101925085831115610262575f5ffd5b602085015b83811015610286576102788161012b565b835260209283019201610267565b5095945050505050565b5f82601f83011261029f575f5ffd5b81356001600160401b038111156102b8576102b861007b565b6102cb601f8201601f19166020016100fb565b8181528460208386010111156102df575f5ffd5b816020850160208301375f918101602001919091529392505050565b5f82601f83011261030a575f5ffd5b813561031861023c826101fd565b8082825260208201915060208360051b860101925085831115610339575f5ffd5b602085015b838110156102865780356001600160401b0381111561035b575f5ffd5b86016060818903601f19011215610370575f5ffd5b6103786100b7565b6103846020830161012b565b815260408201356001600160401b0381111561039e575f5ffd5b6103ad8a602083860101610290565b60208301525060608201356001600160401b038111156103cb575f5ffd5b6020818401019250506060828a0312156103e3575f5ffd5b6103eb61008f565b6103f58a84610143565b815260408301356001600160401b0381111561040f575f5ffd5b80840193505089601f840112610423575f5ffd5b823561043161023c826101fd565b8082825260208201915060208360051b87010192508c831115610452575f5ffd5b6020860195505b8286101561048d5785356bffffffffffffffffffffffff8116811461047c575f5ffd5b825260209586019590910190610459565b80602085015250505080604083015250808552505060208301925060208101905061033e565b5f5f604083850312156104c4575f5ffd5b8235915060208301356001600160401b038111156104e0575f5ffd5b830161014081860312156104f2575f5ffd5b6104fa6100d9565b6105038261012b565b81526020828101359082015261051c8660408401610143565b604082015261052e86608084016101bf565b60608201526101008201356001600160401b0381111561054c575f5ffd5b6105588782850161021f565b6080830152506101208201356001600160401b03811115610577575f5ffd5b610583878285016102fb565b60a08301525080925050509250929050565b5f602082840312156105a5575f5ffd5b5035919050565b80356001600160a01b038116811461013e575f5ffd5b5f5f5f83850360808112156105d5575f5ffd5b6105de856105ac565b93506040601f19820112156105f1575f5ffd5b506105fa61008f565b610606602086016105ac565b81526106146040860161012b565b6020820152915060608401356001600160401b03811115610633575f5ffd5b61063f86828701610290565b915050925092509256fea2646970667358221220f623caa1f69eae53919b285ab11b9f96fc0fe2590a89374f49d34da5cff91deb64736f6c634300081b0033",
}

// AVSTaskHookABI is the input ABI used to generate the binding from.
// Deprecated: Use AVSTaskHookMetaData.ABI instead.
var AVSTaskHookABI = AVSTaskHookMetaData.ABI

// AVSTaskHookBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AVSTaskHookMetaData.Bin instead.
var AVSTaskHookBin = AVSTaskHookMetaData.Bin

// DeployAVSTaskHook deploys a new Ethereum contract, binding an instance of AVSTaskHook to it.
func DeployAVSTaskHook(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *AVSTaskHook, error) {
	parsed, err := AVSTaskHookMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AVSTaskHookBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AVSTaskHook{AVSTaskHookCaller: AVSTaskHookCaller{contract: contract}, AVSTaskHookTransactor: AVSTaskHookTransactor{contract: contract}, AVSTaskHookFilterer: AVSTaskHookFilterer{contract: contract}}, nil
}

// AVSTaskHook is an auto generated Go binding around an Ethereum contract.
type AVSTaskHook struct {
	AVSTaskHookCaller     // Read-only binding to the contract
	AVSTaskHookTransactor // Write-only binding to the contract
	AVSTaskHookFilterer   // Log filterer for contract events
}

// AVSTaskHookCaller is an auto generated read-only Go binding around an Ethereum contract.
type AVSTaskHookCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AVSTaskHookTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AVSTaskHookTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AVSTaskHookFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AVSTaskHookFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AVSTaskHookSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AVSTaskHookSession struct {
	Contract     *AVSTaskHook      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AVSTaskHookCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AVSTaskHookCallerSession struct {
	Contract *AVSTaskHookCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AVSTaskHookTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AVSTaskHookTransactorSession struct {
	Contract     *AVSTaskHookTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AVSTaskHookRaw is an auto generated low-level Go binding around an Ethereum contract.
type AVSTaskHookRaw struct {
	Contract *AVSTaskHook // Generic contract binding to access the raw methods on
}

// AVSTaskHookCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AVSTaskHookCallerRaw struct {
	Contract *AVSTaskHookCaller // Generic read-only contract binding to access the raw methods on
}

// AVSTaskHookTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AVSTaskHookTransactorRaw struct {
	Contract *AVSTaskHookTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAVSTaskHook creates a new instance of AVSTaskHook, bound to a specific deployed contract.
func NewAVSTaskHook(address common.Address, backend bind.ContractBackend) (*AVSTaskHook, error) {
	contract, err := bindAVSTaskHook(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AVSTaskHook{AVSTaskHookCaller: AVSTaskHookCaller{contract: contract}, AVSTaskHookTransactor: AVSTaskHookTransactor{contract: contract}, AVSTaskHookFilterer: AVSTaskHookFilterer{contract: contract}}, nil
}

// NewAVSTaskHookCaller creates a new read-only instance of AVSTaskHook, bound to a specific deployed contract.
func NewAVSTaskHookCaller(address common.Address, caller bind.ContractCaller) (*AVSTaskHookCaller, error) {
	contract, err := bindAVSTaskHook(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AVSTaskHookCaller{contract: contract}, nil
}

// NewAVSTaskHookTransactor creates a new write-only instance of AVSTaskHook, bound to a specific deployed contract.
func NewAVSTaskHookTransactor(address common.Address, transactor bind.ContractTransactor) (*AVSTaskHookTransactor, error) {
	contract, err := bindAVSTaskHook(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AVSTaskHookTransactor{contract: contract}, nil
}

// NewAVSTaskHookFilterer creates a new log filterer instance of AVSTaskHook, bound to a specific deployed contract.
func NewAVSTaskHookFilterer(address common.Address, filterer bind.ContractFilterer) (*AVSTaskHookFilterer, error) {
	contract, err := bindAVSTaskHook(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AVSTaskHookFilterer{contract: contract}, nil
}

// bindAVSTaskHook binds a generic wrapper to an already deployed contract.
func bindAVSTaskHook(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AVSTaskHookMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AVSTaskHook *AVSTaskHookRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AVSTaskHook.Contract.AVSTaskHookCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AVSTaskHook *AVSTaskHookRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.AVSTaskHookTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AVSTaskHook *AVSTaskHookRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.AVSTaskHookTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AVSTaskHook *AVSTaskHookCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AVSTaskHook.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AVSTaskHook *AVSTaskHookTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AVSTaskHook *AVSTaskHookTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.contract.Transact(opts, method, params...)
}

// ValidatePreTaskCreation is a free data retrieval call binding the contract method 0xe507027a.
//
// Solidity: function validatePreTaskCreation(address , (address,uint32) , bytes ) view returns()
func (_AVSTaskHook *AVSTaskHookCaller) ValidatePreTaskCreation(opts *bind.CallOpts, arg0 common.Address, arg1 OperatorSet, arg2 []byte) error {
	var out []interface{}
	err := _AVSTaskHook.contract.Call(opts, &out, "validatePreTaskCreation", arg0, arg1, arg2)

	if err != nil {
		return err
	}

	return err

}

// ValidatePreTaskCreation is a free data retrieval call binding the contract method 0xe507027a.
//
// Solidity: function validatePreTaskCreation(address , (address,uint32) , bytes ) view returns()
func (_AVSTaskHook *AVSTaskHookSession) ValidatePreTaskCreation(arg0 common.Address, arg1 OperatorSet, arg2 []byte) error {
	return _AVSTaskHook.Contract.ValidatePreTaskCreation(&_AVSTaskHook.CallOpts, arg0, arg1, arg2)
}

// ValidatePreTaskCreation is a free data retrieval call binding the contract method 0xe507027a.
//
// Solidity: function validatePreTaskCreation(address , (address,uint32) , bytes ) view returns()
func (_AVSTaskHook *AVSTaskHookCallerSession) ValidatePreTaskCreation(arg0 common.Address, arg1 OperatorSet, arg2 []byte) error {
	return _AVSTaskHook.Contract.ValidatePreTaskCreation(&_AVSTaskHook.CallOpts, arg0, arg1, arg2)
}

// ValidatePostTaskCreation is a paid mutator transaction binding the contract method 0x8679c781.
//
// Solidity: function validatePostTaskCreation(bytes32 ) returns()
func (_AVSTaskHook *AVSTaskHookTransactor) ValidatePostTaskCreation(opts *bind.TransactOpts, arg0 [32]byte) (*types.Transaction, error) {
	return _AVSTaskHook.contract.Transact(opts, "validatePostTaskCreation", arg0)
}

// ValidatePostTaskCreation is a paid mutator transaction binding the contract method 0x8679c781.
//
// Solidity: function validatePostTaskCreation(bytes32 ) returns()
func (_AVSTaskHook *AVSTaskHookSession) ValidatePostTaskCreation(arg0 [32]byte) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.ValidatePostTaskCreation(&_AVSTaskHook.TransactOpts, arg0)
}

// ValidatePostTaskCreation is a paid mutator transaction binding the contract method 0x8679c781.
//
// Solidity: function validatePostTaskCreation(bytes32 ) returns()
func (_AVSTaskHook *AVSTaskHookTransactorSession) ValidatePostTaskCreation(arg0 [32]byte) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.ValidatePostTaskCreation(&_AVSTaskHook.TransactOpts, arg0)
}

// ValidateTaskResultSubmission is a paid mutator transaction binding the contract method 0x1e5e8313.
//
// Solidity: function validateTaskResultSubmission(bytes32 , (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) ) returns()
func (_AVSTaskHook *AVSTaskHookTransactor) ValidateTaskResultSubmission(opts *bind.TransactOpts, arg0 [32]byte, arg1 IBN254CertificateVerifierBN254Certificate) (*types.Transaction, error) {
	return _AVSTaskHook.contract.Transact(opts, "validateTaskResultSubmission", arg0, arg1)
}

// ValidateTaskResultSubmission is a paid mutator transaction binding the contract method 0x1e5e8313.
//
// Solidity: function validateTaskResultSubmission(bytes32 , (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) ) returns()
func (_AVSTaskHook *AVSTaskHookSession) ValidateTaskResultSubmission(arg0 [32]byte, arg1 IBN254CertificateVerifierBN254Certificate) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.ValidateTaskResultSubmission(&_AVSTaskHook.TransactOpts, arg0, arg1)
}

// ValidateTaskResultSubmission is a paid mutator transaction binding the contract method 0x1e5e8313.
//
// Solidity: function validateTaskResultSubmission(bytes32 , (uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) ) returns()
func (_AVSTaskHook *AVSTaskHookTransactorSession) ValidateTaskResultSubmission(arg0 [32]byte, arg1 IBN254CertificateVerifierBN254Certificate) (*types.Transaction, error) {
	return _AVSTaskHook.Contract.ValidateTaskResultSubmission(&_AVSTaskHook.TransactOpts, arg0, arg1)
}
