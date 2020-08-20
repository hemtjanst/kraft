package kaifa

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

const (
	frameTag uint8 = 0x7e
	// 16 bit frame format field = "0101SLLL_LLLLLLLL"
	frameFormatMask uint8 = 0xF0
	frameFormat     uint8 = 0xA0
	frameLengthMask uint8 = 0b111
)

type Phase struct {
	Index   int
	Current float64
	Voltage float64
}

type Message struct {
	header                 Header
	meta                   Meta
	checksum               uint16
	Timestamp              time.Time
	Version                *string    `json:",omitempty"`
	MeterID                *string    `json:",omitempty"`
	MeterType              *string    `json:",omitempty"`
	ActivePowerPositive    *int       `json:",omitempty"`
	ActivePowerNegative    *int       `json:",omitempty"`
	ReactivePowerPositive  *int       `json:",omitempty"`
	ReactivePowerNegative  *int       `json:",omitempty"`
	Phases                 []Phase    `json:",omitempty"`
	EnergyTimestamp        *time.Time `json:",omitempty"`
	ActiveEnergyPositive   *int       `json:",omitempty"`
	ActiveEnergyNegative   *int       `json:",omitempty"`
	ReactiveEnergyPositive *int       `json:",omitempty"`
	ReactiveEnergyNegative *int       `json:",omitempty"`
}

type Header struct {
	Format       uint8
	Separator    bool
	Length       uint16
	DestAddr     uint8
	SrcAddr      uint16
	ControlField uint8
	Checksum     uint16
}

type Meta struct {
	LsapDest   uint8
	LsapSrc    uint8
	LlcQuality uint8
	Meta       []byte
}

func (h *Header) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return data, errors.New("too few bytes (header)")
	}
	h.Format = data[0] & frameFormatMask
	h.Separator = (data[0] & ^(frameFormatMask | frameLengthMask)) > 0
	h.Length = (uint16(data[0]&frameLengthMask) << 8) | uint16(data[1])
	h.DestAddr = data[2]
	h.SrcAddr = (uint16(data[3]) << 8) | uint16(data[4])
	h.ControlField = data[5]
	h.Checksum = (uint16(data[6]) << 8) | uint16(data[7])
	return data[8:], nil
}

func (m *Meta) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return data, errors.New("too few bytes (meta)")
	}
	m.LsapDest = data[0]
	m.LsapSrc = data[1]
	m.LlcQuality = data[2]
	m.Meta = data[3:8]
	return data[8:], nil
}

func unmarshalLvString(data []byte) (string, []byte, error) {
	b, data, err := unmarshalLvData(data)
	return string(b), data, err
}

func unmarshalLvData(data []byte) ([]byte, []byte, error) {
	if len(data) < 2 {
		return nil, nil, errors.New("too few bytes (lvdata)")
	}
	if data[0] != 0x09 {
		return nil, nil, fmt.Errorf("lvdata: wrong item type, expected 0x09, got 0x%02X", data[0])
	}
	ln := int(data[1])
	data = data[2:]
	if ln == 0 {
		return nil, data, nil
	}
	if ln > len(data) {
		return nil, nil, fmt.Errorf("lvdata: length exceeds frame size: %d, remaining: %X", ln, data)
	}

	return data[:ln], data[ln:], nil

}

func unmarshalUint8(data []byte) (int, []byte, error) {
	if len(data) < 2 {
		return 0, nil, errors.New("too few bytes (uint8)")
	}
	if data[0] != 0x02 {
		return 0, nil, fmt.Errorf("uint8: wrong item type, expected 0x02, got 0x%02X", data[0])
	}
	return int(data[1]), data[2:], nil
}

func unmarshalInt32(data []byte) (int, []byte, error) {
	if len(data) < 5 {
		return 0, nil, errors.New("too few bytes (uint8)")
	}
	if data[0] != 0x06 {
		return 0, nil, fmt.Errorf("uint8: wrong item type, expected 0x02, got 0x%02X", data[0])
	}
	return int(int32(binary.BigEndian.Uint32(data[1:5]))), data[5:], nil
}

func parseTimestamp(data []byte) (time.Time, error) {
	if len(data) < 8 {
		return time.Time{}, fmt.Errorf("too few bytes (timestamp)")
	}
	ts := time.Date(
		int(binary.BigEndian.Uint16(data[0:2])), // year
		time.Month(data[2]),                     // month
		int(data[3]),                            // day
		int(data[5]),                            // hour
		int(data[6]),                            // minute
		int(data[7]),                            // second
		0,                                       // nsec
		time.Local,
	)
	return ts, nil
}
