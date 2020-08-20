# Kraft ![GitHub release](https://img.shields.io/github/release/hemtjanst/kraft.svg) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/hemtjan.st/kraft/)

Kraft reads ModBus data from an energy meter and publishes it as a Hemtjänst device

Currently only supports Kaifa energy meters over serial

## Usage

Command line options:

```
Usage of kraft:
  -device string
        Serial device (default "/dev/ttyUSB0")
  -mqtt.address string
        Address to MQTT endpoint (default "localhost:1883")
  -mqtt.ca string
        Path to CA certificate
  -mqtt.cert string
        Path to Client certificate
  -mqtt.cn string
        Common name of server certificate (usually the hostname)
  -mqtt.key string
        Path to Client certificate key
  -mqtt.password string
        MQTT Password
  -mqtt.tls
        Enable TLS
  -mqtt.tls-insecure
        Disable TLS certificate validation
  -mqtt.username string
        MQTT Username
  -name string
        Name of hemtjanst device (default "House Power Meter")
  -speed int
        Baud rate of serial port (default 2400)
  -topic string
        Topic of hemtjanst device (default "powerMeter/house")
  -topic.announce string
        Announce topic for Hemtjänst (default "announce")
  -topic.discover string
        Discover topic for Hemtjänst (default "discover")
  -topic.leave string
        Leave topic for hemtjänst (default "leave")                                                                                                                                               -device string                                                                                                                                                    Serial device (default "/dev/ttyUSB0")                                                                                                                -mqtt.address string                                                                                                                                              Address to MQTT endpoint (default "localhost:1883")                                                                                                   -mqtt.ca string                                                                                                                                                   Path to CA certificate                                                                                                                                -mqtt.cert string                                                                                                                                                 Path to Client certificate                                                                                                                            -mqtt.cn string                                                                                                                                                   Common name of server certificate (usually the hostname)                                                                                              -mqtt.key string                                                                                                                                                  Path to Client certificate key                                                                                                                        -mqtt.password string                                                                                                                                             MQTT Password                                                                                                                                         -mqtt.tls                                                                                                                                                         Enable TLS                                                                                                                                            -mqtt.tls-insecure                                                                                                                                                Disable TLS certificate validation                                                                                                                    -mqtt.username string                                                                                                                                             MQTT Username                                                                                                                                         -name string                                                                                                                                                      Name of hemtjanst device (default "House Power Meter")                                                                                                -speed int                                                                                                                                                        Baud rate of serial port (default 2400)                                                                                                               -topic string                                                                                                                                                     Topic of hemtjanst device (default "powerMeter/house")                                                                                                -topic.announce string                                                                                                                                            Announce topic for Hemtjänst (default "announce")                                                                                                     -topic.discover string                                                                                                                                            Discover topic for Hemtjänst (default "discover")                                                                                                     -topic.leave string                                                                                                                                               Leave topic for hemtjänst (default "leave")   
```
