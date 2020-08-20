package kaifa

import (
	"fmt"
)

func Unmarshal(data []byte) (*Message, error) {

	var err error
	m := &Message{}

	if data, err = m.header.Unmarshal(data); err != nil {
		return m, err
	}

	if data, err = m.meta.Unmarshal(data); err != nil {
		return m, err
	}

	var tsData []byte
	if tsData, data, err = unmarshalLvData(data); err != nil {
		return m, err
	}

	if m.Timestamp, err = parseTimestamp(tsData); err != nil {
		return m, err
	}
	var itemCount int
	energy := false
	fullHdr := true
	phases := 0

	if itemCount, data, err = unmarshalUint8(data); err != nil {
		return m, err
	}

	switch itemCount {
	case 0:
		return m, nil
	case 1:
		if len(data) < 4 {
			return m, fmt.Errorf("too few bytes (pwr_act_pos)")
		}
		activePowerPositive, _, err := unmarshalInt32(data)
		if err != nil {
			return m, err
		}
		fullHdr = false
		m.ActivePowerPositive = &activePowerPositive
	case 9:
		phases = 1
	case 13:
		phases = 3
	case 14:
		phases = 1
		energy = true
	case 18:
		phases = 3
		energy = true
	default:
		return m, fmt.Errorf("unsupported number of items: %d", itemCount)
	}

	if fullHdr {
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

		for i := 0; i < phases; i++ {
			var cur int
			if cur, data, err = unmarshalInt32(data); err != nil {
				return m, err
			}
			m.Phases[i].Index = i + 1
			m.Phases[i].Current = float64(cur) / 1000
		}

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
	data = data[2:]
	if len(data) > 0 {
		return m, fmt.Errorf("trailing data: %X", data)
	}

	return m, nil

}
