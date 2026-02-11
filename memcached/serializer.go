package memcached

import (
	"encoding/json"
	"fmt"

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
