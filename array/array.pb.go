// Code generated by protoc-gen-go. DO NOT EDIT.
// source: array.proto

package array

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Array32 struct {
	// compatibility guarantee:
	//     reserved field number: 1, 2, 3, 4
	//     reserved field name: Cnt, Bitmaps, Offsets, Elts
	//
	Cnt     int32    `protobuf:"varint,1,opt,name=Cnt,proto3" json:"Cnt,omitempty"`
	Bitmaps []uint64 `protobuf:"varint,2,rep,packed,name=Bitmaps,proto3" json:"Bitmaps,omitempty"`
	Offsets []int32  `protobuf:"varint,3,rep,packed,name=Offsets,proto3" json:"Offsets,omitempty"`
	Elts    []byte   `protobuf:"bytes,4,opt,name=Elts,proto3" json:"Elts,omitempty"`
	// Flags provides options
	//
	// Since 0.5.4
	Flags uint32 `protobuf:"varint,10,opt,name=Flags,proto3" json:"Flags,omitempty"`
	// EltWidth set width of elt in bits.
	//
	// Since 0.5.4
	EltWidth int32 `protobuf:"varint,20,opt,name=EltWidth,proto3" json:"EltWidth,omitempty"`
	// BMElts is optimized for elt itself is a bitmap.
	//
	// Since 0.5.4
	BMElts *Bits `protobuf:"bytes,30,opt,name=BMElts,proto3" json:"BMElts,omitempty"`
	// EltIndex is a bitmap in which a "1" indicate element start.
	// To get the start position of i-th element:
	//     s := EltIndex.Select(i)
	//
	// To get the end position of i-th element:
	//     e := EltIndex.Select(i+1)
	//
	// Since 0.5.10
	EltIndex             *Bits    `protobuf:"bytes,40,opt,name=EltIndex,proto3" json:"EltIndex,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Array32) Reset()         { *m = Array32{} }
func (m *Array32) String() string { return proto.CompactTextString(m) }
func (*Array32) ProtoMessage()    {}
func (*Array32) Descriptor() ([]byte, []int) {
	return fileDescriptor_array_3c19ab00b3340202, []int{0}
}
func (m *Array32) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Array32.Unmarshal(m, b)
}
func (m *Array32) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Array32.Marshal(b, m, deterministic)
}
func (dst *Array32) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Array32.Merge(dst, src)
}
func (m *Array32) XXX_Size() int {
	return xxx_messageInfo_Array32.Size(m)
}
func (m *Array32) XXX_DiscardUnknown() {
	xxx_messageInfo_Array32.DiscardUnknown(m)
}

var xxx_messageInfo_Array32 proto.InternalMessageInfo

func (m *Array32) GetCnt() int32 {
	if m != nil {
		return m.Cnt
	}
	return 0
}

func (m *Array32) GetBitmaps() []uint64 {
	if m != nil {
		return m.Bitmaps
	}
	return nil
}

func (m *Array32) GetOffsets() []int32 {
	if m != nil {
		return m.Offsets
	}
	return nil
}

func (m *Array32) GetElts() []byte {
	if m != nil {
		return m.Elts
	}
	return nil
}

func (m *Array32) GetFlags() uint32 {
	if m != nil {
		return m.Flags
	}
	return 0
}

func (m *Array32) GetEltWidth() int32 {
	if m != nil {
		return m.EltWidth
	}
	return 0
}

func (m *Array32) GetBMElts() *Bits {
	if m != nil {
		return m.BMElts
	}
	return nil
}

func (m *Array32) GetEltIndex() *Bits {
	if m != nil {
		return m.EltIndex
	}
	return nil
}

func init() {
	proto.RegisterType((*Array32)(nil), "Array32")
}

func init() { proto.RegisterFile("array.proto", fileDescriptor_array_3c19ab00b3340202) }

var fileDescriptor_array_3c19ab00b3340202 = []byte{
	// 209 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4e, 0x2c, 0x2a, 0x4a,
	0xac, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x97, 0xe2, 0x49, 0xca, 0x2c, 0xc9, 0x4d, 0x2c, 0x80,
	0xf0, 0x94, 0xae, 0x33, 0x72, 0xb1, 0x3b, 0x82, 0x64, 0x8d, 0x8d, 0x84, 0x04, 0xb8, 0x98, 0x9d,
	0xf3, 0x4a, 0x24, 0x18, 0x15, 0x18, 0x35, 0x58, 0x83, 0x40, 0x4c, 0x21, 0x09, 0x2e, 0x76, 0x27,
	0xb0, 0xea, 0x62, 0x09, 0x26, 0x05, 0x66, 0x0d, 0x96, 0x20, 0x18, 0x17, 0x24, 0xe3, 0x9f, 0x96,
	0x56, 0x9c, 0x5a, 0x52, 0x2c, 0xc1, 0xac, 0xc0, 0xac, 0xc1, 0x1a, 0x04, 0xe3, 0x0a, 0x09, 0x71,
	0xb1, 0xb8, 0xe6, 0x94, 0x14, 0x4b, 0xb0, 0x28, 0x30, 0x6a, 0xf0, 0x04, 0x81, 0xd9, 0x42, 0x22,
	0x5c, 0xac, 0x6e, 0x39, 0x89, 0xe9, 0xc5, 0x12, 0x5c, 0x0a, 0x8c, 0x1a, 0xbc, 0x41, 0x10, 0x8e,
	0x90, 0x14, 0x17, 0x87, 0x6b, 0x4e, 0x49, 0x78, 0x66, 0x4a, 0x49, 0x86, 0x84, 0x08, 0xd8, 0x52,
	0x38, 0x5f, 0x48, 0x96, 0x8b, 0xcd, 0xc9, 0x17, 0x6c, 0x8e, 0x9c, 0x02, 0xa3, 0x06, 0xb7, 0x11,
	0xab, 0x9e, 0x53, 0x66, 0x49, 0x71, 0x10, 0x54, 0x50, 0x48, 0x11, 0xac, 0xd5, 0x33, 0x2f, 0x25,
	0xb5, 0x42, 0x42, 0x03, 0x59, 0x01, 0x5c, 0xd8, 0x89, 0x3d, 0x8a, 0x15, 0xec, 0xed, 0x24, 0x36,
	0xb0, 0x4f, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xd7, 0xb1, 0x28, 0xbc, 0x06, 0x01, 0x00,
	0x00,
}
