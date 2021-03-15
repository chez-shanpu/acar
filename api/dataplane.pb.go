// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.15.5
// source: dataplane.proto

package api

import (
	reflect "reflect"
	sync "sync"

	types "github.com/chez-shanpu/acar/api/types"
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SRInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SrcAddr string   `protobuf:"bytes,1,opt,name=src_addr,json=srcAddr,proto3" json:"src_addr,omitempty"`
	DstAddr string   `protobuf:"bytes,2,opt,name=dst_addr,json=dstAddr,proto3" json:"dst_addr,omitempty"`
	SidList []string `protobuf:"bytes,3,rep,name=sid_list,json=sidList,proto3" json:"sid_list,omitempty"`
}

func (x *SRInfo) Reset() {
	*x = SRInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dataplane_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SRInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SRInfo) ProtoMessage() {}

func (x *SRInfo) ProtoReflect() protoreflect.Message {
	mi := &file_dataplane_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SRInfo.ProtoReflect.Descriptor instead.
func (*SRInfo) Descriptor() ([]byte, []int) {
	return file_dataplane_proto_rawDescGZIP(), []int{0}
}

func (x *SRInfo) GetSrcAddr() string {
	if x != nil {
		return x.SrcAddr
	}
	return ""
}

func (x *SRInfo) GetDstAddr() string {
	if x != nil {
		return x.DstAddr
	}
	return ""
}

func (x *SRInfo) GetSidList() []string {
	if x != nil {
		return x.SidList
	}
	return nil
}

var File_dataplane_proto protoreflect.FileDescriptor

var file_dataplane_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x64, 0x61, 0x74, 0x61, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x61, 0x63, 0x61, 0x72, 0x1a, 0x12, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x59, 0x0a, 0x06, 0x53,
	0x52, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x72, 0x63, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x72, 0x63, 0x41, 0x64, 0x64, 0x72,
	0x12, 0x19, 0x0a, 0x08, 0x64, 0x73, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x64, 0x73, 0x74, 0x41, 0x64, 0x64, 0x72, 0x12, 0x19, 0x0a, 0x08, 0x73,
	0x69, 0x64, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x73,
	0x69, 0x64, 0x4c, 0x69, 0x73, 0x74, 0x32, 0x3e, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x50, 0x6c,
	0x61, 0x6e, 0x65, 0x12, 0x31, 0x0a, 0x0d, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x53, 0x52, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x12, 0x0c, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x53, 0x52, 0x49, 0x6e,
	0x66, 0x6f, 0x1a, 0x12, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x21, 0x5a, 0x1f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x65, 0x7a, 0x2d, 0x73, 0x68, 0x61, 0x6e, 0x70, 0x75,
	0x2f, 0x61, 0x63, 0x61, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_dataplane_proto_rawDescOnce sync.Once
	file_dataplane_proto_rawDescData = file_dataplane_proto_rawDesc
)

func file_dataplane_proto_rawDescGZIP() []byte {
	file_dataplane_proto_rawDescOnce.Do(func() {
		file_dataplane_proto_rawDescData = protoimpl.X.CompressGZIP(file_dataplane_proto_rawDescData)
	})
	return file_dataplane_proto_rawDescData
}

var file_dataplane_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_dataplane_proto_goTypes = []interface{}{
	(*SRInfo)(nil),       // 0: acar.SRInfo
	(*types.Result)(nil), // 1: acar.types.Result
}
var file_dataplane_proto_depIdxs = []int32{
	0, // 0: acar.DataPlane.ApplySRPolicy:input_type -> acar.SRInfo
	1, // 1: acar.DataPlane.ApplySRPolicy:output_type -> acar.types.Result
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_dataplane_proto_init() }
func file_dataplane_proto_init() {
	if File_dataplane_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dataplane_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SRInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_dataplane_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_dataplane_proto_goTypes,
		DependencyIndexes: file_dataplane_proto_depIdxs,
		MessageInfos:      file_dataplane_proto_msgTypes,
	}.Build()
	File_dataplane_proto = out.File
	file_dataplane_proto_rawDesc = nil
	file_dataplane_proto_goTypes = nil
	file_dataplane_proto_depIdxs = nil
}
