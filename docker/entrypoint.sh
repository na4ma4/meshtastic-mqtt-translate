#!/bin/bash

MQTT_BROKER=$(bashio::services mqtt "host")
MQTT_USERNAME=$(bashio::services mqtt "username")
MQTT_PASSWORD=$(bashio::services mqtt "password")

export MQTT_BROKER MQTT_USERNAME MQTT_PASSWORD

echo "Starting Meshtastic MQTT Relay with the following settings:"
echo "MQTT Host: ${MQTT_BROKER}"
echo "MQTT User: ${MQTT_USERNAME}"
# Note: Not printing MQTT_PASSWORD for security reasons

tini -- /meshtastic-mqtt-relay "${@}"
