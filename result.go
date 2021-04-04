package powtcha

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/powtcha/powtcha/constants"
	"strings"
	"time"
)

type Diagnostic struct {
	Implementation byte
	Runtime uint16
}

type Result struct {
	Puzzle Puzzle
	Solutions []Solution
	Diagnostic *Diagnostic
}

func (r *Result) Valid(appId uint32) bool {
	if appId > 0 && r.Puzzle.AppID != appId {
		return false
	}
	if time.Now().Sub(time.Unix(int64(r.Puzzle.Timestamp), 0)) > time.Duration(r.Puzzle.Expiry) * time.Minute {
		return false
	}
	T := r.Puzzle.T()
	header, err := r.Puzzle.Marshal()
	if err != nil {
		return false
	}
	duplicates := make(map[string]bool)
	valid := true
	for _, solution := range r.Solutions {
		if _, ok := duplicates[string(solution[:])]; ok {
			valid = false
			break
		}
		duplicates[string(solution[:])] = true
		if !solution.Valid(T, header) {
			valid = false
			break
		}
	}
	return valid
}

func (r *Result) Marshal() ([]byte, []byte) {
	solutions := make([]byte, len(r.Solutions) * 8)
	for i, solution := range r.Solutions {
		copy(solutions[i*8:], solution[:])
	}
	if r.Diagnostic == nil {
		return solutions, nil
	}
	diag := make([]byte, 3)
	diag[0] = r.Diagnostic.Implementation
	diag[1] = byte(r.Diagnostic.Runtime >> 8)
	diag[2] = byte(r.Diagnostic.Runtime)
	return solutions, diag
}

func (r *Result) Encode(secret []byte, encodedPuzzle string) (string, error) {
	var err error
	header := encodedPuzzle
	if header == "" {
		header, err = r.Puzzle.Encode(secret)
		if err != nil {
			return "", err
		}
	}
	solutions, diag := r.Marshal()
	return header +
		constants.EncodeSperationChar +
		base64.URLEncoding.EncodeToString(solutions) +
		constants.EncodeSperationChar +
		base64.URLEncoding.EncodeToString(diag), nil
}

func DecodeResult(data string, secret []byte) (*Result, error) {
	puzzle, err := DecodePuzzle(data, secret)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(data, constants.EncodeSperationChar)
	if len(parts) < 3 {
		return nil, errors.New("too few parts for a result")
	}

	sols, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid b64 encoded solutions")
	}
	if len(sols) % 8 != 0 {
		return nil, errors.New("weird solution length")
	}
	if len(sols) / 8 != int(puzzle.Problems) {
		return nil, errors.New("number of solutions doesn't match puzzle")
	}

	solutions := make([]Solution, puzzle.Problems)
	var s Solution
	for i := 0; i < int(puzzle.Problems); i++ {
		copy(s[:], sols[i*8:i*8+8])
		solutions[i] = s
	}

	result := Result{
		Puzzle:    puzzle,
		Solutions: solutions,
	}

	if len(parts) > 3 {
		if diag, err := base64.URLEncoding.DecodeString(parts[3]); err == nil {
			result.Diagnostic = &Diagnostic{
				Implementation: diag[0],
				Runtime:        binary.BigEndian.Uint16(diag[1:3]),
			}
		}
	}

	return &result, nil
}
