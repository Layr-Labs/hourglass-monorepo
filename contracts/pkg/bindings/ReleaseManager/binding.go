// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ReleaseManager

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

// ReleaseManagerMetaData contains all meta data concerning the ReleaseManager contract.
var ReleaseManagerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"allPromotedArtifacts\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"artifactExists\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"artifacts\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"deregister\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.Artifact\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLatestPromotedArtifact\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.PromotedArtifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotedArtifactAtBlock\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIReleaseManager.PromotedArtifact\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotedArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.PromotedArtifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionCheckpointAt\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"pos\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"artifactIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionCheckpointCount\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionHistory\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.PromotedArtifact[]\",\"components\":[{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"promotedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPromotionStatusAtBlock\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_permissionController\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"permissionController\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPermissionController\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"promoteArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"promotions\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.ArtifactPromotion[]\",\"components\":[{\"name\":\"promotionStatus\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"operatorSetIds\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}]},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"publishArtifacts\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_artifacts\",\"type\":\"tuple[]\",\"internalType\":\"structIReleaseManager.Artifact[]\",\"components\":[{\"name\":\"artifactType\",\"type\":\"uint8\",\"internalType\":\"enumArtifactType\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"internalType\":\"enumOperatingSystem\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"publishedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"register\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registeredAVS\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updatePromotionStatus\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"newStatus\",\"type\":\"uint8\",\"internalType\":\"enumPromotionStatus\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AVSDeregistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AVSRegistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ArtifactPublished\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"registryUrl\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"architecture\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumArchitecture\"},{\"name\":\"os\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOperatingSystem\"},{\"name\":\"artifactType\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumArtifactType\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ArtifactsPromoted\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"version\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"deploymentDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"digests\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"statuses\",\"type\":\"uint8[]\",\"indexed\":false,\"internalType\":\"enumPromotionStatus[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PromotionStatusUpdated\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"digest\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"operatorSetId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"oldStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumPromotionStatus\"},{\"name\":\"newStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumPromotionStatus\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AVSAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AVSNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArrayLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArtifactNotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDeadline\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Unauthorized\",\"inputs\":[]}]",
	Bin: "0x6080806040523460195760015f55612655908161001e8239f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c806320eb2c99146101745780633446128f1461016f5780633642b6ff1461016a5780634420e48614610165578063447039bd146101605780634657e26a1461015b578063715018a614610156578063716d0d9b146101515780637d708ea61461014c57806384ac33ec146101475780638da5cb5b1461014257806397f1c2c91461013d578063a6210e1814610138578063a63e3a3714610133578063ae7389201461012e578063b2daca5c14610129578063b501f96f14610124578063bf2d8e071461011f578063c4d66de81461011a578063d2d104ef14610115578063eba04f8a146101105763f2fde38b1461010b575f80fd5b61123b565b6110e7565b611057565b610efb565b610ebe565b610e8a565b610e5b565b610ddd565b610d08565b610aaa565b610a16565b6109da565b610954565b61076d565b6106c6565b61063b565b610613565b610581565b6104cf565b6103fc565b61038d565b6102a1565b600435906001600160a01b038216820361018f57565b5f80fd5b805180835260209291819084018484015e5f828201840152601f01601f1916010190565b634e487b7160e01b5f52602160045260245ffd5b600411156101d557565b6101b7565b9060048210156101d55752565b908151815260a08061022f61020b602086015160c0602087015260c0860190610193565b61021d604087015160408701906101da565b60608601518582036060870152610193565b9360808101516080850152015191015290565b602081016020825282518091526040820191602060408360051b8301019401925f915b83831061027457505050505090565b9091929394602080610292600193603f1986820301875289516101e7565b97019301930191939290610265565b3461018f57604036600319011261018f576102ba610179565b6001600160a01b03165f9081526069602090815260408083206024358452909152902080546102e8816112e6565b916102f66040519384610bda565b81835260208301905f5260205f205f915b838310610320576040518061031c8782610242565b0390f35b6006602060019260405161033381610b84565b85548152610342858701610bfb565b8382015261035a60ff600288015416604083016112fd565b61036660038701610bfb565b606082015260048601546080820152600586015460a0820152815201920192019190610307565b3461018f57604036600319011261018f576103a6610179565b6024359060018060a01b03165f52606b60205260405f20905f52602052602060405f2054604051908152f35b606090600319011261018f576004356001600160a01b038116810361018f57906024359060443590565b3461018f5761040a366103d2565b60018060a01b0383165f52606b60205260405f20825f5260205260405f205411156104975761031c61046d6104686104759361045b63ffffffff9660018060a01b03165f52606b60205260405f2090565b905f5260205260405f2090565b611fa7565b9390916113ae565b604080519490911684526001600160e01b039290921660208401528291820190565b60405162461bcd60e51b815260206004820152601060248201526f24b73b30b634b2103837b9b4ba34b7b760811b6044820152606490fd5b3461018f57602036600319011261018f576104e8610179565b6104f23382612028565b15610561576001600160a01b03165f8181526067602052604090205460ff1661055257805f52606760205260405f20600160ff198254161790557f2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e5f80a2005b63886f069560e01b5f5260045ffd5b6282b42960e81b5f5260045ffd5b60208101929161057f91906101da565b565b3461018f57608036600319011261018f576105fa61059d610179565b60405160609190911b6bffffffffffffffffffffffff19166020820190815260243560348301526044356054830152606435916105e781607481015b03601f198101835282610bda565b5190205f52606c60205260405f206120cb565b60048110156101d55761031c906040519182918261056f565b3461018f575f36600319011261018f576066546040516001600160a01b039091168152602090f35b3461018f575f36600319011261018f576106536121bc565b603480546001600160a01b031981169091555f906001600160a01b03167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e08280a3005b9181601f8401121561018f578235916001600160401b03831161018f576020808501948460051b01011161018f57565b3461018f57608036600319011261018f576106df610179565b6024356001600160401b03811161018f576106fe903690600401610696565b604492919235916001600160401b03831161018f573660238401121561018f578260040135916001600160401b03831161018f57366024848601011161018f5761074f9460246064359501926113eb565b005b60643590600482101561018f57565b3590600482101561018f57565b3461018f57608036600319011261018f57610786610179565b60243590604435610795610751565b9261079e61225c565b6001600160a01b0383165f8181526067602052604090205460ff1615610945576107c83385612028565b15610561576108026107f66107f18561045b8860018060a01b03165f52606b60205260405f2090565b612374565b6001600160e01b031690565b946108218461045b8760018060a01b03165f52606960205260405f2090565b95865415610853575f96846108368383610d76565b505414610862575b50505050505050156108535761074f60015f55565b630c03e12160e31b5f5260045ffd5b7f1341ac2513ee1dd81316f76f3a7840608cd4dc5a3302d2fb45ea1ad24d35567193949596975060026108996108a0938593610d76565b5001611967565b60405160609690961b6bffffffffffffffffffffffff19166020870190815260348701869052605487018590526001966108dd81607481016105d9565b519020610925826109176108ff6107f66107f1865f52606c60205260405f2090565b93610909856101cb565b5f52606c60205260405f2090565b610920826101cb565b6122e4565b505061093660405192839283611b7b565b0390a45f80808080808061083e565b635ae33df360e11b5f5260045ffd5b3461018f57602036600319011261018f5761096d610179565b6109773382612028565b15610561576001600160a01b03165f8181526067602052604090205460ff161561094557805f5260676020526109b460405f2060ff198154169055565b7ff7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab75f80a2005b3461018f575f36600319011261018f576034546040516001600160a01b039091168152602090f35b906020610a139281815201906101e7565b90565b3461018f57604036600319011261018f57610a2f610179565b60243590610a3b611b99565b5060018060a01b031690815f52606b60205260405f20815f5260205260018060e01b03610a6a60405f20612374565b16915f52606960205260405f20905f52602052610a8960405f20611309565b8051156108535761031c91610a9d91611826565b5160405191829182610a02565b3461018f57610af1610b0f610b0a610ac1366103d2565b949091610acc611b99565b5060018060a01b031694855f52606b60205260405f20835f5260205260405f206120cb565b935f52606960205260405f20905f5260205260405f2090565b611309565b80518015908115610b2d575b506108535761031c91610a9d91611826565b90508210155f610b1b565b90600182811c92168015610b66575b6020831014610b5257565b634e487b7160e01b5f52602260045260245ffd5b91607f1691610b47565b634e487b7160e01b5f52604160045260245ffd5b60c081019081106001600160401b03821117610b9f57604052565b610b70565b608081019081106001600160401b03821117610b9f57604052565b604081019081106001600160401b03821117610b9f57604052565b90601f801991011681019081106001600160401b03821117610b9f57604052565b9060405191825f825492610c0e84610b38565b8084529360018116908115610c775750600114610c33575b5061057f92500383610bda565b90505f9291925260205f20905f915b818310610c5b57505090602061057f928201015f610c26565b6020919350806001915483858901015201910190918492610c42565b90506020925061057f94915060ff191682840152151560051b8201015f610c26565b600211156101d557565b600311156101d557565b9060038210156101d55752565b96959491610d0393610ced9160a09693610cd381610c99565b8a52610cde81610c99565b60208a01526040890190610cad565b606087015260c0608087015260c0860190610193565b930152565b3461018f57602036600319011261018f576004355f52606860205260405f20805461031c6001830154926003610d4060028301610bfb565b91015490604051948460ff879660101c169060ff808260081c16911687610cba565b634e487b7160e01b5f52603260045260245ffd5b8054821015610d8f575f52600660205f20910201905f90565b610d62565b929093610dc6610dbb610dd39460a0979a99989a875260c0602088015260c0870190610193565b9260408601906101da565b8382036060850152610193565b9460808201520152565b3461018f57610deb366103d2565b9160018060a01b03165f52606960205260405f20905f5260205260405f2090815481101561018f57610e1c91610d76565b50805461031c610e2e60018401610bfb565b9260ff60028201541690610e4460038201610bfb565b600560048301549201549260405196879687610d94565b3461018f57602036600319011261018f576004355f52606a602052602060ff60405f2054166040519015158152f35b3461018f57604036600319011261018f5761031c610eb2610ea9610179565b60243590611c0a565b60405191829182610242565b3461018f57602036600319011261018f576001600160a01b03610edf610179565b165f526067602052602060ff60405f2054166040519015158152f35b3461018f57602036600319011261018f57610f14610179565b610f6460015491610f49610f33610f2f8560ff9060081c1690565b1590565b80948195610fe2575b8115610fc2575b50611cc2565b82610f5b600160ff1981541617600155565b610fa957611d25565b610f6a57005b610f7a60015461ff001916600155565b604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb384740249890602090a1005b610fbd61010060015461ff00191617600155565b611d25565b303b15915081610fd4575b505f610f43565b60ff1660011490505f610fcd565b600160ff8216109150610f3c565b602081528151610fff81610c99565b6020820152602082015161101281610c99565b604082015261102960408301516060830190610cad565b6060820151608082015260c060a061104e6080850151838386015260e0850190610193565b93015191015290565b3461018f57604036600319011261018f576110a9611073610179565b602435905f60a060405161108681610b84565b8281528260208201528260408201528260608201526060608082015201526122b0565b805f52606a60205260ff60405f20541615610853576110db6110d661031c925f52606860205260405f2090565b611d65565b60405191829182610ff0565b3461018f57604036600319011261018f57611100610179565b6024356001600160401b03811161018f5761111f903690600401610696565b909161112961225c565b6001600160a01b0381165f8181526067602052604090205460ff1615610945576111533383612028565b15610561575f5b83811061116a5761074f60015f55565b8061118061117b6001938789611dd4565b611e18565b4260a0820152837f622e1c25f0e4dcedadb24c8f0bbe6ceaa3776cbeb58b17a2d5d8ac8ae31a7502606083016111f36111e66111bd83518b6122b0565b6111d8876111d3835f52606860205260405f2090565b611ebb565b5f52606a60205260405f2090565b805460ff19166001179055565b5192608081015161123260208301519261120c84610c99565b60408101519061121b82610ca3565b519061122682610c99565b60405194859485611f43565b0390a30161115a565b3461018f57602036600319011261018f57611254610179565b61125c6121bc565b6001600160a01b038116156112745761074f90612214565b60405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608490fd5b6040519061057f60c083610bda565b6040519061057f604083610bda565b6001600160401b038111610b9f5760051b60200190565b60048210156101d55752565b908154611315816112e6565b926113236040519485610bda565b81845260208401905f5260205f205f915b8383106113415750505050565b6006602060019260405161135481610b84565b85548152611363858701610bfb565b8382015261137b60ff600288015416604083016112fd565b61138760038701610bfb565b606082015260048601546080820152600586015460a0820152815201920192019190611334565b156113b557565b60405162461bcd60e51b815260206004820152600e60248201526d4e6f20636865636b706f696e747360901b6044820152606490fd5b9390929491946113f961225c565b6001600160a01b0385165f8181526067602052604090205490959060ff1615610945576114263382612028565b15610561574284111561168d57939561143e8161169c565b946114488261169c565b975f905b8382106114a057505050506114877f7fd55cb6307da041fb4711b42eb59c940bb12e76dc208c9be19be9abde37815d94939261149492611ae5565b9560405193849384611af9565b0390a361057f60015f55565b6114bd6114b88386849d999a979c9d9b98969b6116ce565b611750565b9860208a01966114df610f2f6114d86111d88d8c51906122b0565b5460ff1690565b6108535790919a8a9660409c89516114f78784611826565b526115158951611506816101cb565b6115108887611826565b6112fd565b5f60608a019e8f9a01995b518051821015611673578f9695949392918f8f908f8f928f918f8f956115c6611659956116489361158f6116659a61155a8d60019f611826565b5197611588875195518d519061156f826101cb565b6115776112c8565b9788526020880152604087016112fd565b36916116f0565b606083015260808201524260a08201526001600160a01b0386165f9081526069602052604090206115c190869061045b565b61197f565b6116126115f06115ea8561045b8860018060a01b03165f52606960205260405f2090565b54611ad2565b6001600160a01b0386165f908152606b6020526040902061092090869061045b565b505051916105d960405193849260208401968791605493916001600160601b03199060601b168352601483015260348201520190565b5190205f52606c60205260405f2090565b905190610917826101cb565b505001909192939495611520565b50509597509995989750995099509060010190929161144c565b631da7447960e21b5f5260045ffd5b906116a6826112e6565b6116b36040519182610bda565b82815280926116c4601f19916112e6565b0190602036910137565b9190811015610d8f5760051b81013590607e198136030182121561018f570190565b9291926001600160401b038211610b9f5760405191611719601f8201601f191660200184610bda565b82948184528183011161018f578281602093845f960137010152565b9080601f8301121561018f57816020610a13933591016116f0565b60808136031261018f576040519061176782610ba4565b61177081610760565b82526020810135602083015260408101356001600160401b03811161018f5761179c9036908301611735565b60408301526060810135906001600160401b03821161018f570136601f8201121561018f578035906117cd826112e6565b916117db6040519384610bda565b80835260208084019160051b8301019136831161018f57602001905b82821061180957505050606082015290565b81358152602091820191016117f7565b805115610d8f5760200190565b8051821015610d8f5760209160051b010190565b634e487b7160e01b5f525f60045260245ffd5b601f821161185a57505050565b5f5260205f20906020601f840160051c83019310611892575b601f0160051c01905b818110611887575050565b5f815560010161187c565b9091508190611873565b91909182516001600160401b038111610b9f576118c3816118bd8454610b38565b8461184d565b6020601f82116001146119025781906118f39394955f926118f7575b50508160011b915f199060031b1c19161790565b9055565b015190505f806118df565b601f19821690611915845f5260205f2090565b915f5b81811061194f57509583600195969710611937575b505050811b019055565b01515f1960f88460031b161c191690555f808061192d565b9192602060018192868b015181550194019201611918565b9060048110156101d55760ff80198354169116179055565b8054600160401b811015610b9f5761199c91600182018155610d76565b919091611ab957805182556001820160208201518051906001600160401b038211610b9f576119d5826119cf8554610b38565b8561184d565b602090601f8311600114611a4957826005959360a09593611a0a935f926118f75750508160011b915f199060031b1c19161790565b90555b611a276040820151611a1e816101cb565b60028601611967565b611a3860608201516003860161189c565b608081015160048501550151910155565b90601f19831691611a5d855f5260205f2090565b925f5b818110611aa157509260019285926005989660a0989610611a89575b505050811b019055611a0d565b01515f1960f88460031b161c191690555f8080611a7c565b92936020600181928786015181550195019301611a60565b61183a565b634e487b7160e01b5f52601160045260245ffd5b5f19810191908211611ae057565b611abe565b81604051928392833781015f815203902090565b90606082019082526060602083015282518091526020608083019301905f5b818110611b65575050506040818303910152602080835192838152019201905f5b818110611b465750505090565b90919260208082611b5a60019488516101da565b019401929101611b39565b8251855260209485019490920191600101611b18565b91602061057f929493611b928160408101976101da565b01906101da565b60405190611ba682610b84565b5f60a083828152606060208201528260408201526060808201528260808201520152565b60408051909190611bdb8382610bda565b6001815291601f1901825f5b828110611bf357505050565b602090611bfe611b99565b82828501015201611be7565b60018060a01b031690815f52606b60205260405f20815f5260205260018060e01b03611c3860405f20612374565b16915f52606960205260405f20905f52602052611c5760405f20611309565b90815115611c8957611c7190611c6b611bca565b92611826565b51611c7b82611819565b52611c8581611819565b5090565b5050604051611c99602082610bda565b5f81525f805b818110611cab57505090565b602090611cb6611b99565b82828601015201611c9f565b15611cc957565b60405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b6064820152608490fd5b611d3f60ff60015460081c16611d3a8161239e565b61239e565b611d4833612214565b60018060a01b03166001600160601b0360a01b6066541617606655565b90604051611d7281610b84565b809260ff8154818116611d8481610c99565b8452818160081c16611d9581610c99565b602085015260101c1660038110156101d55760a091600391604085015260018101546060850152611dc860028201610bfb565b60808501520154910152565b9190811015610d8f5760051b8101359060be198136030182121561018f570190565b6002111561018f57565b359061057f82611df6565b3590600382101561018f57565b60c08136031261018f5760405190611e2f82610b84565b8035611e3a81611df6565b8252611e4860208201611e00565b6020830152611e5960408201611e0b565b6040830152606081013560608301526080810135906001600160401b03821161018f57611e8b60a09236908301611735565b6080840152013560a082015290565b9060038110156101d55762ff000082549160101b169062ff00001916179055565b908051611ec781610c99565b611ed081610c99565b60ff801984541691161780835561ff006020830151611eee81610c99565b611ef781610c99565b60081b169061ff001916178255604081015160038110156101d557600391611f2160a09285611e9a565b60608101516001850155611f3c60808201516002860161189c565b0151910155565b90606092959493611f5f611f7892608085526080850190610193565b96611f6981610c99565b60208401526040830190610cad565b611f8183610c99565b0152565b90604051611f9281610bbf565b602081935463ffffffff81168352811c910152565b80549081611fb95750505f905f905f90565b815f19810111611ae0575f525f199060205f20010190602060405192611fde84610bbf565b5463ffffffff811680855290821c91909301819052600192916001600160e01b0390911690565b9081602091031261018f5751801515810361018f5790565b6040513d5f823e3d90fd5b6001600160a01b03818116908316146120c4576066546001600160a01b03169182612054575050505f90565b604051639100674560e01b81526001600160a01b0392831660048201529116602482015290602090829060449082905afa9081156120bf575f91612096575090565b610a13915060203d6020116120b8575b6120b08183610bda565b810190612005565b503d6120a6565b61201d565b5050600190565b9043811015612178576120dd906123fe565b81549063ffffffff165f5b828110612124575050806120fd57505f919050565b61211f9161210d61211892611ad2565b905f5260205f200190565b5460201c90565b6107f6565b90918082169080831860011c8201809211611ae057845f528363ffffffff6121568460205f200163ffffffff90541690565b1611156121665750915b906120e8565b92915061217290612466565b90612160565b606460405162461bcd60e51b815260206004820152602060248201527f436865636b706f696e74733a20626c6f636b206e6f7420796574206d696e65646044820152fd5b6034546001600160a01b031633036121d057565b606460405162461bcd60e51b815260206004820152602060248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152fd5b603480546001600160a01b039283166001600160a01b0319821681179092559091167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e05f80a3565b60025f541461226b5760025f55565b60405162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c006044820152606490fd5b906040519060208201926001600160601b03199060601b1683526034820152603481526122de605482610bda565b51902090565b6122ed436123fe565b6001600160e01b03831161231f5761230e926001600160e01b031691612509565b6001600160e01b0391821692911690565b60405162461bcd60e51b815260206004820152602760248201527f53616665436173743a2076616c756520646f65736e27742066697420696e20326044820152663234206269747360c81b6064820152608490fd5b805490816123825750505f90565b815f19810111611ae0575f525f199060205f2001015460201c90565b156123a557565b60405162461bcd60e51b815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201526a6e697469616c697a696e6760a81b6064820152608490fd5b63ffffffff81116124125763ffffffff1690565b60405162461bcd60e51b815260206004820152602660248201527f53616665436173743a2076616c756520646f65736e27742066697420696e203360448201526532206269747360d01b6064820152608490fd5b9060018201809211611ae057565b908154600160401b811015610b9f5760018101808455811015610d8f575f92835260209283902082519284015190931b63ffffffff191663ffffffff9290921691909117910155565b156124c457565b60405162461bcd60e51b815260206004820152601b60248201527f436865636b706f696e743a2064656372656173696e67206b65797300000000006044820152606490fd5b909291928382548015155f146125f55792602092918461254161253c6125316125bb98611ad2565b855f5260205f200190565b611f85565b9363ffffffff612566612558875163ffffffff1690565b8284169283911611156124bd565b612580612577875163ffffffff1690565b63ffffffff1690565b036125bf57506125ad9261210d61259692611ad2565b9063ffffffff82549181199060201b169116179055565b01516001600160e01b031690565b9190565b9150506125f0916125dd6125d16112d7565b63ffffffff9093168352565b6001600160e01b03881682860152612474565b6125ad565b505061261a916126066125d16112d7565b6001600160e01b0385166020830152612474565b5f919056fea2646970667358221220e5c773b15719eb1cd3fb0d27f7382159d493a805b7cbaa5cfdb65e8975cc9ba664736f6c634300081b0033",
}

// ReleaseManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use ReleaseManagerMetaData.ABI instead.
var ReleaseManagerABI = ReleaseManagerMetaData.ABI

// ReleaseManagerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ReleaseManagerMetaData.Bin instead.
var ReleaseManagerBin = ReleaseManagerMetaData.Bin

// DeployReleaseManager deploys a new Ethereum contract, binding an instance of ReleaseManager to it.
func DeployReleaseManager(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ReleaseManager, error) {
	parsed, err := ReleaseManagerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ReleaseManagerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ReleaseManager{ReleaseManagerCaller: ReleaseManagerCaller{contract: contract}, ReleaseManagerTransactor: ReleaseManagerTransactor{contract: contract}, ReleaseManagerFilterer: ReleaseManagerFilterer{contract: contract}}, nil
}

// ReleaseManager is an auto generated Go binding around an Ethereum contract.
type ReleaseManager struct {
	ReleaseManagerCaller     // Read-only binding to the contract
	ReleaseManagerTransactor // Write-only binding to the contract
	ReleaseManagerFilterer   // Log filterer for contract events
}

// ReleaseManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReleaseManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReleaseManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReleaseManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReleaseManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ReleaseManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReleaseManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReleaseManagerSession struct {
	Contract     *ReleaseManager   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ReleaseManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReleaseManagerCallerSession struct {
	Contract *ReleaseManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ReleaseManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReleaseManagerTransactorSession struct {
	Contract     *ReleaseManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ReleaseManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReleaseManagerRaw struct {
	Contract *ReleaseManager // Generic contract binding to access the raw methods on
}

// ReleaseManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReleaseManagerCallerRaw struct {
	Contract *ReleaseManagerCaller // Generic read-only contract binding to access the raw methods on
}

// ReleaseManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReleaseManagerTransactorRaw struct {
	Contract *ReleaseManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReleaseManager creates a new instance of ReleaseManager, bound to a specific deployed contract.
func NewReleaseManager(address common.Address, backend bind.ContractBackend) (*ReleaseManager, error) {
	contract, err := bindReleaseManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ReleaseManager{ReleaseManagerCaller: ReleaseManagerCaller{contract: contract}, ReleaseManagerTransactor: ReleaseManagerTransactor{contract: contract}, ReleaseManagerFilterer: ReleaseManagerFilterer{contract: contract}}, nil
}

// NewReleaseManagerCaller creates a new read-only instance of ReleaseManager, bound to a specific deployed contract.
func NewReleaseManagerCaller(address common.Address, caller bind.ContractCaller) (*ReleaseManagerCaller, error) {
	contract, err := bindReleaseManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerCaller{contract: contract}, nil
}

// NewReleaseManagerTransactor creates a new write-only instance of ReleaseManager, bound to a specific deployed contract.
func NewReleaseManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*ReleaseManagerTransactor, error) {
	contract, err := bindReleaseManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerTransactor{contract: contract}, nil
}

// NewReleaseManagerFilterer creates a new log filterer instance of ReleaseManager, bound to a specific deployed contract.
func NewReleaseManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*ReleaseManagerFilterer, error) {
	contract, err := bindReleaseManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerFilterer{contract: contract}, nil
}

// bindReleaseManager binds a generic wrapper to an already deployed contract.
func bindReleaseManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ReleaseManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReleaseManager *ReleaseManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ReleaseManager.Contract.ReleaseManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReleaseManager *ReleaseManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReleaseManager.Contract.ReleaseManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReleaseManager *ReleaseManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReleaseManager.Contract.ReleaseManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReleaseManager *ReleaseManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ReleaseManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReleaseManager *ReleaseManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReleaseManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReleaseManager *ReleaseManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReleaseManager.Contract.contract.Transact(opts, method, params...)
}

// AllPromotedArtifacts is a free data retrieval call binding the contract method 0xae738920.
//
// Solidity: function allPromotedArtifacts(address , bytes32 , uint256 ) view returns(bytes32 digest, string registryUrl, uint8 status, string version, uint256 deploymentDeadline, uint256 promotedAt)
func (_ReleaseManager *ReleaseManagerCaller) AllPromotedArtifacts(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte, arg2 *big.Int) (struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "allPromotedArtifacts", arg0, arg1, arg2)

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
func (_ReleaseManager *ReleaseManagerSession) AllPromotedArtifacts(arg0 common.Address, arg1 [32]byte, arg2 *big.Int) (struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}, error) {
	return _ReleaseManager.Contract.AllPromotedArtifacts(&_ReleaseManager.CallOpts, arg0, arg1, arg2)
}

// AllPromotedArtifacts is a free data retrieval call binding the contract method 0xae738920.
//
// Solidity: function allPromotedArtifacts(address , bytes32 , uint256 ) view returns(bytes32 digest, string registryUrl, uint8 status, string version, uint256 deploymentDeadline, uint256 promotedAt)
func (_ReleaseManager *ReleaseManagerCallerSession) AllPromotedArtifacts(arg0 common.Address, arg1 [32]byte, arg2 *big.Int) (struct {
	Digest             [32]byte
	RegistryUrl        string
	Status             uint8
	Version            string
	DeploymentDeadline *big.Int
	PromotedAt         *big.Int
}, error) {
	return _ReleaseManager.Contract.AllPromotedArtifacts(&_ReleaseManager.CallOpts, arg0, arg1, arg2)
}

// ArtifactExists is a free data retrieval call binding the contract method 0xb2daca5c.
//
// Solidity: function artifactExists(bytes32 ) view returns(bool)
func (_ReleaseManager *ReleaseManagerCaller) ArtifactExists(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "artifactExists", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ArtifactExists is a free data retrieval call binding the contract method 0xb2daca5c.
//
// Solidity: function artifactExists(bytes32 ) view returns(bool)
func (_ReleaseManager *ReleaseManagerSession) ArtifactExists(arg0 [32]byte) (bool, error) {
	return _ReleaseManager.Contract.ArtifactExists(&_ReleaseManager.CallOpts, arg0)
}

// ArtifactExists is a free data retrieval call binding the contract method 0xb2daca5c.
//
// Solidity: function artifactExists(bytes32 ) view returns(bool)
func (_ReleaseManager *ReleaseManagerCallerSession) ArtifactExists(arg0 [32]byte) (bool, error) {
	return _ReleaseManager.Contract.ArtifactExists(&_ReleaseManager.CallOpts, arg0)
}

// Artifacts is a free data retrieval call binding the contract method 0xa63e3a37.
//
// Solidity: function artifacts(bytes32 ) view returns(uint8 artifactType, uint8 architecture, uint8 os, bytes32 digest, string registryUrl, uint256 publishedAt)
func (_ReleaseManager *ReleaseManagerCaller) Artifacts(opts *bind.CallOpts, arg0 [32]byte) (struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "artifacts", arg0)

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
func (_ReleaseManager *ReleaseManagerSession) Artifacts(arg0 [32]byte) (struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}, error) {
	return _ReleaseManager.Contract.Artifacts(&_ReleaseManager.CallOpts, arg0)
}

// Artifacts is a free data retrieval call binding the contract method 0xa63e3a37.
//
// Solidity: function artifacts(bytes32 ) view returns(uint8 artifactType, uint8 architecture, uint8 os, bytes32 digest, string registryUrl, uint256 publishedAt)
func (_ReleaseManager *ReleaseManagerCallerSession) Artifacts(arg0 [32]byte) (struct {
	ArtifactType uint8
	Architecture uint8
	Os           uint8
	Digest       [32]byte
	RegistryUrl  string
	PublishedAt  *big.Int
}, error) {
	return _ReleaseManager.Contract.Artifacts(&_ReleaseManager.CallOpts, arg0)
}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_ReleaseManager *ReleaseManagerCaller) GetArtifact(opts *bind.CallOpts, avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getArtifact", avs, digest)

	if err != nil {
		return *new(IReleaseManagerArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerArtifact)).(*IReleaseManagerArtifact)

	return out0, err

}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_ReleaseManager *ReleaseManagerSession) GetArtifact(avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	return _ReleaseManager.Contract.GetArtifact(&_ReleaseManager.CallOpts, avs, digest)
}

// GetArtifact is a free data retrieval call binding the contract method 0xd2d104ef.
//
// Solidity: function getArtifact(address avs, bytes32 digest) view returns((uint8,uint8,uint8,bytes32,string,uint256))
func (_ReleaseManager *ReleaseManagerCallerSession) GetArtifact(avs common.Address, digest [32]byte) (IReleaseManagerArtifact, error) {
	return _ReleaseManager.Contract.GetArtifact(&_ReleaseManager.CallOpts, avs, digest)
}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManager *ReleaseManagerCaller) GetLatestPromotedArtifact(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getLatestPromotedArtifact", avs, operatorSetId)

	if err != nil {
		return *new(IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerPromotedArtifact)).(*IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManager *ReleaseManagerSession) GetLatestPromotedArtifact(avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetLatestPromotedArtifact(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetLatestPromotedArtifact is a free data retrieval call binding the contract method 0x97f1c2c9.
//
// Solidity: function getLatestPromotedArtifact(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManager *ReleaseManagerCallerSession) GetLatestPromotedArtifact(avs common.Address, operatorSetId [32]byte) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetLatestPromotedArtifact(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManager *ReleaseManagerCaller) GetPromotedArtifactAtBlock(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getPromotedArtifactAtBlock", avs, operatorSetId, blockNumber)

	if err != nil {
		return *new(IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new(IReleaseManagerPromotedArtifact)).(*IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManager *ReleaseManagerSession) GetPromotedArtifactAtBlock(avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetPromotedArtifactAtBlock(&_ReleaseManager.CallOpts, avs, operatorSetId, blockNumber)
}

// GetPromotedArtifactAtBlock is a free data retrieval call binding the contract method 0xa6210e18.
//
// Solidity: function getPromotedArtifactAtBlock(address avs, bytes32 operatorSetId, uint256 blockNumber) view returns((bytes32,string,uint8,string,uint256,uint256))
func (_ReleaseManager *ReleaseManagerCallerSession) GetPromotedArtifactAtBlock(avs common.Address, operatorSetId [32]byte, blockNumber *big.Int) (IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetPromotedArtifactAtBlock(&_ReleaseManager.CallOpts, avs, operatorSetId, blockNumber)
}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManager *ReleaseManagerCaller) GetPromotedArtifacts(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getPromotedArtifacts", avs, operatorSetId)

	if err != nil {
		return *new([]IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]IReleaseManagerPromotedArtifact)).(*[]IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManager *ReleaseManagerSession) GetPromotedArtifacts(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetPromotedArtifacts(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotedArtifacts is a free data retrieval call binding the contract method 0xb501f96f.
//
// Solidity: function getPromotedArtifacts(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManager *ReleaseManagerCallerSession) GetPromotedArtifacts(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetPromotedArtifacts(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_ReleaseManager *ReleaseManagerCaller) GetPromotionCheckpointAt(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getPromotionCheckpointAt", avs, operatorSetId, pos)

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
func (_ReleaseManager *ReleaseManagerSession) GetPromotionCheckpointAt(avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	return _ReleaseManager.Contract.GetPromotionCheckpointAt(&_ReleaseManager.CallOpts, avs, operatorSetId, pos)
}

// GetPromotionCheckpointAt is a free data retrieval call binding the contract method 0x3642b6ff.
//
// Solidity: function getPromotionCheckpointAt(address avs, bytes32 operatorSetId, uint256 pos) view returns(uint256 blockNumber, uint256 artifactIndex)
func (_ReleaseManager *ReleaseManagerCallerSession) GetPromotionCheckpointAt(avs common.Address, operatorSetId [32]byte, pos *big.Int) (struct {
	BlockNumber   *big.Int
	ArtifactIndex *big.Int
}, error) {
	return _ReleaseManager.Contract.GetPromotionCheckpointAt(&_ReleaseManager.CallOpts, avs, operatorSetId, pos)
}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_ReleaseManager *ReleaseManagerCaller) GetPromotionCheckpointCount(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getPromotionCheckpointCount", avs, operatorSetId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_ReleaseManager *ReleaseManagerSession) GetPromotionCheckpointCount(avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	return _ReleaseManager.Contract.GetPromotionCheckpointCount(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionCheckpointCount is a free data retrieval call binding the contract method 0x3446128f.
//
// Solidity: function getPromotionCheckpointCount(address avs, bytes32 operatorSetId) view returns(uint256)
func (_ReleaseManager *ReleaseManagerCallerSession) GetPromotionCheckpointCount(avs common.Address, operatorSetId [32]byte) (*big.Int, error) {
	return _ReleaseManager.Contract.GetPromotionCheckpointCount(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManager *ReleaseManagerCaller) GetPromotionHistory(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getPromotionHistory", avs, operatorSetId)

	if err != nil {
		return *new([]IReleaseManagerPromotedArtifact), err
	}

	out0 := *abi.ConvertType(out[0], new([]IReleaseManagerPromotedArtifact)).(*[]IReleaseManagerPromotedArtifact)

	return out0, err

}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManager *ReleaseManagerSession) GetPromotionHistory(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetPromotionHistory(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionHistory is a free data retrieval call binding the contract method 0x20eb2c99.
//
// Solidity: function getPromotionHistory(address avs, bytes32 operatorSetId) view returns((bytes32,string,uint8,string,uint256,uint256)[])
func (_ReleaseManager *ReleaseManagerCallerSession) GetPromotionHistory(avs common.Address, operatorSetId [32]byte) ([]IReleaseManagerPromotedArtifact, error) {
	return _ReleaseManager.Contract.GetPromotionHistory(&_ReleaseManager.CallOpts, avs, operatorSetId)
}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_ReleaseManager *ReleaseManagerCaller) GetPromotionStatusAtBlock(opts *bind.CallOpts, avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "getPromotionStatusAtBlock", avs, operatorSetId, digest, blockNumber)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_ReleaseManager *ReleaseManagerSession) GetPromotionStatusAtBlock(avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	return _ReleaseManager.Contract.GetPromotionStatusAtBlock(&_ReleaseManager.CallOpts, avs, operatorSetId, digest, blockNumber)
}

// GetPromotionStatusAtBlock is a free data retrieval call binding the contract method 0x447039bd.
//
// Solidity: function getPromotionStatusAtBlock(address avs, bytes32 operatorSetId, bytes32 digest, uint256 blockNumber) view returns(uint8)
func (_ReleaseManager *ReleaseManagerCallerSession) GetPromotionStatusAtBlock(avs common.Address, operatorSetId [32]byte, digest [32]byte, blockNumber *big.Int) (uint8, error) {
	return _ReleaseManager.Contract.GetPromotionStatusAtBlock(&_ReleaseManager.CallOpts, avs, operatorSetId, digest, blockNumber)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ReleaseManager *ReleaseManagerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ReleaseManager *ReleaseManagerSession) Owner() (common.Address, error) {
	return _ReleaseManager.Contract.Owner(&_ReleaseManager.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ReleaseManager *ReleaseManagerCallerSession) Owner() (common.Address, error) {
	return _ReleaseManager.Contract.Owner(&_ReleaseManager.CallOpts)
}

// PermissionController is a free data retrieval call binding the contract method 0x4657e26a.
//
// Solidity: function permissionController() view returns(address)
func (_ReleaseManager *ReleaseManagerCaller) PermissionController(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "permissionController")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PermissionController is a free data retrieval call binding the contract method 0x4657e26a.
//
// Solidity: function permissionController() view returns(address)
func (_ReleaseManager *ReleaseManagerSession) PermissionController() (common.Address, error) {
	return _ReleaseManager.Contract.PermissionController(&_ReleaseManager.CallOpts)
}

// PermissionController is a free data retrieval call binding the contract method 0x4657e26a.
//
// Solidity: function permissionController() view returns(address)
func (_ReleaseManager *ReleaseManagerCallerSession) PermissionController() (common.Address, error) {
	return _ReleaseManager.Contract.PermissionController(&_ReleaseManager.CallOpts)
}

// RegisteredAVS is a free data retrieval call binding the contract method 0xbf2d8e07.
//
// Solidity: function registeredAVS(address ) view returns(bool)
func (_ReleaseManager *ReleaseManagerCaller) RegisteredAVS(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ReleaseManager.contract.Call(opts, &out, "registeredAVS", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RegisteredAVS is a free data retrieval call binding the contract method 0xbf2d8e07.
//
// Solidity: function registeredAVS(address ) view returns(bool)
func (_ReleaseManager *ReleaseManagerSession) RegisteredAVS(arg0 common.Address) (bool, error) {
	return _ReleaseManager.Contract.RegisteredAVS(&_ReleaseManager.CallOpts, arg0)
}

// RegisteredAVS is a free data retrieval call binding the contract method 0xbf2d8e07.
//
// Solidity: function registeredAVS(address ) view returns(bool)
func (_ReleaseManager *ReleaseManagerCallerSession) RegisteredAVS(arg0 common.Address) (bool, error) {
	return _ReleaseManager.Contract.RegisteredAVS(&_ReleaseManager.CallOpts, arg0)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_ReleaseManager *ReleaseManagerTransactor) Deregister(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "deregister", avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_ReleaseManager *ReleaseManagerSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.Deregister(&_ReleaseManager.TransactOpts, avs)
}

// Deregister is a paid mutator transaction binding the contract method 0x84ac33ec.
//
// Solidity: function deregister(address avs) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) Deregister(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.Deregister(&_ReleaseManager.TransactOpts, avs)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _permissionController) returns()
func (_ReleaseManager *ReleaseManagerTransactor) Initialize(opts *bind.TransactOpts, _permissionController common.Address) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "initialize", _permissionController)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _permissionController) returns()
func (_ReleaseManager *ReleaseManagerSession) Initialize(_permissionController common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.Initialize(&_ReleaseManager.TransactOpts, _permissionController)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _permissionController) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) Initialize(_permissionController common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.Initialize(&_ReleaseManager.TransactOpts, _permissionController)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_ReleaseManager *ReleaseManagerTransactor) PromoteArtifacts(opts *bind.TransactOpts, avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "promoteArtifacts", avs, promotions, version, deploymentDeadline)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_ReleaseManager *ReleaseManagerSession) PromoteArtifacts(avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _ReleaseManager.Contract.PromoteArtifacts(&_ReleaseManager.TransactOpts, avs, promotions, version, deploymentDeadline)
}

// PromoteArtifacts is a paid mutator transaction binding the contract method 0x716d0d9b.
//
// Solidity: function promoteArtifacts(address avs, (uint8,bytes32,string,bytes32[])[] promotions, string version, uint256 deploymentDeadline) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) PromoteArtifacts(avs common.Address, promotions []IReleaseManagerArtifactPromotion, version string, deploymentDeadline *big.Int) (*types.Transaction, error) {
	return _ReleaseManager.Contract.PromoteArtifacts(&_ReleaseManager.TransactOpts, avs, promotions, version, deploymentDeadline)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] _artifacts) returns()
func (_ReleaseManager *ReleaseManagerTransactor) PublishArtifacts(opts *bind.TransactOpts, avs common.Address, _artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "publishArtifacts", avs, _artifacts)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] _artifacts) returns()
func (_ReleaseManager *ReleaseManagerSession) PublishArtifacts(avs common.Address, _artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _ReleaseManager.Contract.PublishArtifacts(&_ReleaseManager.TransactOpts, avs, _artifacts)
}

// PublishArtifacts is a paid mutator transaction binding the contract method 0xeba04f8a.
//
// Solidity: function publishArtifacts(address avs, (uint8,uint8,uint8,bytes32,string,uint256)[] _artifacts) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) PublishArtifacts(avs common.Address, _artifacts []IReleaseManagerArtifact) (*types.Transaction, error) {
	return _ReleaseManager.Contract.PublishArtifacts(&_ReleaseManager.TransactOpts, avs, _artifacts)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_ReleaseManager *ReleaseManagerTransactor) Register(opts *bind.TransactOpts, avs common.Address) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "register", avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_ReleaseManager *ReleaseManagerSession) Register(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.Register(&_ReleaseManager.TransactOpts, avs)
}

// Register is a paid mutator transaction binding the contract method 0x4420e486.
//
// Solidity: function register(address avs) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) Register(avs common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.Register(&_ReleaseManager.TransactOpts, avs)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ReleaseManager *ReleaseManagerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ReleaseManager *ReleaseManagerSession) RenounceOwnership() (*types.Transaction, error) {
	return _ReleaseManager.Contract.RenounceOwnership(&_ReleaseManager.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ReleaseManager.Contract.RenounceOwnership(&_ReleaseManager.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReleaseManager *ReleaseManagerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReleaseManager *ReleaseManagerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.TransferOwnership(&_ReleaseManager.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ReleaseManager.Contract.TransferOwnership(&_ReleaseManager.TransactOpts, newOwner)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_ReleaseManager *ReleaseManagerTransactor) UpdatePromotionStatus(opts *bind.TransactOpts, avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _ReleaseManager.contract.Transact(opts, "updatePromotionStatus", avs, digest, operatorSetId, newStatus)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_ReleaseManager *ReleaseManagerSession) UpdatePromotionStatus(avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _ReleaseManager.Contract.UpdatePromotionStatus(&_ReleaseManager.TransactOpts, avs, digest, operatorSetId, newStatus)
}

// UpdatePromotionStatus is a paid mutator transaction binding the contract method 0x7d708ea6.
//
// Solidity: function updatePromotionStatus(address avs, bytes32 digest, bytes32 operatorSetId, uint8 newStatus) returns()
func (_ReleaseManager *ReleaseManagerTransactorSession) UpdatePromotionStatus(avs common.Address, digest [32]byte, operatorSetId [32]byte, newStatus uint8) (*types.Transaction, error) {
	return _ReleaseManager.Contract.UpdatePromotionStatus(&_ReleaseManager.TransactOpts, avs, digest, operatorSetId, newStatus)
}

// ReleaseManagerAVSDeregisteredIterator is returned from FilterAVSDeregistered and is used to iterate over the raw logs and unpacked data for AVSDeregistered events raised by the ReleaseManager contract.
type ReleaseManagerAVSDeregisteredIterator struct {
	Event *ReleaseManagerAVSDeregistered // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerAVSDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerAVSDeregistered)
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
		it.Event = new(ReleaseManagerAVSDeregistered)
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
func (it *ReleaseManagerAVSDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerAVSDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerAVSDeregistered represents a AVSDeregistered event raised by the ReleaseManager contract.
type ReleaseManagerAVSDeregistered struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAVSDeregistered is a free log retrieval operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_ReleaseManager *ReleaseManagerFilterer) FilterAVSDeregistered(opts *bind.FilterOpts, avs []common.Address) (*ReleaseManagerAVSDeregisteredIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "AVSDeregistered", avsRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerAVSDeregisteredIterator{contract: _ReleaseManager.contract, event: "AVSDeregistered", logs: logs, sub: sub}, nil
}

// WatchAVSDeregistered is a free log subscription operation binding the contract event 0xf7cd17cf5978e63a941e1b110c3afd213843bff041513266571af35a6cec8ab7.
//
// Solidity: event AVSDeregistered(address indexed avs)
func (_ReleaseManager *ReleaseManagerFilterer) WatchAVSDeregistered(opts *bind.WatchOpts, sink chan<- *ReleaseManagerAVSDeregistered, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "AVSDeregistered", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerAVSDeregistered)
				if err := _ReleaseManager.contract.UnpackLog(event, "AVSDeregistered", log); err != nil {
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
func (_ReleaseManager *ReleaseManagerFilterer) ParseAVSDeregistered(log types.Log) (*ReleaseManagerAVSDeregistered, error) {
	event := new(ReleaseManagerAVSDeregistered)
	if err := _ReleaseManager.contract.UnpackLog(event, "AVSDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerAVSRegisteredIterator is returned from FilterAVSRegistered and is used to iterate over the raw logs and unpacked data for AVSRegistered events raised by the ReleaseManager contract.
type ReleaseManagerAVSRegisteredIterator struct {
	Event *ReleaseManagerAVSRegistered // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerAVSRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerAVSRegistered)
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
		it.Event = new(ReleaseManagerAVSRegistered)
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
func (it *ReleaseManagerAVSRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerAVSRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerAVSRegistered represents a AVSRegistered event raised by the ReleaseManager contract.
type ReleaseManagerAVSRegistered struct {
	Avs common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAVSRegistered is a free log retrieval operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_ReleaseManager *ReleaseManagerFilterer) FilterAVSRegistered(opts *bind.FilterOpts, avs []common.Address) (*ReleaseManagerAVSRegisteredIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "AVSRegistered", avsRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerAVSRegisteredIterator{contract: _ReleaseManager.contract, event: "AVSRegistered", logs: logs, sub: sub}, nil
}

// WatchAVSRegistered is a free log subscription operation binding the contract event 0x2c7ccee1b83a57ffa52bfd71692c05a6b8b9dc9b1e73a6d25c78bab22a98b06e.
//
// Solidity: event AVSRegistered(address indexed avs)
func (_ReleaseManager *ReleaseManagerFilterer) WatchAVSRegistered(opts *bind.WatchOpts, sink chan<- *ReleaseManagerAVSRegistered, avs []common.Address) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "AVSRegistered", avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerAVSRegistered)
				if err := _ReleaseManager.contract.UnpackLog(event, "AVSRegistered", log); err != nil {
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
func (_ReleaseManager *ReleaseManagerFilterer) ParseAVSRegistered(log types.Log) (*ReleaseManagerAVSRegistered, error) {
	event := new(ReleaseManagerAVSRegistered)
	if err := _ReleaseManager.contract.UnpackLog(event, "AVSRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerArtifactPublishedIterator is returned from FilterArtifactPublished and is used to iterate over the raw logs and unpacked data for ArtifactPublished events raised by the ReleaseManager contract.
type ReleaseManagerArtifactPublishedIterator struct {
	Event *ReleaseManagerArtifactPublished // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerArtifactPublishedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerArtifactPublished)
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
		it.Event = new(ReleaseManagerArtifactPublished)
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
func (it *ReleaseManagerArtifactPublishedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerArtifactPublishedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerArtifactPublished represents a ArtifactPublished event raised by the ReleaseManager contract.
type ReleaseManagerArtifactPublished struct {
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
func (_ReleaseManager *ReleaseManagerFilterer) FilterArtifactPublished(opts *bind.FilterOpts, avs []common.Address, digest [][32]byte) (*ReleaseManagerArtifactPublishedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "ArtifactPublished", avsRule, digestRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerArtifactPublishedIterator{contract: _ReleaseManager.contract, event: "ArtifactPublished", logs: logs, sub: sub}, nil
}

// WatchArtifactPublished is a free log subscription operation binding the contract event 0x622e1c25f0e4dcedadb24c8f0bbe6ceaa3776cbeb58b17a2d5d8ac8ae31a7502.
//
// Solidity: event ArtifactPublished(address indexed avs, bytes32 indexed digest, string registryUrl, uint8 architecture, uint8 os, uint8 artifactType)
func (_ReleaseManager *ReleaseManagerFilterer) WatchArtifactPublished(opts *bind.WatchOpts, sink chan<- *ReleaseManagerArtifactPublished, avs []common.Address, digest [][32]byte) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var digestRule []interface{}
	for _, digestItem := range digest {
		digestRule = append(digestRule, digestItem)
	}

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "ArtifactPublished", avsRule, digestRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerArtifactPublished)
				if err := _ReleaseManager.contract.UnpackLog(event, "ArtifactPublished", log); err != nil {
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
func (_ReleaseManager *ReleaseManagerFilterer) ParseArtifactPublished(log types.Log) (*ReleaseManagerArtifactPublished, error) {
	event := new(ReleaseManagerArtifactPublished)
	if err := _ReleaseManager.contract.UnpackLog(event, "ArtifactPublished", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerArtifactsPromotedIterator is returned from FilterArtifactsPromoted and is used to iterate over the raw logs and unpacked data for ArtifactsPromoted events raised by the ReleaseManager contract.
type ReleaseManagerArtifactsPromotedIterator struct {
	Event *ReleaseManagerArtifactsPromoted // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerArtifactsPromotedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerArtifactsPromoted)
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
		it.Event = new(ReleaseManagerArtifactsPromoted)
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
func (it *ReleaseManagerArtifactsPromotedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerArtifactsPromotedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerArtifactsPromoted represents a ArtifactsPromoted event raised by the ReleaseManager contract.
type ReleaseManagerArtifactsPromoted struct {
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
func (_ReleaseManager *ReleaseManagerFilterer) FilterArtifactsPromoted(opts *bind.FilterOpts, avs []common.Address, version []string) (*ReleaseManagerArtifactsPromotedIterator, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "ArtifactsPromoted", avsRule, versionRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerArtifactsPromotedIterator{contract: _ReleaseManager.contract, event: "ArtifactsPromoted", logs: logs, sub: sub}, nil
}

// WatchArtifactsPromoted is a free log subscription operation binding the contract event 0x7fd55cb6307da041fb4711b42eb59c940bb12e76dc208c9be19be9abde37815d.
//
// Solidity: event ArtifactsPromoted(address indexed avs, string indexed version, uint256 deploymentDeadline, bytes32[] digests, uint8[] statuses)
func (_ReleaseManager *ReleaseManagerFilterer) WatchArtifactsPromoted(opts *bind.WatchOpts, sink chan<- *ReleaseManagerArtifactsPromoted, avs []common.Address, version []string) (event.Subscription, error) {

	var avsRule []interface{}
	for _, avsItem := range avs {
		avsRule = append(avsRule, avsItem)
	}
	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "ArtifactsPromoted", avsRule, versionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerArtifactsPromoted)
				if err := _ReleaseManager.contract.UnpackLog(event, "ArtifactsPromoted", log); err != nil {
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
func (_ReleaseManager *ReleaseManagerFilterer) ParseArtifactsPromoted(log types.Log) (*ReleaseManagerArtifactsPromoted, error) {
	event := new(ReleaseManagerArtifactsPromoted)
	if err := _ReleaseManager.contract.UnpackLog(event, "ArtifactsPromoted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ReleaseManager contract.
type ReleaseManagerInitializedIterator struct {
	Event *ReleaseManagerInitialized // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerInitialized)
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
		it.Event = new(ReleaseManagerInitialized)
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
func (it *ReleaseManagerInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerInitialized represents a Initialized event raised by the ReleaseManager contract.
type ReleaseManagerInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ReleaseManager *ReleaseManagerFilterer) FilterInitialized(opts *bind.FilterOpts) (*ReleaseManagerInitializedIterator, error) {

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerInitializedIterator{contract: _ReleaseManager.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ReleaseManager *ReleaseManagerFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ReleaseManagerInitialized) (event.Subscription, error) {

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerInitialized)
				if err := _ReleaseManager.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ReleaseManager *ReleaseManagerFilterer) ParseInitialized(log types.Log) (*ReleaseManagerInitialized, error) {
	event := new(ReleaseManagerInitialized)
	if err := _ReleaseManager.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ReleaseManager contract.
type ReleaseManagerOwnershipTransferredIterator struct {
	Event *ReleaseManagerOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerOwnershipTransferred)
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
		it.Event = new(ReleaseManagerOwnershipTransferred)
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
func (it *ReleaseManagerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerOwnershipTransferred represents a OwnershipTransferred event raised by the ReleaseManager contract.
type ReleaseManagerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ReleaseManager *ReleaseManagerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ReleaseManagerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerOwnershipTransferredIterator{contract: _ReleaseManager.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ReleaseManager *ReleaseManagerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ReleaseManagerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerOwnershipTransferred)
				if err := _ReleaseManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ReleaseManager *ReleaseManagerFilterer) ParseOwnershipTransferred(log types.Log) (*ReleaseManagerOwnershipTransferred, error) {
	event := new(ReleaseManagerOwnershipTransferred)
	if err := _ReleaseManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ReleaseManagerPromotionStatusUpdatedIterator is returned from FilterPromotionStatusUpdated and is used to iterate over the raw logs and unpacked data for PromotionStatusUpdated events raised by the ReleaseManager contract.
type ReleaseManagerPromotionStatusUpdatedIterator struct {
	Event *ReleaseManagerPromotionStatusUpdated // Event containing the contract specifics and raw log

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
func (it *ReleaseManagerPromotionStatusUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReleaseManagerPromotionStatusUpdated)
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
		it.Event = new(ReleaseManagerPromotionStatusUpdated)
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
func (it *ReleaseManagerPromotionStatusUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReleaseManagerPromotionStatusUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReleaseManagerPromotionStatusUpdated represents a PromotionStatusUpdated event raised by the ReleaseManager contract.
type ReleaseManagerPromotionStatusUpdated struct {
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
func (_ReleaseManager *ReleaseManagerFilterer) FilterPromotionStatusUpdated(opts *bind.FilterOpts, avs []common.Address, digest [][32]byte, operatorSetId [][32]byte) (*ReleaseManagerPromotionStatusUpdatedIterator, error) {

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

	logs, sub, err := _ReleaseManager.contract.FilterLogs(opts, "PromotionStatusUpdated", avsRule, digestRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return &ReleaseManagerPromotionStatusUpdatedIterator{contract: _ReleaseManager.contract, event: "PromotionStatusUpdated", logs: logs, sub: sub}, nil
}

// WatchPromotionStatusUpdated is a free log subscription operation binding the contract event 0x1341ac2513ee1dd81316f76f3a7840608cd4dc5a3302d2fb45ea1ad24d355671.
//
// Solidity: event PromotionStatusUpdated(address indexed avs, bytes32 indexed digest, bytes32 indexed operatorSetId, uint8 oldStatus, uint8 newStatus)
func (_ReleaseManager *ReleaseManagerFilterer) WatchPromotionStatusUpdated(opts *bind.WatchOpts, sink chan<- *ReleaseManagerPromotionStatusUpdated, avs []common.Address, digest [][32]byte, operatorSetId [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _ReleaseManager.contract.WatchLogs(opts, "PromotionStatusUpdated", avsRule, digestRule, operatorSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReleaseManagerPromotionStatusUpdated)
				if err := _ReleaseManager.contract.UnpackLog(event, "PromotionStatusUpdated", log); err != nil {
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
func (_ReleaseManager *ReleaseManagerFilterer) ParsePromotionStatusUpdated(log types.Log) (*ReleaseManagerPromotionStatusUpdated, error) {
	event := new(ReleaseManagerPromotionStatusUpdated)
	if err := _ReleaseManager.contract.UnpackLog(event, "PromotionStatusUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
