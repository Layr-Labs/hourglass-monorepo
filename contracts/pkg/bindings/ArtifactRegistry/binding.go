// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ArtifactRegistry

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

// ArtifactRegistryStorageArtifact is an auto generated low-level Go binding around an user-defined struct.
type ArtifactRegistryStorageArtifact struct {
	Digest      []byte
	RegistryUrl []byte
}

// ArtifactRegistryMetaData contains all meta data concerning the ArtifactRegistry contract.
var ArtifactRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"associateOperatorWithAVS\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"avsAddresses\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLatestArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLatestArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structArtifactRegistryStorage.Artifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"listArtifacts\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structArtifactRegistryStorage.Artifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorAvs\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"publishArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registries\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"avsId\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"PublishedArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes\",\"indexed\":true,\"internalType\":\"bytes\"},{\"name\":\"newArtifact\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structArtifactRegistryStorage.Artifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"previousArtifact\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structArtifactRegistryStorage.Artifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"registryUrl\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false}]",
	Bin: "0x6080604052348015600e575f5ffd5b50611a858061001c5f395ff3fe608060405234801561000f575f5ffd5b5060043610610085575f3560e01c8063741cb6f611610058578063741cb6f614610101578063caed80df14610121578063ce216c4f14610141578063f18d167714610154575f5ffd5b80630b26cc771461008957806346257b7a146100b9578063643d1528146100d957806367ef0db1146100ec575b5f5ffd5b61009c610097366004611436565b610174565b6040516001600160a01b0390911681526020015b60405180910390f35b6100cc6100c736600461145e565b6101a8565b6040516100b091906114ac565b61009c6100e736600461150f565b61050e565b6100ff6100fa366004611526565b610536565b005b61011461010f36600461159c565b610645565b6040516100b09190611621565b61013461012f36600461145e565b61088d565b6040516100b09190611633565b6100ff61014f366004611645565b610928565b61016761016236600461145e565b610eed565b6040516100b091906116c5565b6002602052815f5260405f20818154811061018d575f80fd5b5f918252602090912001546001600160a01b03169150829050565b6001600160a01b0381165f90815260208181526040808320600201805482518185028101850190935280835260609493849084015b82821015610285578382905f5260205f200180546101fa9061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546102269061171c565b80156102715780601f1061024857610100808354040283529160200191610271565b820191905f5260205f20905b81548152906001019060200180831161025457829003601f168201915b5050505050815260200190600101906101dd565b509293505f9250829150505b8251811015610316575f5f5f876001600160a01b03166001600160a01b031681526020019081526020015f206001018483815181106102d2576102d2611754565b60200260200101516040516102e79190611768565b90815260405190819003602001902060010154111561030e578161030a81611792565b9250505b600101610291565b505f8167ffffffffffffffff811115610331576103316117aa565b60405190808252806020026020018201604052801561036457816020015b606081526020019060019003908161034f5790505b5090505f805b8451811015610503575f5f5f896001600160a01b03166001600160a01b031681526020019081526020015f206001018683815181106103ab576103ab611754565b60200260200101516040516103c09190611768565b9081526020016040518091039020600101805480602002602001604051908101604052809291908181526020015f905b82821015610498578382905f5260205f2001805461040d9061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546104399061171c565b80156104845780601f1061045b57610100808354040283529160200191610484565b820191905f5260205f20905b81548152906001019060200180831161046757829003601f168201915b5050505050815260200190600101906103f0565b5050505090505f815111156104fa5780600182516104b691906117be565b815181106104c6576104c6611754565b60200260200101518484815181106104e0576104e0611754565b602002602001018190525082806104f690611792565b9350505b5060010161036a565b509095945050505050565b6001818154811061051d575f80fd5b5f918252602090912001546001600160a01b0316905081565b6001600160a01b0381165f90815260208190526040812080546105589061171c565b9050116105a15760405162461bcd60e51b8152602060048201526012602482015271105594c8191bd95cc81b9bdd08195e1a5cdd60721b60448201526064015b60405180910390fd5b6001600160a01b0382165f908152600260205260408120905b815481101561060657826001600160a01b03168282815481106105df576105df611754565b5f918252602090912001546001600160a01b0316036105fe5750505050565b6001016105ba565b50506001600160a01b039182165f9081526002602090815260408220805460018101825590835291200180546001600160a01b03191691909216179055565b60408051808201909152606080825260208201525f83838080601f0160208091040260200160405190810160405280939291908181526020018383808284375f9201829052506001600160a01b038a1681526020819052604080822090519596509094600190910193506106bc9250859150611768565b908152604051908190036020019020600181015490915061072a5760405162461bcd60e51b815260206004820152602260248201527f4e6f2061727469666163747320666f722074686973206f70657261746f722073604482015261195d60f21b6064820152608401610598565b6001818101545f9161073b916117be565b90505f604051806040016040528084600101848154811061075e5761075e611754565b905f5260205f200180546107719061171c565b80601f016020809104026020016040519081016040528092919081815260200182805461079d9061171c565b80156107e85780601f106107bf576101008083540402835291602001916107e8565b820191905f5260205f20905b8154815290600101906020018083116107cb57829003601f168201915b50505050508152602001845f0180546108009061171c565b80601f016020809104026020016040519081016040528092919081815260200182805461082c9061171c565b80156108775780601f1061084e57610100808354040283529160200191610877565b820191905f5260205f20905b81548152906001019060200180831161085a57829003601f168201915b5050509190925250909998505050505050505050565b5f602081905290815260409020805481906108a79061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546108d39061171c565b801561091e5780601f106108f55761010080835404028352916020019161091e565b820191905f5260205f20905b81548152906001019060200180831161090157829003601f168201915b5050505050905081565b5f60405180604001604052806014815260200173191959985d5b1d0b5c9959da5cdd1c9e4b5d5c9b60621b81525090505f85858080601f0160208091040260200160405190810160405280939291908181526020018383808284375f9201829052506040805160606020601f8c01819004028201810183529181018a8152969750919591945084935090915088908890819085018382808284375f92018290525093855250505060209182018690526040805180820182526060808252818501526001600160a01b038d1683529282905280822090519394509192909160010190610a14908690611768565b908152604051908190036020019020600101541115610c40576001600160a01b0389165f908152602081905260408082209051600191820190610a58908790611768565b90815260405190819003602001902060010154610a7591906117be565b90505f5f5f8c6001600160a01b03166001600160a01b031681526020019081526020015f2060010185604051610aab9190611768565b90815260200160405180910390206001018281548110610acd57610acd611754565b905f5260205f20018054610ae09061171c565b80601f0160208091040260200160405190810160405280929190818152602001828054610b0c9061171c565b8015610b575780601f10610b2e57610100808354040283529160200191610b57565b820191905f5260205f20905b815481529060010190602001808311610b3a57829003601f168201915b5050505050905060405180604001604052808281526020015f5f8e6001600160a01b03166001600160a01b031681526020019081526020015f2060010187604051610ba29190611768565b9081526040519081900360200190208054610bbc9061171c565b80601f0160208091040260200160405190810160405280929190818152602001828054610be89061171c565b8015610c335780601f10610c0a57610100808354040283529160200191610c33565b820191905f5260205f20905b815481529060010190602001808311610c1657829003601f168201915b5050505050815250925050505b6001600160a01b0389165f9081526020819052604090208054610c629061171c565b90505f03610d03576040516bffffffffffffffffffffffff1960608b901b16602082015260340160408051601f198184030181529181526001600160a01b038b165f90815260208190522090610cb89082611823565b506001805480820182555f919091527fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf60180546001600160a01b0319166001600160a01b038b161790555b5f805b6001600160a01b038b165f90815260208190526040902060020154811015610d995784805190602001205f5f8d6001600160a01b03166001600160a01b031681526020019081526020015f206002018281548110610d6657610d66611754565b905f5260205f2001604051610d7b91906118de565b604051809103902003610d915760019150610d99565b600101610d06565b5080610dd3576001600160a01b038a165f9081526020818152604082206002018054600181018255908352912001610dd18582611823565b505b845f5f8c6001600160a01b03166001600160a01b031681526020019081526020015f2060010185604051610e079190611768565b90815260405190819003602001902090610e219082611823565b505f5f8b6001600160a01b03166001600160a01b031681526020019081526020015f2060010184604051610e559190611768565b908152604051602091819003820190206001908101805491820181555f9081529190912001610e8587898361194f565b508888604051610e96929190611a09565b60405180910390208a6001600160a01b03167f84d083fc00f2f83818ed6f62e52ebfae84c6e4183fadc0d5ef74070bdb19968a8585604051610ed9929190611a18565b60405180910390a350505050505050505050565b6001600160a01b0381165f908152600260209081526040808320805482518185028101850190935280835260609493830182828015610f5357602002820191905f5260205f20905b81546001600160a01b03168152600190910190602001808311610f35575b509394505f935083925050505b82518110156110cf575f838281518110610f7c57610f7c611754565b602002602001015190505f5f90505b6001600160a01b0382165f908152602081905260409020600201548110156110c5576001600160a01b0382165f908152602081905260408120600201805483908110610fd957610fd9611754565b905f5260205f20018054610fec9061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546110189061171c565b80156110635780601f1061103a57610100808354040283529160200191611063565b820191905f5260205f20905b81548152906001019060200180831161104657829003601f168201915b505050505090505f5f846001600160a01b03166001600160a01b031681526020019081526020015f206001018160405161109d9190611768565b908152604051908190036020019020600101546110ba9086611a3c565b945050600101610f8b565b5050600101610f60565b505f8167ffffffffffffffff8111156110ea576110ea6117aa565b60405190808252806020026020018201604052801561112f57816020015b60408051808201909152606080825260208201528152602001906001900390816111085790505b5090505f805b8451811015610503575f85828151811061115157611151611754565b602002602001015190505f5f90505b6001600160a01b0382165f90815260208190526040902060020154811015611411576001600160a01b0382165f9081526020819052604081206002018054839081106111ae576111ae611754565b905f5260205f200180546111c19061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546111ed9061171c565b80156112385780601f1061120f57610100808354040283529160200191611238565b820191905f5260205f20905b81548152906001019060200180831161121b57829003601f168201915b505050505090505f5f5f856001600160a01b03166001600160a01b031681526020019081526020015f20600101826040516112739190611768565b90815260405190819003602001902090505f5b60018201548110156114065760405180604001604052808360010183815481106112b2576112b2611754565b905f5260205f200180546112c59061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546112f19061171c565b801561133c5780601f106113135761010080835404028352916020019161133c565b820191905f5260205f20905b81548152906001019060200180831161131f57829003601f168201915b50505050508152602001835f0180546113549061171c565b80601f01602080910402602001604051908101604052809291908181526020018280546113809061171c565b80156113cb5780601f106113a2576101008083540402835291602001916113cb565b820191905f5260205f20905b8154815290600101906020018083116113ae57829003601f168201915b50505050508152508888815181106113e5576113e5611754565b602002602001018190525086806113fb90611792565b975050600101611286565b505050600101611160565b5050600101611135565b80356001600160a01b0381168114611431575f5ffd5b919050565b5f5f60408385031215611447575f5ffd5b6114508361141b565b946020939093013593505050565b5f6020828403121561146e575f5ffd5b6114778261141b565b9392505050565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b5f602082016020835280845180835260408501915060408160051b8601019250602086015f5b8281101561150357603f198786030184526114ee85835161147e565b945060209384019391909101906001016114d2565b50929695505050505050565b5f6020828403121561151f575f5ffd5b5035919050565b5f5f60408385031215611537575f5ffd5b6115408361141b565b915061154e6020840161141b565b90509250929050565b5f5f83601f840112611567575f5ffd5b50813567ffffffffffffffff81111561157e575f5ffd5b602083019150836020828501011115611595575f5ffd5b9250929050565b5f5f5f604084860312156115ae575f5ffd5b6115b78461141b565b9250602084013567ffffffffffffffff8111156115d2575f5ffd5b6115de86828701611557565b9497909650939450505050565b5f8151604084526115ff604085018261147e565b905060208301518482036020860152611618828261147e565b95945050505050565b602081525f61147760208301846115eb565b602081525f611477602083018461147e565b5f5f5f5f5f60608688031215611659575f5ffd5b6116628661141b565b9450602086013567ffffffffffffffff81111561167d575f5ffd5b61168988828901611557565b909550935050604086013567ffffffffffffffff8111156116a8575f5ffd5b6116b488828901611557565b969995985093965092949392505050565b5f602082016020835280845180835260408501915060408160051b8601019250602086015f5b8281101561150357603f198786030184526117078583516115eb565b945060209384019391909101906001016116eb565b600181811c9082168061173057607f821691505b60208210810361174e57634e487b7160e01b5f52602260045260245ffd5b50919050565b634e487b7160e01b5f52603260045260245ffd5b5f82518060208501845e5f920191825250919050565b634e487b7160e01b5f52601160045260245ffd5b5f600182016117a3576117a361177e565b5060010190565b634e487b7160e01b5f52604160045260245ffd5b818103818111156117d1576117d161177e565b92915050565b601f82111561181e57805f5260205f20601f840160051c810160208510156117fc5750805b601f840160051c820191505b8181101561181b575f8155600101611808565b50505b505050565b815167ffffffffffffffff81111561183d5761183d6117aa565b6118518161184b845461171c565b846117d7565b6020601f821160018114611883575f831561186c5750848201515b5f19600385901b1c1916600184901b17845561181b565b5f84815260208120601f198516915b828110156118b25787850151825560209485019460019092019101611892565b50848210156118cf57868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b5f5f83546118eb8161171c565b600182168015611902576001811461191757611944565b60ff1983168652811515820286019350611944565b865f5260205f205f5b8381101561193c57815488820152600190910190602001611920565b505081860193505b509195945050505050565b67ffffffffffffffff831115611967576119676117aa565b61197b83611975835461171c565b836117d7565b5f601f8411600181146119ac575f85156119955750838201355b5f19600387901b1c1916600186901b17835561181b565b5f83815260208120601f198716915b828110156119db57868501358255602094850194600190920191016119bb565b50868210156119f7575f1960f88860031b161c19848701351681555b505060018560011b0183555050505050565b818382375f9101908152919050565b604081525f611a2a60408301856115eb565b828103602084015261161881856115eb565b808201808211156117d1576117d161177e56fea264697066735822122081cc6f5a2b98e110aa322a615a9d7aa858de167abcefd3f3831711a7b4353dbe64736f6c634300081b0033",
}

// ArtifactRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ArtifactRegistryMetaData.ABI instead.
var ArtifactRegistryABI = ArtifactRegistryMetaData.ABI

// ArtifactRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ArtifactRegistryMetaData.Bin instead.
var ArtifactRegistryBin = ArtifactRegistryMetaData.Bin

// DeployArtifactRegistry deploys a new Ethereum contract, binding an instance of ArtifactRegistry to it.
func DeployArtifactRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ArtifactRegistry, error) {
	parsed, err := ArtifactRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ArtifactRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ArtifactRegistry{ArtifactRegistryCaller: ArtifactRegistryCaller{contract: contract}, ArtifactRegistryTransactor: ArtifactRegistryTransactor{contract: contract}, ArtifactRegistryFilterer: ArtifactRegistryFilterer{contract: contract}}, nil
}

// ArtifactRegistry is an auto generated Go binding around an Ethereum contract.
type ArtifactRegistry struct {
	ArtifactRegistryCaller     // Read-only binding to the contract
	ArtifactRegistryTransactor // Write-only binding to the contract
	ArtifactRegistryFilterer   // Log filterer for contract events
}

// ArtifactRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ArtifactRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArtifactRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ArtifactRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArtifactRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ArtifactRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ArtifactRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ArtifactRegistrySession struct {
	Contract     *ArtifactRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ArtifactRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ArtifactRegistryCallerSession struct {
	Contract *ArtifactRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// ArtifactRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ArtifactRegistryTransactorSession struct {
	Contract     *ArtifactRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// ArtifactRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ArtifactRegistryRaw struct {
	Contract *ArtifactRegistry // Generic contract binding to access the raw methods on
}

// ArtifactRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ArtifactRegistryCallerRaw struct {
	Contract *ArtifactRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ArtifactRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ArtifactRegistryTransactorRaw struct {
	Contract *ArtifactRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewArtifactRegistry creates a new instance of ArtifactRegistry, bound to a specific deployed contract.
func NewArtifactRegistry(address common.Address, backend bind.ContractBackend) (*ArtifactRegistry, error) {
	contract, err := bindArtifactRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistry{ArtifactRegistryCaller: ArtifactRegistryCaller{contract: contract}, ArtifactRegistryTransactor: ArtifactRegistryTransactor{contract: contract}, ArtifactRegistryFilterer: ArtifactRegistryFilterer{contract: contract}}, nil
}

// NewArtifactRegistryCaller creates a new read-only instance of ArtifactRegistry, bound to a specific deployed contract.
func NewArtifactRegistryCaller(address common.Address, caller bind.ContractCaller) (*ArtifactRegistryCaller, error) {
	contract, err := bindArtifactRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryCaller{contract: contract}, nil
}

// NewArtifactRegistryTransactor creates a new write-only instance of ArtifactRegistry, bound to a specific deployed contract.
func NewArtifactRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ArtifactRegistryTransactor, error) {
	contract, err := bindArtifactRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryTransactor{contract: contract}, nil
}

// NewArtifactRegistryFilterer creates a new log filterer instance of ArtifactRegistry, bound to a specific deployed contract.
func NewArtifactRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ArtifactRegistryFilterer, error) {
	contract, err := bindArtifactRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryFilterer{contract: contract}, nil
}

// bindArtifactRegistry binds a generic wrapper to an already deployed contract.
func bindArtifactRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ArtifactRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ArtifactRegistry *ArtifactRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ArtifactRegistry.Contract.ArtifactRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ArtifactRegistry *ArtifactRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.ArtifactRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ArtifactRegistry *ArtifactRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.ArtifactRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ArtifactRegistry *ArtifactRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ArtifactRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ArtifactRegistry *ArtifactRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ArtifactRegistry *ArtifactRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.contract.Transact(opts, method, params...)
}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_ArtifactRegistry *ArtifactRegistryCaller) AvsAddresses(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ArtifactRegistry.contract.Call(opts, &out, "avsAddresses", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_ArtifactRegistry *ArtifactRegistrySession) AvsAddresses(arg0 *big.Int) (common.Address, error) {
	return _ArtifactRegistry.Contract.AvsAddresses(&_ArtifactRegistry.CallOpts, arg0)
}

// AvsAddresses is a free data retrieval call binding the contract method 0x643d1528.
//
// Solidity: function avsAddresses(uint256 ) view returns(address)
func (_ArtifactRegistry *ArtifactRegistryCallerSession) AvsAddresses(arg0 *big.Int) (common.Address, error) {
	return _ArtifactRegistry.Contract.AvsAddresses(&_ArtifactRegistry.CallOpts, arg0)
}

// GetLatestArtifact is a free data retrieval call binding the contract method 0x46257b7a.
//
// Solidity: function getLatestArtifact(address avs) view returns(bytes[])
func (_ArtifactRegistry *ArtifactRegistryCaller) GetLatestArtifact(opts *bind.CallOpts, avs common.Address) ([][]byte, error) {
	var out []interface{}
	err := _ArtifactRegistry.contract.Call(opts, &out, "getLatestArtifact", avs)

	if err != nil {
		return *new([][]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)

	return out0, err

}

// GetLatestArtifact is a free data retrieval call binding the contract method 0x46257b7a.
//
// Solidity: function getLatestArtifact(address avs) view returns(bytes[])
func (_ArtifactRegistry *ArtifactRegistrySession) GetLatestArtifact(avs common.Address) ([][]byte, error) {
	return _ArtifactRegistry.Contract.GetLatestArtifact(&_ArtifactRegistry.CallOpts, avs)
}

// GetLatestArtifact is a free data retrieval call binding the contract method 0x46257b7a.
//
// Solidity: function getLatestArtifact(address avs) view returns(bytes[])
func (_ArtifactRegistry *ArtifactRegistryCallerSession) GetLatestArtifact(avs common.Address) ([][]byte, error) {
	return _ArtifactRegistry.Contract.GetLatestArtifact(&_ArtifactRegistry.CallOpts, avs)
}

// GetLatestArtifact0 is a free data retrieval call binding the contract method 0x741cb6f6.
//
// Solidity: function getLatestArtifact(address avs, bytes operatorSetId) view returns((bytes,bytes))
func (_ArtifactRegistry *ArtifactRegistryCaller) GetLatestArtifact0(opts *bind.CallOpts, avs common.Address, operatorSetId []byte) (ArtifactRegistryStorageArtifact, error) {
	var out []interface{}
	err := _ArtifactRegistry.contract.Call(opts, &out, "getLatestArtifact0", avs, operatorSetId)

	if err != nil {
		return *new(ArtifactRegistryStorageArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(ArtifactRegistryStorageArtifact)).(*ArtifactRegistryStorageArtifact)

	return out0, err

}

// GetLatestArtifact0 is a free data retrieval call binding the contract method 0x741cb6f6.
//
// Solidity: function getLatestArtifact(address avs, bytes operatorSetId) view returns((bytes,bytes))
func (_ArtifactRegistry *ArtifactRegistrySession) GetLatestArtifact0(avs common.Address, operatorSetId []byte) (ArtifactRegistryStorageArtifact, error) {
	return _ArtifactRegistry.Contract.GetLatestArtifact0(&_ArtifactRegistry.CallOpts, avs, operatorSetId)
}

// GetLatestArtifact0 is a free data retrieval call binding the contract method 0x741cb6f6.
//
// Solidity: function getLatestArtifact(address avs, bytes operatorSetId) view returns((bytes,bytes))
func (_ArtifactRegistry *ArtifactRegistryCallerSession) GetLatestArtifact0(avs common.Address, operatorSetId []byte) (ArtifactRegistryStorageArtifact, error) {
	return _ArtifactRegistry.Contract.GetLatestArtifact0(&_ArtifactRegistry.CallOpts, avs, operatorSetId)
}

// ListArtifacts is a free data retrieval call binding the contract method 0xf18d1677.
//
// Solidity: function listArtifacts(address operator) view returns((bytes,bytes)[])
func (_ArtifactRegistry *ArtifactRegistryCaller) ListArtifacts(opts *bind.CallOpts, operator common.Address) ([]ArtifactRegistryStorageArtifact, error) {
	var out []interface{}
	err := _ArtifactRegistry.contract.Call(opts, &out, "listArtifacts", operator)

	if err != nil {
		return *new([]ArtifactRegistryStorageArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]ArtifactRegistryStorageArtifact)).(*[]ArtifactRegistryStorageArtifact)

	return out0, err

}

// ListArtifacts is a free data retrieval call binding the contract method 0xf18d1677.
//
// Solidity: function listArtifacts(address operator) view returns((bytes,bytes)[])
func (_ArtifactRegistry *ArtifactRegistrySession) ListArtifacts(operator common.Address) ([]ArtifactRegistryStorageArtifact, error) {
	return _ArtifactRegistry.Contract.ListArtifacts(&_ArtifactRegistry.CallOpts, operator)
}

// ListArtifacts is a free data retrieval call binding the contract method 0xf18d1677.
//
// Solidity: function listArtifacts(address operator) view returns((bytes,bytes)[])
func (_ArtifactRegistry *ArtifactRegistryCallerSession) ListArtifacts(operator common.Address) ([]ArtifactRegistryStorageArtifact, error) {
	return _ArtifactRegistry.Contract.ListArtifacts(&_ArtifactRegistry.CallOpts, operator)
}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_ArtifactRegistry *ArtifactRegistryCaller) OperatorAvs(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ArtifactRegistry.contract.Call(opts, &out, "operatorAvs", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_ArtifactRegistry *ArtifactRegistrySession) OperatorAvs(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _ArtifactRegistry.Contract.OperatorAvs(&_ArtifactRegistry.CallOpts, arg0, arg1)
}

// OperatorAvs is a free data retrieval call binding the contract method 0x0b26cc77.
//
// Solidity: function operatorAvs(address , uint256 ) view returns(address)
func (_ArtifactRegistry *ArtifactRegistryCallerSession) OperatorAvs(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _ArtifactRegistry.Contract.OperatorAvs(&_ArtifactRegistry.CallOpts, arg0, arg1)
}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_ArtifactRegistry *ArtifactRegistryCaller) Registries(opts *bind.CallOpts, arg0 common.Address) ([]byte, error) {
	var out []interface{}
	err := _ArtifactRegistry.contract.Call(opts, &out, "registries", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_ArtifactRegistry *ArtifactRegistrySession) Registries(arg0 common.Address) ([]byte, error) {
	return _ArtifactRegistry.Contract.Registries(&_ArtifactRegistry.CallOpts, arg0)
}

// Registries is a free data retrieval call binding the contract method 0xcaed80df.
//
// Solidity: function registries(address ) view returns(bytes avsId)
func (_ArtifactRegistry *ArtifactRegistryCallerSession) Registries(arg0 common.Address) ([]byte, error) {
	return _ArtifactRegistry.Contract.Registries(&_ArtifactRegistry.CallOpts, arg0)
}

// AssociateOperatorWithAVS is a paid mutator transaction binding the contract method 0x67ef0db1.
//
// Solidity: function associateOperatorWithAVS(address operator, address avs) returns()
func (_ArtifactRegistry *ArtifactRegistryTransactor) AssociateOperatorWithAVS(opts *bind.TransactOpts, operator common.Address, avs common.Address) (*types.Transaction, error) {
	return _ArtifactRegistry.contract.Transact(opts, "associateOperatorWithAVS", operator, avs)
}

// AssociateOperatorWithAVS is a paid mutator transaction binding the contract method 0x67ef0db1.
//
// Solidity: function associateOperatorWithAVS(address operator, address avs) returns()
func (_ArtifactRegistry *ArtifactRegistrySession) AssociateOperatorWithAVS(operator common.Address, avs common.Address) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.AssociateOperatorWithAVS(&_ArtifactRegistry.TransactOpts, operator, avs)
}

// AssociateOperatorWithAVS is a paid mutator transaction binding the contract method 0x67ef0db1.
//
// Solidity: function associateOperatorWithAVS(address operator, address avs) returns()
func (_ArtifactRegistry *ArtifactRegistryTransactorSession) AssociateOperatorWithAVS(operator common.Address, avs common.Address) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.AssociateOperatorWithAVS(&_ArtifactRegistry.TransactOpts, operator, avs)
}

// PublishArtifact is a paid mutator transaction binding the contract method 0xce216c4f.
//
// Solidity: function publishArtifact(address avs, bytes operatorSetId, bytes digest) returns()
func (_ArtifactRegistry *ArtifactRegistryTransactor) PublishArtifact(opts *bind.TransactOpts, avs common.Address, operatorSetId []byte, digest []byte) (*types.Transaction, error) {
	return _ArtifactRegistry.contract.Transact(opts, "publishArtifact", avs, operatorSetId, digest)
}

// PublishArtifact is a paid mutator transaction binding the contract method 0xce216c4f.
//
// Solidity: function publishArtifact(address avs, bytes operatorSetId, bytes digest) returns()
func (_ArtifactRegistry *ArtifactRegistrySession) PublishArtifact(avs common.Address, operatorSetId []byte, digest []byte) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.PublishArtifact(&_ArtifactRegistry.TransactOpts, avs, operatorSetId, digest)
}

// PublishArtifact is a paid mutator transaction binding the contract method 0xce216c4f.
//
// Solidity: function publishArtifact(address avs, bytes operatorSetId, bytes digest) returns()
func (_ArtifactRegistry *ArtifactRegistryTransactorSession) PublishArtifact(avs common.Address, operatorSetId []byte, digest []byte) (*types.Transaction, error) {
	return _ArtifactRegistry.Contract.PublishArtifact(&_ArtifactRegistry.TransactOpts, avs, operatorSetId, digest)
}

// ArtifactRegistryPublishedArtifactIterator is returned from FilterPublishedArtifact and is used to iterate over the raw logs and unpacked data for PublishedArtifact events raised by the ArtifactRegistry contract.
type ArtifactRegistryPublishedArtifactIterator struct {
	Event *ArtifactRegistryPublishedArtifact // Event containing the contract specifics and raw log

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
func (it *ArtifactRegistryPublishedArtifactIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ArtifactRegistryPublishedArtifact)
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
		it.Event = new(ArtifactRegistryPublishedArtifact)
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
func (it *ArtifactRegistryPublishedArtifactIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ArtifactRegistryPublishedArtifactIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ArtifactRegistryPublishedArtifact represents a PublishedArtifact event raised by the ArtifactRegistry contract.
type ArtifactRegistryPublishedArtifact struct {
	Avs              common.Address
	OperatorSetId    common.Hash
	NewArtifact      ArtifactRegistryStorageArtifact
	PreviousArtifact ArtifactRegistryStorageArtifact
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterPublishedArtifact is a free log retrieval operation binding the contract event 0x84d083fc00f2f83818ed6f62e52ebfae84c6e4183fadc0d5ef74070bdb19968a.
//
// Solidity: event PublishedArtifact(address indexed avs, bytes indexed operatorSetId, (bytes,bytes) newArtifact, (bytes,bytes) previousArtifact)
func (_ArtifactRegistry *ArtifactRegistryFilterer) FilterPublishedArtifact(opts *bind.FilterOpts, avs []common.Address, operatorSetId [][]byte) (*ArtifactRegistryPublishedArtifactIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _ArtifactRegistry.contract.FilterLogs(opts, "PublishedArtifact", avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ArtifactRegistryPublishedArtifactIterator{contract: _ArtifactRegistry.contract, event: "PublishedArtifact", logs: logs, sub: sub}, nil
}

// WatchPublishedArtifact is a free log subscription operation binding the contract event 0x84d083fc00f2f83818ed6f62e52ebfae84c6e4183fadc0d5ef74070bdb19968a.
//
// Solidity: event PublishedArtifact(address indexed avs, bytes indexed operatorSetId, (bytes,bytes) newArtifact, (bytes,bytes) previousArtifact)
func (_ArtifactRegistry *ArtifactRegistryFilterer) WatchPublishedArtifact(opts *bind.WatchOpts, sink chan<- *ArtifactRegistryPublishedArtifact, avs []common.Address, operatorSetId [][]byte) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var operatorSetIdRule []interface{}
	for _, operatorSetIdItem := range operatorSetId {
		operatorSetIdRule = append(operatorSetIdRule, operatorSetIdItem)
	}

	logs, sub, err := _ArtifactRegistry.contract.WatchLogs(opts, "PublishedArtifact", avsRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ArtifactRegistryPublishedArtifact)
				if err := _ArtifactRegistry.contract.UnpackLog(event, "PublishedArtifact", log); err != nil {
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

// ParsePublishedArtifact is a log parse operation binding the contract event 0x84d083fc00f2f83818ed6f62e52ebfae84c6e4183fadc0d5ef74070bdb19968a.
//
// Solidity: event PublishedArtifact(address indexed avs, bytes indexed operatorSetId, (bytes,bytes) newArtifact, (bytes,bytes) previousArtifact)
func (_ArtifactRegistry *ArtifactRegistryFilterer) ParsePublishedArtifact(log types.Log) (*ArtifactRegistryPublishedArtifact, error) {
	event := new(ArtifactRegistryPublishedArtifact)
	if err := _ArtifactRegistry.contract.UnpackLog(event, "PublishedArtifact", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
