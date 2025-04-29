// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package IBN254CertificateVerifier

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

// IBN254CertificateVerifierMetaData contains all meta data concerning the IBN254CertificateVerifier contract.
var IBN254CertificateVerifierMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"maxOperatorTableStaleness\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verifyCertificate\",\"inputs\":[{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]}],\"outputs\":[{\"name\":\"signedStakes\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifyCertificateNominal\",\"inputs\":[{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]},{\"name\":\"totalStakeNominalThresholds\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifyCertificateProportion\",\"inputs\":[{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"structIBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]},{\"name\":\"totalStakeProportionThresholds\",\"type\":\"uint16[]\",\"internalType\":\"uint16[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"}]",
}

// IBN254CertificateVerifierABI is the input ABI used to generate the binding from.
// Deprecated: Use IBN254CertificateVerifierMetaData.ABI instead.
var IBN254CertificateVerifierABI = IBN254CertificateVerifierMetaData.ABI

// IBN254CertificateVerifier is an auto generated Go binding around an Ethereum contract.
type IBN254CertificateVerifier struct {
	IBN254CertificateVerifierCaller     // Read-only binding to the contract
	IBN254CertificateVerifierTransactor // Write-only binding to the contract
	IBN254CertificateVerifierFilterer   // Log filterer for contract events
}

// IBN254CertificateVerifierCaller is an auto generated read-only Go binding around an Ethereum contract.
type IBN254CertificateVerifierCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBN254CertificateVerifierTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IBN254CertificateVerifierTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBN254CertificateVerifierFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IBN254CertificateVerifierFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IBN254CertificateVerifierSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IBN254CertificateVerifierSession struct {
	Contract     *IBN254CertificateVerifier // Generic contract binding to set the session for
	CallOpts     bind.CallOpts              // Call options to use throughout this session
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// IBN254CertificateVerifierCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IBN254CertificateVerifierCallerSession struct {
	Contract *IBN254CertificateVerifierCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                    // Call options to use throughout this session
}

// IBN254CertificateVerifierTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IBN254CertificateVerifierTransactorSession struct {
	Contract     *IBN254CertificateVerifierTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// IBN254CertificateVerifierRaw is an auto generated low-level Go binding around an Ethereum contract.
type IBN254CertificateVerifierRaw struct {
	Contract *IBN254CertificateVerifier // Generic contract binding to access the raw methods on
}

// IBN254CertificateVerifierCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IBN254CertificateVerifierCallerRaw struct {
	Contract *IBN254CertificateVerifierCaller // Generic read-only contract binding to access the raw methods on
}

// IBN254CertificateVerifierTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IBN254CertificateVerifierTransactorRaw struct {
	Contract *IBN254CertificateVerifierTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIBN254CertificateVerifier creates a new instance of IBN254CertificateVerifier, bound to a specific deployed contract.
func NewIBN254CertificateVerifier(address common.Address, backend bind.ContractBackend) (*IBN254CertificateVerifier, error) {
	contract, err := bindIBN254CertificateVerifier(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IBN254CertificateVerifier{IBN254CertificateVerifierCaller: IBN254CertificateVerifierCaller{contract: contract}, IBN254CertificateVerifierTransactor: IBN254CertificateVerifierTransactor{contract: contract}, IBN254CertificateVerifierFilterer: IBN254CertificateVerifierFilterer{contract: contract}}, nil
}

// NewIBN254CertificateVerifierCaller creates a new read-only instance of IBN254CertificateVerifier, bound to a specific deployed contract.
func NewIBN254CertificateVerifierCaller(address common.Address, caller bind.ContractCaller) (*IBN254CertificateVerifierCaller, error) {
	contract, err := bindIBN254CertificateVerifier(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IBN254CertificateVerifierCaller{contract: contract}, nil
}

// NewIBN254CertificateVerifierTransactor creates a new write-only instance of IBN254CertificateVerifier, bound to a specific deployed contract.
func NewIBN254CertificateVerifierTransactor(address common.Address, transactor bind.ContractTransactor) (*IBN254CertificateVerifierTransactor, error) {
	contract, err := bindIBN254CertificateVerifier(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IBN254CertificateVerifierTransactor{contract: contract}, nil
}

// NewIBN254CertificateVerifierFilterer creates a new log filterer instance of IBN254CertificateVerifier, bound to a specific deployed contract.
func NewIBN254CertificateVerifierFilterer(address common.Address, filterer bind.ContractFilterer) (*IBN254CertificateVerifierFilterer, error) {
	contract, err := bindIBN254CertificateVerifier(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IBN254CertificateVerifierFilterer{contract: contract}, nil
}

// bindIBN254CertificateVerifier binds a generic wrapper to an already deployed contract.
func bindIBN254CertificateVerifier(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IBN254CertificateVerifierMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBN254CertificateVerifier *IBN254CertificateVerifierRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBN254CertificateVerifier.Contract.IBN254CertificateVerifierCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBN254CertificateVerifier *IBN254CertificateVerifierRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBN254CertificateVerifier.Contract.IBN254CertificateVerifierTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBN254CertificateVerifier *IBN254CertificateVerifierRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBN254CertificateVerifier.Contract.IBN254CertificateVerifierTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IBN254CertificateVerifier.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IBN254CertificateVerifier *IBN254CertificateVerifierTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBN254CertificateVerifier.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IBN254CertificateVerifier *IBN254CertificateVerifierTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IBN254CertificateVerifier.Contract.contract.Transact(opts, method, params...)
}

// VerifyCertificate is a free data retrieval call binding the contract method 0x7192a81c.
//
// Solidity: function verifyCertificate((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert) view returns(uint96[] signedStakes)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCaller) VerifyCertificate(opts *bind.CallOpts, cert IBN254CertificateVerifierBN254Certificate) ([]*big.Int, error) {
	var out []interface{}
	err := _IBN254CertificateVerifier.contract.Call(opts, &out, "verifyCertificate", cert)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// VerifyCertificate is a free data retrieval call binding the contract method 0x7192a81c.
//
// Solidity: function verifyCertificate((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert) view returns(uint96[] signedStakes)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierSession) VerifyCertificate(cert IBN254CertificateVerifierBN254Certificate) ([]*big.Int, error) {
	return _IBN254CertificateVerifier.Contract.VerifyCertificate(&_IBN254CertificateVerifier.CallOpts, cert)
}

// VerifyCertificate is a free data retrieval call binding the contract method 0x7192a81c.
//
// Solidity: function verifyCertificate((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert) view returns(uint96[] signedStakes)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCallerSession) VerifyCertificate(cert IBN254CertificateVerifierBN254Certificate) ([]*big.Int, error) {
	return _IBN254CertificateVerifier.Contract.VerifyCertificate(&_IBN254CertificateVerifier.CallOpts, cert)
}

// VerifyCertificateNominal is a free data retrieval call binding the contract method 0xd8d201fe.
//
// Solidity: function verifyCertificateNominal((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, uint96[] totalStakeNominalThresholds) view returns(bool)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCaller) VerifyCertificateNominal(opts *bind.CallOpts, cert IBN254CertificateVerifierBN254Certificate, totalStakeNominalThresholds []*big.Int) (bool, error) {
	var out []interface{}
	err := _IBN254CertificateVerifier.contract.Call(opts, &out, "verifyCertificateNominal", cert, totalStakeNominalThresholds)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyCertificateNominal is a free data retrieval call binding the contract method 0xd8d201fe.
//
// Solidity: function verifyCertificateNominal((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, uint96[] totalStakeNominalThresholds) view returns(bool)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierSession) VerifyCertificateNominal(cert IBN254CertificateVerifierBN254Certificate, totalStakeNominalThresholds []*big.Int) (bool, error) {
	return _IBN254CertificateVerifier.Contract.VerifyCertificateNominal(&_IBN254CertificateVerifier.CallOpts, cert, totalStakeNominalThresholds)
}

// VerifyCertificateNominal is a free data retrieval call binding the contract method 0xd8d201fe.
//
// Solidity: function verifyCertificateNominal((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, uint96[] totalStakeNominalThresholds) view returns(bool)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCallerSession) VerifyCertificateNominal(cert IBN254CertificateVerifierBN254Certificate, totalStakeNominalThresholds []*big.Int) (bool, error) {
	return _IBN254CertificateVerifier.Contract.VerifyCertificateNominal(&_IBN254CertificateVerifier.CallOpts, cert, totalStakeNominalThresholds)
}

// VerifyCertificateProportion is a free data retrieval call binding the contract method 0xc89fbdbe.
//
// Solidity: function verifyCertificateProportion((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, uint16[] totalStakeProportionThresholds) view returns(bool)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCaller) VerifyCertificateProportion(opts *bind.CallOpts, cert IBN254CertificateVerifierBN254Certificate, totalStakeProportionThresholds []uint16) (bool, error) {
	var out []interface{}
	err := _IBN254CertificateVerifier.contract.Call(opts, &out, "verifyCertificateProportion", cert, totalStakeProportionThresholds)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyCertificateProportion is a free data retrieval call binding the contract method 0xc89fbdbe.
//
// Solidity: function verifyCertificateProportion((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, uint16[] totalStakeProportionThresholds) view returns(bool)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierSession) VerifyCertificateProportion(cert IBN254CertificateVerifierBN254Certificate, totalStakeProportionThresholds []uint16) (bool, error) {
	return _IBN254CertificateVerifier.Contract.VerifyCertificateProportion(&_IBN254CertificateVerifier.CallOpts, cert, totalStakeProportionThresholds)
}

// VerifyCertificateProportion is a free data retrieval call binding the contract method 0xc89fbdbe.
//
// Solidity: function verifyCertificateProportion((uint32,bytes32,(uint256,uint256),(uint256[2],uint256[2]),uint32[],(uint32,bytes,((uint256,uint256),uint96[]))[]) cert, uint16[] totalStakeProportionThresholds) view returns(bool)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierCallerSession) VerifyCertificateProportion(cert IBN254CertificateVerifierBN254Certificate, totalStakeProportionThresholds []uint16) (bool, error) {
	return _IBN254CertificateVerifier.Contract.VerifyCertificateProportion(&_IBN254CertificateVerifier.CallOpts, cert, totalStakeProportionThresholds)
}

// MaxOperatorTableStaleness is a paid mutator transaction binding the contract method 0xb697a30b.
//
// Solidity: function maxOperatorTableStaleness() returns(uint32)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierTransactor) MaxOperatorTableStaleness(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IBN254CertificateVerifier.contract.Transact(opts, "maxOperatorTableStaleness")
}

// MaxOperatorTableStaleness is a paid mutator transaction binding the contract method 0xb697a30b.
//
// Solidity: function maxOperatorTableStaleness() returns(uint32)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierSession) MaxOperatorTableStaleness() (*types.Transaction, error) {
	return _IBN254CertificateVerifier.Contract.MaxOperatorTableStaleness(&_IBN254CertificateVerifier.TransactOpts)
}

// MaxOperatorTableStaleness is a paid mutator transaction binding the contract method 0xb697a30b.
//
// Solidity: function maxOperatorTableStaleness() returns(uint32)
func (_IBN254CertificateVerifier *IBN254CertificateVerifierTransactorSession) MaxOperatorTableStaleness() (*types.Transaction, error) {
	return _IBN254CertificateVerifier.Contract.MaxOperatorTableStaleness(&_IBN254CertificateVerifier.TransactOpts)
}
