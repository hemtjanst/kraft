package kaifa

import (
	"encoding/binary"
	"io"
)

type Buffer []byte

func NewBuffer(data []byte) *Buffer {
	buf := Buffer(data)
	return &buf
}

var (
	order = binary.BigEndian
)

func (b *Buffer) Len() int {
	return len(*b)
}

func (b *Buffer) ReadRaw(tv ...interface{}) error {
	for _, t := range tv {
		switch t.(type) {
		case *uint8:
			if len(*b) < 1 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*uint8)) = (*b)[0]
			*b = (*b)[1:]
		case *int8:
			if len(*b) < 1 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*int8)) = int8((*b)[0])
			*b = (*b)[1:]
		case *uint16:
			if len(*b) < 2 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*uint16)) = order.Uint16((*b)[0:2])
			*b = (*b)[2:]
		case *int16:
			if len(*b) < 2 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*int16)) = int16(order.Uint16((*b)[0:2]))
			*b = (*b)[2:]
		case *uint32:
			if len(*b) < 4 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*uint32)) = order.Uint32((*b)[0:4])
			*b = (*b)[4:]
		case *int32:
			if len(*b) < 4 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*int32)) = int32(order.Uint32((*b)[0:4]))
			*b = (*b)[4:]
		case *uint64:
			if len(*b) < 8 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*uint64)) = order.Uint64((*b)[0:8])
			*b = (*b)[8:]
		case *int64:
			if len(*b) < 8 {
				return io.ErrUnexpectedEOF
			}
			*(t.(*int64)) = int64(order.Uint64((*b)[0:8]))
			*b = (*b)[8:]
		case []byte:
			ln := len(t.([]byte))
			if len(*b) < ln {
				return io.ErrUnexpectedEOF
			}
			for i := 0; i < ln; i++ {
				t.([]byte)[i] = (*b)[i]
			}
			*b = (*b)[ln:]
		default:
			return ErrUnsupportedtype
		}
	}
	return nil
}

func (b *Buffer) ReadType(tv ...interface{}) error {
	for _, t := range tv {
		var typ uint8
		if err := b.ReadRaw(&typ); err != nil {
			return err
		}
		switch t.(type) {
		case *uint8, *int8:
			if typ != 0x02 {
				return ErrWrongType
			}
			if err := b.ReadRaw(t); err != nil {
				return err
			}
		case *uint32, *int32:
			if typ != 0x06 {
				return ErrWrongType
			}
			if err := b.ReadRaw(t); err != nil {
				return err
			}
		case *[]byte:
			var ln uint8
			if err := b.ReadRaw(&ln); err != nil {
				return err
			}
			dat := make([]byte, int(ln))
			if err := b.ReadRaw(dat); err != nil {
				return err
			}
			*(t.(*[]byte)) = dat
		case *string:
			var ln uint8
			if err := b.ReadRaw(&ln); err != nil {
				return err
			}
			dat := make([]byte, int(ln))
			if err := b.ReadRaw(dat); err != nil {
				return err
			}
			*(t.(*string)) = string(dat)
		default:
			return ErrUnsupportedtype
		}
	}
	return nil
}
