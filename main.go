package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/tarm/serial"
	"hemtjan.st/kraft/kaifa"
	"io"
	"lib.hemtjan.st/client"
	"lib.hemtjan.st/device"
	"lib.hemtjan.st/feature"
	"lib.hemtjan.st/transport/mqtt"
	"log"
	"os"
	"time"
)

const (
	currentPower         = string(feature.CurrentPower)
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
	name := flag.String("name", "House Power Meter", "Name of hemtjanst device")

	mqFlags := mqtt.MustFlags(flag.String, flag.Bool)
	flag.Parse()

	ctx := context.Background()
	mq, err := mqtt.New(ctx, mqFlags())
	if err != nil {
		log.Fatalf("connecting to mqtt: %v", err)
	}
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

	pushData := func(msg *kaifa.Message) {
		if d == nil {
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
			_ = d.Feature(currentPower).Update(fmt.Sprintf("%d", *msg.ActivePowerPositive))
		}
		if msg.ActivePowerNegative != nil {
			_ = d.Feature(currentPowerProduced).Update(fmt.Sprintf("%d", *msg.ActivePowerNegative))
		}

		for _, ph := range msg.Phases {
			_ = d.Feature(fmt.Sprintf(phaseCurrent, ph.Index)).Update(fmt.Sprintf("%.3f", ph.Current))
			_ = d.Feature(fmt.Sprintf(phaseVoltage, ph.Index)).Update(fmt.Sprintf("%.1f", ph.Voltage))
		}

		if msg.ActiveEnergyPositive != nil {
			_ = d.Feature(energyUsed).Update(fmt.Sprintf("%.3f", float64(*msg.ActiveEnergyPositive)/1000))
		}
		if msg.ActiveEnergyNegative != nil {
			_ = d.Feature(energyProduced).Update(fmt.Sprintf("%.3f", float64(*msg.ActiveEnergyNegative)/1000))
		}

	}

	cfg := &serial.Config{
		Name:   *serialDevice,
		Baud:   *baudFlag,
		Parity: serial.ParityEven,
		Size:   8,
	}
	s, err := serial.OpenPort(cfg)
	if err != nil {
		log.Fatalf("error opening %s: %v", *serialDevice, err)
	}

	r := kaifa.NewReader(s)

	for {
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
			log.Printf("Error unmarshalling frame: %v\nData: %X", err, fr)
			continue
		}

		pushData(msg)
	}
}
