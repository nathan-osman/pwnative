package protocol

import (
	"encoding/binary"
	"testing"

	"github.com/nitroshare/compare"
)

const (
	fakeID          = 1
	fakeOpcode      = 2
	fakeSize        = 3
	fakeSeq         = 4
	fakeNumFileDesc = 5
)

var (
	headerBigEndian = Header{
		0, 0, 0, 1,
		2, 0, 0, 3,
		0, 0, 0, 4,
		0, 0, 0, 5,
	}
	headerLittleEndian = Header{
		1, 0, 0, 0,
		2, 3, 0, 0,
		4, 0, 0, 0,
		5, 0, 0, 0,
	}
)

func TestEncode(t *testing.T) {
	for _, v := range []struct {
		name           string
		isLittleEndian bool
		output         Header
	}{
		{
			name:           "big endian",
			isLittleEndian: false,
			output:         headerBigEndian,
		},
		{
			name:           "little endian",
			isLittleEndian: true,
			output:         headerLittleEndian,
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			isLittleEndian = v.isLittleEndian
			if isLittleEndian {
				nativeEndian = binary.LittleEndian
			} else {
				nativeEndian = binary.BigEndian
			}
			h := &Header{}
			h.SetID(fakeID)
			h.SetOpcode(fakeOpcode)
			h.SetSize(fakeSize)
			h.SetSeq(fakeSeq)
			h.SetNumFileDesc(fakeNumFileDesc)
			compare.Compare(t, *h, v.output, true)
		})
	}
}

func TestDecode(t *testing.T) {
	for _, v := range []struct {
		name           string
		isLittleEndian bool
		input          Header
	}{
		{
			name:           "big endian",
			isLittleEndian: false,
			input:          headerBigEndian,
		},
		{
			name:           "little endian",
			isLittleEndian: true,
			input:          headerLittleEndian,
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			isLittleEndian = v.isLittleEndian
			if isLittleEndian {
				nativeEndian = binary.LittleEndian
			} else {
				nativeEndian = binary.BigEndian
			}
			compare.Compare(t, v.input.ID(), fakeID, true)
			compare.Compare(t, v.input.Opcode(), fakeOpcode, true)
			compare.Compare(t, v.input.Size(), fakeSize, true)
			compare.Compare(t, v.input.Seq(), fakeSeq, true)
			compare.Compare(t, v.input.NumFileDesc(), fakeNumFileDesc, true)
		})
	}
}
