## 0.2.2

### Changed
- **Breaking**: Renamed project from `meshtastic-mqtt-bin-to-json` to `meshtastic-mqtt-translate`
  - Updated all Go package paths to use new repository name
  - Updated protobuf package references throughout codebase
  - Docker images now use new repository name
- Regenerated all protobuf files with updated package paths
- Updated Go module path to `github.com/na4ma4/meshtastic-mqtt-translate`

### Added
- Home Assistant addon integration
  - Added Home Assistant configuration reader
  - Added entrypoint script for Home Assistant addon support
  - Integrated with Home Assistant's MQTT service discovery

### Fixed
- Resolved issue with Docker container not running as root
- Improved logging output for better debugging
- Fixed GitHub Actions workflow permissions for enhanced security

## 0.2.1

- Initial Release
