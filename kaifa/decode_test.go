package kaifa

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	testFrame = []byte{
		0x01,       // DestAddr
		0x00, 0x01, // SrcAddr
		0x10,       // ControlField
		0x56, 0x1b, // Hdr Checksum
		0xe6,                         // LSAP Dest
		0xe7,                         // LSAP Src
		0x00,                         // LLC Quality
		0x0f, 0x40, 0x00, 0x00, 0x00, // Meta
		0x09, 0x0c, // Data, length 12 (Timestamp)
		0x07, 0xe4, 0x08, 0x14, // Timestamp YYMD
		0x04,             // Day of week?
		0x0b, 0x1b, 0x0f, // Timestamp HMS
		0xff, 0x80, 0x00, 0x00, // Timestamp extra (millisec+offset?)
		0x02, 0x12, // Number of fields = 18
		0x09, 0x07, // Data, Length 7 (Version)
		'K', 'F', 'M', '_', '0', '0', '1',
		0x09, 0x10, // Data, length 16 (MeterID)
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5', '6',
		0x09, 0x08, // Data, length 8 (MeterType/Model)
		'M', 'A', '3', '0', '4', 'H', '4', 'D',
		0x06, 0x00, 0x00, 0x00, 0x00, // int32 - active power positive
		0x06, 0x00, 0x00, 0x0a, 0xa9, // int32 - active power negative
		0x06, 0x00, 0x00, 0x00, 0x00, // int32 - reactive power positive
		0x06, 0x00, 0x00, 0x01, 0xef, // int32 - reactive power negative
		0x06, 0x00, 0x00, 0x0e, 0x4b, // int32 - phase1 current (mA)
		0x06, 0x00, 0x00, 0x0f, 0x1c, // int32 - phase2 current (mA)
		0x06, 0x00, 0x00, 0x11, 0xd1, // int32 - phase3 current (mA)
		0x06, 0x00, 0x00, 0x09, 0x16, // int32 - phase1 voltage (1/10 V)
		0x06, 0x00, 0x00, 0x09, 0x0e, // int32 - phase2 voltage (1/10 V)
		0x06, 0x00, 0x00, 0x09, 0x06, // int32 - phase3 voltage (1/10 V)
		0x09, 0x0c, // Data, length 12 (meter timestamp)
		0x07, 0xe4, 0x08, 0x14, 0x04, 0x0b, 0x1b, 0x0f, 0xff, 0x80, 0x00, 0x00,
		0x06, 0x01, 0xe7, 0xbc, 0xb1, // accum active energy positive
		0x06, 0x00, 0x86, 0x97, 0xef, // accum active energy negative
		0x06, 0x00, 0x01, 0x3e, 0x98, // accum reactive energy positive
		0x06, 0x00, 0x49, 0x14, 0x1b, // accum reactive energy negative
		0x2e, 0x88, // footer checksum
	}
	testData = append(
		[]byte{frameTag, 0xa0, byte(len(testFrame) + 2)},
		append(testFrame, frameTag)...,
	)
)

func TestReader(t *testing.T) {
	r := bytes.NewReader(testData)
	d := NewReader(r)
	fr, err := d.ReadFrame()
	if err != nil {
		t.Error(err)
	}
	msg, err := Unmarshal(fr)
	if err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	cmpTime, _ := time.Parse(time.RFC3339, "2020-08-20T11:27:15+02:00")

	assert.Equal(t, cmpTime, msg.Timestamp)
	assert.Equal(t, "KFM_001", *msg.Version)
	assert.Equal(t, "1234567890123456", *msg.MeterID)
	assert.Equal(t, "MA304H4D", *msg.MeterType)
	assert.Equal(t, 0, *msg.ActivePowerPositive)
	assert.Equal(t, 2729, *msg.ActivePowerNegative)
	assert.Equal(t, 0, *msg.ReactivePowerPositive)
	assert.Equal(t, 495, *msg.ReactivePowerNegative)
	assert.Len(t, msg.Phases, 3)
	assert.Equal(t, 1, msg.Phases[0].Index)
	assert.Equal(t, 2, msg.Phases[1].Index)
	assert.Equal(t, 3, msg.Phases[2].Index)
	assert.Equal(t, 3.659, msg.Phases[0].Current)
	assert.Equal(t, 3.868, msg.Phases[1].Current)
	assert.Equal(t, 4.561, msg.Phases[2].Current)
	assert.Equal(t, 232.6, msg.Phases[0].Voltage)
	assert.Equal(t, 231.8, msg.Phases[1].Voltage)
	assert.Equal(t, 231.0, msg.Phases[2].Voltage)
	assert.Equal(t, cmpTime, *msg.EnergyTimestamp)
	assert.Equal(t, 31964337, *msg.ActiveEnergyPositive)
	assert.Equal(t, 8820719, *msg.ActiveEnergyNegative)
	assert.Equal(t, 81560, *msg.ReactiveEnergyPositive)
	assert.Equal(t, 4789275, *msg.ReactiveEnergyNegative)
}
