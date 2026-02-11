package igbinary

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by the decoder.
var (
	// ErrDataTooShort is returned when the input data is shorter than the minimum
	// required length (5 bytes: 4-byte header + at least 1 type byte).
	ErrDataTooShort = errors.New("igbinary: data too short")

	// ErrInvalidHeader is returned when the 4-byte header does not match the
	// expected igbinary format version.
	ErrInvalidHeader = errors.New("igbinary: invalid header")

	// ErrUnexpectedEnd is returned when the decoder reaches the end of data
	// while still expecting more bytes.
	ErrUnexpectedEnd = errors.New("igbinary: unexpected end of data")

	// ErrUnknownType is returned when the decoder encounters an unrecognized
	// type code byte.
	ErrUnknownType = errors.New("igbinary: unknown type code")

	// ErrStringIDOutOfRange is returned when a string back-reference ID exceeds
	// the number of strings seen so far.
	ErrStringIDOutOfRange = errors.New("igbinary: string ID out of range")

	// ErrInvalidObjectProperties is returned when an object's property count
	// is not encoded as an array type code.
	ErrInvalidObjectProperties = errors.New("igbinary: invalid object property encoding")

	// ErrInvalidSerializedData is returned when a serialized object's data length
	// is not encoded as a string type code.
	ErrInvalidSerializedData = errors.New("igbinary: invalid serialized data encoding")

	// ErrUnsupportedArrayKey is returned when an array key uses an unsupported
	// type code (only string and integer keys are supported).
	ErrUnsupportedArrayKey = errors.New("igbinary: unsupported array key type")
)

// DecodeError wraps a sentinel error with positional context about where
// in the binary stream the error occurred.
type DecodeError struct {
	// Err is the underlying sentinel error.
	Err error
	// Pos is the byte offset in the input where the error was detected.
	Pos int
	// Detail provides additional context about the error.
	Detail string
}

// Error returns a human-readable description of the decode error.
func (e *DecodeError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s at pos %d: %s", e.Err.Error(), e.Pos, e.Detail)
	}
	return fmt.Sprintf("%s at pos %d", e.Err.Error(), e.Pos)
}

// Unwrap returns the underlying sentinel error, enabling errors.Is() matching.
func (e *DecodeError) Unwrap() error {
	return e.Err
}

// newError creates a DecodeError with position and optional detail.
func newError(err error, pos int, detail string) *DecodeError {
	return &DecodeError{Err: err, Pos: pos, Detail: detail}
}
