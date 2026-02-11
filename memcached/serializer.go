package memcached

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	igbinary "github.com/RezaKargar/go-igbinary"
)

// Serializer handles deserialization of cache values.
//
// Implement this interface to add support for custom serialization formats
// (e.g., msgpack, protobuf, php serialize).
type Serializer interface {
	// Deserialize converts raw bytes into a Go value.
	Deserialize(data []byte) (any, error)
}

// IgbinarySerializer deserializes igbinary-encoded data using the
// [github.com/RezaKargar/go-igbinary] package.
type IgbinarySerializer struct{}

// Deserialize decodes igbinary data into native Go types.
func (s *IgbinarySerializer) Deserialize(data []byte) (any, error) {
	return igbinary.Decode(data)
}

// StringSerializer returns the raw bytes as a Go string (no transformation).
type StringSerializer struct{}

// Deserialize returns the data as a string.
func (s *StringSerializer) Deserialize(data []byte) (any, error) {
	return string(data), nil
}

// LongSerializer deserializes integer values stored as their string representation.
//
// PHP's memcached extension stores integers this way when the value is a scalar
// (flags type = FlagLong). For example, the integer 42 is stored as the bytes "42".
type LongSerializer struct{}

// Deserialize parses a string-encoded integer into int64.
func (s *LongSerializer) Deserialize(data []byte) (any, error) {
	str := strings.TrimSpace(string(data))
	if str == "" {
		return int64(0), nil
	}
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("memcached: long deserialize %q: %w", str, err)
	}
	return v, nil
}

// DoubleSerializer deserializes float values stored as their string representation.
//
// PHP's memcached extension stores floats this way when the value is a scalar
// (flags type = FlagDouble). For example, 3.14 is stored as the bytes "3.14".
type DoubleSerializer struct{}

// Deserialize parses a string-encoded float into float64.
func (s *DoubleSerializer) Deserialize(data []byte) (any, error) {
	str := strings.TrimSpace(string(data))
	if str == "" {
		return float64(0), nil
	}
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, fmt.Errorf("memcached: double deserialize %q: %w", str, err)
	}
	return v, nil
}

// BoolSerializer deserializes boolean values stored by PHP's memcached extension.
//
// PHP stores true as "1" and false as "" (empty bytes) with flags type = FlagBool.
type BoolSerializer struct{}

// Deserialize parses a PHP boolean value into a Go bool.
func (s *BoolSerializer) Deserialize(data []byte) (any, error) {
	str := string(data)
	return str == "1", nil
}

// JSONSerializer deserializes JSON-encoded data into native Go types.
type JSONSerializer struct{}

// Deserialize decodes JSON data into Go values (map, slice, string, number, bool, nil).
func (s *JSONSerializer) Deserialize(data []byte) (any, error) {
	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("memcached: json deserialize: %w", err)
	}
	return result, nil
}
