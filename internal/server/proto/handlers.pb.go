// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: internal/server/proto/handlers.proto

package proto

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

type Metric struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Value         float64                `protobuf:"fixed64,3,opt,name=value,proto3" json:"value,omitempty"`
	Delta         int64                  `protobuf:"varint,4,opt,name=delta,proto3" json:"delta,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Metric) Reset() {
	*x = Metric{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{0}
}

func (x *Metric) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Metric) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Metric) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *Metric) GetDelta() int64 {
	if x != nil {
		return x.Delta
	}
	return 0
}

type PostUpdateRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metric        *Metric                `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostUpdateRequest) Reset() {
	*x = PostUpdateRequest{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostUpdateRequest) ProtoMessage() {}

func (x *PostUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostUpdateRequest.ProtoReflect.Descriptor instead.
func (*PostUpdateRequest) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{1}
}

func (x *PostUpdateRequest) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

type PostUpdateResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         string                 `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"` // ошибка
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostUpdateResponse) Reset() {
	*x = PostUpdateResponse{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostUpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostUpdateResponse) ProtoMessage() {}

func (x *PostUpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostUpdateResponse.ProtoReflect.Descriptor instead.
func (*PostUpdateResponse) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{2}
}

func (x *PostUpdateResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type PostUpdatesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metric        []*Metric              `protobuf:"bytes,1,rep,name=metric,proto3" json:"metric,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostUpdatesRequest) Reset() {
	*x = PostUpdatesRequest{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostUpdatesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostUpdatesRequest) ProtoMessage() {}

func (x *PostUpdatesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostUpdatesRequest.ProtoReflect.Descriptor instead.
func (*PostUpdatesRequest) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{3}
}

func (x *PostUpdatesRequest) GetMetric() []*Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

type PostUpdatesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         string                 `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"` // ошибка
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PostUpdatesResponse) Reset() {
	*x = PostUpdatesResponse{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostUpdatesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostUpdatesResponse) ProtoMessage() {}

func (x *PostUpdatesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostUpdatesResponse.ProtoReflect.Descriptor instead.
func (*PostUpdatesResponse) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{4}
}

func (x *PostUpdatesResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type GetValueRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetValueRequest) Reset() {
	*x = GetValueRequest{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetValueRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetValueRequest) ProtoMessage() {}

func (x *GetValueRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetValueRequest.ProtoReflect.Descriptor instead.
func (*GetValueRequest) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{5}
}

func (x *GetValueRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetValueResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metric        *Metric                `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetValueResponse) Reset() {
	*x = GetValueResponse{}
	mi := &file_internal_server_proto_handlers_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetValueResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetValueResponse) ProtoMessage() {}

func (x *GetValueResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_proto_handlers_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetValueResponse.ProtoReflect.Descriptor instead.
func (*GetValueResponse) Descriptor() ([]byte, []int) {
	return file_internal_server_proto_handlers_proto_rawDescGZIP(), []int{6}
}

func (x *GetValueResponse) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

var File_internal_server_proto_handlers_proto protoreflect.FileDescriptor

const file_internal_server_proto_handlers_proto_rawDesc = "" +
	"\n" +
	"$internal/server/proto/handlers.proto\x12\vserver_grpc\"X\n" +
	"\x06Metric\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\x12\x14\n" +
	"\x05value\x18\x03 \x01(\x01R\x05value\x12\x14\n" +
	"\x05delta\x18\x04 \x01(\x03R\x05delta\"@\n" +
	"\x11PostUpdateRequest\x12+\n" +
	"\x06metric\x18\x01 \x01(\v2\x13.server_grpc.MetricR\x06metric\"*\n" +
	"\x12PostUpdateResponse\x12\x14\n" +
	"\x05error\x18\x01 \x01(\tR\x05error\"A\n" +
	"\x12PostUpdatesRequest\x12+\n" +
	"\x06metric\x18\x01 \x03(\v2\x13.server_grpc.MetricR\x06metric\"+\n" +
	"\x13PostUpdatesResponse\x12\x14\n" +
	"\x05error\x18\x01 \x01(\tR\x05error\"!\n" +
	"\x0fGetValueRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\"?\n" +
	"\x10GetValueResponse\x12+\n" +
	"\x06metric\x18\x01 \x01(\v2\x13.server_grpc.MetricR\x06metric2\xf4\x01\n" +
	"\bHandlers\x12M\n" +
	"\n" +
	"PostUpdate\x12\x1e.server_grpc.PostUpdateRequest\x1a\x1f.server_grpc.PostUpdateResponse\x12P\n" +
	"\vPostUpdates\x12\x1f.server_grpc.PostUpdatesRequest\x1a .server_grpc.PostUpdatesResponse\x12G\n" +
	"\bGetValue\x12\x1c.server_grpc.GetValueRequest\x1a\x1d.server_grpc.GetValueResponseB\x17Z\x15internal/server/protob\x06proto3"

var (
	file_internal_server_proto_handlers_proto_rawDescOnce sync.Once
	file_internal_server_proto_handlers_proto_rawDescData []byte
)

func file_internal_server_proto_handlers_proto_rawDescGZIP() []byte {
	file_internal_server_proto_handlers_proto_rawDescOnce.Do(func() {
		file_internal_server_proto_handlers_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_internal_server_proto_handlers_proto_rawDesc), len(file_internal_server_proto_handlers_proto_rawDesc)))
	})
	return file_internal_server_proto_handlers_proto_rawDescData
}

var file_internal_server_proto_handlers_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_internal_server_proto_handlers_proto_goTypes = []any{
	(*Metric)(nil),              // 0: server_grpc.Metric
	(*PostUpdateRequest)(nil),   // 1: server_grpc.PostUpdateRequest
	(*PostUpdateResponse)(nil),  // 2: server_grpc.PostUpdateResponse
	(*PostUpdatesRequest)(nil),  // 3: server_grpc.PostUpdatesRequest
	(*PostUpdatesResponse)(nil), // 4: server_grpc.PostUpdatesResponse
	(*GetValueRequest)(nil),     // 5: server_grpc.GetValueRequest
	(*GetValueResponse)(nil),    // 6: server_grpc.GetValueResponse
}
var file_internal_server_proto_handlers_proto_depIdxs = []int32{
	0, // 0: server_grpc.PostUpdateRequest.metric:type_name -> server_grpc.Metric
	0, // 1: server_grpc.PostUpdatesRequest.metric:type_name -> server_grpc.Metric
	0, // 2: server_grpc.GetValueResponse.metric:type_name -> server_grpc.Metric
	1, // 3: server_grpc.Handlers.PostUpdate:input_type -> server_grpc.PostUpdateRequest
	3, // 4: server_grpc.Handlers.PostUpdates:input_type -> server_grpc.PostUpdatesRequest
	5, // 5: server_grpc.Handlers.GetValue:input_type -> server_grpc.GetValueRequest
	2, // 6: server_grpc.Handlers.PostUpdate:output_type -> server_grpc.PostUpdateResponse
	4, // 7: server_grpc.Handlers.PostUpdates:output_type -> server_grpc.PostUpdatesResponse
	6, // 8: server_grpc.Handlers.GetValue:output_type -> server_grpc.GetValueResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_internal_server_proto_handlers_proto_init() }
func file_internal_server_proto_handlers_proto_init() {
	if File_internal_server_proto_handlers_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_internal_server_proto_handlers_proto_rawDesc), len(file_internal_server_proto_handlers_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_server_proto_handlers_proto_goTypes,
		DependencyIndexes: file_internal_server_proto_handlers_proto_depIdxs,
		MessageInfos:      file_internal_server_proto_handlers_proto_msgTypes,
	}.Build()
	File_internal_server_proto_handlers_proto = out.File
	file_internal_server_proto_handlers_proto_goTypes = nil
	file_internal_server_proto_handlers_proto_depIdxs = nil
}
