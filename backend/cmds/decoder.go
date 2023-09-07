package cmds

import "encoding/hex"

func defaultDecoder(input []byte) ([]byte, error) {
	temp := []byte{}
	for i := 0; i < len(input); i += 33 {
		temp = append(temp, input[i:i+32]...)
	}

	return hex.DecodeString(string(temp))
}
