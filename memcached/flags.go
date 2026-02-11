// Package memcached provides a codec for decoding PHP memcached cache entries in Go.
//
// PHP's memcached PECL extension stores each cache value alongside a 32-bit flags
// field that encodes the serializer type and compression algorithm used. This package
// abstracts the two-stage pipeline (decompress then deserialize) behind pluggable
// interfaces, with built-in support for igbinary, JSON, FastLZ, and Zlib.
//
// # Quick Start
//
//	codec := memcached.NewCodec()
//	val, err := codec.Decode(item.Value, item.Flags)
package memcached

import "fmt"

// PHP memcached flag constants.
//
// These match the PHP memcached PECL extension's internal flag format.
// See: https://github.com/php-memcached-dev/php-memcached/blob/master/php_memcached_private.h

// Serializer type flags (lower 4 bits of the flags field).
const (
	FlagString     uint32 = 0 // Raw string
	FlagLong       uint32 = 1 // Integer
	FlagDouble     uint32 = 2 // Float
	FlagBool       uint32 = 3 // Boolean
	FlagSerialized uint32 = 4 // PHP serialize()
	FlagIgbinary   uint32 = 5 // igbinary
	FlagJSON       uint32 = 6 // JSON
	FlagMsgpack    uint32 = 7 // Msgpack
)

// Compression flags (bit flags in the upper nibble of the low byte).
const (
	FlagCompressed uint32 = 1 << 4 // Data is compressed (16)
	FlagZlib       uint32 = 1 << 5 // zlib compression (32)
	FlagFastlz     uint32 = 1 << 6 // FastLZ compression (64)
	FlagZstd       uint32 = 1 << 7 // Zstandard compression (128)
)

// Masks for extracting parts of the flags.
const (
	SerializerMask  uint32 = 0x0F // Lower 4 bits: serializer type
	CompressionMask uint32 = FlagZlib | FlagFastlz | FlagZstd
)

// SerializerType extracts the serializer type from cache flags.
func SerializerType(flags uint32) uint32 {
	return flags & SerializerMask
}

// IsCompressed returns true if the compressed flag is set.
func IsCompressed(flags uint32) bool {
	return flags&FlagCompressed != 0
}

// SerializerName returns a human-readable name for the serializer type.
func SerializerName(flags uint32) string {
	switch SerializerType(flags) {
	case FlagString:
		return "string"
	case FlagLong:
		return "long"
	case FlagDouble:
		return "double"
	case FlagBool:
		return "bool"
	case FlagSerialized:
		return "php_serialize"
	case FlagIgbinary:
		return "igbinary"
	case FlagJSON:
		return "json"
	case FlagMsgpack:
		return "msgpack"
	default:
		return fmt.Sprintf("unknown(%d)", SerializerType(flags))
	}
}

// CompressionName returns a human-readable name for the compression type.
func CompressionName(flags uint32) string {
	if !IsCompressed(flags) {
		return "none"
	}
	switch {
	case flags&FlagFastlz != 0:
		return "fastlz"
	case flags&FlagZlib != 0:
		return "zlib"
	case flags&FlagZstd != 0:
		return "zstd"
	default:
		return "unknown"
	}
}

// ExplainFlags returns a human-readable explanation of cache flags.
func ExplainFlags(flags uint32) string {
	return fmt.Sprintf("type=%s compressed=%v compression=%s (raw=0x%08x)",
		SerializerName(flags), IsCompressed(flags), CompressionName(flags), flags)
}
