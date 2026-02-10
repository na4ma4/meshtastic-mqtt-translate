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

## 0.4.0 (2026-02-10)


### Features

* Add fanout relay feature to Meshtastic MQTT Relay ([dcc1e9d](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/dcc1e9df0c4e9eb06d92093072bc0a2f5452057b))
* Add fanout relay feature to Meshtastic MQTT Relay ([#5](https://github.com/na4ma4/meshtastic-mqtt-translate/issues/5)) ([aeb444c](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/aeb444cee14981169f796045e6719d75d037dfec))
* Add message repeat command to store-query tool ([ac3a54b](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/ac3a54b6d869119af079e20452f774a2fa0d56bd))
* Add message repeat command to store-query tool ([#7](https://github.com/na4ma4/meshtastic-mqtt-translate/issues/7)) ([03ee4c6](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/03ee4c6c9cb40e850e88caeb60ec94ae58374b97))
* **goreleaser:** migrate to goreleaser for releases ([637a7d8](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/637a7d856e865f78182b5060d9ec9b08f739e28a))
* **hass-package:** initial commit of hass package files ([d69fa74](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/d69fa74db11a0a42f3fe74bc4e56e126d8028930))
* **hass-package:** initial commit of hass package files ([e5d98d5](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/e5d98d59aaf52e898503b08f5d930ec170a8c6a6))
* **meshtastic-mqtt-translate:** add health check endpoint and update README ([08eea17](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/08eea178942e9ea685a899d272bb6c81d191320e))
* **mqtt-retain-flag:** add retain flag option for MQTT messages ([5af48c0](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/5af48c018f71a076de9e4dca189432a440062058))
* **mqtt-retain-flag:** add retain flag option for MQTT messages ([#10](https://github.com/na4ma4/meshtastic-mqtt-translate/issues/10)) ([cbc0964](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/cbc0964b3ead81ca1ea9ebbc6bfe886fbfaee2d6))
* **release-please:** add release-please workflow ([2ab0091](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/2ab0091f076dc0955ae495ed76eb325cc4983909))
* split special metrics into separate topics ([0d1cc32](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/0d1cc323332038dd5c473366b2b3de16e8a4a637))
* split special metrics into separate topics ([#8](https://github.com/na4ma4/meshtastic-mqtt-translate/issues/8)) ([3136abf](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/3136abfa454196b028ca5836501d21b3e9977680))


### Bug Fixes

* enable optional message store feature ([b42c90e](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/b42c90e39cff158442a369b5709053b74657f0ef))
* **github-actions:** release secrets ([61f5018](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/61f5018fde7b887e88419691e474abe8c9f927d0))
* update fanout relay config and logging ([9b46bb4](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/9b46bb4a0851e04ddfdd005aa349b66b51649d92))
* update fanout relay config and logging ([#6](https://github.com/na4ma4/meshtastic-mqtt-translate/issues/6)) ([63d81b6](https://github.com/na4ma4/meshtastic-mqtt-translate/commit/63d81b6ef701b1a6dc4841e220f1b44596d7cf54))

## 0.2.1

- Initial Release
