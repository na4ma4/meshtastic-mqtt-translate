//nolint:protogetter // copying structures
package translator

import (
	"encoding/json"

	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"
)

// DeviceMetrics - Key native device metrics such as battery level.
type DeviceMetrics struct {
	// 0-100 (>100 means powered)
	BatteryLevel *uint32 `json:"battery_level,omitempty"`
	// Voltage measured
	Voltage *SpecialFloat64 `json:"voltage,omitempty"`
	// Utilization for the current channel, including well formed TX, RX and malformed RX (aka noise).
	ChannelUtilization *SpecialFloat64 `json:"channel_utilization,omitempty"`
	// Percent of airtime for transmission used within the last hour.
	AirUtilTx *SpecialFloat64 `json:"air_util_tx,omitempty"`
	// How long the device has been running since the last reboot (in seconds)
	UptimeSeconds *uint32 `json:"uptime_seconds,omitempty"`
}
type TelemetryDeviceMetrics struct {
	Time          *uint32        `json:"time,omitempty"`
	DeviceMetrics *DeviceMetrics `json:"device_metrics,omitempty"`
}

func NewDeviceMetrics(in *meshtastic.DeviceMetrics) *TelemetryDeviceMetrics {
	if in == nil {
		return nil
	}
	return &TelemetryDeviceMetrics{
		DeviceMetrics: &DeviceMetrics{
			BatteryLevel:       in.BatteryLevel,
			Voltage:            specialPtrFloat(in.Voltage),
			ChannelUtilization: specialPtrFloat(in.ChannelUtilization),
			AirUtilTx:          specialPtrFloat(in.AirUtilTx),
			UptimeSeconds:      in.UptimeSeconds,
		},
	}
}

func (p *TelemetryDeviceMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// LocalStats - Local device mesh statistics.
type LocalStats struct {
	// How long the device has been running since the last reboot (in seconds)
	UptimeSeconds uint32 `json:"uptime_seconds,omitempty"`
	// Utilization for the current channel, including well formed TX, RX and malformed RX (aka noise).
	ChannelUtilization float64 `json:"channel_utilization,omitempty"`
	// Percent of airtime for transmission used within the last hour.
	AirUtilTx float64 `json:"air_util_tx,omitempty"`
	// Number of packets sent
	NumPacketsTx uint32 `json:"num_packets_tx,omitempty"`
	// Number of packets received (both good and bad)
	NumPacketsRx uint32 `json:"num_packets_rx,omitempty"`
	// Number of packets received that are malformed or violate the protocol
	NumPacketsRxBad uint32 `json:"num_packets_rx_bad,omitempty"`
	// Number of nodes online (in the past 2 hours)
	NumOnlineNodes uint32 `json:"num_online_nodes,omitempty"`
	// Number of nodes total
	NumTotalNodes uint32 `json:"num_total_nodes,omitempty"`
	// Number of received packets that were duplicates (due to multiple nodes relaying).
	// If this number is high, there are nodes in the mesh relaying packets when it's unnecessary, for example due to the ROUTER/REPEATER role.
	NumRxDupe uint32 `json:"num_rx_dupe,omitempty"`
	// Number of packets we transmitted that were a relay for others (not originating from ourselves).
	NumTxRelay uint32 `json:"num_tx_relay,omitempty"`
	// Number of times we canceled a packet to be relayed, because someone else did it before us.
	// This will always be zero for ROUTERs/REPEATERs. If this number is high, some other node(s) is/are relaying faster than you.
	NumTxRelayCanceled uint32 `json:"num_tx_relay_canceled,omitempty"`
	// Number of bytes used in the heap
	HeapTotalBytes uint32 `json:"heap_total_bytes,omitempty"`
	// Number of bytes free in the heap
	HeapFreeBytes uint32 `json:"heap_free_bytes,omitempty"`
	// Number of packets that were dropped because the transmit queue was full.
	NumTxDropped uint32 `json:"num_tx_dropped,omitempty"`
}
type TelemetryLocalStats struct {
	LocalStats *LocalStats `json:"local_stats,omitempty"`
}

func NewLocalStats(in *meshtastic.LocalStats) *TelemetryLocalStats {
	if in == nil {
		return nil
	}
	return &TelemetryLocalStats{
		LocalStats: &LocalStats{
			UptimeSeconds:      in.UptimeSeconds,
			ChannelUtilization: float64(in.ChannelUtilization),
			AirUtilTx:          float64(in.AirUtilTx),
			NumPacketsTx:       in.NumPacketsTx,
			NumPacketsRx:       in.NumPacketsRx,
			NumPacketsRxBad:    in.NumPacketsRxBad,
			NumOnlineNodes:     in.NumOnlineNodes,
			NumTotalNodes:      in.NumTotalNodes,
			NumRxDupe:          in.NumRxDupe,
			NumTxRelay:         in.NumTxRelay,
			NumTxRelayCanceled: in.NumTxRelayCanceled,
			HeapTotalBytes:     in.HeapTotalBytes,
			HeapFreeBytes:      in.HeapFreeBytes,
			NumTxDropped:       in.NumTxDropped,
		},
	}
}

func (p *TelemetryLocalStats) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// PowerMetrics - Power Metrics (voltage / current / etc).
type PowerMetrics struct {
	// Channel 1 Voltage
	Ch1Voltage *float64 `json:"ch1_voltage,omitempty"`
	// Channel 1 Current
	Ch1Current *float64 `json:"ch1_current,omitempty"`
	// Channel 2 Voltage
	Ch2Voltage *float64 `json:"ch2_voltage,omitempty"`
	// Channel 2 Current
	Ch2Current *float64 `json:"ch2_current,omitempty"`
	// Channel 3 Voltage
	Ch3Voltage *float64 `json:"ch3_voltage,omitempty"`
	// Channel 3 Current
	Ch3Current *float64 `json:"ch3_current,omitempty"`
	// Channel 4 Voltage
	Ch4Voltage *float64 `json:"ch4_voltage,omitempty"`
	// Channel 4 Current
	Ch4Current *float64 `json:"ch4_current,omitempty"`
	// Channel 5 Voltage
	Ch5Voltage *float64 `json:"ch5_voltage,omitempty"`
	// Channel 5 Current
	Ch5Current *float64 `json:"ch5_current,omitempty"`
	// Channel 6 Voltage
	Ch6Voltage *float64 `json:"ch6_voltage,omitempty"`
	// Channel 6 Current
	Ch6Current *float64 `json:"ch6_current,omitempty"`
	// Channel 7 Voltage
	Ch7Voltage *float64 `json:"ch7_voltage,omitempty"`
	// Channel 7 Current
	Ch7Current *float64 `json:"ch7_current,omitempty"`
	// Channel 8 Voltage
	Ch8Voltage *float64 `json:"ch8_voltage,omitempty"`
	// Channel 8 Current
	Ch8Current *float64 `json:"ch8_current,omitempty"`
}
type TelemetryPowerMetrics struct {
	PowerMetrics *PowerMetrics `json:"power_metrics,omitempty"`
}

func NewPowerMetrics(in *meshtastic.PowerMetrics) *TelemetryPowerMetrics {
	if in == nil {
		return nil
	}
	return &TelemetryPowerMetrics{
		PowerMetrics: &PowerMetrics{
			Ch1Voltage: ptrFloat(in.Ch1Voltage),
			Ch1Current: ptrFloat(in.Ch1Current),
			Ch2Voltage: ptrFloat(in.Ch2Voltage),
			Ch2Current: ptrFloat(in.Ch2Current),
			Ch3Voltage: ptrFloat(in.Ch3Voltage),
			Ch3Current: ptrFloat(in.Ch3Current),
			Ch4Voltage: ptrFloat(in.Ch4Voltage),
			Ch4Current: ptrFloat(in.Ch4Current),
			Ch5Voltage: ptrFloat(in.Ch5Voltage),
			Ch5Current: ptrFloat(in.Ch5Current),
			Ch6Voltage: ptrFloat(in.Ch6Voltage),
			Ch6Current: ptrFloat(in.Ch6Current),
			Ch7Voltage: ptrFloat(in.Ch7Voltage),
			Ch7Current: ptrFloat(in.Ch7Current),
			Ch8Voltage: ptrFloat(in.Ch8Voltage),
			Ch8Current: ptrFloat(in.Ch8Current),
		},
	}
}

func (p *TelemetryPowerMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// HostMetrics - Linux host metrics.
type HostMetrics struct {
	// Host system uptime
	UptimeSeconds uint32 `json:"uptime_seconds,omitempty"`
	// Host system free memory
	FreememBytes uint64 `json:"freemem_bytes,omitempty"`
	// Host system disk space free for /
	Diskfree1Bytes uint64 `json:"diskfree1_bytes,omitempty"`
	// Secondary system disk space free
	Diskfree2Bytes *uint64 `json:"diskfree2_bytes,omitempty"`
	// Tertiary disk space free
	Diskfree3Bytes *uint64 `json:"diskfree3_bytes,omitempty"`
	// Host system one minute load in 1/100ths
	Load1 uint32 `json:"load1,omitempty"`
	// Host system five minute load  in 1/100ths
	Load5 uint32 `json:"load5,omitempty"`
	// Host system fifteen minute load  in 1/100ths
	Load15 uint32 `json:"load15,omitempty"`
	// Optional User-provided string for arbitrary host system information
	// that doesn't make sense as a dedicated entry.
	UserString *string `json:"user_string,omitempty"`
}
type TelemetryHostMetrics struct {
	HostMetrics *HostMetrics `json:"host_metrics,omitempty"`
}

func NewHostMetrics(in *meshtastic.HostMetrics) *TelemetryHostMetrics {
	if in == nil {
		return nil
	}
	return &TelemetryHostMetrics{
		HostMetrics: &HostMetrics{
			UptimeSeconds:  in.UptimeSeconds,
			FreememBytes:   in.FreememBytes,
			Diskfree1Bytes: in.Diskfree1Bytes,
			Diskfree2Bytes: in.Diskfree2Bytes,
			Diskfree3Bytes: in.Diskfree3Bytes,
			Load1:          in.Load1,
			Load5:          in.Load5,
			Load15:         in.Load15,
			UserString:     in.UserString,
		},
	}
}

func (p *TelemetryHostMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// EnvironmentMetrics - Weather station or other environmental metrics.
type EnvironmentMetrics struct {
	// Temperature measured
	Temperature *SpecialFloat64 `json:"temperature,omitempty"`
	// Relative humidity percent measured
	RelativeHumidity *SpecialFloat64 `json:"relative_humidity,omitempty"`
	// Barometric pressure in hPA measured
	BarometricPressure *SpecialFloat64 `json:"barometric_pressure,omitempty"`
	// Gas resistance in MOhm measured
	GasResistance *SpecialFloat64 `json:"gas_resistance,omitempty"`
	// Voltage measured (To be depreciated in favor of PowerMetrics in Meshtastic 3.x)
	Voltage *SpecialFloat64 `json:"voltage,omitempty"`
	// Current measured (To be depreciated in favor of PowerMetrics in Meshtastic 3.x)
	Current *SpecialFloat64 `json:"current,omitempty"`
	// relative scale IAQ value as measured by Bosch BME680 . value 0-500.
	// Belongs to Air Quality but is not particle but VOC measurement. Other VOC values can also be put in here.
	Iaq *uint32 `json:"iaq,omitempty"`
	// RCWL9620 Doppler Radar Distance Sensor, used for water level detection. Float value in mm.
	Distance *SpecialFloat64 `json:"distance,omitempty"`
	// VEML7700 high accuracy ambient light(Lux) digital 16-bit resolution sensor.
	Lux *SpecialFloat64 `json:"lux,omitempty"`
	// VEML7700 high accuracy white light(irradiance) not calibrated digital 16-bit resolution sensor.
	WhiteLux *SpecialFloat64 `json:"white_lux,omitempty"`
	// Infrared lux
	IrLux *SpecialFloat64 `json:"ir_lux,omitempty"`
	// Ultraviolet lux
	UvLux *SpecialFloat64 `json:"uv_lux,omitempty"`
	// Wind direction in degrees
	// 0 degrees = North, 90 = East, etc...
	WindDirection *uint32 `json:"wind_direction,omitempty"`
	// Wind speed in m/s
	WindSpeed *SpecialFloat64 `json:"wind_speed,omitempty"`
	// Weight in KG
	Weight *SpecialFloat64 `json:"weight,omitempty"`
	// Wind gust in m/s
	WindGust *SpecialFloat64 `json:"wind_gust,omitempty"`
	// Wind lull in m/s
	WindLull *SpecialFloat64 `json:"wind_lull,omitempty"`
	// Radiation in µR/h
	Radiation *SpecialFloat64 `json:"radiation,omitempty"`
	// Rainfall in the last hour in mm
	Rainfall1H *SpecialFloat64 `json:"rainfall_1h,omitempty"`
	// Rainfall in the last 24 hours in mm
	Rainfall24H *SpecialFloat64 `json:"rainfall_24h,omitempty"`
	// Soil moisture measured (% 1-100)
	SoilMoisture *uint32 `json:"soil_moisture,omitempty"`
	// Soil temperature measured (*C)
	SoilTemperature *SpecialFloat64 `json:"soil_temperature,omitempty"`
}
type TelemetryEnvironmentMetrics struct {
	EnvironmentMetrics *EnvironmentMetrics `json:"environment_metrics,omitempty"`
}

func NewEnvironmentMetrics(in *meshtastic.EnvironmentMetrics) *TelemetryEnvironmentMetrics {
	if in == nil {
		return nil
	}
	return &TelemetryEnvironmentMetrics{
		EnvironmentMetrics: &EnvironmentMetrics{
			Temperature:        specialPtrFloat(in.Temperature),
			RelativeHumidity:   specialPtrFloat(in.RelativeHumidity),
			BarometricPressure: specialPtrFloat(in.BarometricPressure),
			GasResistance:      specialPtrFloat(in.GasResistance),
			Voltage:            specialPtrFloat(in.Voltage),
			Current:            specialPtrFloat(in.Current),
			Iaq:                in.Iaq,
			Distance:           specialPtrFloat(in.Distance),
			Lux:                specialPtrFloat(in.Lux),
			WhiteLux:           specialPtrFloat(in.WhiteLux),
			IrLux:              specialPtrFloat(in.IrLux),
			UvLux:              specialPtrFloat(in.UvLux),
			WindDirection:      in.WindDirection,
			WindSpeed:          specialPtrFloat(in.WindSpeed),
			Weight:             specialPtrFloat(in.Weight),
			WindGust:           specialPtrFloat(in.WindGust),
			WindLull:           specialPtrFloat(in.WindLull),
			Radiation:          specialPtrFloat(in.Radiation),
			Rainfall1H:         specialPtrFloat(in.Rainfall_1H),
			Rainfall24H:        specialPtrFloat(in.Rainfall_24H),
			SoilMoisture:       in.SoilMoisture,
			SoilTemperature:    specialPtrFloat(in.SoilTemperature),
		},
	}
}

func (p *TelemetryEnvironmentMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// AirQualityMetrics - Air quality metrics.
type AirQualityMetrics struct {
	// Concentration Units Standard PM1.0 in ug/m3
	Pm10Standard *uint32 `json:"pm10_standard,omitempty"`
	// Concentration Units Standard PM2.5 in ug/m3
	Pm25Standard *uint32 `json:"pm25_standard,omitempty"`
	// Concentration Units Standard PM10.0 in ug/m3
	Pm100Standard *uint32 `json:"pm100_standard,omitempty"`
	// Concentration Units Environmental PM1.0 in ug/m3
	Pm10Environmental *uint32 `json:"pm10_environmental,omitempty"`
	// Concentration Units Environmental PM2.5 in ug/m3
	Pm25Environmental *uint32 `json:"pm25_environmental,omitempty"`
	// Concentration Units Environmental PM10.0 in ug/m3
	Pm100Environmental *uint32 `json:"pm100_environmental,omitempty"`
	// 0.3um Particle Count in #/0.1l
	Particles03Um *uint32 `json:"particles_03um,omitempty"`
	// 0.5um Particle Count in #/0.1l
	Particles05Um *uint32 `json:"particles_05um,omitempty"`
	// 1.0um Particle Count in #/0.1l
	Particles10Um *uint32 `json:"particles_10um,omitempty"`
	// 2.5um Particle Count in #/0.1l
	Particles25Um *uint32 `json:"particles_25um,omitempty"`
	// 5.0um Particle Count in #/0.1l
	Particles50Um *uint32 `json:"particles_50um,omitempty"`
	// 10.0um Particle Count in #/0.1l
	Particles100Um *uint32 `json:"particles_100um,omitempty"`
	// CO2 concentration in ppm
	Co2 *uint32 `json:"co2,omitempty"`
	// CO2 sensor temperature in degC
	Co2Temperature *float64 `json:"co2_temperature,omitempty"`
	// CO2 sensor relative humidity in %
	Co2Humidity *float64 `json:"co2_humidity,omitempty"`
	// Formaldehyde sensor formaldehyde concentration in ppb
	FormFormaldehyde *float64 `json:"form_formaldehyde,omitempty"`
	// Formaldehyde sensor relative humidity in %RH
	FormHumidity *float64 `json:"form_humidity,omitempty"`
	// Formaldehyde sensor temperature in degrees Celsius
	FormTemperature *float64 `json:"form_temperature,omitempty"`
	// Concentration Units Standard PM4.0 in ug/m3
	Pm40Standard *uint32 `json:"pm40_standard,omitempty"`
	// 4.0um Particle Count in #/0.1l
	Particles40Um *uint32 `json:"particles_40um,omitempty"`
	// PM Sensor Temperature
	PmTemperature *float64 `json:"pm_temperature,omitempty"`
	// PM Sensor humidity
	PmHumidity *float64 `json:"pm_humidity,omitempty"`
	// PM Sensor VOC Index
	PmVocIdx *float64 `json:"pm_voc_idx,omitempty"`
	// PM Sensor NOx Index
	PmNoxIdx *float64 `json:"pm_nox_idx,omitempty"`
	// Typical Particle Size in um
	ParticlesTps *float64 `json:"particles_tps,omitempty"`
}
type TelemetryAirQualityMetrics struct {
	AirQualityMetrics *AirQualityMetrics `json:"air_quality_metrics,omitempty"`
}

func NewAirQualityMetrics(in *meshtastic.AirQualityMetrics) *TelemetryAirQualityMetrics {
	if in == nil {
		return nil
	}
	return &TelemetryAirQualityMetrics{
		AirQualityMetrics: &AirQualityMetrics{
			Pm10Standard:       in.Pm10Standard,
			Pm25Standard:       in.Pm25Standard,
			Pm100Standard:      in.Pm100Standard,
			Pm10Environmental:  in.Pm10Environmental,
			Pm25Environmental:  in.Pm25Environmental,
			Pm100Environmental: in.Pm100Environmental,
			Particles03Um:      in.Particles_03Um,
			Particles05Um:      in.Particles_05Um,
			Particles10Um:      in.Particles_10Um,
			Particles25Um:      in.Particles_25Um,
			Particles50Um:      in.Particles_50Um,
			Particles100Um:     in.Particles_100Um,
			Co2:                in.Co2,
			Co2Temperature:     ptrFloat(in.Co2Temperature),
			Co2Humidity:        ptrFloat(in.Co2Humidity),
			FormFormaldehyde:   ptrFloat(in.FormFormaldehyde),
			FormHumidity:       ptrFloat(in.FormHumidity),
			FormTemperature:    ptrFloat(in.FormTemperature),
			Pm40Standard:       in.Pm40Standard,
			Particles40Um:      in.Particles_40Um,
			PmTemperature:      ptrFloat(in.PmTemperature),
			PmHumidity:         ptrFloat(in.PmHumidity),
			PmVocIdx:           ptrFloat(in.PmVocIdx),
			PmNoxIdx:           ptrFloat(in.PmNoxIdx),
			ParticlesTps:       ptrFloat(in.ParticlesTps),
		},
	}
}

func (p *TelemetryAirQualityMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
