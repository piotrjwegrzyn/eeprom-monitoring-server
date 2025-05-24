package monitor

import (
	"encoding/hex"
)

type Decoder func([]byte) (Eeprom, error)

func DefaultDecoder() Decoder {
	return func(input []byte) (Eeprom, error) {
		var temp []byte
		for i := 0; i < len(input); i += 33 {
			temp = append(temp, input[i:i+32]...)
		}

		return hex.DecodeString(string(temp))
	}
}
