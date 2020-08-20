package kaifa

import (
	"errors"
	"time"
)

const (
	// Each frame starts and ends with 0x7E
	frameTag uint8 = 0x7e

	// Half of next byte after frameTag is the frame format
	frameFormatMask uint8 = 0xF0
	frameFormat     uint8 = 0xA0

	// The last three bits of same byte contains the upper bits
	// of the frame length
	frameLengthMask uint8 = 0b111
)

type Phase struct {
	Index   int
	Current float64
	Voltage float64
}

type Message struct {
	header    Header
	meta      Meta
	checksum  uint16
	Timestamp time.Time
	// Version of the MBus protocol
	Version *string `json:",omitempty"`
	// MeterID is the serial number
	MeterID *string `json:",omitempty"`
	// MeterType is the model number
	MeterType *string `json:",omitempty"`
	// ActivePowerPositive is the amount of power currently drawn from the grid
	ActivePowerPositive *int `json:",omitempty"`
	// ActivePowerNegative is the amount of power currently exported to the grid
	ActivePowerNegative   *int `json:",omitempty"`
	ReactivePowerPositive *int `json:",omitempty"`
	ReactivePowerNegative *int `json:",omitempty"`
	// Phases contains per-phase information. Normally there are 0, 1 or 3 phases present
	Phases []Phase `json:",omitempty"`
	// EnergyTimestamp is the timestamp at which the Energy values below where last read
	EnergyTimestamp *time.Time `json:",omitempty"`
	// ActiveEnergyPositive is the accumulated amount of energy drawn from the grid (in Wh)
	ActiveEnergyPositive *int `json:",omitempty"`
	// ActiveEnergyNegative is the accumulated amount of energy exported to the grid (in Wh)
	ActiveEnergyNegative   *int `json:",omitempty"`
	ReactiveEnergyPositive *int `json:",omitempty"`
	ReactiveEnergyNegative *int `json:",omitempty"`
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
