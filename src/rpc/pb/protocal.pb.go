// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocal.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type AuctionType int32

const (
	AuctionType_FIRST_PRICE  AuctionType = 0
	AuctionType_SECOND_PRICE AuctionType = 1
	AuctionType_FIXED_PRICE  AuctionType = 2
)

var AuctionType_name = map[int32]string{
	0: "FIRST_PRICE",
	1: "SECOND_PRICE",
	2: "FIXED_PRICE",
}

var AuctionType_value = map[string]int32{
	"FIRST_PRICE":  0,
	"SECOND_PRICE": 1,
	"FIXED_PRICE":  2,
}

func (x AuctionType) String() string {
	return proto.EnumName(AuctionType_name, int32(x))
}

func (AuctionType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{0}
}

type VarintMsg struct {
	ArgI32               int32       `protobuf:"varint,1,opt,name=argI32,proto3" json:"argI32,omitempty"`
	ArgI64               int64       `protobuf:"varint,2,opt,name=argI64,proto3" json:"argI64,omitempty"`
	ArgUI32              uint32      `protobuf:"varint,3,opt,name=argUI32,proto3" json:"argUI32,omitempty"`
	ArgUI64              uint64      `protobuf:"varint,4,opt,name=argUI64,proto3" json:"argUI64,omitempty"`
	ArgSI32              int32       `protobuf:"zigzag32,5,opt,name=argSI32,proto3" json:"argSI32,omitempty"`
	ArgSI64              int64       `protobuf:"zigzag64,6,opt,name=argSI64,proto3" json:"argSI64,omitempty"`
	ArgBool              []bool      `protobuf:"varint,7,rep,packed,name=argBool,proto3" json:"argBool,omitempty"`
	ArgEnum              AuctionType `protobuf:"varint,8,opt,name=argEnum,proto3,enum=pb.AuctionType" json:"argEnum,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *VarintMsg) Reset()         { *m = VarintMsg{} }
func (m *VarintMsg) String() string { return proto.CompactTextString(m) }
func (*VarintMsg) ProtoMessage()    {}
func (*VarintMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{0}
}

func (m *VarintMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VarintMsg.Unmarshal(m, b)
}
func (m *VarintMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VarintMsg.Marshal(b, m, deterministic)
}
func (m *VarintMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VarintMsg.Merge(m, src)
}
func (m *VarintMsg) XXX_Size() int {
	return xxx_messageInfo_VarintMsg.Size(m)
}
func (m *VarintMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_VarintMsg.DiscardUnknown(m)
}

var xxx_messageInfo_VarintMsg proto.InternalMessageInfo

func (m *VarintMsg) GetArgI32() int32 {
	if m != nil {
		return m.ArgI32
	}
	return 0
}

func (m *VarintMsg) GetArgI64() int64 {
	if m != nil {
		return m.ArgI64
	}
	return 0
}

func (m *VarintMsg) GetArgUI32() uint32 {
	if m != nil {
		return m.ArgUI32
	}
	return 0
}

func (m *VarintMsg) GetArgUI64() uint64 {
	if m != nil {
		return m.ArgUI64
	}
	return 0
}

func (m *VarintMsg) GetArgSI32() int32 {
	if m != nil {
		return m.ArgSI32
	}
	return 0
}

func (m *VarintMsg) GetArgSI64() int64 {
	if m != nil {
		return m.ArgSI64
	}
	return 0
}

func (m *VarintMsg) GetArgBool() []bool {
	if m != nil {
		return m.ArgBool
	}
	return nil
}

func (m *VarintMsg) GetArgEnum() AuctionType {
	if m != nil {
		return m.ArgEnum
	}
	return AuctionType_FIRST_PRICE
}

type Bit64 struct {
	ArgFixed64           uint64   `protobuf:"fixed64,1,opt,name=argFixed64,proto3" json:"argFixed64,omitempty"`
	ArgSFixed64          int64    `protobuf:"fixed64,2,opt,name=argSFixed64,proto3" json:"argSFixed64,omitempty"`
	ArgDouble            float64  `protobuf:"fixed64,3,opt,name=argDouble,proto3" json:"argDouble,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Bit64) Reset()         { *m = Bit64{} }
func (m *Bit64) String() string { return proto.CompactTextString(m) }
func (*Bit64) ProtoMessage()    {}
func (*Bit64) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{1}
}

func (m *Bit64) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Bit64.Unmarshal(m, b)
}
func (m *Bit64) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Bit64.Marshal(b, m, deterministic)
}
func (m *Bit64) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Bit64.Merge(m, src)
}
func (m *Bit64) XXX_Size() int {
	return xxx_messageInfo_Bit64.Size(m)
}
func (m *Bit64) XXX_DiscardUnknown() {
	xxx_messageInfo_Bit64.DiscardUnknown(m)
}

var xxx_messageInfo_Bit64 proto.InternalMessageInfo

func (m *Bit64) GetArgFixed64() uint64 {
	if m != nil {
		return m.ArgFixed64
	}
	return 0
}

func (m *Bit64) GetArgSFixed64() int64 {
	if m != nil {
		return m.ArgSFixed64
	}
	return 0
}

func (m *Bit64) GetArgDouble() float64 {
	if m != nil {
		return m.ArgDouble
	}
	return 0
}

type Bit32 struct {
	ArgFixed32           uint32   `protobuf:"fixed32,1,opt,name=argFixed32,proto3" json:"argFixed32,omitempty"`
	ArgSFixed32          int32    `protobuf:"fixed32,2,opt,name=argSFixed32,proto3" json:"argSFixed32,omitempty"`
	ArgFloat             float32  `protobuf:"fixed32,3,opt,name=argFloat,proto3" json:"argFloat,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Bit32) Reset()         { *m = Bit32{} }
func (m *Bit32) String() string { return proto.CompactTextString(m) }
func (*Bit32) ProtoMessage()    {}
func (*Bit32) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{2}
}

func (m *Bit32) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Bit32.Unmarshal(m, b)
}
func (m *Bit32) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Bit32.Marshal(b, m, deterministic)
}
func (m *Bit32) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Bit32.Merge(m, src)
}
func (m *Bit32) XXX_Size() int {
	return xxx_messageInfo_Bit32.Size(m)
}
func (m *Bit32) XXX_DiscardUnknown() {
	xxx_messageInfo_Bit32.DiscardUnknown(m)
}

var xxx_messageInfo_Bit32 proto.InternalMessageInfo

func (m *Bit32) GetArgFixed32() uint32 {
	if m != nil {
		return m.ArgFixed32
	}
	return 0
}

func (m *Bit32) GetArgSFixed32() int32 {
	if m != nil {
		return m.ArgSFixed32
	}
	return 0
}

func (m *Bit32) GetArgFloat() float32 {
	if m != nil {
		return m.ArgFloat
	}
	return 0
}

type Simple struct {
	ArgI32               int32    `protobuf:"varint,1,opt,name=argI32,proto3" json:"argI32,omitempty"`
	ArgUI32              uint32   `protobuf:"varint,2,opt,name=argUI32,proto3" json:"argUI32,omitempty"`
	ArgBool              bool     `protobuf:"varint,3,opt,name=argBool,proto3" json:"argBool,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Simple) Reset()         { *m = Simple{} }
func (m *Simple) String() string { return proto.CompactTextString(m) }
func (*Simple) ProtoMessage()    {}
func (*Simple) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{3}
}

func (m *Simple) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Simple.Unmarshal(m, b)
}
func (m *Simple) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Simple.Marshal(b, m, deterministic)
}
func (m *Simple) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Simple.Merge(m, src)
}
func (m *Simple) XXX_Size() int {
	return xxx_messageInfo_Simple.Size(m)
}
func (m *Simple) XXX_DiscardUnknown() {
	xxx_messageInfo_Simple.DiscardUnknown(m)
}

var xxx_messageInfo_Simple proto.InternalMessageInfo

func (m *Simple) GetArgI32() int32 {
	if m != nil {
		return m.ArgI32
	}
	return 0
}

func (m *Simple) GetArgUI32() uint32 {
	if m != nil {
		return m.ArgUI32
	}
	return 0
}

func (m *Simple) GetArgBool() bool {
	if m != nil {
		return m.ArgBool
	}
	return false
}

type Repeat struct {
	ArgBoolList          []bool    `protobuf:"varint,1,rep,packed,name=argBoolList,proto3" json:"argBoolList,omitempty"`
	ArgI32List           []int32   `protobuf:"varint,2,rep,packed,name=argI32List,proto3" json:"argI32List,omitempty"`
	ArgUI32List          []uint32  `protobuf:"varint,3,rep,packed,name=argUI32List,proto3" json:"argUI32List,omitempty"`
	ArgSI32List          []int32   `protobuf:"zigzag32,4,rep,packed,name=argSI32List,proto3" json:"argSI32List,omitempty"`
	ArgStrList           []string  `protobuf:"bytes,5,rep,name=argStrList,proto3" json:"argStrList,omitempty"`
	ArgByList            [][]byte  `protobuf:"bytes,6,rep,name=argByList,proto3" json:"argByList,omitempty"`
	ArgSimple            []*Simple `protobuf:"bytes,7,rep,name=argSimple,proto3" json:"argSimple,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *Repeat) Reset()         { *m = Repeat{} }
func (m *Repeat) String() string { return proto.CompactTextString(m) }
func (*Repeat) ProtoMessage()    {}
func (*Repeat) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{4}
}

func (m *Repeat) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Repeat.Unmarshal(m, b)
}
func (m *Repeat) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Repeat.Marshal(b, m, deterministic)
}
func (m *Repeat) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Repeat.Merge(m, src)
}
func (m *Repeat) XXX_Size() int {
	return xxx_messageInfo_Repeat.Size(m)
}
func (m *Repeat) XXX_DiscardUnknown() {
	xxx_messageInfo_Repeat.DiscardUnknown(m)
}

var xxx_messageInfo_Repeat proto.InternalMessageInfo

func (m *Repeat) GetArgBoolList() []bool {
	if m != nil {
		return m.ArgBoolList
	}
	return nil
}

func (m *Repeat) GetArgI32List() []int32 {
	if m != nil {
		return m.ArgI32List
	}
	return nil
}

func (m *Repeat) GetArgUI32List() []uint32 {
	if m != nil {
		return m.ArgUI32List
	}
	return nil
}

func (m *Repeat) GetArgSI32List() []int32 {
	if m != nil {
		return m.ArgSI32List
	}
	return nil
}

func (m *Repeat) GetArgStrList() []string {
	if m != nil {
		return m.ArgStrList
	}
	return nil
}

func (m *Repeat) GetArgByList() [][]byte {
	if m != nil {
		return m.ArgByList
	}
	return nil
}

func (m *Repeat) GetArgSimple() []*Simple {
	if m != nil {
		return m.ArgSimple
	}
	return nil
}

type LenPayload struct {
	ArgMap               map[string]int32 `protobuf:"bytes,1,rep,name=argMap,proto3" json:"argMap,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ArgStr               string           `protobuf:"bytes,2,opt,name=argStr,proto3" json:"argStr,omitempty"`
	ArgBytes             []byte           `protobuf:"bytes,3,opt,name=argBytes,proto3" json:"argBytes,omitempty"`
	ArgBit64             *Bit64           `protobuf:"bytes,4,opt,name=argBit64,proto3" json:"argBit64,omitempty"`
	ArgBit32             *Bit32           `protobuf:"bytes,5,opt,name=argBit32,proto3" json:"argBit32,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *LenPayload) Reset()         { *m = LenPayload{} }
func (m *LenPayload) String() string { return proto.CompactTextString(m) }
func (*LenPayload) ProtoMessage()    {}
func (*LenPayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{5}
}

func (m *LenPayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LenPayload.Unmarshal(m, b)
}
func (m *LenPayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LenPayload.Marshal(b, m, deterministic)
}
func (m *LenPayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LenPayload.Merge(m, src)
}
func (m *LenPayload) XXX_Size() int {
	return xxx_messageInfo_LenPayload.Size(m)
}
func (m *LenPayload) XXX_DiscardUnknown() {
	xxx_messageInfo_LenPayload.DiscardUnknown(m)
}

var xxx_messageInfo_LenPayload proto.InternalMessageInfo

func (m *LenPayload) GetArgMap() map[string]int32 {
	if m != nil {
		return m.ArgMap
	}
	return nil
}

func (m *LenPayload) GetArgStr() string {
	if m != nil {
		return m.ArgStr
	}
	return ""
}

func (m *LenPayload) GetArgBytes() []byte {
	if m != nil {
		return m.ArgBytes
	}
	return nil
}

func (m *LenPayload) GetArgBit64() *Bit64 {
	if m != nil {
		return m.ArgBit64
	}
	return nil
}

func (m *LenPayload) GetArgBit32() *Bit32 {
	if m != nil {
		return m.ArgBit32
	}
	return nil
}

func init() {
	proto.RegisterEnum("pb.AuctionType", AuctionType_name, AuctionType_value)
	proto.RegisterType((*VarintMsg)(nil), "pb.VarintMsg")
	proto.RegisterType((*Bit64)(nil), "pb.Bit64")
	proto.RegisterType((*Bit32)(nil), "pb.Bit32")
	proto.RegisterType((*Simple)(nil), "pb.Simple")
	proto.RegisterType((*Repeat)(nil), "pb.Repeat")
	proto.RegisterType((*LenPayload)(nil), "pb.LenPayload")
	proto.RegisterMapType((map[string]int32)(nil), "pb.LenPayload.ArgMapEntry")
}

func init() { proto.RegisterFile("protocal.proto", fileDescriptor_80e19dc549ac3013) }

var fileDescriptor_80e19dc549ac3013 = []byte{
	// 556 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x54, 0x4f, 0x6b, 0xdb, 0x30,
	0x14, 0x9f, 0xec, 0xda, 0x6d, 0xe4, 0xfe, 0x49, 0xc5, 0x18, 0xa6, 0x8c, 0x21, 0x0c, 0x03, 0x6d,
	0x87, 0x1c, 0xec, 0x10, 0xb6, 0xdd, 0x9a, 0x36, 0x85, 0x40, 0xbb, 0x95, 0xe7, 0x76, 0xec, 0x36,
	0x94, 0xd6, 0x18, 0x33, 0x37, 0x36, 0x8e, 0x33, 0xe6, 0x6f, 0xbc, 0x4f, 0xb0, 0xdb, 0x60, 0xe8,
	0x45, 0xb2, 0x35, 0xd8, 0x6e, 0xfe, 0xfd, 0x91, 0x9e, 0xde, 0xef, 0x49, 0xa6, 0xc7, 0x75, 0x53,
	0xb5, 0xd5, 0x83, 0x2c, 0x27, 0xf8, 0xc1, 0x9c, 0x7a, 0x15, 0xfd, 0x24, 0x74, 0xf4, 0x59, 0x36,
	0xc5, 0xba, 0xbd, 0xd9, 0xe4, 0xec, 0x05, 0xf5, 0x65, 0x93, 0x2f, 0x93, 0x38, 0x24, 0x9c, 0x08,
	0x0f, 0x34, 0x32, 0xfc, 0x6c, 0x1a, 0x3a, 0x9c, 0x08, 0x17, 0x34, 0x62, 0x21, 0xdd, 0x97, 0x4d,
	0x7e, 0xaf, 0x16, 0xb8, 0x9c, 0x88, 0x23, 0x30, 0xb0, 0x57, 0x66, 0xd3, 0x70, 0x8f, 0x13, 0xb1,
	0x07, 0x06, 0x6a, 0x25, 0x55, 0x6b, 0x3c, 0x4e, 0xc4, 0x29, 0x18, 0xd8, 0x2b, 0xb3, 0x69, 0xe8,
	0x73, 0x22, 0x18, 0x18, 0xa8, 0x95, 0x79, 0x55, 0x95, 0xe1, 0x3e, 0x77, 0xc5, 0x01, 0x18, 0xc8,
	0xde, 0xa0, 0xb2, 0x58, 0x6f, 0x9f, 0xc2, 0x03, 0x4e, 0xc4, 0x71, 0x7c, 0x32, 0xa9, 0x57, 0x93,
	0xf3, 0xed, 0x43, 0x5b, 0x54, 0xeb, 0xbb, 0xae, 0xce, 0xc0, 0xe8, 0x51, 0x4e, 0xbd, 0x79, 0xd1,
	0xce, 0xa6, 0xec, 0x15, 0xa5, 0xb2, 0xc9, 0xaf, 0x8a, 0x1f, 0xd9, 0xe3, 0x6c, 0x8a, 0x9d, 0xfa,
	0x60, 0x31, 0x8c, 0xd3, 0x40, 0x15, 0x36, 0x06, 0xd5, 0xf2, 0x18, 0x6c, 0x8a, 0xbd, 0xa4, 0x23,
	0xd9, 0xe4, 0x97, 0xd5, 0x76, 0x55, 0x66, 0xd8, 0x39, 0x81, 0x81, 0x88, 0x32, 0x2c, 0x94, 0xc4,
	0x76, 0x21, 0x1d, 0xe9, 0x3e, 0x58, 0xcc, 0x5f, 0x85, 0x92, 0x18, 0x0b, 0x9d, 0x80, 0x4d, 0xb1,
	0x33, 0x7a, 0xa0, 0xfc, 0x65, 0x25, 0x5b, 0xac, 0xe3, 0x40, 0x8f, 0xa3, 0x3b, 0xea, 0xa7, 0xc5,
	0x53, 0x5d, 0x66, 0xff, 0x1d, 0x9b, 0x35, 0x1e, 0xe7, 0x5f, 0xe3, 0xc1, 0x40, 0xd5, 0xb6, 0x43,
	0xa0, 0xd1, 0x2f, 0x42, 0x7d, 0xc8, 0xea, 0x4c, 0xb6, 0xfa, 0x78, 0x8a, 0xbd, 0x2e, 0x36, 0x6d,
	0x48, 0x30, 0x79, 0x9b, 0xd2, 0x0d, 0x2e, 0x93, 0x18, 0x0d, 0x0e, 0x77, 0x85, 0x07, 0x16, 0xa3,
	0x77, 0xb8, 0x37, 0x06, 0x97, 0xbb, 0xe2, 0x08, 0x6c, 0xca, 0x44, 0x60, 0x1c, 0x7b, 0xdc, 0x15,
	0xa7, 0x60, 0x53, 0xba, 0x46, 0xda, 0x36, 0x68, 0xf0, 0xb8, 0x2b, 0x46, 0x60, 0x31, 0x7a, 0x16,
	0xf3, 0x0e, 0x65, 0x9f, 0xbb, 0xe2, 0x10, 0x06, 0x82, 0x09, 0x54, 0x77, 0x39, 0xe1, 0xdd, 0x09,
	0x62, 0xaa, 0x6e, 0xc8, 0x8e, 0x81, 0x41, 0x8c, 0x7e, 0x13, 0x4a, 0xaf, 0xb3, 0xf5, 0xad, 0xec,
	0xca, 0x4a, 0x3e, 0xb2, 0x18, 0x33, 0xbd, 0x91, 0x35, 0xf6, 0x1d, 0xc4, 0x67, 0x6a, 0xd5, 0xa0,
	0x4f, 0xce, 0x51, 0x5c, 0xac, 0xdb, 0xa6, 0x03, 0xed, 0xd4, 0x73, 0x48, 0xdb, 0x06, 0xe3, 0x1e,
	0x81, 0x46, 0x7a, 0x8a, 0xf3, 0xae, 0xcd, 0x36, 0x18, 0xf7, 0x21, 0xf4, 0x98, 0xbd, 0xde, 0x69,
	0xea, 0x62, 0xe2, 0x4b, 0x09, 0xe2, 0x91, 0xaa, 0x84, 0x04, 0xf4, 0xd2, 0x60, 0xd3, 0xcf, 0x66,
	0xb0, 0x25, 0x31, 0xf4, 0xd2, 0xd9, 0x7b, 0x1a, 0x58, 0x07, 0x63, 0x63, 0xea, 0x7e, 0xcb, 0x3a,
	0xbc, 0x15, 0x23, 0x50, 0x9f, 0xec, 0x39, 0xf5, 0xbe, 0xcb, 0x72, 0x9b, 0xe1, 0x09, 0x3d, 0xd8,
	0x81, 0x0f, 0xce, 0x3b, 0xf2, 0xf6, 0x9c, 0x06, 0xd6, 0xb3, 0x61, 0x27, 0x34, 0xb8, 0x5a, 0x42,
	0x7a, 0xf7, 0xf5, 0x16, 0x96, 0x17, 0x8b, 0xf1, 0x33, 0x36, 0xa6, 0x87, 0xe9, 0xe2, 0xe2, 0xd3,
	0xc7, 0x4b, 0xcd, 0x90, 0x9d, 0xe5, 0xcb, 0xc2, 0x10, 0xce, 0xca, 0xc7, 0xff, 0x4a, 0xf2, 0x27,
	0x00, 0x00, 0xff, 0xff, 0xc1, 0x4f, 0xe3, 0x1b, 0x69, 0x04, 0x00, 0x00,
}
