# Meshtastic MQTT Translate

[![CI](https://github.com/na4ma4/meshtastic-mqtt-translate/actions/workflows/ci.yml/badge.svg)](https://github.com/na4ma4/meshtastic-mqtt-translate/actions/workflows/ci.yml)
[![Docker](https://github.com/na4ma4/meshtastic-mqtt-translate/actions/workflows/docker-release.yml/badge.svg)](https://github.com/na4ma4/meshtastic-mqtt-translate/actions/workflows/docker-release.yml)
[![Go Report Card](https://goreportcard.com/badge/meshtastic-mqtt-translate)](https://goreportcard.com/report/meshtastic-mqtt-translate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance Go application that bridges Meshtastic's binary MQTT protocol with JSON, making Meshtastic mesh network data easily accessible for home automation, monitoring, and analytics platforms.

## Features

- 🔄 **Real-time Translation**: Converts Meshtastic Protocol Buffer messages to human-readable JSON
- 📊 **Multiple Storage Backends**: Optional message archiving to SQLite, MySQL, or PostgreSQL
- 🏥 **Health Monitoring**: Built-in HTTP health check endpoint for container orchestration
- 🔌 **MQTT Native**: Direct MQTT-to-MQTT relay with no intermediate components
- 🚀 **High Performance**: Lightweight, concurrent message processing
- 🐳 **Docker Ready**: Official Docker images with multi-architecture support
- 🔒 **Secure**: Supports MQTT authentication and TLS connections

## Supported Message Types

The translator currently decodes the following Meshtastic message types:

- ✅ **TELEMETRY_APP**: Device, Environment, Air Quality, Power, Host, and Local Stats
- ✅ **NODEINFO_APP**: Node information and user details
- ✅ **POSITION_APP**: GPS location data
- ✅ **TEXT_MESSAGE_APP**: Text messages
- ✅ **STORE_FORWARD_APP**: Store and forward protocol messages
- ✅ **TRACEROUTE_APP**: Network route tracing
- ✅ **ROUTING_APP**: Mesh routing protocol messages
- 🔐 **Encrypted Messages**: Passed through with metadata (payload remains encrypted)

## Quick Start

### Using Docker (Recommended)

```bash
docker run -d \
  --name meshtastic-mqtt-relay \
  -e MQTT_BROKER="tcp://mqtt.example.com:1883" \
  -e MQTT_TOPIC="msh/US/2/e/#" \
  -e MQTT_USERNAME="your-username" \
  -e MQTT_PASSWORD="your-password" \
  ghcr.io/na4ma4/meshtastic-mqtt-relay:latest
```

### Using Docker Compose

```yaml
version: '3.8'

services:
  meshtastic-mqtt-relay:
    image: ghcr.io/na4ma4/meshtastic-mqtt-relay:latest
    container_name: meshtastic-mqtt-relay
    environment:
      MQTT_BROKER: "tcp://mqtt.example.com:1883"
      MQTT_TOPIC: "msh/US/2/e/#"
      MQTT_USERNAME: "your-username"
      MQTT_PASSWORD: "your-password"
      MQTT_CLIENTID: "meshtastic-relay"
      DEBUG: "false"
      # Optional: Enable message archiving
      # STORE_DSN: "sqlite:///data/messages.db"
    ports:
      - "8099:8099"
    restart: unless-stopped
    # volumes:
    #   - ./data:/data  # Uncomment if using SQLite storage
```

### Binary Installation

```bash
# Install from source
go install github.com/na4ma4/meshtastic-mqtt-translate/cmd/meshtastic-mqtt-relay@latest

# Run
meshtastic-mqtt-relay \
  --broker tcp://mqtt.example.com:1883 \
  --topic "msh/US/2/e/#" \
  --username your-username \
  --password your-password
```

## Configuration

### Command Line Flags

```bash
Flags:
  -b, --broker string     Source MQTT broker URL (default "tcp://localhost:1883")
  -c, --clientid string   MQTT client ID (default "meshtastic-mqtt-relay")
  -d, --debug             Debug output
  -n, --dry-run           Dry run mode (optional)
  -o, --dsn string        Data store DSN (optional)
  -p, --password string   MQTT password (optional)
  -t, --topic string      MQTT topic to subscribe to (default "msh/ANZ/2/e/#")
  -u, --username string   MQTT username (optional)
  -h, --help              Help for meshtastic-mqtt-relay
```

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MQTT_BROKER` | MQTT broker URL | `tcp://localhost:1883` | `tcp://mqtt.example.com:1883` |
| `MQTT_TOPIC` | Topic pattern to subscribe to | `msh/ANZ/2/e/#` | `msh/US/2/e/#` |
| `MQTT_USERNAME` | MQTT authentication username | - | `meshtastic-user` |
| `MQTT_PASSWORD` | MQTT authentication password | - | `your-secure-password` |
| `MQTT_CLIENTID` | MQTT client identifier | `meshtastic-mqtt-relay` | `my-relay-01` |
| `DEBUG` | Enable debug logging | `false` | `true` |
| `MQTT_DRY_RUN` | Test mode without publishing | `false` | `true` |
| `STORE_DSN` | Database connection string | - | See Storage Options below |
| `HEALTHCHECK_PORT` | Health check HTTP port | `8099` | `8080` |

### Topic Patterns

Meshtastic uses standardized MQTT topics:

```
msh/<REGION>/<MODEM_PRESET>/e/<CHANNEL>/<GATEWAY_ID>
```

**Subscribe to:**
- Specific region: `msh/US/2/e/#` (US region, LongFast)
- All regions: `msh/+/+/e/#`
- Specific channel: `msh/US/2/e/LongFast/#`

**Publishes to:**
- Converts `/e/` (encrypted/binary) to `/json/` in topic path
- Example: `msh/US/2/e/LongFast/!12345678` → `msh/US/2/json/LongFast/!12345678`

### Storage Options

Enable optional message archiving by setting the `STORE_DSN` environment variable:

#### SQLite (File-based)
```bash
STORE_DSN="sqlite:///data/meshtastic.db"
```

#### PostgreSQL
```bash
STORE_DSN="postgresql://user:password@localhost:5432/meshtastic?sslmode=disable"
```

#### MySQL
```bash
STORE_DSN="mysql://user:password@tcp(localhost:3306)/meshtastic?charset=utf8mb4&parseTime=True"
```

## Output Format

### Example Input (Binary Protocol Buffer)
```
Topic: msh/US/2/e/LongFast/!44be043f
Payload: [binary protobuf data]
```

### Example Output (JSON)
```json
{
  "channel": 0,
  "from": 1151991839,
  "id": 2863291514,
  "payload": {
    "channel": 0,
    "errorReason": "",
    "hopLimit": 3,
    "id": "!44be043f",
    "longName": "My Meshtastic Node",
    "macaddr": "JFh8Ws8g",
    "publicKey": "cvXV8K79XIXmmWnS3eppEHPKmiDslQG4bc4rsJnEE34=",
    "role": "CLIENT",
    "shortName": "MN01"
  },
  "rssi": -52,
  "sender": "!44be043f",
  "snr": 8.75,
  "timestamp": 1699999999,
  "to": 4294967295,
  "type": "NODEINFO_APP",
  "hops_away": 0,
  "hop_start": 3
}
```

## Health Check

The application exposes a health check endpoint on port 8099 (configurable):

```bash
curl http://localhost:8099/

# Response when healthy:
{
  "status": true,
  "source_connected": true,
  "dest_connected": true
}
```

This endpoint is used by Docker's `HEALTHCHECK` and can be integrated with:
- Kubernetes liveness/readiness probes
- Docker Swarm health checks
- Load balancers
- Monitoring systems

## Use Cases

### Home Assistant Integration

```yaml
# configuration.yaml
mqtt:
  sensor:
    - name: "Node AirUtilTX"
      device:
        model: "SEEED_SOLAR_NODE"
        name: "MyNode"
      unique_id: "meshtastic_hoth_airutiltx"
      state_topic: "msh/ANZ/2/json/MediumFast/!44be043f"
      state_class: measurement
      value_template: >-
        {% if value_json.from == 3697797612 and 
          value_json.payload is defined and 
          value_json.payload.device_metrics is defined and
          value_json.payload.device_metrics.air_util_tx is defined %}
        {{ (value_json.payload.device_metrics.air_util_tx | float) | round(2) }}
        {% else %}
        {{ this.state }}
        {% endif %}
      unit_of_measurement: "%"
```

### Node-RED Processing

Subscribe to `msh/+/+/json/#` and process JSON messages directly with function nodes.

### Data Analytics

Archive messages to PostgreSQL and run SQL queries:

```sql
SELECT 
  from_node,
  COUNT(*) as message_count,
  AVG(rssi) as avg_rssi
FROM messages
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY from_node;
```

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/na4ma4/meshtastic-mqtt-translate.git
cd meshtastic-mqtt-translate

# Build binary
go build -o meshtastic-mqtt-relay ./cmd/meshtastic-mqtt-relay

# Run tests
go test ./...

# Build Docker image
docker build -t meshtastic-mqtt-relay .
```

### Project Structure

```
.
├── cmd/
│   └── meshtastic-mqtt-relay/  # Main application
├── internal/
│   ├── health/                  # Health check HTTP server
│   ├── mainconfig/              # Configuration management
│   ├── relay/                   # Core MQTT relay logic
│   ├── store/                   # Database storage backends
│   └── translator/              # Message type decoders
├── pkg/
│   └── meshtastic/              # Generated protobuf code
└── testdata/                    # Test fixtures
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

### No Messages Received

- Verify MQTT broker connectivity: `mosquitto_sub -h mqtt.example.com -t 'msh/#' -v`
- Check topic pattern matches your region/channel
- Ensure MQTT credentials are correct
- Enable debug logging: `-d` or `DEBUG=true`

### Health Check Failing

- Verify port 8099 is accessible
- Check MQTT connections in health response
- Review logs for connection errors

### Messages Not Decoded

- Some message types may not be fully implemented (see TODO comments in code)
- Encrypted messages will show type but no payload content
- Enable debug logging to see raw protobuf data

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Meshtastic Project](https://meshtastic.org/) - Open source mesh networking platform
- [Eclipse Paho](https://www.eclipse.org/paho/) - MQTT client library
- Protocol Buffer definitions from the Meshtastic project

## Related Projects

- [Meshtastic](https://github.com/meshtastic/meshtastic) - Main Meshtastic firmware
- [Meshtastic Python](https://github.com/meshtastic/python) - Official Python CLI/library
- [MQTT Explorer](http://mqtt-explorer.com/) - Useful for debugging MQTT topics

## Support

- 📖 [Meshtastic Documentation](https://meshtastic.org/docs/)
- 💬 [Meshtastic Discord](https://discord.gg/meshtastic)
- 🐛 [Issue Tracker](https://github.com/na4ma4/meshtastic-mqtt-translate/issues)
