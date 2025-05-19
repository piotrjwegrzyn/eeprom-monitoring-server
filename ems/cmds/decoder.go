package cmds

import (
	"encoding/hex"
	"math"
)

func defaultDecoder(input []byte) (Eeprom, error) {
	var temp []byte
	for i := 0; i < len(input); i += 33 {
		temp = append(temp, input[i:i+32]...)
	}

	return hex.DecodeString(string(temp))
}

func microWatt01ToDbm(mw01 uint16) float64 {
	return 10 * math.Log10(float64(mw01)/10000)
}
