package protocol

import (
	"encoding/binary"
)

// Detect platform endianness

var (
	isLittleEndian                  = false
	nativeEndian   binary.ByteOrder = binary.BigEndian
)

func init() {
	var (
		b        = make([]byte, 2)
		x uint16 = 1
	)
	binary.NativeEndian.PutUint16(b, x)
	if b[0] == 1 {
		isLittleEndian = true
		nativeEndian = binary.LittleEndian
	}
}

// Header stores the message header, which consists of five fields packed into
// sixteen bytes. Each field has a corresponding method for encoding and
// decoding the value in the header.
type Header [16]byte

// ID decodes the proxy / resource ID.
func (h Header) ID() uint32 {
	return nativeEndian.Uint32(h[:4])
}

// SetID encodes the proxy / resource ID.
func (h *Header) SetID(id uint32) {
	nativeEndian.PutUint32(h[0:4], id)
}

// Opcode decodes the opcode of the message.
func (h Header) Opcode() uint8 {
	return h[4]
}

// SetOpcode encodes the opcode of the message.
func (h *Header) SetOpcode(opcode uint8) {
	h[4] = opcode
}

// Size decodes the payload size.
func (h Header) Size() uint32 {
	b := [4]byte{}
	if isLittleEndian {
		copy(b[:3], h[5:8])
	} else {
		copy(b[1:], h[5:8])
	}
	return nativeEndian.Uint32(b[:])
}

// SetSize encodes the payload size.
func (h *Header) SetSize(size uint32) {
	b := [4]byte{}
	nativeEndian.PutUint32(b[:], size)
	if isLittleEndian {
		copy(h[5:8], b[:3])
	} else {
		copy(h[5:8], b[1:])
	}
}

// Seq decodes the message sequence number.
func (h Header) Seq() uint32 {
	return nativeEndian.Uint32(h[8:12])
}

// SetSeq encodes the message sequence number.
func (h *Header) SetSeq(seq uint32) {
	nativeEndian.PutUint32(h[8:12], seq)
}

// NumFileDesc decodes the number of file descriptors in the message.
func (h Header) NumFileDesc() uint32 {
	return nativeEndian.Uint32(h[12:16])
}

// SetNumFileDesc encodes the number of file descriptors in the message.
func (h *Header) SetNumFileDesc(n uint32) {
	nativeEndian.PutUint32(h[12:16], n)
}
