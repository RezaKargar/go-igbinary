package igbinary

import (
	"fmt"
	"math"
)

// Decode decodes igbinary-serialized data into a Go value.
//
// The data must include the 4-byte igbinary header (00 00 00 02 for version 2).
// Returns the decoded value or an error if the data is malformed.
//
// This is a convenience wrapper around [Decoder.Decode] using default options.
func Decode(data []byte) (any, error) {
	return defaultDecoder.Decode(data)
}

// defaultDecoder is the package-level decoder with default options.
var defaultDecoder = NewDecoder()

// Option configures a [Decoder].
type Option func(*Decoder)

// WithStrictMode enables strict decoding mode. In strict mode, the decoder
// returns errors for conditions that would otherwise be silently handled
// (e.g., unresolved references return an error instead of nil).
func WithStrictMode(strict bool) Option {
	return func(d *Decoder) {
		d.strict = strict
	}
}

// Decoder decodes igbinary-serialized binary data into Go values.
//
// A Decoder is safe for concurrent use: each call to [Decoder.Decode] creates
// its own internal state. The Decoder itself only holds configuration.
type Decoder struct {
	strict bool
}

// NewDecoder creates a new Decoder with the given options.
//
//	dec := igbinary.NewDecoder(
//	    igbinary.WithStrictMode(true),
//	)
func NewDecoder(opts ...Option) *Decoder {
	d := &Decoder{}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Decode decodes igbinary-serialized data into a Go value.
//
// The data must include the 4-byte igbinary header (00 00 00 02 for version 2).
func (d *Decoder) Decode(data []byte) (any, error) {
	if len(data) < 5 {
		return nil, newError(ErrDataTooShort,
			0, fmt.Sprintf("%d bytes (need at least 5)", len(data)))
	}

	// Validate header: 00 00 00 02
	if data[0] != 0x00 || data[1] != 0x00 || data[2] != 0x00 || data[3] != FormatVersion {
		return nil, newError(ErrInvalidHeader,
			0, fmt.Sprintf("got %02x %02x %02x %02x, want 00 00 00 %02x",
				data[0], data[1], data[2], data[3], FormatVersion))
	}

	r := &reader{
		data:   data,
		pos:    4, // skip header
		strict: d.strict,
	}

	return r.decodeValue()
}

// reader holds the mutable state for a single decode operation.
type reader struct {
	data    []byte
	pos     int
	strings []string // string deduplication table
	values  []any    // compound value reference table (arrays and objects)
	strict  bool
}

// --- Low-level read primitives ---

func (r *reader) readByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, newError(ErrUnexpectedEnd, r.pos, "")
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *reader) readBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, newError(ErrUnexpectedEnd, r.pos,
			fmt.Sprintf("need %d bytes, have %d", n, len(r.data)-r.pos))
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *reader) readUint8() (uint8, error) {
	return r.readByte()
}

func (r *reader) readUint16() (uint16, error) {
	b, err := r.readBytes(2)
	if err != nil {
		return 0, err
	}
	return uint16(b[0])<<8 | uint16(b[1]), nil
}

func (r *reader) readUint32() (uint32, error) {
	b, err := r.readBytes(4)
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

func (r *reader) readUint64() (uint64, error) {
	b, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7]), nil
}

// --- Value decoding ---

func (r *reader) decodeValue() (any, error) {
	code, err := r.readByte()
	if err != nil {
		return nil, err
	}

	switch code {
	case TypeNil:
		return nil, nil

	case TypeBoolFalse:
		return false, nil

	case TypeBoolTrue:
		return true, nil

	// Positive integers
	case TypePosInt8:
		v, err := r.readUint8()
		return int64(v), err
	case TypePosInt16:
		v, err := r.readUint16()
		return int64(v), err
	case TypePosInt32:
		v, err := r.readUint32()
		return int64(v), err
	case TypePosInt64:
		v, err := r.readUint64()
		return int64(v), err

	// Negative integers
	case TypeNegInt8:
		v, err := r.readUint8()
		return -int64(v), err
	case TypeNegInt16:
		v, err := r.readUint16()
		return -int64(v), err
	case TypeNegInt32:
		v, err := r.readUint32()
		return -int64(v), err
	case TypeNegInt64:
		v, err := r.readUint64()
		return -int64(v), err

	// Float
	case TypeDouble:
		v, err := r.readUint64()
		if err != nil {
			return nil, err
		}
		return math.Float64frombits(v), nil

	// Strings
	case TypeStringEmpty:
		// Empty strings are NOT registered in the dedup table.
		// PHP igbinary uses type_string_empty as a special marker
		// that does not occupy a slot in the string table.
		return "", nil
	case TypeString8:
		return r.decodeNewString8()
	case TypeString16:
		return r.decodeNewString16()
	case TypeString32:
		return r.decodeNewString32()

	// String back-references
	case TypeStringID8:
		v, err := r.readUint8()
		if err != nil {
			return nil, err
		}
		return r.lookupString(int(v))
	case TypeStringID16:
		v, err := r.readUint16()
		if err != nil {
			return nil, err
		}
		return r.lookupString(int(v))
	case TypeStringID32:
		v, err := r.readUint32()
		if err != nil {
			return nil, err
		}
		return r.lookupString(int(v))

	// Arrays
	case TypeArray8:
		v, err := r.readUint8()
		if err != nil {
			return nil, err
		}
		return r.decodeArray(int(v))
	case TypeArray16:
		v, err := r.readUint16()
		if err != nil {
			return nil, err
		}
		return r.decodeArray(int(v))
	case TypeArray32:
		v, err := r.readUint32()
		if err != nil {
			return nil, err
		}
		return r.decodeArray(int(v))

	// Objects with inline class name
	case TypeObject8, TypeObject16, TypeObject32:
		return r.decodeObject(code)

	// Objects with class name back-reference
	case TypeObjectID8, TypeObjectID16, TypeObjectID32:
		return r.decodeObjectByID(code)

	// Serializable objects
	case TypeObjectSer8, TypeObjectSer16, TypeObjectSer32:
		return r.decodeObjectSerialized(code)

	// Array/object references
	case TypeArrayRef8, TypeArrayRef16, TypeArrayRef32:
		return r.decodeRef(code)
	case TypeObjectRef8, TypeObjectRef16, TypeObjectRef32:
		return r.decodeRef(code)

	// Simple reference (&$var)
	case TypeSimpleRef:
		if r.strict {
			return nil, newError(ErrUnknownType, r.pos-1,
				"simple references are not fully supported in strict mode")
		}
		return nil, nil

	default:
		return nil, newError(ErrUnknownType, r.pos-1,
			fmt.Sprintf("0x%02x", code))
	}
}

// --- String helpers ---

func (r *reader) decodeNewString8() (string, error) {
	length, err := r.readUint8()
	if err != nil {
		return "", err
	}
	return r.readAndRegisterString(int(length))
}

func (r *reader) decodeNewString16() (string, error) {
	length, err := r.readUint16()
	if err != nil {
		return "", err
	}
	return r.readAndRegisterString(int(length))
}

func (r *reader) decodeNewString32() (string, error) {
	length, err := r.readUint32()
	if err != nil {
		return "", err
	}
	return r.readAndRegisterString(int(length))
}

func (r *reader) readAndRegisterString(length int) (string, error) {
	b, err := r.readBytes(length)
	if err != nil {
		return "", err
	}
	s := string(b)
	r.strings = append(r.strings, s)
	return s, nil
}

func (r *reader) lookupString(id int) (string, error) {
	if id < 0 || id >= len(r.strings) {
		return "", newError(ErrStringIDOutOfRange, r.pos,
			fmt.Sprintf("ID %d, table size %d", id, len(r.strings)))
	}
	return r.strings[id], nil
}

func (r *reader) lookupValue(id int) (any, error) {
	if id < 0 || id >= len(r.values) {
		return nil, newError(ErrValueRefOutOfRange, r.pos,
			fmt.Sprintf("ID %d, table size %d", id, len(r.values)))
	}
	return r.values[id], nil
}

// --- Array decoding ---

func (r *reader) decodeArray(size int) (any, error) {
	m := make(map[string]any, size)
	// Register in the values table before populating so that back-references
	// from nested values can resolve to this map (handles circular refs).
	r.values = append(r.values, m)

	for i := 0; i < size; i++ {
		key, err := r.decodeArrayKey()
		if err != nil {
			return nil, fmt.Errorf("array key %d: %w", i, err)
		}

		val, err := r.decodeValue()
		if err != nil {
			return nil, fmt.Errorf("array value for key %q: %w", key, err)
		}

		m[key] = val
	}

	return m, nil
}

func (r *reader) decodeArrayKey() (string, error) {
	code, err := r.readByte()
	if err != nil {
		return "", err
	}

	switch code {
	// String keys
	case TypeStringEmpty:
		// Empty strings are NOT registered in the dedup table.
		return "", nil
	case TypeString8:
		return r.decodeNewString8()
	case TypeString16:
		return r.decodeNewString16()
	case TypeString32:
		return r.decodeNewString32()
	case TypeStringID8:
		v, err := r.readUint8()
		if err != nil {
			return "", err
		}
		return r.lookupString(int(v))
	case TypeStringID16:
		v, err := r.readUint16()
		if err != nil {
			return "", err
		}
		return r.lookupString(int(v))
	case TypeStringID32:
		v, err := r.readUint32()
		if err != nil {
			return "", err
		}
		return r.lookupString(int(v))

	// Integer keys (converted to string representation)
	case TypePosInt8:
		v, err := r.readUint8()
		return fmt.Sprintf("%d", v), err
	case TypeNegInt8:
		v, err := r.readUint8()
		return fmt.Sprintf("-%d", v), err
	case TypePosInt16:
		v, err := r.readUint16()
		return fmt.Sprintf("%d", v), err
	case TypeNegInt16:
		v, err := r.readUint16()
		return fmt.Sprintf("-%d", v), err
	case TypePosInt32:
		v, err := r.readUint32()
		return fmt.Sprintf("%d", v), err
	case TypeNegInt32:
		v, err := r.readUint32()
		return fmt.Sprintf("-%d", v), err
	case TypePosInt64:
		v, err := r.readUint64()
		return fmt.Sprintf("%d", v), err
	case TypeNegInt64:
		v, err := r.readUint64()
		return fmt.Sprintf("-%d", v), err

	default:
		return "", newError(ErrUnsupportedArrayKey, r.pos-1,
			fmt.Sprintf("0x%02x", code))
	}
}

// --- Object decoding ---

func (r *reader) decodeObject(code byte) (any, error) {
	var nameLen int
	var err error
	switch code {
	case TypeObject8:
		v, e := r.readUint8()
		nameLen, err = int(v), e
	case TypeObject16:
		v, e := r.readUint16()
		nameLen, err = int(v), e
	case TypeObject32:
		v, e := r.readUint32()
		nameLen, err = int(v), e
	}
	if err != nil {
		return nil, err
	}

	className, err := r.readAndRegisterString(nameLen)
	if err != nil {
		return nil, err
	}

	return r.decodeObjectProperties(className)
}

func (r *reader) decodeObjectByID(code byte) (any, error) {
	var classID int
	var err error
	switch code {
	case TypeObjectID8:
		v, e := r.readUint8()
		classID, err = int(v), e
	case TypeObjectID16:
		v, e := r.readUint16()
		classID, err = int(v), e
	case TypeObjectID32:
		v, e := r.readUint32()
		classID, err = int(v), e
	}
	if err != nil {
		return nil, err
	}

	className, err := r.lookupString(classID)
	if err != nil {
		return nil, fmt.Errorf("object class ID: %w", err)
	}

	return r.decodeObjectProperties(className)
}

// decodeObjectProperties reads the property count and key-value pairs for an object.
func (r *reader) decodeObjectProperties(className string) (any, error) {
	propCountCode, err := r.readByte()
	if err != nil {
		return nil, err
	}

	var propCount int
	switch propCountCode {
	case TypeArray8:
		v, e := r.readUint8()
		propCount, err = int(v), e
	case TypeArray16:
		v, e := r.readUint16()
		propCount, err = int(v), e
	case TypeArray32:
		v, e := r.readUint32()
		propCount, err = int(v), e
	default:
		return nil, newError(ErrInvalidObjectProperties, r.pos-1,
			fmt.Sprintf("expected array type code, got 0x%02x", propCountCode))
	}
	if err != nil {
		return nil, err
	}

	m := make(map[string]any, propCount+1)
	m[ClassKey] = className
	// Register in the values table before populating so that back-references
	// from nested values can resolve to this object.
	r.values = append(r.values, m)

	for i := 0; i < propCount; i++ {
		key, keyErr := r.decodeArrayKey()
		if keyErr != nil {
			return nil, fmt.Errorf("object %q property key %d: %w", className, i, keyErr)
		}
		val, valErr := r.decodeValue()
		if valErr != nil {
			return nil, fmt.Errorf("object %q property %q: %w", className, key, valErr)
		}
		m[key] = val
	}

	return m, nil
}

func (r *reader) decodeObjectSerialized(code byte) (any, error) {
	var nameLen int
	var err error
	switch code {
	case TypeObjectSer8:
		v, e := r.readUint8()
		nameLen, err = int(v), e
	case TypeObjectSer16:
		v, e := r.readUint16()
		nameLen, err = int(v), e
	case TypeObjectSer32:
		v, e := r.readUint32()
		nameLen, err = int(v), e
	}
	if err != nil {
		return nil, err
	}

	className, err := r.readAndRegisterString(nameLen)
	if err != nil {
		return nil, err
	}

	dataLenCode, err := r.readByte()
	if err != nil {
		return nil, err
	}

	var dataLen int
	switch dataLenCode {
	case TypeString8:
		v, e := r.readUint8()
		dataLen, err = int(v), e
	case TypeString16:
		v, e := r.readUint16()
		dataLen, err = int(v), e
	case TypeString32:
		v, e := r.readUint32()
		dataLen, err = int(v), e
	default:
		return nil, newError(ErrInvalidSerializedData, r.pos-1,
			fmt.Sprintf("expected string type code, got 0x%02x", dataLenCode))
	}
	if err != nil {
		return nil, err
	}

	raw, err := r.readBytes(dataLen)
	if err != nil {
		return nil, err
	}

	m := map[string]any{
		ClassKey:          className,
		SerializedDataKey: string(raw),
	}
	r.values = append(r.values, m)
	return m, nil
}

// --- Reference decoding ---

func (r *reader) decodeRef(code byte) (any, error) {
	var id int
	var err error
	switch code {
	case TypeArrayRef8, TypeObjectRef8:
		v, e := r.readUint8()
		id, err = int(v), e
	case TypeArrayRef16, TypeObjectRef16:
		v, e := r.readUint16()
		id, err = int(v), e
	case TypeArrayRef32, TypeObjectRef32:
		v, e := r.readUint32()
		id, err = int(v), e
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.lookupValue(id)
}
