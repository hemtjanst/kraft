package kaifa

import (
	"encoding/binary"
	"fmt"
	"time"
)

func Unmarshal(data []byte) (*Message, error) {

	var err error
	m := &Message{}

	// Read header and metadata. Currently not used or exported
	if data, err = m.header.Unmarshal(data); err != nil {
		return m, err
	}
	if data, err = m.meta.Unmarshal(data); err != nil {
		return m, err
	}

	// Read timestamp blob as LV Data
	var tsData []byte
	if tsData, data, err = unmarshalLvData(data); err != nil {
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

	var itemCount int

	if itemCount, data, err = unmarshalUint8(data); err != nil {
		return m, err
	}

	switch itemCount {
	case 0:
		// No items
		return m, nil
	case 1:
		// For a single item only the Active Power imported from the grid is reported
		activePowerPositive, _, err := unmarshalInt32(data)
		if err != nil {
			return m, err
		}
		m.ActivePowerPositive = &activePowerPositive

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
		// Version, Serial number and Model is read as LV strings
		var version, meterID, meterType string
		if version, data, err = unmarshalLvString(data); err != nil {
			return m, err
		}
		m.Version = &version
		if meterID, data, err = unmarshalLvString(data); err != nil {
			return m, err
		}
		m.MeterID = &meterID
		if meterType, data, err = unmarshalLvString(data); err != nil {
			return m, err
		}
		m.MeterType = &meterType

		// Active and reactive power values
		var actPos, actNeg, reactPos, reactNeg int
		if actPos, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ActivePowerPositive = &actPos
		if actNeg, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ActivePowerNegative = &actNeg
		if reactPos, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ReactivePowerPositive = &reactPos
		if reactNeg, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ReactivePowerNegative = &reactNeg
	}

	if phases > 0 {
		m.Phases = make([]Phase, phases)

		// First is current (Amperes) for each phase
		for i := 0; i < phases; i++ {
			var cur int
			if cur, data, err = unmarshalInt32(data); err != nil {
				return m, err
			}
			m.Phases[i].Index = i + 1
			m.Phases[i].Current = float64(cur) / 1000
		}

		// Then the voltage for each phase
		for i := 0; i < phases; i++ {
			var voltage int
			if voltage, data, err = unmarshalInt32(data); err != nil {
				return m, err
			}
			m.Phases[i].Voltage = float64(voltage) / 10
		}
	}
	if energy {
		if tsData, data, err = unmarshalLvData(data); err != nil {
			return m, err
		}
		if ts, err := parseTimestamp(tsData); err != nil {
			return m, err
		} else {
			m.EnergyTimestamp = &ts
		}

		var actPos, actNeg, reactPos, reactNeg int
		if actPos, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ActiveEnergyPositive = &actPos
		if actNeg, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ActiveEnergyNegative = &actNeg
		if reactPos, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ReactiveEnergyPositive = &reactPos
		if reactNeg, data, err = unmarshalInt32(data); err != nil {
			return m, err
		}
		m.ReactiveEnergyNegative = &reactNeg
	}

	if len(data) < 2 {
		return m, fmt.Errorf("too few bytes (checksum)")
	}
	m.checksum = (uint16(data[0]) << 8) | uint16(data[1])
	// TODO: Validate checksum
	data = data[2:]
	if len(data) > 0 {
		return m, fmt.Errorf("trailing data: %X", data)
	}

	return m, nil

}

func unmarshalLvString(data []byte) (string, []byte, error) {
	b, data, err := unmarshalLvData(data)
	return string(b), data, err
}

func unmarshalLvData(data []byte) ([]byte, []byte, error) {
	if len(data) < 2 {
		return nil, nil, fmt.Errorf("too few bytes (lvdata)")
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
		return 0, nil, fmt.Errorf("too few bytes (uint8)")
	}
	if data[0] != 0x02 {
		return 0, nil, fmt.Errorf("uint8: wrong item type, expected 0x02, got 0x%02X", data[0])
	}
	return int(data[1]), data[2:], nil
}

func unmarshalInt32(data []byte) (int, []byte, error) {
	if len(data) < 5 {
		return 0, nil, fmt.Errorf("too few bytes (uint8)")
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
