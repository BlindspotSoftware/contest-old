// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: contest/v1/grpclistener.proto

package contestlistener

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StartJobRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Requestor string `protobuf:"bytes,1,opt,name=requestor,proto3" json:"requestor,omitempty"`
	Job       []byte `protobuf:"bytes,2,opt,name=job,proto3" json:"job,omitempty"`
}

func (x *StartJobRequest) Reset() {
	*x = StartJobRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contest_v1_grpclistener_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartJobRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartJobRequest) ProtoMessage() {}

func (x *StartJobRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contest_v1_grpclistener_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartJobRequest.ProtoReflect.Descriptor instead.
func (*StartJobRequest) Descriptor() ([]byte, []int) {
	return file_contest_v1_grpclistener_proto_rawDescGZIP(), []int{0}
}

func (x *StartJobRequest) GetRequestor() string {
	if x != nil {
		return x.Requestor
	}
	return ""
}

func (x *StartJobRequest) GetJob() []byte {
	if x != nil {
		return x.Job
	}
	return nil
}

type StartJobResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId int32  `protobuf:"varint,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Error string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StartJobResponse) Reset() {
	*x = StartJobResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contest_v1_grpclistener_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartJobResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartJobResponse) ProtoMessage() {}

func (x *StartJobResponse) ProtoReflect() protoreflect.Message {
	mi := &file_contest_v1_grpclistener_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartJobResponse.ProtoReflect.Descriptor instead.
func (*StartJobResponse) Descriptor() ([]byte, []int) {
	return file_contest_v1_grpclistener_proto_rawDescGZIP(), []int{1}
}

func (x *StartJobResponse) GetJobId() int32 {
	if x != nil {
		return x.JobId
	}
	return 0
}

func (x *StartJobResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type StatusJobRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId     int32  `protobuf:"varint,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Requestor string `protobuf:"bytes,2,opt,name=requestor,proto3" json:"requestor,omitempty"`
}

func (x *StatusJobRequest) Reset() {
	*x = StatusJobRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contest_v1_grpclistener_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusJobRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusJobRequest) ProtoMessage() {}

func (x *StatusJobRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contest_v1_grpclistener_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusJobRequest.ProtoReflect.Descriptor instead.
func (*StatusJobRequest) Descriptor() ([]byte, []int) {
	return file_contest_v1_grpclistener_proto_rawDescGZIP(), []int{2}
}

func (x *StatusJobRequest) GetJobId() int32 {
	if x != nil {
		return x.JobId
	}
	return 0
}

func (x *StatusJobRequest) GetRequestor() string {
	if x != nil {
		return x.Requestor
	}
	return ""
}

type StatusJobResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Error  string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Report []byte `protobuf:"bytes,3,opt,name=report,proto3" json:"report,omitempty"`
	Log    []byte `protobuf:"bytes,4,opt,name=log,proto3" json:"log,omitempty"`
}

func (x *StatusJobResponse) Reset() {
	*x = StatusJobResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contest_v1_grpclistener_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusJobResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusJobResponse) ProtoMessage() {}

func (x *StatusJobResponse) ProtoReflect() protoreflect.Message {
	mi := &file_contest_v1_grpclistener_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusJobResponse.ProtoReflect.Descriptor instead.
func (*StatusJobResponse) Descriptor() ([]byte, []int) {
	return file_contest_v1_grpclistener_proto_rawDescGZIP(), []int{3}
}

func (x *StatusJobResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *StatusJobResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *StatusJobResponse) GetReport() []byte {
	if x != nil {
		return x.Report
	}
	return nil
}

func (x *StatusJobResponse) GetLog() []byte {
	if x != nil {
		return x.Log
	}
	return nil
}

var File_contest_v1_grpclistener_proto protoreflect.FileDescriptor

var file_contest_v1_grpclistener_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x67, 0x72, 0x70,
	0x63, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x22, 0x41, 0x0a, 0x0f, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c,
	0x0a, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x12, 0x10, 0x0a, 0x03,
	0x6a, 0x6f, 0x62, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6a, 0x6f, 0x62, 0x22, 0x3f,
	0x0a, 0x10, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22,
	0x47, 0x0a, 0x10, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x22, 0x6b, 0x0a, 0x11, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x03, 0x6c, 0x6f, 0x67, 0x32, 0xa7, 0x01, 0x0a, 0x0e, 0x43, 0x6f, 0x6e, 0x54, 0x65, 0x73,
	0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x47, 0x0a, 0x08, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x4a, 0x6f, 0x62, 0x12, 0x1b, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1c, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x4c, 0x0a, 0x09, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4a, 0x6f, 0x62, 0x12, 0x1c,
	0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42,
	0x13, 0x5a, 0x11, 0x2e, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x6c, 0x69, 0x73, 0x74,
	0x65, 0x6e, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_contest_v1_grpclistener_proto_rawDescOnce sync.Once
	file_contest_v1_grpclistener_proto_rawDescData = file_contest_v1_grpclistener_proto_rawDesc
)

func file_contest_v1_grpclistener_proto_rawDescGZIP() []byte {
	file_contest_v1_grpclistener_proto_rawDescOnce.Do(func() {
		file_contest_v1_grpclistener_proto_rawDescData = protoimpl.X.CompressGZIP(file_contest_v1_grpclistener_proto_rawDescData)
	})
	return file_contest_v1_grpclistener_proto_rawDescData
}

var file_contest_v1_grpclistener_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_contest_v1_grpclistener_proto_goTypes = []interface{}{
	(*StartJobRequest)(nil),   // 0: contest.v1.StartJobRequest
	(*StartJobResponse)(nil),  // 1: contest.v1.StartJobResponse
	(*StatusJobRequest)(nil),  // 2: contest.v1.StatusJobRequest
	(*StatusJobResponse)(nil), // 3: contest.v1.StatusJobResponse
}
var file_contest_v1_grpclistener_proto_depIdxs = []int32{
	0, // 0: contest.v1.ConTestService.StartJob:input_type -> contest.v1.StartJobRequest
	2, // 1: contest.v1.ConTestService.StatusJob:input_type -> contest.v1.StatusJobRequest
	1, // 2: contest.v1.ConTestService.StartJob:output_type -> contest.v1.StartJobResponse
	3, // 3: contest.v1.ConTestService.StatusJob:output_type -> contest.v1.StatusJobResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_contest_v1_grpclistener_proto_init() }
func file_contest_v1_grpclistener_proto_init() {
	if File_contest_v1_grpclistener_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_contest_v1_grpclistener_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartJobRequest); i {
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
		file_contest_v1_grpclistener_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartJobResponse); i {
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
		file_contest_v1_grpclistener_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusJobRequest); i {
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
		file_contest_v1_grpclistener_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusJobResponse); i {
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
			RawDescriptor: file_contest_v1_grpclistener_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_contest_v1_grpclistener_proto_goTypes,
		DependencyIndexes: file_contest_v1_grpclistener_proto_depIdxs,
		MessageInfos:      file_contest_v1_grpclistener_proto_msgTypes,
	}.Build()
	File_contest_v1_grpclistener_proto = out.File
	file_contest_v1_grpclistener_proto_rawDesc = nil
	file_contest_v1_grpclistener_proto_goTypes = nil
	file_contest_v1_grpclistener_proto_depIdxs = nil
}
