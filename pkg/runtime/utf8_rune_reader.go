package runtime

import (
	"errors"
	"io"
	"unicode/utf8"
)

// ErrInvalidUTF8Encoding is an error used when decoding a UTF-8 rune fails.
var ErrInvalidUTF8Encoding = errors.New("invalid UTF-8 encoding")

// UTF8RuneReader extracts individual runes from the source provided. It does decode UTF-8 encoded runes as needed.
// It keeps track of the overall byte offset of the rune from the start.
// It keeps track of the line and column the rune is located in.
// It can safely be copied by value to retain the reader state as needed.
type UTF8RuneReader struct {
	source     []byte
	byteOffset int

	currRune     rune
	currRuneSize int
	line         int
	column       int

	err error
}

// NewUTF8RuneReader creates a new instance with the given source.
func NewUTF8RuneReader(source []byte) UTF8RuneReader {
	return UTF8RuneReader{
		source:   source,
		currRune: utf8.RuneError,
		line:     1, // we need to start out on line 1 instead of 0
		column:   1, // we need to start out on column 1 instead of 0
	}
}

// Next consumes the next UTF-8 rune from the source.
// It returns true when it successfully consumed a rune.
// It returns false when no more runes are available to consume.
// Note that in the case of invalid UTF-8 encodings, it still returns true. Check the rune value for utf8.RuneError and
// the Err() return value to detect those situations.
func (r *UTF8RuneReader) Next() bool {
	r.err = nil
	r.byteOffset += r.currRuneSize
	if r.currRuneSize > 0 {
		r.column++
	}

	if r.currRune == '\n' {
		// the last rune was a newline, update our bookkeeping
		r.line++
		r.column = 1
	}

	if r.byteOffset >= len(r.source) {
		r.currRune = utf8.RuneError
		r.currRuneSize = 0
		r.err = io.EOF
		return false
	}

	// we initially assume that we are dealing with ASCII runes
	r.currRune, r.currRuneSize = rune(r.source[r.byteOffset]), 1
	if r.currRune >= utf8.RuneSelf {
		// we now noticed that we are not dealing with ASCII runes, so we need to decode the next UTF-8 rune
		r.currRune, r.currRuneSize = utf8.DecodeRune(r.source[r.byteOffset:])
		if r.currRune == utf8.RuneError && r.currRuneSize == 1 {
			r.err = ErrInvalidUTF8Encoding
		}
	}
	return true
}

// Err returns the error which occurred on the last call to Next().
// It will return io.EOF after Next() has returned false.
func (r *UTF8RuneReader) Err() error {
	return r.err
}

// Rune returns the rune which was consumed on the last call to Next().
// It will return utf8.RuneError after Next() has returned false or an invalid encoded UTF-8 rune has been consumed.
func (r *UTF8RuneReader) Rune() rune {
	return r.currRune
}

// RuneSize returns the size of the rune in bytes.
func (r *UTF8RuneReader) RuneSize() int {
	return r.currRuneSize
}

// ByteOffset returns the start of the rune in bytes from the source start.
// It returns the location after the last byte after the call to Next() returned false.
func (r *UTF8RuneReader) ByteOffset() int {
	return r.byteOffset
}

// Line returns the line the rune is located on.
func (r *UTF8RuneReader) Line() int {
	return r.line
}

// Column returns the column the rune is located on.
func (r *UTF8RuneReader) Column() int {
	return r.column
}

// Lexeme returns the bytes which make up the lexeme.
func (r *UTF8RuneReader) Lexeme(start, end int) []byte {
	return r.source[start:end]
}
