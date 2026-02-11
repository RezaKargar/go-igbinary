// Package igbinary provides a pure Go decoder for PHP's igbinary serialization format.
//
// igbinary is a compact binary serializer for PHP values that replaces PHP's standard
// serialize() with a faster, smaller binary representation. It is commonly used with
// PHP's memcached extension to store serialized data in cache.
//
// This package allows Go programs to decode igbinary-serialized data produced by PHP,
// making it possible to read PHP memcached cache entries, session data, or any other
// igbinary-encoded payloads directly from Go.
//
// # Type Mapping
//
// PHP types are decoded into Go types as follows:
//
//   - PHP array       -> map[string]any  (integer keys are converted to string keys)
//   - PHP string      -> string
//   - PHP integer     -> int64
//   - PHP float       -> float64
//   - PHP boolean     -> bool
//   - PHP NULL        -> nil
//   - PHP object      -> map[string]any  (class name stored under "__class" key)
//
// # Quick Start
//
//	data := []byte{0x00, 0x00, 0x00, 0x02, 0x06, 0x2a} // igbinary-encoded int(42)
//	val, err := igbinary.Decode(data)
//	// val == int64(42)
//
// # Decoder Options
//
// For advanced usage, create a [Decoder] with options:
//
//	dec := igbinary.NewDecoder(
//	    igbinary.WithStrictMode(true),
//	)
//	val, err := dec.Decode(data)
//
// # Sub-packages
//
// The [github.com/RezaKargar/go-igbinary/memcached] sub-package provides a full
// PHP memcached interop layer that handles decompression (FastLZ, Zlib) and flag-based
// dispatch on top of this decoder.
package igbinary
