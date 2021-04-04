package powtcha

import (
	"encoding/binary"
	"golang.org/x/crypto/blake2b"
)

type Solution [8]byte

func (s *Solution) Valid(T uint32, header []byte) bool {
	solution := make([]byte, 128)
	copy(solution, header)
	for h := 0; h < 8; h++ {
		solution[120 + h] = s[h]
	}
	h := blake2b.Sum256(solution)
	x := binary.LittleEndian.Uint32(h[:4])
	return x < T
}
