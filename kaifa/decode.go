package kaifa

import (
	"encoding/binary"
	"fmt"
	"time"
)

func Unmarshal(data []byte) (*Message, error) {
	buf := NewBuffer(data)

	var err error
	m := &Message{}

	if err := buf.ReadRaw(&m.header.Length); err != nil {
		return m, err
	}
	b0 := uint8((m.header.Length & 0xF800) >> 8)
	m.header.Length = m.header.Length & 0x07FF
	m.header.Separator = (b0 & 0x08) > 0
	m.header.Format = b0 & 0xF0

	err = buf.ReadRaw(
		&m.header.DestAddr,
		&m.header.SrcAddr,
		&m.header.ControlField,
		&m.header.Checksum,
		&m.meta.LsapDest,
		&m.meta.LsapSrc,
		&m.meta.LlcQuality,
	)

	if err != nil {
		return m, err
	}
	m.meta.Meta = make([]byte, 5)
	if err := buf.ReadRaw(m.meta.Meta); err != nil {
		return m, err
	}

	// Read timestamp blob as LV Data
	var tsData []byte
	if err = buf.ReadType(&tsData); err != nil {
		return m, err
	}
	if m.Timestamp, err = parseTimestamp(tsData); err != nil {
		return m, err
	}

	// Variables that controls what to look for based on number of items
	// the meter reports.
	energy := false
	fullHdr := true
	phases := 0

	var itemCount uint8
	if err := buf.ReadType(&itemCount); err != nil {
		return m, err
	}

	switch itemCount {
	case 0:
		// No items
		return m, nil
	case 1:
		// For a single item only the Active Power imported from the grid is reported
		if err := buf.ReadType(m.ActivePowerPositive); err != nil {
			return m, err
		}

		// Skip reading the full header
		fullHdr = false
	case 9:
		// 9 items = Single phase, no accumulated energy
		phases = 1
	case 13:
		// 13 items = Three phases, no accumulated energy
		phases = 3
	case 14:
		// 14 items = Single phase with accumulated energy
		phases = 1
		energy = true
	case 18:
		// 18 items = Three phases with accumulated energy
		phases = 3
		energy = true
	default:
		return m, fmt.Errorf("unsupported number of items: %d", itemCount)
	}

	if fullHdr {
		m.Version = new(string)
		m.MeterID = new(string)
		m.MeterType = new(string)
		m.ActivePowerPositive = new(int32)
		m.ActivePowerNegative = new(int32)
		m.ReactivePowerPositive = new(int32)
		m.ReactivePowerNegative = new(int32)

		err := buf.ReadType(
			m.Version,
			m.MeterID,
			m.MeterType,
			m.ActivePowerPositive,
			m.ActivePowerNegative,
			m.ReactivePowerPositive,
			m.ReactivePowerNegative,
		)
		if err != nil {
			return m, err
		}
	}

	if phases > 0 {
		m.Phases = make([]Phase, phases)

		// First is current (Amperes) for each phase
		for i := 0; i < phases; i++ {
			var cur int32
			if err := buf.ReadType(&cur); err != nil {
				return m, err
			}
			m.Phases[i].Index = i + 1
			m.Phases[i].Current = float64(cur) / 1000
		}

		// Then the voltage for each phase
		for i := 0; i < phases; i++ {
			var voltage int32
			if err := buf.ReadType(&voltage); err != nil {
				return m, err
			}
			m.Phases[i].Voltage = float64(voltage) / 10
		}
	}
	if energy {
		var tsData []byte
		if err := buf.ReadType(&tsData); err != nil {
			return m, err
		}
		if ts, err := parseTimestamp(tsData); err != nil {
			return m, err
		} else {
			m.EnergyTimestamp = &ts
		}

		m.ActiveEnergyPositive = new(int32)
		m.ActiveEnergyNegative = new(int32)
		m.ReactiveEnergyPositive = new(int32)
		m.ReactiveEnergyNegative = new(int32)
		err := buf.ReadType(
			m.ActiveEnergyPositive,
			m.ActiveEnergyNegative,
			m.ReactiveEnergyPositive,
			m.ReactiveEnergyNegative,
		)
		if err != nil {
			return m, err
		}
	}

	if err := buf.ReadRaw(&m.checksum); err != nil {
		return m, err
	}

	if buf.Len() > 0 {
		return m, fmt.Errorf("trailing data: %X", data)
	}

	return m, nil

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
