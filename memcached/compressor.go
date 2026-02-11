package memcached

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"

	fastlz "github.com/dgryski/go-fastlz"
)

// Compressor handles decompression of cache values.
//
// Implement this interface to add support for custom compression algorithms
// (e.g., zstd, lz4, snappy).
type Compressor interface {
	// Decompress decompresses data and returns the original bytes.
	Decompress(data []byte) ([]byte, error)
}

// FastlzCompressor decompresses data using the FastLZ algorithm.
//
// This is the default compressor used by PHP's memcached extension.
type FastlzCompressor struct{}

// Decompress decompresses FastLZ-compressed data.
// Expects the go-fastlz framing format (4-byte uncompressed length prefix).
func (c *FastlzCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("memcached: fastlz data too short: %d bytes", len(data))
	}
	result, err := fastlz.Decode(nil, data)
	if err != nil {
		return nil, fmt.Errorf("memcached: fastlz decompress: %w", err)
	}
	return result, nil
}

// ZlibCompressor decompresses data using the zlib algorithm.
type ZlibCompressor struct {
	// LengthPrefixed indicates whether the data has a 4-byte little-endian
	// uncompressed length prefix before the zlib stream. Defaults to true,
	// matching the PHP memcached extension format.
	LengthPrefixed bool
}

// NewZlibCompressor creates a ZlibCompressor.
// If lengthPrefixed is true, expects 4-byte LE uncompressed length + zlib stream.
func NewZlibCompressor(lengthPrefixed bool) *ZlibCompressor {
	return &ZlibCompressor{LengthPrefixed: lengthPrefixed}
}

// Decompress decompresses zlib-compressed data.
func (c *ZlibCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("memcached: zlib data too short: %d bytes", len(data))
	}

	var uncompressedLen uint32
	zlibData := data

	if c.LengthPrefixed {
		uncompressedLen = binary.LittleEndian.Uint32(data[:4])
		zlibData = data[4:]
	}

	r, err := zlib.NewReader(bytes.NewReader(zlibData))
	if err != nil {
		// Fallback: try without length prefix
		if c.LengthPrefixed {
			r, err = zlib.NewReader(bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("memcached: zlib reader: %w", err)
			}
		} else {
			return nil, fmt.Errorf("memcached: zlib reader: %w", err)
		}
	}
	defer func(r io.ReadCloser) {
		_ = r.Close() //nolint:errcheck
	}(r)

	bufSize := int(uncompressedLen)
	if bufSize == 0 {
		bufSize = len(data) * 2
	}
	result := make([]byte, 0, bufSize)
	buf := make([]byte, 4096)
	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("memcached: zlib read: %w", readErr)
		}
	}
	return result, nil
}
