// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: platformd/proxy/v1alpha1/service.proto

package v1alpha1

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateListenersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkloadID string `protobuf:"bytes,1,opt,name=workloadID,proto3" json:"workloadID,omitempty"`
	Ip         string `protobuf:"bytes,2,opt,name=ip,proto3" json:"ip,omitempty"`
}

func (x *CreateListenersRequest) Reset() {
	*x = CreateListenersRequest{}
	mi := &file_platformd_proxy_v1alpha1_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateListenersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateListenersRequest) ProtoMessage() {}

func (x *CreateListenersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_platformd_proxy_v1alpha1_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateListenersRequest.ProtoReflect.Descriptor instead.
func (*CreateListenersRequest) Descriptor() ([]byte, []int) {
	return file_platformd_proxy_v1alpha1_service_proto_rawDescGZIP(), []int{0}
}

func (x *CreateListenersRequest) GetWorkloadID() string {
	if x != nil {
		return x.WorkloadID
	}
	return ""
}

func (x *CreateListenersRequest) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

type CreateListenersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateListenersResponse) Reset() {
	*x = CreateListenersResponse{}
	mi := &file_platformd_proxy_v1alpha1_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateListenersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateListenersResponse) ProtoMessage() {}

func (x *CreateListenersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_platformd_proxy_v1alpha1_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateListenersResponse.ProtoReflect.Descriptor instead.
func (*CreateListenersResponse) Descriptor() ([]byte, []int) {
	return file_platformd_proxy_v1alpha1_service_proto_rawDescGZIP(), []int{1}
}

var File_platformd_proxy_v1alpha1_service_proto protoreflect.FileDescriptor

var file_platformd_proxy_v1alpha1_service_proto_rawDesc = []byte{
	0x0a, 0x26, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f,
	0x72, 0x6d, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x22, 0x48, 0x0a, 0x16, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x73, 0x74,
	0x65, 0x6e, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a,
	0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x44, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x22, 0x19, 0x0a, 0x17,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x86, 0x01, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x78,
	0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x76, 0x0a, 0x0f, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x73, 0x12, 0x30, 0x2e, 0x70, 0x6c,
	0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x73,
	0x74, 0x65, 0x6e, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e,
	0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4c,
	0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x47, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x2f, 0x65, 0x78, 0x70, 0x6c, 0x6f,
	0x72, 0x65, 0x72, 0x2d, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_platformd_proxy_v1alpha1_service_proto_rawDescOnce sync.Once
	file_platformd_proxy_v1alpha1_service_proto_rawDescData = file_platformd_proxy_v1alpha1_service_proto_rawDesc
)

func file_platformd_proxy_v1alpha1_service_proto_rawDescGZIP() []byte {
	file_platformd_proxy_v1alpha1_service_proto_rawDescOnce.Do(func() {
		file_platformd_proxy_v1alpha1_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_platformd_proxy_v1alpha1_service_proto_rawDescData)
	})
	return file_platformd_proxy_v1alpha1_service_proto_rawDescData
}

var file_platformd_proxy_v1alpha1_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_platformd_proxy_v1alpha1_service_proto_goTypes = []any{
	(*CreateListenersRequest)(nil),  // 0: platformd.proxy.v1alpha1.CreateListenersRequest
	(*CreateListenersResponse)(nil), // 1: platformd.proxy.v1alpha1.CreateListenersResponse
}
var file_platformd_proxy_v1alpha1_service_proto_depIdxs = []int32{
	0, // 0: platformd.proxy.v1alpha1.ProxyService.CreateListeners:input_type -> platformd.proxy.v1alpha1.CreateListenersRequest
	1, // 1: platformd.proxy.v1alpha1.ProxyService.CreateListeners:output_type -> platformd.proxy.v1alpha1.CreateListenersResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_platformd_proxy_v1alpha1_service_proto_init() }
func file_platformd_proxy_v1alpha1_service_proto_init() {
	if File_platformd_proxy_v1alpha1_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_platformd_proxy_v1alpha1_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_platformd_proxy_v1alpha1_service_proto_goTypes,
		DependencyIndexes: file_platformd_proxy_v1alpha1_service_proto_depIdxs,
		MessageInfos:      file_platformd_proxy_v1alpha1_service_proto_msgTypes,
	}.Build()
	File_platformd_proxy_v1alpha1_service_proto = out.File
	file_platformd_proxy_v1alpha1_service_proto_rawDesc = nil
	file_platformd_proxy_v1alpha1_service_proto_goTypes = nil
	file_platformd_proxy_v1alpha1_service_proto_depIdxs = nil
}
