# meshtastic-mqtt-bin-to-json

A Golang application that connects to an MQTT broker to receive Meshtastic messages in binary format (Protocol Buffers), converts them to JSON, and publishes them to another MQTT broker.

## Features

- Subscribes to Meshtastic MQTT topics with binary (protobuf) messages
- Decodes Meshtastic ServiceEnvelope and MeshPacket messages
- Converts binary messages to human-readable JSON format
- Publishes converted messages to a destination MQTT broker
- Supports authentication with username/password
- Configurable source and destination brokers
- Minimal memory footprint and efficient processing

## Installation

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (protoc) - only needed for development
- Docker (optional, for containerized deployment)

### Build from source

```bash
git clone https://github.com/na4ma4/meshtastic-mqtt-bin-to-json.git
cd meshtastic-mqtt-bin-to-json
make build
```

Or manually:

```bash
go build -o meshtastic-mqtt-relay ./cmd/meshtastic-mqtt-relay
```

### Docker

Build the Docker image:

```bash
docker build -t meshtastic-mqtt-relay .
```

Run with Docker:

```bash
docker run --rm meshtastic-mqtt-relay \
  -source-broker tcp://mqtt.meshtastic.org:1883 \
  -source-topic "msh/US/2/json/#" \
  -dest-broker tcp://your-broker:1883
```

Or use docker-compose:

```bash
docker-compose up -d
```

## Usage

### Command-line Options

```
Usage of meshtastic-mqtt-relay:
  -source-broker string
        Source MQTT broker URL (default "tcp://localhost:1883")
  -source-client-id string
        Source MQTT client ID (default "meshtastic-relay-source")
  -source-topic string
        Source MQTT topic to subscribe to (default "msh/+/json/+/+")
  -dest-broker string
        Destination MQTT broker URL (default "tcp://localhost:1884")
  -dest-client-id string
        Destination MQTT client ID (default "meshtastic-relay-dest")
  -dest-topic string
        Destination MQTT topic prefix (default "meshtastic/json")
  -username string
        MQTT username (optional)
  -password string
        MQTT password (optional)
```

### Examples

**Basic usage with default settings:**

```bash
./meshtastic-mqtt-relay
```

**Connect to a specific Meshtastic MQTT broker:**

```bash
./meshtastic-mqtt-relay \
  -source-broker tcp://mqtt.meshtastic.org:1883 \
  -source-topic msh/US/2/json/# \
  -dest-broker tcp://localhost:1883 \
  -dest-topic meshtastic/converted
```

**With authentication:**

```bash
./meshtastic-mqtt-relay \
  -source-broker tcp://mqtt.example.com:1883 \
  -username myuser \
  -password mypassword
```

**Using TLS:**

```bash
./meshtastic-mqtt-relay \
  -source-broker tls://mqtt.example.com:8883 \
  -dest-broker tls://localhost:8883
```

## How It Works

1. The application connects to the source MQTT broker and subscribes to the specified topic pattern
2. When a message arrives, it's decoded from the Meshtastic protobuf format (ServiceEnvelope)
3. The binary message is converted to a structured JSON format
4. The JSON message is published to the destination broker with the original topic appended to the destination prefix

### Message Format

**Input (Binary/Protobuf):**
- Meshtastic ServiceEnvelope containing MeshPacket

**Output (JSON):**
```json
{
  "packet": {
    "from": 123456789,
    "to": 987654321,
    "channel": 0,
    "id": 12345,
    "rxTime": 1699435200,
    "rxSnr": 8.5,
    "rxRssi": -85,
    "hopLimit": 3,
    "wantAck": false,
    "viaMqtt": false,
    "hopStart": 3,
    "decoded": {
      "portnum": "TEXT_MESSAGE_APP",
      "payload": "SGVsbG8gV29ybGQ=",
      "wantResponse": false
    }
  },
  "channelId": "LongFast",
  "gatewayId": "!12345678"
}
```

## Development

### Building

```bash
make build
```

### Generating protobuf code

If you modify the protobuf definitions:

```bash
make proto
```

### Running tests

```bash
make test
```

### Cleaning build artifacts

```bash
make clean
```

## Architecture

```
┌─────────────────┐      Binary/Protobuf      ┌──────────────────┐
│  Source MQTT    │ ──────────────────────────>│  meshtastic-     │
│  Broker         │    (Meshtastic messages)   │  mqtt-relay      │
│ (Meshtastic)    │                            │                  │
└─────────────────┘                            │  - Deserialize   │
                                               │  - Convert       │
                                               │  - Publish       │
                                               └────────┬─────────┘
                                                        │
                                                  JSON Format
                                                        │
                                                        v
                                               ┌─────────────────┐
                                               │ Destination     │
                                               │ MQTT Broker     │
                                               └─────────────────┘
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See LICENSE file for details.

## References

- [Meshtastic](https://meshtastic.org/)
- [Meshtastic Protobufs](https://buf.build/meshtastic/protobufs)
- [MQTT](https://mqtt.org/)

