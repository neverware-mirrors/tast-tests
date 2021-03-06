// Code generated by protoc-gen-go. DO NOT EDIT.
// source: file_system.proto

package baserpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ReadDirRequest struct {
	// Path to the directory to read the content of.
	Dir                  string   `protobuf:"bytes,1,opt,name=dir,proto3" json:"dir,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadDirRequest) Reset()         { *m = ReadDirRequest{} }
func (m *ReadDirRequest) String() string { return proto.CompactTextString(m) }
func (*ReadDirRequest) ProtoMessage()    {}
func (*ReadDirRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{0}
}

func (m *ReadDirRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadDirRequest.Unmarshal(m, b)
}
func (m *ReadDirRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadDirRequest.Marshal(b, m, deterministic)
}
func (m *ReadDirRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadDirRequest.Merge(m, src)
}
func (m *ReadDirRequest) XXX_Size() int {
	return xxx_messageInfo_ReadDirRequest.Size(m)
}
func (m *ReadDirRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadDirRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReadDirRequest proto.InternalMessageInfo

func (m *ReadDirRequest) GetDir() string {
	if m != nil {
		return m.Dir
	}
	return ""
}

type ReadDirResponse struct {
	Error *Error `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	// List of files under the directory, sorted by filename.
	Files                []*FileInfo `protobuf:"bytes,2,rep,name=files,proto3" json:"files,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ReadDirResponse) Reset()         { *m = ReadDirResponse{} }
func (m *ReadDirResponse) String() string { return proto.CompactTextString(m) }
func (*ReadDirResponse) ProtoMessage()    {}
func (*ReadDirResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{1}
}

func (m *ReadDirResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadDirResponse.Unmarshal(m, b)
}
func (m *ReadDirResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadDirResponse.Marshal(b, m, deterministic)
}
func (m *ReadDirResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadDirResponse.Merge(m, src)
}
func (m *ReadDirResponse) XXX_Size() int {
	return xxx_messageInfo_ReadDirResponse.Size(m)
}
func (m *ReadDirResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadDirResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ReadDirResponse proto.InternalMessageInfo

func (m *ReadDirResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func (m *ReadDirResponse) GetFiles() []*FileInfo {
	if m != nil {
		return m.Files
	}
	return nil
}

type StatRequest struct {
	// File path to the file to get file information.
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatRequest) Reset()         { *m = StatRequest{} }
func (m *StatRequest) String() string { return proto.CompactTextString(m) }
func (*StatRequest) ProtoMessage()    {}
func (*StatRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{2}
}

func (m *StatRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatRequest.Unmarshal(m, b)
}
func (m *StatRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatRequest.Marshal(b, m, deterministic)
}
func (m *StatRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatRequest.Merge(m, src)
}
func (m *StatRequest) XXX_Size() int {
	return xxx_messageInfo_StatRequest.Size(m)
}
func (m *StatRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StatRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StatRequest proto.InternalMessageInfo

func (m *StatRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type StatResponse struct {
	Error                *Error    `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Info                 *FileInfo `protobuf:"bytes,2,opt,name=info,proto3" json:"info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *StatResponse) Reset()         { *m = StatResponse{} }
func (m *StatResponse) String() string { return proto.CompactTextString(m) }
func (*StatResponse) ProtoMessage()    {}
func (*StatResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{3}
}

func (m *StatResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatResponse.Unmarshal(m, b)
}
func (m *StatResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatResponse.Marshal(b, m, deterministic)
}
func (m *StatResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatResponse.Merge(m, src)
}
func (m *StatResponse) XXX_Size() int {
	return xxx_messageInfo_StatResponse.Size(m)
}
func (m *StatResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StatResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StatResponse proto.InternalMessageInfo

func (m *StatResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func (m *StatResponse) GetInfo() *FileInfo {
	if m != nil {
		return m.Info
	}
	return nil
}

type ReadFileRequest struct {
	// File path to the file to be read.
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadFileRequest) Reset()         { *m = ReadFileRequest{} }
func (m *ReadFileRequest) String() string { return proto.CompactTextString(m) }
func (*ReadFileRequest) ProtoMessage()    {}
func (*ReadFileRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{4}
}

func (m *ReadFileRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadFileRequest.Unmarshal(m, b)
}
func (m *ReadFileRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadFileRequest.Marshal(b, m, deterministic)
}
func (m *ReadFileRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadFileRequest.Merge(m, src)
}
func (m *ReadFileRequest) XXX_Size() int {
	return xxx_messageInfo_ReadFileRequest.Size(m)
}
func (m *ReadFileRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadFileRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReadFileRequest proto.InternalMessageInfo

func (m *ReadFileRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type ReadFileResponse struct {
	Error                *Error   `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Content              []byte   `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadFileResponse) Reset()         { *m = ReadFileResponse{} }
func (m *ReadFileResponse) String() string { return proto.CompactTextString(m) }
func (*ReadFileResponse) ProtoMessage()    {}
func (*ReadFileResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{5}
}

func (m *ReadFileResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadFileResponse.Unmarshal(m, b)
}
func (m *ReadFileResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadFileResponse.Marshal(b, m, deterministic)
}
func (m *ReadFileResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadFileResponse.Merge(m, src)
}
func (m *ReadFileResponse) XXX_Size() int {
	return xxx_messageInfo_ReadFileResponse.Size(m)
}
func (m *ReadFileResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadFileResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ReadFileResponse proto.InternalMessageInfo

func (m *ReadFileResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func (m *ReadFileResponse) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

// FileInfo contains attributes of a file.
type FileInfo struct {
	Name                 string               `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Size                 uint64               `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	Mode                 uint64               `protobuf:"varint,3,opt,name=mode,proto3" json:"mode,omitempty"`
	Modified             *timestamp.Timestamp `protobuf:"bytes,4,opt,name=modified,proto3" json:"modified,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *FileInfo) Reset()         { *m = FileInfo{} }
func (m *FileInfo) String() string { return proto.CompactTextString(m) }
func (*FileInfo) ProtoMessage()    {}
func (*FileInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{6}
}

func (m *FileInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileInfo.Unmarshal(m, b)
}
func (m *FileInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileInfo.Marshal(b, m, deterministic)
}
func (m *FileInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileInfo.Merge(m, src)
}
func (m *FileInfo) XXX_Size() int {
	return xxx_messageInfo_FileInfo.Size(m)
}
func (m *FileInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_FileInfo.DiscardUnknown(m)
}

var xxx_messageInfo_FileInfo proto.InternalMessageInfo

func (m *FileInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FileInfo) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *FileInfo) GetMode() uint64 {
	if m != nil {
		return m.Mode
	}
	return 0
}

func (m *FileInfo) GetModified() *timestamp.Timestamp {
	if m != nil {
		return m.Modified
	}
	return nil
}

type Error struct {
	// Types that are valid to be assigned to Type:
	//	*Error_Errno
	//	*Error_Link
	//	*Error_Path
	//	*Error_Syscall
	//	*Error_Msg
	Type                 isError_Type `protobuf_oneof:"type"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{7}
}

func (m *Error) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Error.Unmarshal(m, b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Error.Marshal(b, m, deterministic)
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return xxx_messageInfo_Error.Size(m)
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

type isError_Type interface {
	isError_Type()
}

type Error_Errno struct {
	Errno uint32 `protobuf:"varint,1,opt,name=errno,proto3,oneof"`
}

type Error_Link struct {
	Link *LinkError `protobuf:"bytes,2,opt,name=link,proto3,oneof"`
}

type Error_Path struct {
	Path *PathError `protobuf:"bytes,3,opt,name=path,proto3,oneof"`
}

type Error_Syscall struct {
	Syscall *SyscallError `protobuf:"bytes,4,opt,name=syscall,proto3,oneof"`
}

type Error_Msg struct {
	Msg string `protobuf:"bytes,5,opt,name=msg,proto3,oneof"`
}

func (*Error_Errno) isError_Type() {}

func (*Error_Link) isError_Type() {}

func (*Error_Path) isError_Type() {}

func (*Error_Syscall) isError_Type() {}

func (*Error_Msg) isError_Type() {}

func (m *Error) GetType() isError_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (m *Error) GetErrno() uint32 {
	if x, ok := m.GetType().(*Error_Errno); ok {
		return x.Errno
	}
	return 0
}

func (m *Error) GetLink() *LinkError {
	if x, ok := m.GetType().(*Error_Link); ok {
		return x.Link
	}
	return nil
}

func (m *Error) GetPath() *PathError {
	if x, ok := m.GetType().(*Error_Path); ok {
		return x.Path
	}
	return nil
}

func (m *Error) GetSyscall() *SyscallError {
	if x, ok := m.GetType().(*Error_Syscall); ok {
		return x.Syscall
	}
	return nil
}

func (m *Error) GetMsg() string {
	if x, ok := m.GetType().(*Error_Msg); ok {
		return x.Msg
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Error) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Error_Errno)(nil),
		(*Error_Link)(nil),
		(*Error_Path)(nil),
		(*Error_Syscall)(nil),
		(*Error_Msg)(nil),
	}
}

type LinkError struct {
	Op                   string   `protobuf:"bytes,1,opt,name=op,proto3" json:"op,omitempty"`
	Old                  string   `protobuf:"bytes,2,opt,name=old,proto3" json:"old,omitempty"`
	New                  string   `protobuf:"bytes,3,opt,name=new,proto3" json:"new,omitempty"`
	Error                *Error   `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LinkError) Reset()         { *m = LinkError{} }
func (m *LinkError) String() string { return proto.CompactTextString(m) }
func (*LinkError) ProtoMessage()    {}
func (*LinkError) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{8}
}

func (m *LinkError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LinkError.Unmarshal(m, b)
}
func (m *LinkError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LinkError.Marshal(b, m, deterministic)
}
func (m *LinkError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LinkError.Merge(m, src)
}
func (m *LinkError) XXX_Size() int {
	return xxx_messageInfo_LinkError.Size(m)
}
func (m *LinkError) XXX_DiscardUnknown() {
	xxx_messageInfo_LinkError.DiscardUnknown(m)
}

var xxx_messageInfo_LinkError proto.InternalMessageInfo

func (m *LinkError) GetOp() string {
	if m != nil {
		return m.Op
	}
	return ""
}

func (m *LinkError) GetOld() string {
	if m != nil {
		return m.Old
	}
	return ""
}

func (m *LinkError) GetNew() string {
	if m != nil {
		return m.New
	}
	return ""
}

func (m *LinkError) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

type PathError struct {
	Op                   string   `protobuf:"bytes,1,opt,name=op,proto3" json:"op,omitempty"`
	Path                 string   `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Error                *Error   `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PathError) Reset()         { *m = PathError{} }
func (m *PathError) String() string { return proto.CompactTextString(m) }
func (*PathError) ProtoMessage()    {}
func (*PathError) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{9}
}

func (m *PathError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PathError.Unmarshal(m, b)
}
func (m *PathError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PathError.Marshal(b, m, deterministic)
}
func (m *PathError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PathError.Merge(m, src)
}
func (m *PathError) XXX_Size() int {
	return xxx_messageInfo_PathError.Size(m)
}
func (m *PathError) XXX_DiscardUnknown() {
	xxx_messageInfo_PathError.DiscardUnknown(m)
}

var xxx_messageInfo_PathError proto.InternalMessageInfo

func (m *PathError) GetOp() string {
	if m != nil {
		return m.Op
	}
	return ""
}

func (m *PathError) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *PathError) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

type SyscallError struct {
	Syscall              string   `protobuf:"bytes,1,opt,name=syscall,proto3" json:"syscall,omitempty"`
	Error                *Error   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SyscallError) Reset()         { *m = SyscallError{} }
func (m *SyscallError) String() string { return proto.CompactTextString(m) }
func (*SyscallError) ProtoMessage()    {}
func (*SyscallError) Descriptor() ([]byte, []int) {
	return fileDescriptor_f798b4e0b3d56780, []int{10}
}

func (m *SyscallError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyscallError.Unmarshal(m, b)
}
func (m *SyscallError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyscallError.Marshal(b, m, deterministic)
}
func (m *SyscallError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyscallError.Merge(m, src)
}
func (m *SyscallError) XXX_Size() int {
	return xxx_messageInfo_SyscallError.Size(m)
}
func (m *SyscallError) XXX_DiscardUnknown() {
	xxx_messageInfo_SyscallError.DiscardUnknown(m)
}

var xxx_messageInfo_SyscallError proto.InternalMessageInfo

func (m *SyscallError) GetSyscall() string {
	if m != nil {
		return m.Syscall
	}
	return ""
}

func (m *SyscallError) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func init() {
	proto.RegisterType((*ReadDirRequest)(nil), "tast.cros.baserpc.ReadDirRequest")
	proto.RegisterType((*ReadDirResponse)(nil), "tast.cros.baserpc.ReadDirResponse")
	proto.RegisterType((*StatRequest)(nil), "tast.cros.baserpc.StatRequest")
	proto.RegisterType((*StatResponse)(nil), "tast.cros.baserpc.StatResponse")
	proto.RegisterType((*ReadFileRequest)(nil), "tast.cros.baserpc.ReadFileRequest")
	proto.RegisterType((*ReadFileResponse)(nil), "tast.cros.baserpc.ReadFileResponse")
	proto.RegisterType((*FileInfo)(nil), "tast.cros.baserpc.FileInfo")
	proto.RegisterType((*Error)(nil), "tast.cros.baserpc.Error")
	proto.RegisterType((*LinkError)(nil), "tast.cros.baserpc.LinkError")
	proto.RegisterType((*PathError)(nil), "tast.cros.baserpc.PathError")
	proto.RegisterType((*SyscallError)(nil), "tast.cros.baserpc.SyscallError")
}

func init() { proto.RegisterFile("file_system.proto", fileDescriptor_f798b4e0b3d56780) }

var fileDescriptor_f798b4e0b3d56780 = []byte{
	// 579 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0x41, 0x6f, 0xd3, 0x4c,
	0x10, 0x4d, 0x1d, 0xa7, 0x49, 0x26, 0xf9, 0xfa, 0xb5, 0x7b, 0x40, 0x56, 0x40, 0xa4, 0x5d, 0x54,
	0xd1, 0x93, 0x2d, 0x82, 0xc4, 0x85, 0x5b, 0x05, 0x28, 0x95, 0x38, 0xa0, 0x0d, 0x48, 0x08, 0x21,
	0x55, 0x8e, 0xbd, 0x49, 0x56, 0xb5, 0xbd, 0xc6, 0xbb, 0x01, 0x85, 0x03, 0x07, 0x7e, 0x29, 0x3f,
	0x05, 0xcd, 0xae, 0xed, 0x14, 0xd5, 0x8d, 0xa2, 0xde, 0x66, 0x67, 0xdf, 0x9b, 0x79, 0x33, 0x3b,
	0xb3, 0x70, 0xb2, 0x10, 0x09, 0xbf, 0x56, 0x1b, 0xa5, 0x79, 0xea, 0xe7, 0x85, 0xd4, 0x92, 0x9c,
	0xe8, 0x50, 0x69, 0x3f, 0x2a, 0xa4, 0xf2, 0xe7, 0xa1, 0xe2, 0x45, 0x1e, 0x8d, 0xc6, 0x4b, 0x29,
	0x97, 0x09, 0x0f, 0x0c, 0x60, 0xbe, 0x5e, 0x04, 0x5a, 0xa4, 0x5c, 0xe9, 0x30, 0xcd, 0x2d, 0x87,
	0x52, 0x38, 0x62, 0x3c, 0x8c, 0xdf, 0x88, 0x82, 0xf1, 0x6f, 0x6b, 0xae, 0x34, 0x39, 0x86, 0x76,
	0x2c, 0x0a, 0xef, 0xe0, 0xf4, 0xe0, 0xa2, 0xcf, 0xd0, 0xa4, 0x1a, 0xfe, 0xaf, 0x31, 0x2a, 0x97,
	0x99, 0xe2, 0xc4, 0x87, 0x0e, 0x2f, 0x0a, 0x69, 0x61, 0x83, 0x89, 0xe7, 0xdf, 0x49, 0xed, 0xbf,
	0xc5, 0x7b, 0x66, 0x61, 0xe4, 0x05, 0x74, 0x50, 0xaf, 0xf2, 0x9c, 0xd3, 0xf6, 0xc5, 0x60, 0xf2,
	0xb8, 0x01, 0xff, 0x4e, 0x24, 0xfc, 0x2a, 0x5b, 0x48, 0x66, 0x91, 0xf4, 0x0c, 0x06, 0x33, 0x1d,
	0xea, 0x4a, 0x16, 0x01, 0x37, 0x0b, 0x53, 0x5e, 0xea, 0x32, 0x36, 0x95, 0x30, 0xb4, 0x90, 0x07,
	0xaa, 0x0a, 0xc0, 0x15, 0xd9, 0x42, 0x7a, 0x8e, 0x81, 0xef, 0x14, 0x65, 0x80, 0xf4, 0xdc, 0x76,
	0x02, 0xbd, 0xbb, 0x74, 0x7d, 0x85, 0xe3, 0x2d, 0xec, 0x81, 0xda, 0x3c, 0xe8, 0x46, 0x32, 0xd3,
	0x3c, 0xd3, 0x46, 0xde, 0x90, 0x55, 0x47, 0xfa, 0x0b, 0x7a, 0x95, 0xac, 0xa6, 0xec, 0xe8, 0x53,
	0xe2, 0x27, 0x37, 0x34, 0x97, 0x19, 0x1b, 0x7d, 0xa9, 0x8c, 0xb9, 0xd7, 0xb6, 0x3e, 0xb4, 0xc9,
	0x2b, 0xe8, 0xa5, 0x32, 0x16, 0x0b, 0xc1, 0x63, 0xcf, 0x35, 0xa2, 0x46, 0xbe, 0x1d, 0x17, 0xbf,
	0x1a, 0x17, 0xff, 0x63, 0x35, 0x2e, 0xac, 0xc6, 0xd2, 0x3f, 0x07, 0xd0, 0x31, 0x52, 0xc9, 0x23,
	0x53, 0x53, 0x26, 0x4d, 0xfa, 0xff, 0xa6, 0x2d, 0x66, 0x8f, 0x64, 0x02, 0x6e, 0x22, 0xb2, 0x9b,
	0xb2, 0xaf, 0x4f, 0x1a, 0x4a, 0x7d, 0x2f, 0xb2, 0x1b, 0x13, 0x63, 0xda, 0x62, 0x06, 0x8b, 0x9c,
	0x3c, 0xd4, 0x2b, 0xa3, 0xb0, 0x99, 0xf3, 0x21, 0xd4, 0xab, 0x9a, 0x83, 0x58, 0xf2, 0x1a, 0xba,
	0x6a, 0xa3, 0xa2, 0x30, 0x49, 0xca, 0x02, 0xc6, 0x0d, 0xb4, 0x99, 0x45, 0x54, 0xcc, 0x8a, 0x41,
	0x08, 0xb4, 0x53, 0xb5, 0xf4, 0x3a, 0xd8, 0xb9, 0x69, 0x8b, 0xe1, 0xe1, 0xf2, 0x10, 0x5c, 0xbd,
	0xc9, 0x71, 0xb0, 0xfa, 0xb5, 0x42, 0x72, 0x04, 0x8e, 0xcc, 0xcb, 0x0e, 0x3b, 0x32, 0xc7, 0x05,
	0x91, 0x49, 0x6c, 0x8a, 0xeb, 0x33, 0x34, 0xd1, 0x93, 0xf1, 0x1f, 0x46, 0x7a, 0x9f, 0xa1, 0xb9,
	0x7d, 0x6d, 0x77, 0xaf, 0xd7, 0xa6, 0xd7, 0xd0, 0xaf, 0xcb, 0xbb, 0x93, 0x90, 0x94, 0xad, 0xb1,
	0x19, 0x6d, 0xe9, 0x75, 0x82, 0xf6, 0x7e, 0x09, 0x3e, 0xc3, 0xf0, 0x76, 0x23, 0x70, 0xbc, 0xaa,
	0xd6, 0xd9, 0x44, 0x75, 0x5f, 0xea, 0xc8, 0xce, 0x5e, 0x91, 0x27, 0xbf, 0x1d, 0x00, 0x9c, 0xc7,
	0x99, 0xf9, 0x8a, 0x08, 0x83, 0x6e, 0xf9, 0x59, 0x90, 0xb3, 0x06, 0xea, 0xbf, 0x9f, 0xcd, 0x88,
	0xee, 0x82, 0xd8, 0xcd, 0xa1, 0x2d, 0x72, 0x05, 0x2e, 0xee, 0x39, 0x79, 0xda, 0xf4, 0xbc, 0xdb,
	0x3f, 0x62, 0x34, 0xbe, 0xf7, 0xbe, 0x0e, 0xf5, 0x09, 0x7a, 0xd5, 0x6a, 0x92, 0xfb, 0x92, 0xdf,
	0x5a, 0xef, 0xd1, 0xb3, 0x9d, 0x98, 0x2a, 0xec, 0xe5, 0xf3, 0x2f, 0xe7, 0xd1, 0xaa, 0x90, 0xa9,
	0x58, 0xa7, 0x52, 0x05, 0x48, 0x09, 0x14, 0x2f, 0xbe, 0x8b, 0x88, 0xab, 0x00, 0xb9, 0x41, 0xc9,
	0x9d, 0x1f, 0x9a, 0xd5, 0x7a, 0xf9, 0x37, 0x00, 0x00, 0xff, 0xff, 0xe7, 0x73, 0xb0, 0x73, 0xbf,
	0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// FileSystemClient is the client API for FileSystem service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FileSystemClient interface {
	// ReadDir returns the content of a directory.
	ReadDir(ctx context.Context, in *ReadDirRequest, opts ...grpc.CallOption) (*ReadDirResponse, error)
	// Stat returns information of a file.
	Stat(ctx context.Context, in *StatRequest, opts ...grpc.CallOption) (*StatResponse, error)
	// ReadFile reads the content of a file.
	ReadFile(ctx context.Context, in *ReadFileRequest, opts ...grpc.CallOption) (*ReadFileResponse, error)
}

type fileSystemClient struct {
	cc *grpc.ClientConn
}

func NewFileSystemClient(cc *grpc.ClientConn) FileSystemClient {
	return &fileSystemClient{cc}
}

func (c *fileSystemClient) ReadDir(ctx context.Context, in *ReadDirRequest, opts ...grpc.CallOption) (*ReadDirResponse, error) {
	out := new(ReadDirResponse)
	err := c.cc.Invoke(ctx, "/tast.cros.baserpc.FileSystem/ReadDir", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileSystemClient) Stat(ctx context.Context, in *StatRequest, opts ...grpc.CallOption) (*StatResponse, error) {
	out := new(StatResponse)
	err := c.cc.Invoke(ctx, "/tast.cros.baserpc.FileSystem/Stat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileSystemClient) ReadFile(ctx context.Context, in *ReadFileRequest, opts ...grpc.CallOption) (*ReadFileResponse, error) {
	out := new(ReadFileResponse)
	err := c.cc.Invoke(ctx, "/tast.cros.baserpc.FileSystem/ReadFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FileSystemServer is the server API for FileSystem service.
type FileSystemServer interface {
	// ReadDir returns the content of a directory.
	ReadDir(context.Context, *ReadDirRequest) (*ReadDirResponse, error)
	// Stat returns information of a file.
	Stat(context.Context, *StatRequest) (*StatResponse, error)
	// ReadFile reads the content of a file.
	ReadFile(context.Context, *ReadFileRequest) (*ReadFileResponse, error)
}

// UnimplementedFileSystemServer can be embedded to have forward compatible implementations.
type UnimplementedFileSystemServer struct {
}

func (*UnimplementedFileSystemServer) ReadDir(ctx context.Context, req *ReadDirRequest) (*ReadDirResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadDir not implemented")
}
func (*UnimplementedFileSystemServer) Stat(ctx context.Context, req *StatRequest) (*StatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stat not implemented")
}
func (*UnimplementedFileSystemServer) ReadFile(ctx context.Context, req *ReadFileRequest) (*ReadFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadFile not implemented")
}

func RegisterFileSystemServer(s *grpc.Server, srv FileSystemServer) {
	s.RegisterService(&_FileSystem_serviceDesc, srv)
}

func _FileSystem_ReadDir_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadDirRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileSystemServer).ReadDir(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.baserpc.FileSystem/ReadDir",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileSystemServer).ReadDir(ctx, req.(*ReadDirRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileSystem_Stat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileSystemServer).Stat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.baserpc.FileSystem/Stat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileSystemServer).Stat(ctx, req.(*StatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileSystem_ReadFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileSystemServer).ReadFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.baserpc.FileSystem/ReadFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileSystemServer).ReadFile(ctx, req.(*ReadFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FileSystem_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tast.cros.baserpc.FileSystem",
	HandlerType: (*FileSystemServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadDir",
			Handler:    _FileSystem_ReadDir_Handler,
		},
		{
			MethodName: "Stat",
			Handler:    _FileSystem_Stat_Handler,
		},
		{
			MethodName: "ReadFile",
			Handler:    _FileSystem_ReadFile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "file_system.proto",
}
