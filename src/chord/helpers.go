package chord

import (
	"bytes"
	"crypto/sha1"
	"math"
	"encoding/binary"
)

type File struct {
	Name	string
	Data	[]byte
}

func GetHash(key string, m int) int {
	var hash int;
	mod := int(math.Exp2(float64(m)))
	byteHash := sha1.Sum([]byte(key))
	buffer := bytes.NewBuffer(byteHash[:m])
	binary.Read(buffer, binary.LittleEndian, &hash)
	return hash % mod
}