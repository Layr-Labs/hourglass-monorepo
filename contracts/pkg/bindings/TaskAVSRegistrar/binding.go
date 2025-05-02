// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package TaskAVSRegistrar

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

// TaskAVSRegistrarMetaData contains all meta data concerning the TaskAVSRegistrar contract.
var TaskAVSRegistrarMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"allocationManager\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"ALLOCATION_MANAGER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAllocationManager\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"AVS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PUBKEY_REGISTRATION_TYPEHASH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculatePubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}],\"outputs\":[],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFromPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSocketByPubkeyHash\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRegisteredPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkey\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToPubkeyHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorToSocket\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyHashToOperator\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyHashToSocket\",\"inputs\":[{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pubkeyRegistrationMessageHash\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsAVS\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateOperatorSocket\",\"inputs\":[{\"name\":\"socket\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NewPubkeyRegistration\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"pubkeyG1\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"pubkeyG2\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structBN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorSocketUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"pubkeyHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"socket\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"BLSPubkeyAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECAddFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECMulFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECPairingFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpModFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAVS\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBLSSignatureOrPrivateKey\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyAllocationManager\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"ZeroPubKey\",\"inputs\":[]}]",
	Bin: "0x6101a0604052348015610010575f5ffd5b5060405161236538038061236583398101604081905261002f916101d6565b81816040518060400160405280601081526020016f2a30b9b5a0ab29a932b3b4b9ba3930b960811b815250604051806040016040528060018152602001603160f81b8152506100875f8361014760201b90919060201c565b61012052610096816001610147565b61014052815160208084019190912060e052815190820120610100524660a05261012260e05161010051604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201529081019290925260608201524660808201523060a08201525f9060c00160405160208183030381529060405280519060200120905090565b60805250503060c0526001600160a01b03918216610160521661018052506103b89050565b5f6020835110156101625761015b83610179565b9050610173565b8161016d84826102a6565b5060ff90505b92915050565b5f5f829050601f815111156101ac578260405163305a27a960e01b81526004016101a39190610360565b60405180910390fd5b80516101b782610395565b179392505050565b6001600160a01b03811681146101d3575f5ffd5b50565b5f5f604083850312156101e7575f5ffd5b82516101f2816101bf565b6020840151909250610203816101bf565b809150509250929050565b634e487b7160e01b5f52604160045260245ffd5b600181811c9082168061023657607f821691505b60208210810361025457634e487b7160e01b5f52602260045260245ffd5b50919050565b601f8211156102a157805f5260205f20601f840160051c8101602085101561027f5750805b601f840160051c820191505b8181101561029e575f815560010161028b565b50505b505050565b81516001600160401b038111156102bf576102bf61020e565b6102d3816102cd8454610222565b8461025a565b6020601f821160018114610305575f83156102ee5750848201515b5f19600385901b1c1916600184901b17845561029e565b5f84815260208120601f198516915b828110156103345787850151825560209485019460019092019101610314565b508482101561035157868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b80516020808301519190811015610254575f1960209190910360031b1b16919050565b60805160a05160c05160e0516101005161012051610140516101605161018051611f2c6104395f395f8181610184015281816103ce0152818161081101526108cd01525f818161033201526107d601525f6106c001525f61069601525f610d2f01525f610d0701525f610c6201525f610c8c01525f610cb60152611f2c5ff3fe608060405234801561000f575f5ffd5b5060043610610126575f3560e01c80639d6f2285116100a9578063c95e97da1161006e578063c95e97da1461031a578063d74a8b611461032d578063de29fac014610354578063e8bb9ae614610373578063fd0d930a1461039b575f5ffd5b80639d6f2285146102975780639feab859146102aa578063a30db098146102d1578063b5265787146102e4578063c63fd50214610307575f5ffd5b806347b314e8116100ef57806347b314e8146101fe57806369e5aa8b1461022657806373447992146102395780637ff81a871461025a57806384b0196e1461027c575f5ffd5b8062a1f4cb1461012a578063303ca9561461016a57806331232bc91461017f57806339c26f42146101be5780633c2a7f4c146101de575b5f5ffd5b61015061013836600461163d565b60046020525f90815260409020805460019091015482565b604080519283526020830191909152015b60405180910390f35b61017d61017836600461169f565b6103c3565b005b6101a67f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b039091168152602001610161565b6101d16101cc36600461163d565b610438565b604051610161919061172d565b6101f16101ec36600461163d565b6104cf565b604051610161919061173f565b6101a661020c366004611756565b5f908152600360205260409020546001600160a01b031690565b6101d1610234366004611756565b6104f9565b61024c61024736600461163d565b610511565b604051908152602001610161565b61026d61026836600461163d565b610575565b604051610161939291906117ac565b610284610689565b60405161016197969594939291906117d9565b6101d16102a5366004611756565b61070f565b61024c7f2bd82124057f0913bc3b772ce7b83e8057c1ad1f3510fc83778be20f10ec5de681565b6101d16102df36600461163d565b6107ae565b6102f76102f236600461163d565b6107d4565b6040519015158152602001610161565b61017d61031536600461186f565b610806565b61017d610328366004611a27565b6108b6565b6101a67f000000000000000000000000000000000000000000000000000000000000000081565b61024c61036236600461163d565b60026020525f908152604090205481565b6101a6610381366004611756565b60036020525f90815260409020546001600160a01b031681565b61024c6103a936600461163d565b6001600160a01b03165f9081526002602052604090205490565b336001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000161461040c576040516323d871a560e01b815260040160405180910390fd5b610415836107d4565b610432576040516366e565df60e01b815260040160405180910390fd5b50505050565b60076020525f90815260409020805461045090611a60565b80601f016020809104026020016040519081016040528092919081815260200182805461047c90611a60565b80156104c75780601f1061049e576101008083540402835291602001916104c7565b820191905f5260205f20905b8154815290600101906020018083116104aa57829003601f168201915b505050505081565b604080518082019091525f80825260208201526104f36104ee83610511565b6109c7565b92915050565b60066020525f90815260409020805461045090611a60565b5f6104f37f2bd82124057f0913bc3b772ce7b83e8057c1ad1f3510fc83778be20f10ec5de68360405160200161055a9291909182526001600160a01b0316602082015260400190565b60405160208183030381529060405280519060200120610a51565b604080518082019091525f80825260208201526105906114fa565b6001600160a01b0383165f81815260046020908152604080832081518083018352815481526001909101548184015293835260059091528082208151608081018084529394938593919291839190820190839060029082845b8154815260200190600101908083116105e957505050918352505060408051808201918290526020909201919060028481019182845b81548152602001906001019080831161061f5750505050508152505090505f61065c876001600160a01b03165f9081526002602052604090205490565b90508061067c576040516325ec6c1f60e01b815260040160405180910390fd5b9196909550909350915050565b5f606080828080836106bb7f000000000000000000000000000000000000000000000000000000000000000083610a7d565b6106e67f00000000000000000000000000000000000000000000000000000000000000006001610a7d565b604080515f80825260208201909252600f60f81b9b939a50919850469750309650945092509050565b5f81815260066020526040902080546060919061072b90611a60565b80601f016020809104026020016040519081016040528092919081815260200182805461075790611a60565b80156107a25780601f10610779576101008083540402835291602001916107a2565b820191905f5260205f20905b81548152906001019060200180831161078557829003601f168201915b50505050509050919050565b6001600160a01b0381165f90815260076020526040902080546060919061072b90611a60565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0390811691161490565b336001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000161461084f576040516323d871a560e01b815260040160405180910390fd5b610858856107d4565b610875576040516366e565df60e01b815260040160405180910390fd5b5f61088282840184611b14565b90505f61089c8883602001516108978b6104cf565b610b27565b90506108ac8882845f0151610b59565b5050505050505050565b6040516379ae50cd60e01b81523360048201525f907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906379ae50cd906024015f60405180830381865afa158015610919573d5f5f3e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526109409190810190611c08565b90505f805b82518110156109895761097383828151811061096357610963611cf5565b60200260200101515f01516107d4565b156109815760019150610989565b600101610945565b50806109a8576040516325ec6c1f60e01b815260040160405180910390fd5b335f818152600260205260409020546109c2919085610b59565b505050565b604080518082019091525f80825260208201525f80806109f45f516020611ed75f395f51905f5286611d09565b90505b610a0081610bda565b90935091505f516020611ed75f395f51905f528283098303610a38576040805180820190915290815260208101919091529392505050565b5f516020611ed75f395f51905f526001820890506109f7565b5f6104f3610a5d610c56565b8360405161190160f01b8152600281019290925260228201526042902090565b606060ff8314610a9757610a9083610d84565b90506104f3565b818054610aa390611a60565b80601f0160208091040260200160405190810160405280929190818152602001828054610acf90611a60565b8015610b1a5780601f10610af157610100808354040283529160200191610b1a565b820191905f5260205f20905b815481529060010190602001808311610afd57829003601f168201915b5050505050905092915050565b6001600160a01b0383165f9081526002602052604090205480610b5257610b4f848484610dc1565b90505b9392505050565b5f828152600660205260409020610b708282611d73565b506001600160a01b0383165f908152600760205260409020610b928282611d73565b5081836001600160a01b03167fa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a83604051610bcd919061172d565b60405180910390a3505050565b5f80805f516020611ed75f395f51905f5260035f516020611ed75f395f51905f52865f516020611ed75f395f51905f52888909090890505f610c4a827f0c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f525f516020611ed75f395f51905f5261107f565b91959194509092505050565b5f306001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016148015610cae57507f000000000000000000000000000000000000000000000000000000000000000046145b15610cd857507f000000000000000000000000000000000000000000000000000000000000000090565b610d7f604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201527f0000000000000000000000000000000000000000000000000000000000000000918101919091527f000000000000000000000000000000000000000000000000000000000000000060608201524660808201523060a08201525f9060c00160405160208183030381529060405280519060200120905090565b905090565b60605f610d90836110f8565b6040805160208082528183019092529192505f91906020820181803683375050509182525060208101929092525090565b60208281015180515f90815290820151909152604090207f52cdd74989082c32bd7b5abbc0e80e69d4c91b6e4cf5bf4dbfa7b61a6845a04b8101610e1857604051630cc7509160e01b815260040160405180910390fd5b6001600160a01b0384165f9081526002602052604081205414610e4e576040516342ee68b560e01b815260040160405180910390fd5b5f818152600360205260409020546001600160a01b031615610e8357604051634c334c9760e11b815260040160405180910390fd5b825180516020918201518286015180519084015160408089015180519087015189518a89015193515f997f30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f000000199610ee6999098909790969095949392909101611e2d565b604051602081830303815290604052805190602001205f1c610f089190611d09565b9050610f80610f2f610f2783876020015161111f90919063ffffffff16565b86519061118f565b610f37611203565b610f76610f6f85610f696040805180820182525f80825260209182015281518083019092526001825260029082015290565b9061111f565b879061118f565b87604001516112c3565b610f9d5760405163a72d026360e01b815260040160405180910390fd5b6020808501516001600160a01b0387165f9081526004835260408082208351815592840151600190930192909255818701516005909352208151610fe4908290600261151f565b506020820151610ffa906002808401919061151f565b5050506001600160a01b0385165f818152600260209081526040808320869055858352600382529182902080546001600160a01b031916841790558601518682015191518593927ff9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e9261106f92909190611e79565b60405180910390a3509392505050565b5f5f61108961155d565b61109161157b565b602080825281810181905260408201819052606082018890526080820187905260a082018690528260c08360056107d05a03fa925082806110ce57fe5b50826110ed5760405163d51edae360e01b815260040160405180910390fd5b505195945050505050565b5f60ff8216601f8111156104f357604051632cd44ac360e21b815260040160405180910390fd5b604080518082019091525f808252602082015261113a611599565b835181526020808501519082015260408082018490525f908360608460076107d05a03fa9050808061116857fe5b508061118757604051632319df1960e11b815260040160405180910390fd5b505092915050565b604080518082019091525f80825260208201526111aa6115b7565b835181526020808501518183015283516040808401919091529084015160608301525f908360808460066107d05a03fa905080806111e457fe5b50806111875760405163d4b68fd760e01b815260040160405180910390fd5b61120b6114fa565b50604080516080810182527f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c28183019081527f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed6060830152815281518083019092527f275dc4a288d1afb3cbb1ac09187524c7db36395df7be3b99e673b13a075a65ec82527f1d9befcd05a5323e6da4d435f3b617cdb3af83285c2df711ef39c01571827f9d60208381019190915281019190915290565b6040805180820182528581526020808201859052825180840190935285835282018390525f916112f16115d5565b5f5b60028110156114a8575f611308826006611eac565b905084826002811061131c5761131c611cf5565b6020020151518361132d835f611ec3565b600c811061133d5761133d611cf5565b602002015284826002811061135457611354611cf5565b6020020151602001518382600161136b9190611ec3565b600c811061137b5761137b611cf5565b602002015283826002811061139257611392611cf5565b60200201515151836113a5836002611ec3565b600c81106113b5576113b5611cf5565b60200201528382600281106113cc576113cc611cf5565b60200201515160016020020151836113e5836003611ec3565b600c81106113f5576113f5611cf5565b602002015283826002811061140c5761140c611cf5565b6020020151602001515f6002811061142657611426611cf5565b602002015183611437836004611ec3565b600c811061144757611447611cf5565b602002015283826002811061145e5761145e611cf5565b60200201516020015160016002811061147957611479611cf5565b60200201518361148a836005611ec3565b600c811061149a5761149a611cf5565b6020020152506001016112f3565b506114b161155d565b5f6020826101808560086107d05a03fa905080806114cb57fe5b50806114ea576040516324ccc79360e21b815260040160405180910390fd5b5051151598975050505050505050565b604051806040016040528061150d6115f4565b815260200161151a6115f4565b905290565b826002810192821561154d579160200282015b8281111561154d578251825591602001919060010190611532565b50611559929150611612565b5090565b60405180602001604052806001906020820280368337509192915050565b6040518060c001604052806006906020820280368337509192915050565b60405180606001604052806003906020820280368337509192915050565b60405180608001604052806004906020820280368337509192915050565b604051806101800160405280600c906020820280368337509192915050565b60405180604001604052806002906020820280368337509192915050565b5b80821115611559575f8155600101611613565b6001600160a01b038116811461163a575f5ffd5b50565b5f6020828403121561164d575f5ffd5b8135610b5281611626565b5f5f83601f840112611668575f5ffd5b5081356001600160401b0381111561167e575f5ffd5b6020830191508360208260051b8501011115611698575f5ffd5b9250929050565b5f5f5f5f606085870312156116b2575f5ffd5b84356116bd81611626565b935060208501356116cd81611626565b925060408501356001600160401b038111156116e7575f5ffd5b6116f387828801611658565b95989497509550505050565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b602081525f610b5260208301846116ff565b8151815260208083015190820152604081016104f3565b5f60208284031215611766575f5ffd5b5035919050565b805f5b6002811015610432578151845260209384019390910190600101611770565b61179a82825161176d565b60208101516109c2604084018261176d565b835181526020808501519082015260e081016117cb604083018561178f565b8260c0830152949350505050565b60ff60f81b8816815260e060208201525f6117f760e08301896116ff565b828103604084015261180981896116ff565b606084018890526001600160a01b038716608085015260a0840186905283810360c0850152845180825260208087019350909101905f5b8181101561185e578351835260209384019390920191600101611840565b50909b9a5050505050505050505050565b5f5f5f5f5f5f60808789031215611884575f5ffd5b863561188f81611626565b9550602087013561189f81611626565b945060408701356001600160401b038111156118b9575f5ffd5b6118c589828a01611658565b90955093505060608701356001600160401b038111156118e3575f5ffd5b8701601f810189136118f3575f5ffd5b80356001600160401b03811115611908575f5ffd5b896020828401011115611919575f5ffd5b60208201935080925050509295509295509295565b634e487b7160e01b5f52604160045260245ffd5b604080519081016001600160401b03811182821017156119645761196461192e565b60405290565b604051606081016001600160401b03811182821017156119645761196461192e565b604051601f8201601f191681016001600160401b03811182821017156119b4576119b461192e565b604052919050565b5f82601f8301126119cb575f5ffd5b81356001600160401b038111156119e4576119e461192e565b6119f7601f8201601f191660200161198c565b818152846020838601011115611a0b575f5ffd5b816020850160208301375f918101602001919091529392505050565b5f60208284031215611a37575f5ffd5b81356001600160401b03811115611a4c575f5ffd5b611a58848285016119bc565b949350505050565b600181811c90821680611a7457607f821691505b602082108103611a9257634e487b7160e01b5f52602260045260245ffd5b50919050565b5f60408284031215611aa8575f5ffd5b611ab0611942565b823581526020928301359281019290925250919050565b5f82601f830112611ad6575f5ffd5b611ade611942565b806040840185811115611aef575f5ffd5b845b81811015611b09578035845260209384019301611af1565b509095945050505050565b5f60208284031215611b24575f5ffd5b81356001600160401b03811115611b39575f5ffd5b8201808403610120811215611b4c575f5ffd5b611b54611942565b82356001600160401b03811115611b69575f5ffd5b611b75878286016119bc565b825250610100601f1983011215611b8a575f5ffd5b611b9261196a565b611b9f8760208601611a98565b8152611bae8760608601611a98565b60208201526080609f1984011215611bc4575f5ffd5b611bcc611942565b9250611bdb8760a08601611ac7565b8352611bea8760e08601611ac7565b60208401528260408201528060208301525080935050505092915050565b5f60208284031215611c18575f5ffd5b81516001600160401b03811115611c2d575f5ffd5b8201601f81018413611c3d575f5ffd5b80516001600160401b03811115611c5657611c5661192e565b611c6560208260051b0161198c565b8082825260208201915060208360061b850101925086831115611c86575f5ffd5b6020840193505b82841015611ceb5760408488031215611ca4575f5ffd5b611cac611942565b8451611cb781611626565b8152602085015163ffffffff81168114611ccf575f5ffd5b8060208301525080835250602082019150604084019350611c8d565b9695505050505050565b634e487b7160e01b5f52603260045260245ffd5b5f82611d2357634e487b7160e01b5f52601260045260245ffd5b500690565b601f8211156109c257805f5260205f20601f840160051c81016020851015611d4d5750805b601f840160051c820191505b81811015611d6c575f8155600101611d59565b5050505050565b81516001600160401b03811115611d8c57611d8c61192e565b611da081611d9a8454611a60565b84611d28565b6020601f821160018114611dd2575f8315611dbb5750848201515b5f19600385901b1c1916600184901b178455611d6c565b5f84815260208120601f198516915b82811015611e015787850151825560209485019460019092019101611de1565b5084821015611e1e57868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b888152876020820152866040820152856060820152611e4f608082018661176d565b611e5c60c082018561176d565b610100810192909252610120820152610140019695505050505050565b825181526020808401519082015260c08101610b52604083018461178f565b634e487b7160e01b5f52601160045260245ffd5b80820281158282048414176104f3576104f3611e98565b808201808211156104f3576104f3611e9856fe30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47a26469706673582212200b5c265e4dc47a3bcb25d8caa1812e0da26e4e5d4356f7176aec689f0ef8f0f764736f6c634300081b0033",
}

// TaskAVSRegistrarABI is the input ABI used to generate the binding from.
// Deprecated: Use TaskAVSRegistrarMetaData.ABI instead.
var TaskAVSRegistrarABI = TaskAVSRegistrarMetaData.ABI

// TaskAVSRegistrarBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TaskAVSRegistrarMetaData.Bin instead.
var TaskAVSRegistrarBin = TaskAVSRegistrarMetaData.Bin

// DeployTaskAVSRegistrar deploys a new Ethereum contract, binding an instance of TaskAVSRegistrar to it.
func DeployTaskAVSRegistrar(auth *bind.TransactOpts, backend bind.ContractBackend, avs common.Address, allocationManager common.Address) (common.Address, *types.Transaction, *TaskAVSRegistrar, error) {
	parsed, err := TaskAVSRegistrarMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TaskAVSRegistrarBin), backend, avs, allocationManager)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TaskAVSRegistrar{TaskAVSRegistrarCaller: TaskAVSRegistrarCaller{contract: contract}, TaskAVSRegistrarTransactor: TaskAVSRegistrarTransactor{contract: contract}, TaskAVSRegistrarFilterer: TaskAVSRegistrarFilterer{contract: contract}}, nil
}

// TaskAVSRegistrar is an auto generated Go binding around an Ethereum contract.
type TaskAVSRegistrar struct {
	TaskAVSRegistrarCaller     // Read-only binding to the contract
	TaskAVSRegistrarTransactor // Write-only binding to the contract
	TaskAVSRegistrarFilterer   // Log filterer for contract events
}

// TaskAVSRegistrarCaller is an auto generated read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TaskAVSRegistrarFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TaskAVSRegistrarSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TaskAVSRegistrarSession struct {
	Contract     *TaskAVSRegistrar // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TaskAVSRegistrarCallerSession struct {
	Contract *TaskAVSRegistrarCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// TaskAVSRegistrarTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TaskAVSRegistrarTransactorSession struct {
	Contract     *TaskAVSRegistrarTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// TaskAVSRegistrarRaw is an auto generated low-level Go binding around an Ethereum contract.
type TaskAVSRegistrarRaw struct {
	Contract *TaskAVSRegistrar // Generic contract binding to access the raw methods on
}

// TaskAVSRegistrarCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TaskAVSRegistrarCallerRaw struct {
	Contract *TaskAVSRegistrarCaller // Generic read-only contract binding to access the raw methods on
}

// TaskAVSRegistrarTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TaskAVSRegistrarTransactorRaw struct {
	Contract *TaskAVSRegistrarTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTaskAVSRegistrar creates a new instance of TaskAVSRegistrar, bound to a specific deployed contract.
func NewTaskAVSRegistrar(address common.Address, backend bind.ContractBackend) (*TaskAVSRegistrar, error) {
	contract, err := bindTaskAVSRegistrar(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrar{TaskAVSRegistrarCaller: TaskAVSRegistrarCaller{contract: contract}, TaskAVSRegistrarTransactor: TaskAVSRegistrarTransactor{contract: contract}, TaskAVSRegistrarFilterer: TaskAVSRegistrarFilterer{contract: contract}}, nil
}

// NewTaskAVSRegistrarCaller creates a new read-only instance of TaskAVSRegistrar, bound to a specific deployed contract.
func NewTaskAVSRegistrarCaller(address common.Address, caller bind.ContractCaller) (*TaskAVSRegistrarCaller, error) {
	contract, err := bindTaskAVSRegistrar(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarCaller{contract: contract}, nil
}

// NewTaskAVSRegistrarTransactor creates a new write-only instance of TaskAVSRegistrar, bound to a specific deployed contract.
func NewTaskAVSRegistrarTransactor(address common.Address, transactor bind.ContractTransactor) (*TaskAVSRegistrarTransactor, error) {
	contract, err := bindTaskAVSRegistrar(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarTransactor{contract: contract}, nil
}

// NewTaskAVSRegistrarFilterer creates a new log filterer instance of TaskAVSRegistrar, bound to a specific deployed contract.
func NewTaskAVSRegistrarFilterer(address common.Address, filterer bind.ContractFilterer) (*TaskAVSRegistrarFilterer, error) {
	contract, err := bindTaskAVSRegistrar(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarFilterer{contract: contract}, nil
}

// bindTaskAVSRegistrar binds a generic wrapper to an already deployed contract.
func bindTaskAVSRegistrar(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TaskAVSRegistrarMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrar *TaskAVSRegistrarRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrar.Contract.TaskAVSRegistrarCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrar *TaskAVSRegistrarRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.TaskAVSRegistrarTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrar *TaskAVSRegistrarRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.TaskAVSRegistrarTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TaskAVSRegistrar.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TaskAVSRegistrar *TaskAVSRegistrarTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TaskAVSRegistrar *TaskAVSRegistrarTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.contract.Transact(opts, method, params...)
}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) ALLOCATIONMANAGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "ALLOCATION_MANAGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrar.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrar.CallOpts)
}

// ALLOCATIONMANAGER is a free data retrieval call binding the contract method 0x31232bc9.
//
// Solidity: function ALLOCATION_MANAGER() view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) ALLOCATIONMANAGER() (common.Address, error) {
	return _TaskAVSRegistrar.Contract.ALLOCATIONMANAGER(&_TaskAVSRegistrar.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) AVS(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "AVS")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrar.Contract.AVS(&_TaskAVSRegistrar.CallOpts)
}

// AVS is a free data retrieval call binding the contract method 0xd74a8b61.
//
// Solidity: function AVS() view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) AVS() (common.Address, error) {
	return _TaskAVSRegistrar.Contract.AVS(&_TaskAVSRegistrar.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) PUBKEYREGISTRATIONTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "PUBKEY_REGISTRATION_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrar.CallOpts)
}

// PUBKEYREGISTRATIONTYPEHASH is a free data retrieval call binding the contract method 0x9feab859.
//
// Solidity: function PUBKEY_REGISTRATION_TYPEHASH() view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) PUBKEYREGISTRATIONTYPEHASH() ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.PUBKEYREGISTRATIONTYPEHASH(&_TaskAVSRegistrar.CallOpts)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) CalculatePubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "calculatePubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// CalculatePubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x73447992.
//
// Solidity: function calculatePubkeyRegistrationMessageHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) CalculatePubkeyRegistrationMessageHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.CalculatePubkeyRegistrationMessageHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// DeregisterOperator is a free data retrieval call binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address , address avs, uint32[] ) view returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) DeregisterOperator(opts *bind.CallOpts, arg0 common.Address, avs common.Address, arg2 []uint32) error {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "deregisterOperator", arg0, avs, arg2)

	if err != nil {
		return err
	}

	return err

}

// DeregisterOperator is a free data retrieval call binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address , address avs, uint32[] ) view returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) DeregisterOperator(arg0 common.Address, avs common.Address, arg2 []uint32) error {
	return _TaskAVSRegistrar.Contract.DeregisterOperator(&_TaskAVSRegistrar.CallOpts, arg0, avs, arg2)
}

// DeregisterOperator is a free data retrieval call binding the contract method 0x303ca956.
//
// Solidity: function deregisterOperator(address , address avs, uint32[] ) view returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) DeregisterOperator(arg0 common.Address, avs common.Address, arg2 []uint32) error {
	return _TaskAVSRegistrar.Contract.DeregisterOperator(&_TaskAVSRegistrar.CallOpts, arg0, avs, arg2)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "eip712Domain")

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
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _TaskAVSRegistrar.Contract.Eip712Domain(&_TaskAVSRegistrar.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _TaskAVSRegistrar.Contract.Eip712Domain(&_TaskAVSRegistrar.CallOpts)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "getOperatorFromPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorFromPubkeyHash(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) GetOperatorPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "getOperatorPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorPubkeyHash is a free data retrieval call binding the contract method 0xfd0d930a.
//
// Solidity: function getOperatorPubkeyHash(address operator) view returns(bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) GetOperatorPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorPubkeyHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) GetOperatorSocketByOperator(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "getOperatorSocketByOperator", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorSocketByOperator is a free data retrieval call binding the contract method 0xa30db098.
//
// Solidity: function getOperatorSocketByOperator(address operator) view returns(string)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) GetOperatorSocketByOperator(operator common.Address) (string, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorSocketByOperator(&_TaskAVSRegistrar.CallOpts, operator)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) GetOperatorSocketByPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "getOperatorSocketByPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetOperatorSocketByPubkeyHash is a free data retrieval call binding the contract method 0x9d6f2285.
//
// Solidity: function getOperatorSocketByPubkeyHash(bytes32 pubkeyHash) view returns(string)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) GetOperatorSocketByPubkeyHash(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrar.Contract.GetOperatorSocketByPubkeyHash(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), (uint256[2],uint256[2]), bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, BN254G2Point, [32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "getRegisteredPubkey", operator)

	if err != nil {
		return *new(BN254G1Point), *new(BN254G2Point), *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)
	out1 := *abi.ConvertType(out[1], new(BN254G2Point)).(*BN254G2Point)
	out2 := *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)

	return out0, out1, out2, err

}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), (uint256[2],uint256[2]), bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, BN254G2Point, [32]byte, error) {
	return _TaskAVSRegistrar.Contract.GetRegisteredPubkey(&_TaskAVSRegistrar.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), (uint256[2],uint256[2]), bytes32)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, BN254G2Point, [32]byte, error) {
	return _TaskAVSRegistrar.Contract.GetRegisteredPubkey(&_TaskAVSRegistrar.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) OperatorToPubkey(opts *bind.CallOpts, operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "operatorToPubkey", operator)

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
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrar.Contract.OperatorToPubkey(&_TaskAVSRegistrar.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address operator) view returns(uint256 X, uint256 Y)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) OperatorToPubkey(operator common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _TaskAVSRegistrar.Contract.OperatorToPubkey(&_TaskAVSRegistrar.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) OperatorToPubkeyHash(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "operatorToPubkeyHash", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address operator) view returns(bytes32 pubkeyHash)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) OperatorToPubkeyHash(operator common.Address) ([32]byte, error) {
	return _TaskAVSRegistrar.Contract.OperatorToPubkeyHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) OperatorToSocket(opts *bind.CallOpts, operator common.Address) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "operatorToSocket", operator)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrar.Contract.OperatorToSocket(&_TaskAVSRegistrar.CallOpts, operator)
}

// OperatorToSocket is a free data retrieval call binding the contract method 0x39c26f42.
//
// Solidity: function operatorToSocket(address operator) view returns(string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) OperatorToSocket(operator common.Address) (string, error) {
	return _TaskAVSRegistrar.Contract.OperatorToSocket(&_TaskAVSRegistrar.CallOpts, operator)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) PubkeyHashToOperator(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "pubkeyHashToOperator", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrar.Contract.PubkeyHashToOperator(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 pubkeyHash) view returns(address operator)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) PubkeyHashToOperator(pubkeyHash [32]byte) (common.Address, error) {
	return _TaskAVSRegistrar.Contract.PubkeyHashToOperator(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) PubkeyHashToSocket(opts *bind.CallOpts, pubkeyHash [32]byte) (string, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "pubkeyHashToSocket", pubkeyHash)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrar.Contract.PubkeyHashToSocket(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// PubkeyHashToSocket is a free data retrieval call binding the contract method 0x69e5aa8b.
//
// Solidity: function pubkeyHashToSocket(bytes32 pubkeyHash) view returns(string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) PubkeyHashToSocket(pubkeyHash [32]byte) (string, error) {
	return _TaskAVSRegistrar.Contract.PubkeyHashToSocket(&_TaskAVSRegistrar.CallOpts, pubkeyHash)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) PubkeyRegistrationMessageHash(opts *bind.CallOpts, operator common.Address) (BN254G1Point, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "pubkeyRegistrationMessageHash", operator)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrar.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// PubkeyRegistrationMessageHash is a free data retrieval call binding the contract method 0x3c2a7f4c.
//
// Solidity: function pubkeyRegistrationMessageHash(address operator) view returns((uint256,uint256))
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) PubkeyRegistrationMessageHash(operator common.Address) (BN254G1Point, error) {
	return _TaskAVSRegistrar.Contract.PubkeyRegistrationMessageHash(&_TaskAVSRegistrar.CallOpts, operator)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrar *TaskAVSRegistrarCaller) SupportsAVS(opts *bind.CallOpts, avs common.Address) (bool, error) {
	var out []interface{}
	err := _TaskAVSRegistrar.contract.Call(opts, &out, "supportsAVS", avs)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrar.Contract.SupportsAVS(&_TaskAVSRegistrar.CallOpts, avs)
}

// SupportsAVS is a free data retrieval call binding the contract method 0xb5265787.
//
// Solidity: function supportsAVS(address avs) view returns(bool)
func (_TaskAVSRegistrar *TaskAVSRegistrarCallerSession) SupportsAVS(avs common.Address) (bool, error) {
	return _TaskAVSRegistrar.Contract.SupportsAVS(&_TaskAVSRegistrar.CallOpts, avs)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] , bytes data) returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, avs common.Address, arg2 []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrar.contract.Transact(opts, "registerOperator", operator, avs, arg2, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] , bytes data) returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) RegisterOperator(operator common.Address, avs common.Address, arg2 []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.RegisterOperator(&_TaskAVSRegistrar.TransactOpts, operator, avs, arg2, data)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0xc63fd502.
//
// Solidity: function registerOperator(address operator, address avs, uint32[] , bytes data) returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarTransactorSession) RegisterOperator(operator common.Address, avs common.Address, arg2 []uint32, data []byte) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.RegisterOperator(&_TaskAVSRegistrar.TransactOpts, operator, avs, arg2, data)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarTransactor) UpdateOperatorSocket(opts *bind.TransactOpts, socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrar.contract.Transact(opts, "updateOperatorSocket", socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.UpdateOperatorSocket(&_TaskAVSRegistrar.TransactOpts, socket)
}

// UpdateOperatorSocket is a paid mutator transaction binding the contract method 0xc95e97da.
//
// Solidity: function updateOperatorSocket(string socket) returns()
func (_TaskAVSRegistrar *TaskAVSRegistrarTransactorSession) UpdateOperatorSocket(socket string) (*types.Transaction, error) {
	return _TaskAVSRegistrar.Contract.UpdateOperatorSocket(&_TaskAVSRegistrar.TransactOpts, socket)
}

// TaskAVSRegistrarEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the TaskAVSRegistrar contract.
type TaskAVSRegistrarEIP712DomainChangedIterator struct {
	Event *TaskAVSRegistrarEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarEIP712DomainChanged)
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
		it.Event = new(TaskAVSRegistrarEIP712DomainChanged)
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
func (it *TaskAVSRegistrarEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarEIP712DomainChanged represents a EIP712DomainChanged event raised by the TaskAVSRegistrar contract.
type TaskAVSRegistrarEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*TaskAVSRegistrarEIP712DomainChangedIterator, error) {

	logs, sub, err := _TaskAVSRegistrar.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarEIP712DomainChangedIterator{contract: _TaskAVSRegistrar.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _TaskAVSRegistrar.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarEIP712DomainChanged)
				if err := _TaskAVSRegistrar.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) ParseEIP712DomainChanged(log types.Log) (*TaskAVSRegistrarEIP712DomainChanged, error) {
	event := new(TaskAVSRegistrarEIP712DomainChanged)
	if err := _TaskAVSRegistrar.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarNewPubkeyRegistrationIterator is returned from FilterNewPubkeyRegistration and is used to iterate over the raw logs and unpacked data for NewPubkeyRegistration events raised by the TaskAVSRegistrar contract.
type TaskAVSRegistrarNewPubkeyRegistrationIterator struct {
	Event *TaskAVSRegistrarNewPubkeyRegistration // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarNewPubkeyRegistrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarNewPubkeyRegistration)
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
		it.Event = new(TaskAVSRegistrarNewPubkeyRegistration)
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
func (it *TaskAVSRegistrarNewPubkeyRegistrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarNewPubkeyRegistrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarNewPubkeyRegistration represents a NewPubkeyRegistration event raised by the TaskAVSRegistrar contract.
type TaskAVSRegistrarNewPubkeyRegistration struct {
	Operator   common.Address
	PubkeyHash [32]byte
	PubkeyG1   BN254G1Point
	PubkeyG2   BN254G2Point
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewPubkeyRegistration is a free log retrieval operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarNewPubkeyRegistrationIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrar.contract.FilterLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarNewPubkeyRegistrationIterator{contract: _TaskAVSRegistrar.contract, event: "NewPubkeyRegistration", logs: logs, sub: sub}, nil
}

// WatchNewPubkeyRegistration is a free log subscription operation binding the contract event 0xf9e46291596d111f263d5bc0e4ee38ae179bde090419c91be27507ce8bc6272e.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, bytes32 indexed pubkeyHash, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarNewPubkeyRegistration, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrar.contract.WatchLogs(opts, "NewPubkeyRegistration", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarNewPubkeyRegistration)
				if err := _TaskAVSRegistrar.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
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
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) ParseNewPubkeyRegistration(log types.Log) (*TaskAVSRegistrarNewPubkeyRegistration, error) {
	event := new(TaskAVSRegistrarNewPubkeyRegistration)
	if err := _TaskAVSRegistrar.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TaskAVSRegistrarOperatorSocketUpdatedIterator is returned from FilterOperatorSocketUpdated and is used to iterate over the raw logs and unpacked data for OperatorSocketUpdated events raised by the TaskAVSRegistrar contract.
type TaskAVSRegistrarOperatorSocketUpdatedIterator struct {
	Event *TaskAVSRegistrarOperatorSocketUpdated // Event containing the contract specifics and raw log

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
func (it *TaskAVSRegistrarOperatorSocketUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TaskAVSRegistrarOperatorSocketUpdated)
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
		it.Event = new(TaskAVSRegistrarOperatorSocketUpdated)
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
func (it *TaskAVSRegistrarOperatorSocketUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TaskAVSRegistrarOperatorSocketUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TaskAVSRegistrarOperatorSocketUpdated represents a OperatorSocketUpdated event raised by the TaskAVSRegistrar contract.
type TaskAVSRegistrarOperatorSocketUpdated struct {
	Operator   common.Address
	PubkeyHash [32]byte
	Socket     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOperatorSocketUpdated is a free log retrieval operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) FilterOperatorSocketUpdated(opts *bind.FilterOpts, operator []common.Address, pubkeyHash [][32]byte) (*TaskAVSRegistrarOperatorSocketUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrar.contract.FilterLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return &TaskAVSRegistrarOperatorSocketUpdatedIterator{contract: _TaskAVSRegistrar.contract, event: "OperatorSocketUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorSocketUpdated is a free log subscription operation binding the contract event 0xa59c022be52f7db360b7c5ce8556c8337ff4784e694a9aec508e6b2eeb8e540a.
//
// Solidity: event OperatorSocketUpdated(address indexed operator, bytes32 indexed pubkeyHash, string socket)
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) WatchOperatorSocketUpdated(opts *bind.WatchOpts, sink chan<- *TaskAVSRegistrarOperatorSocketUpdated, operator []common.Address, pubkeyHash [][32]byte) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var pubkeyHashRule []interface{}
	for _, pubkeyHashItem := range pubkeyHash {
		pubkeyHashRule = append(pubkeyHashRule, pubkeyHashItem)
	}

	logs, sub, err := _TaskAVSRegistrar.contract.WatchLogs(opts, "OperatorSocketUpdated", operatorRule, pubkeyHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TaskAVSRegistrarOperatorSocketUpdated)
				if err := _TaskAVSRegistrar.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
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
func (_TaskAVSRegistrar *TaskAVSRegistrarFilterer) ParseOperatorSocketUpdated(log types.Log) (*TaskAVSRegistrarOperatorSocketUpdated, error) {
	event := new(TaskAVSRegistrarOperatorSocketUpdated)
	if err := _TaskAVSRegistrar.contract.UnpackLog(event, "OperatorSocketUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
