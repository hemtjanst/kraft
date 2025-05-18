package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tarm/serial"
	"hemtjan.st/kraft/kaifa"
	"io"
	"lib.hemtjan.st/client"
	"lib.hemtjan.st/device"
	"lib.hemtjan.st/feature"
	"lib.hemtjan.st/hass"
	"lib.hemtjan.st/transport/mqtt"
	"log"
	"os"
	"time"
)

const (

	// Re-use currentPower from hemtjanst for positive power (i.e. power flowing into the system from the grid)
	currentPower = string(feature.CurrentPower)
	// Define a custom feature for produced power (e.g. if exporting Solar power to the grid)
	currentPowerProduced = "currentPowerProduced"
	energyUsed           = string(feature.EnergyUsed)
	energyProduced       = "energyProduced"
	phaseCurrent         = "phase%dCurrent"
	phaseVoltage         = "phase%dVoltage"
)

func main() {
	serialDevice := flag.String("device", "/dev/ttyUSB0", "Serial device")
	baudFlag := flag.Int("speed", 2400, "Baud rate of serial port")
	topicName := flag.String("topic", "powerMeter/house", "Topic of hemtjanst device")
	name := flag.String("name", "Grid", "Name of device")
	haName := flag.String("hass.name", "grid", "Name of homeassistant device")

	mqFlags := mqtt.MustFlags(flag.String, flag.Bool)
	flag.Parse()

	ctx := context.Background()
	mq, err := mqtt.New(ctx, mqFlags())
	if err != nil {
		log.Fatalf("connecting to mqtt: %v", err)
	}

	// Spawn a goroutine to detect MQTT errors and handle reconnect
	go func() {
		for {
			ok, err := mq.Start()
			if err != nil {
				log.Printf("MQTT Error: %s", err)
			}
			if !ok {
				os.Exit(1)
			}
			time.Sleep(3 * time.Second)
			log.Printf("MQTT: Reconnecting")
		}
	}()

	var d client.Device

	var haDev *hass.Device

	// pushData gets called on each message
	pushData := func(msg *kaifa.Message) {
		if haDev == nil && *haName != "" {
			uniqPrefix := "kaifa_" + *msg.MeterID
			haDev = &hass.Device{
				Device: &hass.DeviceInfo{
					Identifiers:  []string{*msg.MeterID},
					Manufacturer: "Kaifa",
					Model:        *msg.MeterType,
					Name:         *haName,
					SwVersion:    *msg.Version,
					SerialNumber: *msg.MeterID,
				},
				Origin: &hass.Origin{
					Name:       "Kraft",
					SwVersion:  "0.1.1",
					SupportUrl: "https://github.com/hemtjanst/kraft",
				},
				Components: map[string]*hass.Component{},
				StateTopic: "homeassistant/" + *haName + "/state",
			}

			if msg.ActivePowerPositive != nil {
				haDev.Components["input_power"] = &hass.Component{
					Platform:          "sensor",
					Name:              "Input Power",
					UnitOfMeasurement: "W",
					ValueTemplate:     "{{ value_json.ActivePowerPositive }}",
					StateClass:        "measurement",
					DeviceClass:       "power",
					UniqueId:          uniqPrefix + "_input_power",
				}
			}
			if msg.ActivePowerNegative != nil {
				haDev.Components["output_power"] = &hass.Component{
					Platform:          "sensor",
					Name:              "Output Power",
					UnitOfMeasurement: "W",
					ValueTemplate:     "{{ value_json.ActivePowerNegative }}",
					StateClass:        "measurement",
					DeviceClass:       "power",
					UniqueId:          uniqPrefix + "_output_power",
				}
			}
			for idx, ph := range msg.Phases {
				n1 := fmt.Sprintf("phase_%d_current", ph.Index)
				n2 := fmt.Sprintf("phase_%d_voltage", ph.Index)

				haDev.Components[n1] = &hass.Component{
					Platform:          "sensor",
					Name:              fmt.Sprintf("Phase %d Current", ph.Index),
					UnitOfMeasurement: "A",
					ValueTemplate:     fmt.Sprintf("{{ value_json.Phases[%d].Current }}", idx),
					StateClass:        "measurement",
					DeviceClass:       "current",
					UniqueId:          uniqPrefix + "_" + n1,
				}
				haDev.Components[n2] = &hass.Component{
					Platform:          "sensor",
					Name:              fmt.Sprintf("Phase %d Voltage", ph.Index),
					UnitOfMeasurement: "V",
					ValueTemplate:     fmt.Sprintf("{{ value_json.Phases[%d].Voltage }}", idx),
					StateClass:        "measurement",
					DeviceClass:       "voltage",
					UniqueId:          uniqPrefix + "_" + n2,
				}
			}
			if msg.ActiveEnergyPositive != nil {
				haDev.Components["consumed_energy"] = &hass.Component{
					Platform:          "sensor",
					Name:              "Consumed Energy",
					UnitOfMeasurement: "Wh",
					ValueTemplate:     "{{ value_json.ActiveEnergyPositive }}",
					StateClass:        "total_increasing",
					DeviceClass:       "energy",
					UniqueId:          uniqPrefix + "_consumed_energy",
				}
			}
			if msg.ActiveEnergyNegative != nil {
				haDev.Components[*haName+"_returned_energy"] = &hass.Component{
					Platform:          "sensor",
					Name:              "Returned Energy",
					UnitOfMeasurement: "Wh",
					ValueTemplate:     "{{ value_json.ActiveEnergyNegative }}",
					StateClass:        "total_increasing",
					DeviceClass:       "energy",
					UniqueId:          uniqPrefix + "_returned_energy",
				}
			}
			b, err := json.Marshal(haDev)
			if err == nil {
				mq.Publish("homeassistant/device/"+*haName+"/config", b, true)
			}
		}

		if haDev != nil {
			b, err := json.Marshal(msg)
			if err == nil {
				mq.Publish(haDev.StateTopic, b, true)
			}
		}

		if d == nil && *topicName != "" {
			// Device is created once the first message is received
			// since we need to know which features are supported
			// and the model/serial number

			var err error
			info := &device.Info{
				Topic:        *topicName,
				Name:         *name,
				Manufacturer: "Kaifa",
				Features:     map[string]*feature.Info{},
				Type:         "energyMeter",
			}
			if msg.MeterID != nil {
				info.SerialNumber = *msg.MeterID
			}
			if msg.MeterType != nil {
				info.Model = *msg.MeterType
			}

			if msg.ActivePowerPositive != nil {
				info.Features[currentPower] = &feature.Info{}
			}
			if msg.ActivePowerNegative != nil {
				info.Features[currentPowerProduced] = &feature.Info{}
			}

			for _, ph := range msg.Phases {
				info.Features[fmt.Sprintf(phaseCurrent, ph.Index)] = &feature.Info{}
				info.Features[fmt.Sprintf(phaseVoltage, ph.Index)] = &feature.Info{}
			}

			if msg.ActiveEnergyPositive != nil {
				info.Features[energyUsed] = &feature.Info{}
			}
			if msg.ActiveEnergyNegative != nil {
				info.Features[energyProduced] = &feature.Info{}
			}

			d, err = client.NewDevice(info, mq)
			if err != nil {
				log.Fatalf("error creating device: %v", err)
			}
		}

		if msg.ActivePowerPositive != nil {
			// Power imported from the grid in Watts
			if d != nil {
				_ = d.Feature(currentPower).Update(fmt.Sprintf("%d", *msg.ActivePowerPositive))
			}
		}
		if msg.ActivePowerNegative != nil {
			// Power exported to the grid in Watts
			if d != nil {
				_ = d.Feature(currentPowerProduced).Update(fmt.Sprintf("%d", *msg.ActivePowerNegative))
			}
		}

		for _, ph := range msg.Phases {
			if d != nil {
				// Current in Amperes
				_ = d.Feature(fmt.Sprintf(phaseCurrent, ph.Index)).Update(fmt.Sprintf("%.3f", ph.Current))
				// Voltage in Volts
				_ = d.Feature(fmt.Sprintf(phaseVoltage, ph.Index)).Update(fmt.Sprintf("%.1f", ph.Voltage))
			}
		}

		if msg.ActiveEnergyPositive != nil {
			if d != nil {
				// Convert to float and divide by 1000 to get kWh
				_ = d.Feature(energyUsed).Update(fmt.Sprintf("%.3f", float64(*msg.ActiveEnergyPositive)/1000))
			}
		}
		if msg.ActiveEnergyNegative != nil {
			if d != nil {
				// Convert to float and divide by 1000 to get kWh
				_ = d.Feature(energyProduced).Update(fmt.Sprintf("%.3f", float64(*msg.ActiveEnergyNegative)/1000))
			}
		}

	}

	cfg := &serial.Config{
		Name:   *serialDevice,
		Baud:   *baudFlag,
		Parity: serial.ParityEven,
		Size:   8,
	}

	// Open serial to read & discard everything for 200ms to drain incoming buffer
	s, err := serial.OpenPort(cfg)
	if err != nil {
		log.Fatalf("error opening %s: %v", *serialDevice, err)
	}
	go func() {
		_, _ = io.ReadAll(s)
	}()
	time.Sleep(200 * time.Millisecond)
	_ = s.Close()

	s, err = serial.OpenPort(cfg)
	if err != nil {
		log.Fatalf("error opening %s: %v", *serialDevice, err)
	}
	r := kaifa.NewReader(s)

	for {
		// Main loop, keep reading frames until serial closes or program is terminated
		fr, err := r.ReadFrame()
		if err != nil {
			if err == io.EOF {
				log.Printf("EOF from serial device, exiting")
				return
			}
			log.Fatalf("error while reading frame: %v", err)
		}
		msg, err := kaifa.Unmarshal(fr)
		if err != nil {
			log.Fatalf("Error unmarshalling frame: %v\nData: %X", err, fr)
		}

		pushData(msg)
	}
}
