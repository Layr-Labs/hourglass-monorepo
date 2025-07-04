// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: eigenlayer/hourglass/v1/wire/wire.proto

package wire

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AuthenticateSocket struct {
	state                        protoimpl.MessageState `protogen:"open.v1"`
	AggregatorAddress            string                 `protobuf:"bytes,1,opt,name=aggregator_address,json=aggregatorAddress,proto3" json:"aggregator_address,omitempty"`                                      // address of the aggregator that wants to connect
	OperatorSignedNonce          string                 `protobuf:"bytes,2,opt,name=operator_signed_nonce,json=operatorSignedNonce,proto3" json:"operator_signed_nonce,omitempty"`                              // the signed nonce the operator sent back in the handshake
	OperatorSignedNonceSignature string                 `protobuf:"bytes,3,opt,name=operator_signed_nonce_signature,json=operatorSignedNonceSignature,proto3" json:"operator_signed_nonce_signature,omitempty"` // signature of the operator_signed_nonce signed with aggregator key to verify
	unknownFields                protoimpl.UnknownFields
	sizeCache                    protoimpl.SizeCache
}

func (x *AuthenticateSocket) Reset() {
	*x = AuthenticateSocket{}
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthenticateSocket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticateSocket) ProtoMessage() {}

func (x *AuthenticateSocket) ProtoReflect() protoreflect.Message {
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticateSocket.ProtoReflect.Descriptor instead.
func (*AuthenticateSocket) Descriptor() ([]byte, []int) {
	return file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescGZIP(), []int{0}
}

func (x *AuthenticateSocket) GetAggregatorAddress() string {
	if x != nil {
		return x.AggregatorAddress
	}
	return ""
}

func (x *AuthenticateSocket) GetOperatorSignedNonce() string {
	if x != nil {
		return x.OperatorSignedNonce
	}
	return ""
}

func (x *AuthenticateSocket) GetOperatorSignedNonceSignature() string {
	if x != nil {
		return x.OperatorSignedNonceSignature
	}
	return ""
}

type Task struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	TaskId          string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`                            // ID of the task from the origin inbox contract
	OperatorAddress string                 `protobuf:"bytes,2,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"` // ID of the operator that needs to process the message (mainly for debugging)
	ChainId         uint64                 `protobuf:"varint,3,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`                        // ID of the chain the message originated on
	Payload         []byte                 `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`                                        // generic bytes to pass off to the AVS software to execute
	Deadline        uint64                 `protobuf:"varint,5,opt,name=deadline,proto3" json:"deadline,omitempty"`                                     // unix timestamp of when the task needs to be processed by
	TaskSignature   string                 `protobuf:"bytes,6,opt,name=task_signature,json=taskSignature,proto3" json:"task_signature,omitempty"`       // signature of the payload, signed by aggregator
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *Task) Reset() {
	*x = Task{}
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescGZIP(), []int{1}
}

func (x *Task) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *Task) GetOperatorAddress() string {
	if x != nil {
		return x.OperatorAddress
	}
	return ""
}

func (x *Task) GetChainId() uint64 {
	if x != nil {
		return x.ChainId
	}
	return 0
}

func (x *Task) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Task) GetDeadline() uint64 {
	if x != nil {
		return x.Deadline
	}
	return 0
}

func (x *Task) GetTaskSignature() string {
	if x != nil {
		return x.TaskSignature
	}
	return ""
}

type TaskResult struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	TaskId            string                 `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`                                  // ID of the task processed
	OperatorAddress   string                 `protobuf:"bytes,2,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`       // address of the operator that created the result
	Response          []byte                 `protobuf:"bytes,3,opt,name=response,proto3" json:"response,omitempty"`                                            // the provided response
	ResponseSignature []byte                 `protobuf:"bytes,4,opt,name=response_signature,json=responseSignature,proto3" json:"response_signature,omitempty"` // signature of the response using the operator's key
	ChainId           uint64                 `protobuf:"varint,5,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`                              // ID of the chain the message originated on
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *TaskResult) Reset() {
	*x = TaskResult{}
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TaskResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResult) ProtoMessage() {}

func (x *TaskResult) ProtoReflect() protoreflect.Message {
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResult.ProtoReflect.Descriptor instead.
func (*TaskResult) Descriptor() ([]byte, []int) {
	return file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescGZIP(), []int{2}
}

func (x *TaskResult) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *TaskResult) GetOperatorAddress() string {
	if x != nil {
		return x.OperatorAddress
	}
	return ""
}

func (x *TaskResult) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *TaskResult) GetResponseSignature() []byte {
	if x != nil {
		return x.ResponseSignature
	}
	return nil
}

func (x *TaskResult) GetChainId() uint64 {
	if x != nil {
		return x.ChainId
	}
	return 0
}

type HeartbeatPing struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatPing) Reset() {
	*x = HeartbeatPing{}
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatPing) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatPing) ProtoMessage() {}

func (x *HeartbeatPing) ProtoReflect() protoreflect.Message {
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartbeatPing.ProtoReflect.Descriptor instead.
func (*HeartbeatPing) Descriptor() ([]byte, []int) {
	return file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescGZIP(), []int{3}
}

type HeartbeatPong struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CurrentTime   uint64                 `protobuf:"varint,1,opt,name=current_time,json=currentTime,proto3" json:"current_time,omitempty"` // unix timestamp of the current clock time of the worker
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatPong) Reset() {
	*x = HeartbeatPong{}
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatPong) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatPong) ProtoMessage() {}

func (x *HeartbeatPong) ProtoReflect() protoreflect.Message {
	mi := &file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartbeatPong.ProtoReflect.Descriptor instead.
func (*HeartbeatPong) Descriptor() ([]byte, []int) {
	return file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescGZIP(), []int{4}
}

func (x *HeartbeatPong) GetCurrentTime() uint64 {
	if x != nil {
		return x.CurrentTime
	}
	return 0
}

var File_eigenlayer_hourglass_v1_wire_wire_proto protoreflect.FileDescriptor

const file_eigenlayer_hourglass_v1_wire_wire_proto_rawDesc = "" +
	"\n" +
	"'eigenlayer/hourglass/v1/wire/wire.proto\x12\x1ceigenlayer.hourglass.v1.wire\"\xbe\x01\n" +
	"\x12AuthenticateSocket\x12-\n" +
	"\x12aggregator_address\x18\x01 \x01(\tR\x11aggregatorAddress\x122\n" +
	"\x15operator_signed_nonce\x18\x02 \x01(\tR\x13operatorSignedNonce\x12E\n" +
	"\x1foperator_signed_nonce_signature\x18\x03 \x01(\tR\x1coperatorSignedNonceSignature\"\xc2\x01\n" +
	"\x04Task\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\x12)\n" +
	"\x10operator_address\x18\x02 \x01(\tR\x0foperatorAddress\x12\x19\n" +
	"\bchain_id\x18\x03 \x01(\x04R\achainId\x12\x18\n" +
	"\apayload\x18\x04 \x01(\fR\apayload\x12\x1a\n" +
	"\bdeadline\x18\x05 \x01(\x04R\bdeadline\x12%\n" +
	"\x0etask_signature\x18\x06 \x01(\tR\rtaskSignature\"\xb6\x01\n" +
	"\n" +
	"TaskResult\x12\x17\n" +
	"\atask_id\x18\x01 \x01(\tR\x06taskId\x12)\n" +
	"\x10operator_address\x18\x02 \x01(\tR\x0foperatorAddress\x12\x1a\n" +
	"\bresponse\x18\x03 \x01(\fR\bresponse\x12-\n" +
	"\x12response_signature\x18\x04 \x01(\fR\x11responseSignature\x12\x19\n" +
	"\bchain_id\x18\x05 \x01(\x04R\achainId\"\x0f\n" +
	"\rHeartbeatPing\"2\n" +
	"\rHeartbeatPong\x12!\n" +
	"\fcurrent_time\x18\x01 \x01(\x04R\vcurrentTimeB\x98\x02\n" +
	" com.eigenlayer.hourglass.v1.wireB\tWireProtoP\x01ZUgithub.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/wire\xa2\x02\x04EHVW\xaa\x02\x1cEigenlayer.Hourglass.V1.Wire\xca\x02\x1cEigenlayer\\Hourglass\\V1\\Wire\xe2\x02(Eigenlayer\\Hourglass\\V1\\Wire\\GPBMetadata\xea\x02\x1fEigenlayer::Hourglass::V1::Wireb\x06proto3"

var (
	file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescOnce sync.Once
	file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescData []byte
)

func file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescGZIP() []byte {
	file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescOnce.Do(func() {
		file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_eigenlayer_hourglass_v1_wire_wire_proto_rawDesc), len(file_eigenlayer_hourglass_v1_wire_wire_proto_rawDesc)))
	})
	return file_eigenlayer_hourglass_v1_wire_wire_proto_rawDescData
}

var file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_eigenlayer_hourglass_v1_wire_wire_proto_goTypes = []any{
	(*AuthenticateSocket)(nil), // 0: eigenlayer.hourglass.v1.wire.AuthenticateSocket
	(*Task)(nil),               // 1: eigenlayer.hourglass.v1.wire.Task
	(*TaskResult)(nil),         // 2: eigenlayer.hourglass.v1.wire.TaskResult
	(*HeartbeatPing)(nil),      // 3: eigenlayer.hourglass.v1.wire.HeartbeatPing
	(*HeartbeatPong)(nil),      // 4: eigenlayer.hourglass.v1.wire.HeartbeatPong
}
var file_eigenlayer_hourglass_v1_wire_wire_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_eigenlayer_hourglass_v1_wire_wire_proto_init() }
func file_eigenlayer_hourglass_v1_wire_wire_proto_init() {
	if File_eigenlayer_hourglass_v1_wire_wire_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_eigenlayer_hourglass_v1_wire_wire_proto_rawDesc), len(file_eigenlayer_hourglass_v1_wire_wire_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_eigenlayer_hourglass_v1_wire_wire_proto_goTypes,
		DependencyIndexes: file_eigenlayer_hourglass_v1_wire_wire_proto_depIdxs,
		MessageInfos:      file_eigenlayer_hourglass_v1_wire_wire_proto_msgTypes,
	}.Build()
	File_eigenlayer_hourglass_v1_wire_wire_proto = out.File
	file_eigenlayer_hourglass_v1_wire_wire_proto_goTypes = nil
	file_eigenlayer_hourglass_v1_wire_wire_proto_depIdxs = nil
}
