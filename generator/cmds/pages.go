package cmds

import (
	"crypto/md5"
	"math"
	"math/rand"
)

const (
	TempMonAlarmThreshold     float64 = 10.0 // in Celsius degrees, added/substracted as High/Low Alarm
	VccMonAlarmThreshold      float64 = 0.3  // in V, -||-
	OpticalTxRxAlarmThreshold float64 = 1.33 // multiplicable factor for Optics Alarms
)

func (m *Module) PageLow(temperature float64, vcc float64) (page []byte) {
	temperature = temperature + 2*rand.Float64() - 1
	vcc = vcc + rand.Float64()/500 - 0.001
	page = append(page, byte(m.SFF8024Identifier)) // SFF8024Identifier
	page = append(page, byte(m.CmisRevision))      // CmisRevision
	page = append(page, 0x04)                      // MemoryModel + SteppedConfigOnly + MciMaxSpeed
	page = append(page, 0b0101)                    // ModuleState
	page = append(page, make([]byte, 5)...)        // FlagsSummary (Banks and others)
	var flagsIndicator byte = 0
	{
		if vcc < m.VccMonLowWarningThreshold {
			flagsIndicator = flagsIndicator | (0x1 << 7)
		}
		if vcc > m.VccMonHighWarningThreshold {
			flagsIndicator = flagsIndicator | (0x1 << 6)
		}
		if vcc < (m.VccMonLowWarningThreshold - VccMonAlarmThreshold) {
			flagsIndicator = flagsIndicator | (0x1 << 5)
		}
		if vcc > (m.VccMonHighWarningThreshold + VccMonAlarmThreshold) {
			flagsIndicator = flagsIndicator | (0x1 << 4)
		}
		if temperature < m.TempMonLowWarningThreshold {
			flagsIndicator = flagsIndicator | (0x1 << 3)
		}
		if temperature > m.TempMonHighWarningThreshold {
			flagsIndicator = flagsIndicator | (0x1 << 2)
		}
		if temperature < m.TempMonLowWarningThreshold-TempMonAlarmThreshold {
			flagsIndicator = flagsIndicator | (0x1 << 1)
		}
		if temperature > m.TempMonHighWarningThreshold+TempMonAlarmThreshold {
			flagsIndicator = flagsIndicator | 0x1
		}
	}
	page = append(page, flagsIndicator)     // Latched Flags
	page = append(page, make([]byte, 4)...) // Aux and Custom Flags
	tempMonValue := int16(temperature * 256 / 10)
	page = append(page, byte(tempMonValue>>8), byte(tempMonValue&0xFF)) // TempMonValue
	vccMonValue := uint16(vcc * 10000)
	page = append(page, byte(vccMonValue>>8), byte(vccMonValue&0xFF))           // VccMonVoltage
	page = append(page, make([]byte, 14)...)                                    // Aux + Custom + Global Controls
	page = append(page, 0xFF)                                                   // Module Level Masks (Vcc + Temp)
	page = append(page, make([]byte, 6)...)                                     // -||- (Aux + Custom) + CDB status
	page = append(page, byte(m.ModuleRevision>>8), byte(m.ModuleRevision&0xFF)) // Module Active Firmware Version
	page = append(page, make([]byte, 44)...)                                    // Fault Information + Reserved + Custom
	page = append(page, byte(m.MediaType))                                      // MediaType
	for i := 0; i < 8; i++ {
		page = append(page, 0xFF, 0x00, 0x00, 0x00) // AppDescriptors
	}
	page = append(page, make([]byte, 10)...) // Password Facilities + Page Mapping

	return
}

func (m *Module) Page00h() (page []byte) {
	page = append(page, byte(m.SFF8024Identifier))
	vendorName := (append(make([]byte, 0, 16), m.VendorName...))[0:16]
	page = append(page, vendorName...)                                         // VendorName
	page = append(page, 0xCC, 0xFA, 0xCE)                                      // VendorOUI
	page = append(page, append(vendorName[0:4], []byte("xx1234567890")...)...) // VendorPN
	page = append(page, 0x01, 0x23)                                            // VendorRev
	page = append(page, append(vendorName[0:4], []byte("xx1234567890")...)...) // VendorSN
	page = append(page, append([]byte(m.DateCode)[2:8], 0x00, 0x00)...)        // DateCode
	page = append(page, []byte("BEST_MEMES")...)                               // CLEI
	page = append(page, 0b11100000, byte(m.MaxPower))                          // ModulePowerCharacteristics
	page = append(page, 0x00, 0x07)                                            // CableAssemblyLinkLength + ConnectorType
	page = append(page, make([]byte, 6)...)                                    // Copper Cable Attenuation
	page = append(page, 0xfe, 0x00, 0x10)                                      // MediaLaneInformation + Cable Assembly Information + MediaInterfaceTechnology
	page = append(page, make([]byte, 9)...)                                    // Reserved+Custom
	page = append(page, checksum(page[0:93]))                                  // PageChecksum
	page = append(page, make([]byte, 33)...)                                   // Custom

	return
}

func (m *Module) Page01h() (page []byte) {
	page = append(page, byte((m.ModuleRevision>>8)-1), byte((m.ModuleRevision&0xFF)-1)) // ModuleInactiveFirmwareRevision
	page = append(page, byte(m.ModuleRevision>>8), byte(m.ModuleRevision&0xFF))         // ModuleHardwareRevision
	if m.LenghtSMF > 63 || (m.LenghtSMF >= 10 && m.LenghtSMF%10 == 0) {                 // LengthMultiplierSMF
		page = append(page, byte(((m.LenghtSMF/10)&0xFFFFFF)|(0x10<<6)))
	} else {
		page = append(page, byte((m.LenghtSMF&0xFFFFFF)|(0x1<<6)))
	}
	page = append(page, make([]byte, 5)...)                           // LengthOMs + Reserved
	page = append(page, 0x77, 0xDD, 0x00, 0x2F)                       // NominalWavelength + WavelengthTolerance
	page = append(page, 0b1000000)                                    // Supported Pages Advertising
	page = append(page, 0x04, 0x79, 0x00)                             // Durations Advertising + Module Characteristics Advertising
	page = append(page, byte(m.ModuleTempMax), byte(m.ModuleTempMin)) // ModuleTemp
	page = append(page, make([]byte, 7)...)                           // PropagationDelay + OperatingVoltageMin + Others
	page = append(page, 0b1000000)                                    // TransmitterIsTunable
	page = append(page, make([]byte, 3)...)                           // Others
	page = append(page, 0b11, 0b110)                                  // VccMonSupported + TempMonSupported + RxTxOpticalPowerMonSupported
	page = append(page, make([]byte, 6)...)                           // ???
	page = append(page, 0x79, 0x14)                                   // MaxDurationModulePwr + MaxDurationDPTxTurn
	page = append(page, make([]byte, 86)...)                          // MediaLaneAssignment + Custom + Reserved
	page = append(page, checksum(page[2:126]))                        // checksum

	return
}

func (m *Module) Page02h() (page []byte) {
	// Module-Level Monitor Thresholds (Temp)
	tempTemps := []int16{int16((m.TempMonHighWarningThreshold + TempMonAlarmThreshold) * 256),
		int16((m.TempMonLowWarningThreshold - TempMonAlarmThreshold) * 256),
		int16(m.TempMonHighWarningThreshold * 256),
		int16(m.TempMonLowWarningThreshold * 256),
	}
	for _, v := range tempTemps {
		page = append(page, byte(v>>8), byte(v&0xFF))
	}

	// Module-Level Monitor Thresholds (Vcc)
	tempVccs := []uint16{uint16((m.VccMonHighWarningThreshold + VccMonAlarmThreshold) * 10000),
		uint16((m.VccMonLowWarningThreshold - VccMonAlarmThreshold) * 10000),
		uint16(m.VccMonHighWarningThreshold * 10000),
		uint16(m.VccMonLowWarningThreshold * 10000),
	}
	for _, v := range tempVccs {
		page = append(page, byte(v>>8), byte(v&0xFF))
	}

	page = append(page, make([]byte, 32)...) // Aux + Custom

	// Module-Level Monitor Thresholds (OpticalPowerTx)
	tempOpticalTxs := []uint16{uint16(dbmTo01MicroWatt(m.OpticalPowerTxHighWarningThreshold) * OpticalTxRxAlarmThreshold),
		uint16(dbmTo01MicroWatt(m.OpticalPowerTxLowWarningThreshold) / OpticalTxRxAlarmThreshold),
		uint16(dbmTo01MicroWatt(m.OpticalPowerTxHighWarningThreshold)),
		uint16(dbmTo01MicroWatt(m.OpticalPowerTxLowWarningThreshold)),
	}
	for _, v := range tempOpticalTxs {
		page = append(page, byte(v>>8), byte(v&0xFF))
	}

	page = append(page, make([]byte, 8)...) // LaserBiasCurrent

	// Module-Level Monitor Thresholds (OpticalPowerRx)
	tempOpticalRxs := []uint16{uint16(dbmTo01MicroWatt(m.OpticalPowerRxHighWarningThreshold) * OpticalTxRxAlarmThreshold),
		uint16(dbmTo01MicroWatt(m.OpticalPowerRxLowWarningThreshold) / OpticalTxRxAlarmThreshold),
		uint16(dbmTo01MicroWatt(m.OpticalPowerRxHighWarningThreshold)),
		uint16(dbmTo01MicroWatt(m.OpticalPowerRxLowWarningThreshold)),
	}
	for _, v := range tempOpticalRxs {
		page = append(page, byte(v>>8), byte(v&0xFF))
	}

	page = append(page, make([]byte, 55)...)   // Reserved + Custom
	page = append(page, checksum(page[0:126])) // Page checksum

	return
}

func (m *Module) Page04h() (page []byte) {
	page = append(page, byte(1<<5))
	page = append(page, make([]byte, 21)...)    // Unsupported Grids
	page = append(page, 0xFF, 0xEE, 0x00, 0x1E) // GridChannel100GHz
	page = append(page, make([]byte, 44)...)    // Unsupported Grids + FineTuning
	minPwr, maxPwr := uint16(m.ProgOutputPowerMin*100), uint16(m.ProgOutputPowerMax*100)
	page = append(page, byte(minPwr>>8), byte(minPwr&0xFF)) // ProgOutputPowerMin
	page = append(page, byte(maxPwr>>8), byte(maxPwr&0xFF)) // ProgOutputPowerMax
	page = append(page, make([]byte, 53)...)                // Reserved
	page = append(page, checksum(page[0:126]))              // Page checksum

	return
}

func (m *Module) Page11h(txPower float64, rxPower float64) (page []byte) {
	for i := 0; i < 4; i++ {
		page = append(page, 0x44) // DPStateHostLane
	}
	page = append(page, 0xFF)               // OutputStatusRx
	page = append(page, make([]byte, 6)...) // OutputStatusTx + Lane-Specific State Changed Flags

	// Tx Flags
	if txPower > m.OpticalPowerTxHighWarningThreshold*OpticalTxRxAlarmThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	if txPower < m.OpticalPowerTxLowWarningThreshold/OpticalTxRxAlarmThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	if txPower > m.OpticalPowerTxHighWarningThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	if txPower < m.OpticalPowerTxLowWarningThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	page = append(page, make([]byte, 6)...) // LaserBias + LOS + CDRLOL

	// Rx Flags
	if rxPower > m.OpticalPowerRxHighWarningThreshold*OpticalTxRxAlarmThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	if rxPower < m.OpticalPowerRxLowWarningThreshold/OpticalTxRxAlarmThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	if rxPower > m.OpticalPowerRxHighWarningThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}
	if rxPower < m.OpticalPowerRxLowWarningThreshold {
		page = append(page, 0x01)
	} else {
		page = append(page, 0x00)
	}

	page = append(page, 0x00) // OutputStatusChangedFlagRx
	txPower01microW := uint16(dbmTo01MicroWatt(txPower) + 2*rand.Float64() - 1)
	page = append(page, byte(txPower01microW>>8), byte(txPower01microW&0xFF)) // OpticalPowerTx1
	page = append(page, make([]byte, 30)...)                                  // OpticalPowerTx2-8 + LaserBiasTx
	rxPower01microW := uint16(dbmTo01MicroWatt(rxPower) + 2*rand.Float64() - 1)
	page = append(page, byte(rxPower01microW>>8), byte(rxPower01microW&0xFF)) // OpticalPowerRx1
	page = append(page, make([]byte, 14)...)                                  // OpticalPowerRx2-8
	for i := 0; i < 4; i++ {
		page = append(page, 0x11) // ConfigStatusLane
	}
	page = append(page, make([]byte, 8)...)  // AppSelCodeLane
	page = append(page, make([]byte, 26)...) // Indicators for Active Control Set + Data Path Conditions + Reserved
	page = append(page, 0x11)                // MediaLaneToWavelengthMappingTx1
	page = append(page, make([]byte, 7)...)  // MediaLaneToWavelengthMappingTx2-8
	page = append(page, 0x11)                // MediaLaneToWavelengthMappingRx1
	page = append(page, make([]byte, 7)...)  // MediaLaneToWavelengthMappingRx2-8

	return
}

func (m *Module) Page12h() (page []byte) {
	page = append(page, 0b0101<<4)          // GridSpacingTx1
	page = append(page, make([]byte, 7)...) // GridSpacingTx2-8
	channelNumber := int16((m.CurrentLaserFrequencyTxx - 193100000) / 100)
	page = append(page, byte(channelNumber>>8), byte(channelNumber&0xFF)) // ChannelNumberTx1
	page = append(page, make([]byte, 30)...)
	freq := uint32(m.CurrentLaserFrequencyTxx)                                                               // ChannelNumberTx2-8 + FineTuningOffsetTx
	page = append(page, byte((freq>>24)&0xFF), byte((freq>>16)&0xFF), byte((freq>>8)&0xFF), byte(freq&0xFF)) // CurrentLaserFrequencyTx1
	page = append(page, make([]byte, 28)...)
	pwr := int16(m.TargetOutputPowerTxx)
	page = append(page, byte(pwr>>8), byte(pwr&0xFF)) // TargetOutputPowerTxx1
	page = append(page, make([]byte, 54)...)          // TargetOutputPowerTxx2-8

	return
}

func (m *Module) Page25h(osnr float64, temperature float64) (page []byte) {
	page = append(page, make([]byte, 22)...)
	if osnr == 0.0 { // VDM real-time OSNR
		page = append(page, 0x00, 0x00)
	} else {
		modOsnr := uint16(10*osnr + (float64(rand.Intn(2*int(temperature))-int(temperature)) / 3))
		page = append(page, []byte{byte(modOsnr >> 8), byte(modOsnr & 0xFF)}...)
	}
	page = append(page, make([]byte, 104)...)

	return
}

func checksum(data []byte) byte {
	cs := md5.Sum(data)
	return cs[len(cs)-1]
}

func dbmTo01MicroWatt(dbm float64) float64 {
	return math.Pow(10, (dbm+40)/10)
}
