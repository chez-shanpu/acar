// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: monitoring-server.proto

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

type GetNodesParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetNodesParams) Reset() {
	*x = GetNodesParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monitoring_server_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNodesParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNodesParams) ProtoMessage() {}

func (x *GetNodesParams) ProtoReflect() protoreflect.Message {
	mi := &file_monitoring_server_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNodesParams.ProtoReflect.Descriptor instead.
func (*GetNodesParams) Descriptor() ([]byte, []int) {
	return file_monitoring_server_proto_rawDescGZIP(), []int{0}
}

type NodesInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Nodes []*Node `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
}

func (x *NodesInfo) Reset() {
	*x = NodesInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monitoring_server_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodesInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodesInfo) ProtoMessage() {}

func (x *NodesInfo) ProtoReflect() protoreflect.Message {
	mi := &file_monitoring_server_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodesInfo.ProtoReflect.Descriptor instead.
func (*NodesInfo) Descriptor() ([]byte, []int) {
	return file_monitoring_server_proto_rawDescGZIP(), []int{1}
}

func (x *NodesInfo) GetNodes() []*Node {
	if x != nil {
		return x.Nodes
	}
	return nil
}

type Node struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SID       string      `protobuf:"bytes,1,opt,name=SID,proto3" json:"SID,omitempty"`
	LinkCosts []*LinkCost `protobuf:"bytes,2,rep,name=link_costs,json=linkCosts,proto3" json:"link_costs,omitempty"`
}

func (x *Node) Reset() {
	*x = Node{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monitoring_server_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Node) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Node) ProtoMessage() {}

func (x *Node) ProtoReflect() protoreflect.Message {
	mi := &file_monitoring_server_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Node.ProtoReflect.Descriptor instead.
func (*Node) Descriptor() ([]byte, []int) {
	return file_monitoring_server_proto_rawDescGZIP(), []int{2}
}

func (x *Node) GetSID() string {
	if x != nil {
		return x.SID
	}
	return ""
}

func (x *Node) GetLinkCosts() []*LinkCost {
	if x != nil {
		return x.LinkCosts
	}
	return nil
}

type LinkCost struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NextSid string  `protobuf:"bytes,1,opt,name=next_sid,json=nextSid,proto3" json:"next_sid,omitempty"`
	Cost    float64 `protobuf:"fixed64,2,opt,name=cost,proto3" json:"cost,omitempty"`
}

func (x *LinkCost) Reset() {
	*x = LinkCost{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monitoring_server_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LinkCost) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LinkCost) ProtoMessage() {}

func (x *LinkCost) ProtoReflect() protoreflect.Message {
	mi := &file_monitoring_server_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LinkCost.ProtoReflect.Descriptor instead.
func (*LinkCost) Descriptor() ([]byte, []int) {
	return file_monitoring_server_proto_rawDescGZIP(), []int{3}
}

func (x *LinkCost) GetNextSid() string {
	if x != nil {
		return x.NextSid
	}
	return ""
}

func (x *LinkCost) GetCost() float64 {
	if x != nil {
		return x.Cost
	}
	return 0
}

var File_monitoring_server_proto protoreflect.FileDescriptor

var file_monitoring_server_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2d, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x61, 0x63, 0x61, 0x72, 0x1a,
	0x12, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x10, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x22, 0x2d, 0x0a, 0x09, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x20, 0x0a, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0a, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x05, 0x6e,
	0x6f, 0x64, 0x65, 0x73, 0x22, 0x47, 0x0a, 0x04, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x53, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x53, 0x49, 0x44, 0x12, 0x2d,
	0x0a, 0x0a, 0x6c, 0x69, 0x6e, 0x6b, 0x5f, 0x63, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x43, 0x6f,
	0x73, 0x74, 0x52, 0x09, 0x6c, 0x69, 0x6e, 0x6b, 0x43, 0x6f, 0x73, 0x74, 0x73, 0x22, 0x39, 0x0a,
	0x08, 0x4c, 0x69, 0x6e, 0x6b, 0x43, 0x6f, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x6e, 0x65, 0x78,
	0x74, 0x5f, 0x73, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x65, 0x78,
	0x74, 0x53, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x04, 0x63, 0x6f, 0x73, 0x74, 0x32, 0x7b, 0x0a, 0x10, 0x4d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x31, 0x0a, 0x08,
	0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x14, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e,
	0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x0f,
	0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x34, 0x0a, 0x0d, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x6f, 0x64, 0x65, 0x73,
	0x12, 0x0f, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x49, 0x6e, 0x66,
	0x6f, 0x1a, 0x12, 0x2e, 0x61, 0x63, 0x61, 0x72, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x21, 0x5a, 0x1f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x65, 0x7a, 0x2d, 0x73, 0x68, 0x61, 0x6e, 0x70, 0x75, 0x2f,
	0x61, 0x63, 0x61, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_monitoring_server_proto_rawDescOnce sync.Once
	file_monitoring_server_proto_rawDescData = file_monitoring_server_proto_rawDesc
)

func file_monitoring_server_proto_rawDescGZIP() []byte {
	file_monitoring_server_proto_rawDescOnce.Do(func() {
		file_monitoring_server_proto_rawDescData = protoimpl.X.CompressGZIP(file_monitoring_server_proto_rawDescData)
	})
	return file_monitoring_server_proto_rawDescData
}

var file_monitoring_server_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_monitoring_server_proto_goTypes = []interface{}{
	(*GetNodesParams)(nil), // 0: acar.GetNodesParams
	(*NodesInfo)(nil),      // 1: acar.NodesInfo
	(*Node)(nil),           // 2: acar.Node
	(*LinkCost)(nil),       // 3: acar.LinkCost
	(*types.Result)(nil),   // 4: acar.types.Result
}
var file_monitoring_server_proto_depIdxs = []int32{
	2, // 0: acar.NodesInfo.nodes:type_name -> acar.Node
	3, // 1: acar.Node.link_costs:type_name -> acar.LinkCost
	0, // 2: acar.MonitoringServer.GetNodes:input_type -> acar.GetNodesParams
	1, // 3: acar.MonitoringServer.RegisterNodes:input_type -> acar.NodesInfo
	1, // 4: acar.MonitoringServer.GetNodes:output_type -> acar.NodesInfo
	4, // 5: acar.MonitoringServer.RegisterNodes:output_type -> acar.types.Result
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_monitoring_server_proto_init() }
func file_monitoring_server_proto_init() {
	if File_monitoring_server_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_monitoring_server_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNodesParams); i {
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
		file_monitoring_server_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodesInfo); i {
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
		file_monitoring_server_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Node); i {
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
		file_monitoring_server_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LinkCost); i {
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
			RawDescriptor: file_monitoring_server_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_monitoring_server_proto_goTypes,
		DependencyIndexes: file_monitoring_server_proto_depIdxs,
		MessageInfos:      file_monitoring_server_proto_msgTypes,
	}.Build()
	File_monitoring_server_proto = out.File
	file_monitoring_server_proto_rawDesc = nil
	file_monitoring_server_proto_goTypes = nil
	file_monitoring_server_proto_depIdxs = nil
}