package route

import (
	"encoding/binary"
	"errors"
)

// numberToBytes converts the provided number to a byte array
func numberToBytes(n interface{}) []byte {
	switch i := n.(type) {
	case uint64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, i)
		return b
	case uint32:
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, i)
		return b
	case uint16:
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, i)
		return b
	case uint8:
		b := []byte{byte(i)}
		return b
	default:
		panic(errors.New("Unsupported type"))
	}
}
