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
	ArgEnum              AuctionType `protobuf:"varint,7,opt,name=argEnum,proto3,enum=pb.AuctionType" json:"argEnum,omitempty"`
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

type Map struct {
	ArgII                map[int32]int32   `protobuf:"bytes,1,rep,name=argII,proto3" json:"argII,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ArgUI                map[uint32]uint32 `protobuf:"bytes,2,rep,name=argUI,proto3" json:"argUI,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ArgSS                map[string]string `protobuf:"bytes,3,rep,name=argSS,proto3" json:"argSS,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ArgSU                map[string]uint32 `protobuf:"bytes,4,rep,name=argSU,proto3" json:"argSU,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Map) Reset()         { *m = Map{} }
func (m *Map) String() string { return proto.CompactTextString(m) }
func (*Map) ProtoMessage()    {}
func (*Map) Descriptor() ([]byte, []int) {
	return fileDescriptor_80e19dc549ac3013, []int{5}
}

func (m *Map) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Map.Unmarshal(m, b)
}
func (m *Map) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Map.Marshal(b, m, deterministic)
}
func (m *Map) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Map.Merge(m, src)
}
func (m *Map) XXX_Size() int {
	return xxx_messageInfo_Map.Size(m)
}
func (m *Map) XXX_DiscardUnknown() {
	xxx_messageInfo_Map.DiscardUnknown(m)
}

var xxx_messageInfo_Map proto.InternalMessageInfo

func (m *Map) GetArgII() map[int32]int32 {
	if m != nil {
		return m.ArgII
	}
	return nil
}

func (m *Map) GetArgUI() map[uint32]uint32 {
	if m != nil {
		return m.ArgUI
	}
	return nil
}

func (m *Map) GetArgSS() map[string]string {
	if m != nil {
		return m.ArgSS
	}
	return nil
}

func (m *Map) GetArgSU() map[string]uint32 {
	if m != nil {
		return m.ArgSU
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
	return fileDescriptor_80e19dc549ac3013, []int{6}
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
	proto.RegisterType((*Map)(nil), "pb.Map")
	proto.RegisterMapType((map[int32]int32)(nil), "pb.Map.ArgIIEntry")
	proto.RegisterMapType((map[string]string)(nil), "pb.Map.ArgSSEntry")
	proto.RegisterMapType((map[string]uint32)(nil), "pb.Map.ArgSUEntry")
	proto.RegisterMapType((map[uint32]uint32)(nil), "pb.Map.ArgUIEntry")
	proto.RegisterType((*LenPayload)(nil), "pb.LenPayload")
	proto.RegisterMapType((map[string]int32)(nil), "pb.LenPayload.ArgMapEntry")
}

func init() { proto.RegisterFile("protocal.proto", fileDescriptor_80e19dc549ac3013) }

var fileDescriptor_80e19dc549ac3013 = []byte{
	// 649 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0xdf, 0x6a, 0xdb, 0x3c,
	0x14, 0xff, 0x64, 0xd5, 0x4e, 0x23, 0x37, 0x6d, 0x2a, 0x3e, 0x86, 0x09, 0x63, 0x08, 0xc3, 0x40,
	0xdb, 0x45, 0x2e, 0x9c, 0x10, 0xba, 0xdd, 0x35, 0x6d, 0x0a, 0x81, 0x66, 0x2b, 0xc7, 0xf5, 0xd8,
	0xdd, 0x50, 0x5a, 0x13, 0xc2, 0xdc, 0xd8, 0xb8, 0xce, 0x58, 0x1e, 0x64, 0xaf, 0xb5, 0xc7, 0xd9,
	0xdd, 0x60, 0x48, 0x96, 0x62, 0x65, 0xcd, 0xd8, 0xee, 0xfc, 0xfb, 0x73, 0x74, 0x74, 0xce, 0xb1,
	0x0e, 0x39, 0x2e, 0xca, 0xbc, 0xca, 0xef, 0x44, 0xd6, 0x57, 0x1f, 0xd4, 0x29, 0xe6, 0xe1, 0x77,
	0x44, 0xda, 0x1f, 0x44, 0xb9, 0x5c, 0x55, 0xb3, 0xc7, 0x05, 0x7d, 0x46, 0x3c, 0x51, 0x2e, 0xa6,
	0x83, 0x28, 0x40, 0x0c, 0x71, 0x17, 0x34, 0x32, 0xfc, 0x68, 0x18, 0x38, 0x0c, 0x71, 0x0c, 0x1a,
	0xd1, 0x80, 0xb4, 0x44, 0xb9, 0x48, 0x64, 0x00, 0x66, 0x88, 0x77, 0xc0, 0xc0, 0xad, 0x32, 0x1a,
	0x06, 0x07, 0x0c, 0xf1, 0x03, 0x30, 0x50, 0x2b, 0xb1, 0x8c, 0x71, 0x19, 0xe2, 0xa7, 0x60, 0xe0,
	0x56, 0x19, 0x0d, 0x03, 0x8f, 0x21, 0x4e, 0xc1, 0x40, 0xfa, 0x4a, 0x29, 0x93, 0xd5, 0xfa, 0x21,
	0x68, 0x31, 0xc4, 0x8f, 0xa3, 0x93, 0x7e, 0x31, 0xef, 0x9f, 0xaf, 0xef, 0xaa, 0x65, 0xbe, 0xba,
	0xdd, 0x14, 0x29, 0x18, 0x3d, 0x5c, 0x10, 0x77, 0xbc, 0xac, 0x46, 0x43, 0xfa, 0x82, 0x10, 0x51,
	0x2e, 0xae, 0x96, 0x5f, 0xd3, 0xfb, 0xd1, 0x50, 0xd5, 0xe3, 0x81, 0xc5, 0x50, 0x46, 0x7c, 0x79,
	0xbc, 0x31, 0xc8, 0xc2, 0xba, 0x60, 0x53, 0xf4, 0x39, 0x69, 0x8b, 0x72, 0x71, 0x99, 0xaf, 0xe7,
	0x59, 0xaa, 0xea, 0x43, 0xd0, 0x10, 0x61, 0xaa, 0x12, 0x0d, 0x22, 0x3b, 0x91, 0x6e, 0x5c, 0x0b,
	0x2c, 0x66, 0x27, 0xd1, 0x20, 0x52, 0x89, 0x4e, 0xc0, 0xa6, 0x68, 0x8f, 0x1c, 0x4a, 0x7f, 0x96,
	0x8b, 0x4a, 0xe5, 0x71, 0x60, 0x8b, 0xc3, 0x5b, 0xe2, 0xc5, 0xcb, 0x87, 0x22, 0x4b, 0xff, 0x38,
	0x1c, 0x6b, 0x08, 0xce, 0xbe, 0x21, 0x8c, 0xf3, 0x3c, 0x53, 0xc7, 0x1e, 0x82, 0x81, 0xe1, 0x0f,
	0x44, 0x3c, 0x48, 0x8b, 0x54, 0x54, 0xfa, 0x7a, 0x92, 0xbd, 0x5e, 0x3e, 0x56, 0x01, 0x62, 0x98,
	0x1f, 0x82, 0x4d, 0xe9, 0x02, 0xa7, 0x83, 0x48, 0x19, 0x1c, 0x86, 0xb9, 0x0b, 0x16, 0xa3, 0x4f,
	0x48, 0x8c, 0x01, 0x33, 0xcc, 0x3b, 0x60, 0x53, 0xa6, 0x05, 0xc6, 0x71, 0xc0, 0x30, 0x3f, 0x05,
	0x9b, 0xd2, 0x39, 0xe2, 0xaa, 0x54, 0x06, 0x97, 0x61, 0xde, 0x06, 0x8b, 0xd1, 0xb3, 0x18, 0x6f,
	0x94, 0xec, 0x31, 0xcc, 0x8f, 0xa0, 0x21, 0x28, 0x57, 0x6a, 0xdd, 0xa7, 0xa0, 0xc5, 0x30, 0xf7,
	0x23, 0x22, 0xff, 0x90, 0x9a, 0x81, 0x46, 0x0c, 0xbf, 0x61, 0x82, 0x67, 0xa2, 0xa0, 0x9c, 0xb8,
	0xb2, 0x82, 0xa9, 0xaa, 0xd7, 0x8f, 0xa8, 0x74, 0xcf, 0x44, 0xd1, 0x3f, 0x97, 0xe4, 0x64, 0x55,
	0x95, 0x1b, 0xa8, 0x0d, 0xda, 0x99, 0x4c, 0x55, 0xe1, 0xbb, 0xce, 0xc4, 0x72, 0x26, 0xc6, 0x19,
	0xc7, 0xaa, 0x03, 0xbb, 0xce, 0x38, 0x6e, 0x9c, 0x71, 0x6c, 0x9c, 0x89, 0xea, 0xc4, 0x6f, 0xce,
	0xc4, 0x72, 0x26, 0xbd, 0x33, 0x42, 0x9a, 0x2b, 0xd1, 0x2e, 0xc1, 0x9f, 0xd3, 0x8d, 0x9e, 0xbf,
	0xfc, 0xa4, 0xff, 0x13, 0xf7, 0x8b, 0xc8, 0xd6, 0xa9, 0x1a, 0xbd, 0x0b, 0x35, 0x78, 0xeb, 0x9c,
	0x21, 0x1d, 0x99, 0x3c, 0x8d, 0xec, 0xec, 0x89, 0xec, 0x3c, 0x8d, 0xd4, 0x57, 0xb6, 0x23, 0xdb,
	0x7b, 0x22, 0xdb, 0x7b, 0x22, 0x93, 0x7f, 0x8a, 0xb4, 0x73, 0x86, 0x3f, 0x11, 0x21, 0xd7, 0xe9,
	0xea, 0x46, 0x6c, 0xb2, 0x5c, 0xdc, 0xd3, 0x48, 0xfd, 0xeb, 0x33, 0x51, 0xe8, 0xf9, 0xf4, 0x64,
	0x87, 0x1a, 0x5d, 0x36, 0x6a, 0x26, 0x8a, 0xba, 0x53, 0xda, 0xa9, 0xdf, 0x47, 0x5c, 0x95, 0xfa,
	0x5e, 0x1a, 0xe9, 0xd7, 0x35, 0xde, 0x54, 0xe9, 0xa3, 0x7a, 0x06, 0x47, 0xb0, 0xc5, 0xf4, 0x65,
	0xad, 0xc9, 0x85, 0xa1, 0xf6, 0x94, 0x1f, 0xb5, 0x65, 0x26, 0x45, 0xc0, 0x56, 0x6a, 0x6c, 0x7a,
	0x69, 0x35, 0xb6, 0x41, 0x04, 0x5b, 0xa9, 0xf7, 0x86, 0xf8, 0xd6, 0xc5, 0xfe, 0x56, 0xbf, 0x3d,
	0xad, 0xd7, 0xe7, 0xc4, 0xb7, 0xd6, 0x19, 0x3d, 0x21, 0xfe, 0xd5, 0x14, 0xe2, 0xdb, 0x4f, 0x37,
	0x30, 0xbd, 0x98, 0x74, 0xff, 0xa3, 0x5d, 0x72, 0x14, 0x4f, 0x2e, 0xde, 0xbf, 0xbb, 0xd4, 0x0c,
	0xaa, 0x2d, 0x1f, 0x27, 0x86, 0x70, 0xe6, 0x9e, 0xda, 0xea, 0x83, 0x5f, 0x01, 0x00, 0x00, 0xff,
	0xff, 0xd9, 0x44, 0x1a, 0x4f, 0xe7, 0x05, 0x00, 0x00,
}