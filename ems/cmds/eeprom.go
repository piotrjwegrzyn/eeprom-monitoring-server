package cmds

const (
	PAGE_LEN int = 128

	PAGE_LOW_TEMP   int = 0*PAGE_LEN + 0x0E
	PAGE_LOW_VCC    int = 0*PAGE_LEN + 0x10
	PAGE_11H_TX_PWR int = 5*PAGE_LEN + 0x1A
	PAGE_11H_RX_PWR int = 5*PAGE_LEN + 0x3A
	PAGE_25H_OSNR   int = 7*PAGE_LEN + 0x16
)

type Eeprom []byte

func (e Eeprom) Temperature() float64 {
	tempMonValue := int16(e[PAGE_LOW_TEMP])<<8 | int16(e[PAGE_LOW_TEMP+1])

	return float64(tempMonValue) * 10 / 256
}

func (e Eeprom) Voltage() float64 {
	vccMonValue := uint16(e[PAGE_LOW_VCC])<<8 | uint16(e[PAGE_LOW_VCC+1])

	return float64(vccMonValue) / 10000
}

func (e Eeprom) TxPower() float64 {
	txPower01microW := uint16(e[PAGE_11H_TX_PWR])<<8 | uint16(e[PAGE_11H_TX_PWR+1])

	return microWatt01ToDbm(txPower01microW)
}

func (e Eeprom) RxPower() float64 {
	rxPower01microW := uint16(e[PAGE_11H_RX_PWR])<<8 | uint16(e[PAGE_11H_RX_PWR+1])

	return microWatt01ToDbm(rxPower01microW)
}

func (e Eeprom) Osnr() float64 {
	osnr := uint16(e[PAGE_25H_OSNR])<<8 | uint16(e[PAGE_25H_OSNR+1])

	return float64(osnr) / 10
}
