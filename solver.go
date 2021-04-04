package powtcha

import (
	"encoding/binary"
	"errors"
	"github.com/powtcha/powtcha/util"
	"golang.org/x/crypto/blake2b"
)

func SolveStep(header []byte, T uint32, start []byte) (Solution, error) {
	if len(start) != 8 {
		return [8]byte{}, errors.New("invalid start, must be 8 bytes")
	}

	i := util.BytesToUint64(start) + 1

	solution := make([]byte, 128)
	nulls := make([]byte, 0)
	copy(solution, header)
	var r Solution
	for ; i > 0; i++ { // Should roll-over and be zero again -> break condition
		copy(solution[120:], nulls)
		for h := 0; h < 8; h++ {
			solution[120 + h] = byte(i >> (8 * (7 - h)))
		}
		h := blake2b.Sum256(solution)
		x := binary.LittleEndian.Uint32(h[:4])
		if x < T {
			copy(r[:], solution[120:])
			return r, nil
		}
	}
	return [8]byte{}, errors.New("no solution found")
}
