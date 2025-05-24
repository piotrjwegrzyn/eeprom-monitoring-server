package monitor

import "math"

const (
	PageLength int = 128

	PageLowTemp  int = 0*PageLength + 0x0E
	PageLowVcc   int = 0*PageLength + 0x10
	Page11hTxPwr int = 5*PageLength + 0x1A
	Page11hRxPwr int = 5*PageLength + 0x3A
	Page25hOsnr  int = 7*PageLength + 0x16
)

type Eeprom []byte

func (e Eeprom) Temperature() float64 {
	tempMonValue := int16(e[PageLowTemp])<<8 | int16(e[PageLowTemp+1])

	return float64(tempMonValue) * 10 / 256
}

func (e Eeprom) Voltage() float64 {
	vccMonValue := uint16(e[PageLowVcc])<<8 | uint16(e[PageLowVcc+1])

	return float64(vccMonValue) / 10000
}

func (e Eeprom) TxPower() float64 {
	txPower01microW := uint16(e[Page11hTxPwr])<<8 | uint16(e[Page11hTxPwr+1])

	return microWatt01ToDbm(txPower01microW)
}

func (e Eeprom) RxPower() float64 {
	rxPower01microW := uint16(e[Page11hRxPwr])<<8 | uint16(e[Page11hRxPwr+1])

	return microWatt01ToDbm(rxPower01microW)
}

func (e Eeprom) Osnr() float64 {
	osnr := uint16(e[Page25hOsnr])<<8 | uint16(e[Page25hOsnr+1])

	return float64(osnr) / 10
}

func microWatt01ToDbm(mw01 uint16) float64 {
	return 10 * math.Log10(float64(mw01)/10000)
}
