package powtcha

import (
	assertLib "github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestCreatePuzzle(t *testing.T) {
	assert := assertLib.New(t)
	puzzle, err := NewPuzzle( 0, 255, 8, 140)
	log.Println(puzzle.Marshal())
	assert.NoError(err, "unable to create puzzle")
	str, err := puzzle.Encode(nil)
	assert.NoError(err, "unable to encode puzzle")
	log.Println(str)
}

func TestPuzzleSolvable(t *testing.T) {
	assert := assertLib.New(t)
	puzzle, err := NewPuzzle( 0, 5, 12, 140)
	assert.NoError(err, "unable to create puzzle")
	sols, err := puzzle.FindSolutions()
	assert.NoError(err, "unable to find solutions")
	assert.Len(sols, int(puzzle.Problems))
}

func TestPuzzleResultValid(t *testing.T) {
	assert := assertLib.New(t)
	puzzle, err := NewPuzzle( 0, 5, 12, 60)
	assert.NoError(err, "unable to create puzzle")
	res, err := puzzle.GetResult()
	assert.NoError(err, "unable to find solutions")
	assert.Len(res.Solutions, int(puzzle.Problems))
	assert.True(res.Valid(), "invalid solution found")
}

func TestPuzzleResultParsable(t *testing.T) {
	assert := assertLib.New(t)
	puzzle, err := NewPuzzle( 0, 5, 4, 140)
	assert.NoError(err, "unable to create puzzle")
	res, err := puzzle.GetResult()
	assert.NoError(err, "unable to find solutions")
	assert.Len(res.Solutions, int(puzzle.Problems))
	assert.True(res.Valid(), "invalid solution found")
	str, err := res.Encode(nil, "")
	assert.NoError(err, "unable to encode result")
	result, err := DecodeResult(str, nil)
	assert.NoError(err, "unable to decode result")
	assert.True(result.Valid(), "parsed result is now invalid ¯\\_(ツ)_/¯")
	log.Println("calculation took", result.Diagnostic.Runtime, "ms")
}

func TestPuzzleResultParsableFromBrowser(t *testing.T) {
	browser := ""
	if browser == "" {
		return
	}
	assert := assertLib.New(t)
	result, err := DecodeResult(browser, nil)
	log.Println(result.Puzzle.Nonce)
	log.Println(result.Puzzle.Marshal())
	assert.NoError(err, "unable to decode result")
	assert.True(result.Valid(), "parsed result is invalid")
	log.Println("calculation took", result.Diagnostic.Runtime, "ms")
}

func TestDecodePuzzle(t *testing.T) {
	assert := assertLib.New(t)
	puzzle, err := NewPuzzle( 0, 0, 0, 160)
	assert.NoError(err, "unable to create puzzle")
	body, _ := puzzle.Marshal()
	log.Println(body)
	str, err := puzzle.Encode(nil)
	assert.NoError(err, "unable to encode puzzle")
	parsed, err := DecodeResult(str + ".", nil)
	assert.NoError(err, "unable to decode puzzle")
	log.Println(parsed.Puzzle.Nonce)
	assert.Equal(puzzle.Nonce, parsed.Puzzle.Nonce)
}