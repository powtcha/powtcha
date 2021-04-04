package powtcha

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/cryptobyte"
	"math"
	"powtcha/byteparser"
	"powtcha/constants"
	"strings"
	"time"
)

type Puzzle struct {
	Timestamp  uint32 `json:"timestamp"`
	AppID      uint32 `json:"app_id"`
	Version    byte   `json:"version"`
	Expiry     byte   `json:"expiry"`
	Problems   byte   `json:"problems"`
	Difficulty byte   `json:"difficulty"`
	Nonce      []byte `json:"nonce"`
}

func (p *Puzzle) Marshal() ([]byte, error) {
	var buf []byte
	b := cryptobyte.NewBuilder(buf)
	b.AddUint32(p.Timestamp)
	b.AddUint32(p.AppID)
	b.AddUint8(p.Version)
	b.AddUint8(p.Expiry)
	b.AddUint8(p.Problems)
	b.AddUint8(p.Difficulty)
	b.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) {
		b.AddBytes(p.Nonce)
	})
	return b.Bytes()
}

func (p * Puzzle) T() uint32 {
	return uint32(math.Pow(2, (255.999 - float64(p.Difficulty)) / 8.0))
}

func (p * Puzzle) FindSolutions() ([]Solution, error) {
	var sols []Solution
	T := p.T()
	header, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	solution := make([]byte, 128)
	nulls := make([]byte, 0)
	copy(solution, header)
	var r Solution
	for i := uint64(1); i != 0; i++ { // Should roll-over and be zero again -> break condition
		copy(solution[120:], nulls)
		for h := 0; h < 8; h++ {
			solution[120 + h] = byte(i >> (8 * h))
		}
		h := blake2b.Sum256(solution)
		x := binary.LittleEndian.Uint32(h[:4])
		if x < T {
			copy(r[:], solution[120:])
			sols = append(sols, r)
			if byte(len(sols)) >= p.Problems {
				break
			}
		}
	}
	return sols, nil
}

func (p *Puzzle) GetResult() (*Result, error) {
	start := time.Now()
	solutions, err := p.FindSolutions()
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)
	millis := elapsed.Milliseconds()
	if millis > 65535 {
		millis = 65535
	}
	return &Result{
		Puzzle:     *p,
		Solutions:  solutions,
		Diagnostic: &Diagnostic{
			Implementation: 0,
			Runtime:        uint16(millis),
		},
	}, nil
}


func NewPuzzle(appID uint32, validity time.Duration, problems byte, difficulty byte) (*Puzzle, error) {
	expiry := validity / time.Minute
	if expiry < 1 {
		expiry = 1
	}
	if expiry > 255 {
		return nil, errors.New("max expiry is 255 minutes")
	}
	nonce := make([]byte, 8, 8)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}
	return &Puzzle{
		Timestamp:  uint32(time.Now().Unix()),
		AppID:      appID,
		Version:    1,
		Expiry:     byte(expiry),
		Problems:   problems,
		Difficulty: difficulty,
		Nonce:      nonce,
	}, nil
}

func UnmarshalPuzzle(puzzle *Puzzle, data []byte) (err error) {
	parser := byteparser.NewParser(data)
	if puzzle.Timestamp, err = parser.GetUint32(); err != nil {
		return
	}
	if puzzle.AppID, err = parser.GetUint32(); err != nil {
		return
	}
	if puzzle.Version, err = parser.GetUint8(); err != nil {
		return
	}
	if puzzle.Expiry, err = parser.GetUint8(); err != nil {
		return
	}
	if puzzle.Problems, err = parser.GetUint8(); err != nil {
		return
	}
	if puzzle.Difficulty, err = parser.GetUint8(); err != nil {
		return
	}
	if err = parser.GetUint8LengthPrefixed(func(p *byteparser.Parser) error {
		puzzle.Nonce, err = p.GetBytes(p.Remaining())
		return err
	}); err != nil {
		return
	}

	return nil
}

func (p *Puzzle) Encode(secret []byte) (string, error) {
	body, err := p.Marshal()
	if err != nil {
		return "", err
	}
	sig := SignPuzzle(body, secret)
	return base64.URLEncoding.EncodeToString(sig) +
		constants.EncodeSperationChar +
		base64.URLEncoding.EncodeToString(body), nil
}

func SignPuzzle(data, secret []byte) []byte {
	return hmac.New(sha256.New, secret).Sum(data)
}

func DecodePuzzle(data string, secret []byte) (Puzzle, error) {
	parts := strings.Split(data, constants.EncodeSperationChar)
	if len(parts) < 2 {
		return Puzzle{}, errors.New("too few parts to parse puzzle")
	}

	sig, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return Puzzle{}, errors.New("invalid b64 encoded signature")
	}

	header, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return Puzzle{}, errors.New("invalid b64 encoded puzzle")
	}

	if secret != nil && !bytes.Equal(SignPuzzle(header, secret), sig) {
		return Puzzle{}, errors.New("signature mismatch")
	}

	var puzzle Puzzle
	if err := UnmarshalPuzzle(&puzzle, header); err != nil {
		return Puzzle{}, errors.New("invalid puzzle bytes")
	}

	return puzzle, nil
}



