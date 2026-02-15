package igbinary_test

import (
	"errors"
	"math"
	"testing"

	igbinary "github.com/RezaKargar/go-igbinary"
)

// header is the standard igbinary v2 header prepended to all test payloads.
var header = []byte{0x00, 0x00, 0x00, 0x02}

// makePayload prepends the igbinary header to body bytes.
func makePayload(body ...byte) []byte {
	return append(append([]byte{}, header...), body...)
}

// --- Nil ---

func TestDecodeNil(t *testing.T) {
	data := makePayload(0x00) // TypeNil
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

// --- Booleans ---

func TestDecodeBoolFalse(t *testing.T) {
	data := makePayload(0x04)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualBool(t, val, false)
}

func TestDecodeBoolTrue(t *testing.T) {
	data := makePayload(0x05)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualBool(t, val, true)
}

// --- Positive Integers ---

func TestDecodePosInt8(t *testing.T) {
	data := makePayload(0x06, 42)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 42)
}

func TestDecodePosInt8Zero(t *testing.T) {
	data := makePayload(0x06, 0)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 0)
}

func TestDecodePosInt8Max(t *testing.T) {
	data := makePayload(0x06, 0xFF)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 255)
}

func TestDecodePosInt16(t *testing.T) {
	data := makePayload(0x08, 0x01, 0x00) // 256
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 256)
}

func TestDecodePosInt16Max(t *testing.T) {
	data := makePayload(0x08, 0xFF, 0xFF) // 65535
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 65535)
}

func TestDecodePosInt32(t *testing.T) {
	data := makePayload(0x0A, 0x00, 0x01, 0x00, 0x00) // 65536
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 65536)
}

func TestDecodePosInt32Large(t *testing.T) {
	data := makePayload(0x0A, 0x7F, 0xFF, 0xFF, 0xFF) // 2147483647
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 2147483647)
}

func TestDecodePosInt64(t *testing.T) {
	data := makePayload(0x20, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00) // 4294967296
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, 4294967296)
}

// --- Negative Integers ---

func TestDecodeNegInt8(t *testing.T) {
	data := makePayload(0x07, 5)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, -5)
}

func TestDecodeNegInt16(t *testing.T) {
	data := makePayload(0x09, 0x01, 0x00) // -256
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, -256)
}

func TestDecodeNegInt32(t *testing.T) {
	data := makePayload(0x0B, 0x00, 0x01, 0x00, 0x00) // -65536
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, -65536)
}

func TestDecodeNegInt64(t *testing.T) {
	data := makePayload(0x21, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00) // -4294967296
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualInt64(t, val, -4294967296)
}

// --- Double (float64) ---

func TestDecodeDouble(t *testing.T) {
	// IEEE 754 encoding of 3.14
	bits := math.Float64bits(3.14)
	data := makePayload(0x0C,
		byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
		byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits),
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualFloat64(t, val, 3.14)
}

func TestDecodeDoubleZero(t *testing.T) {
	data := makePayload(0x0C, 0, 0, 0, 0, 0, 0, 0, 0)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualFloat64(t, val, 0.0)
}

func TestDecodeDoubleNegative(t *testing.T) {
	bits := math.Float64bits(-1.5)
	data := makePayload(0x0C,
		byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
		byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits),
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualFloat64(t, val, -1.5)
}

// --- Strings ---

func TestDecodeEmptyString(t *testing.T) {
	data := makePayload(0x0D) // TypeStringEmpty
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualString(t, val, "")
}

func TestDecodeString8(t *testing.T) {
	// TypeString8 + len(5) + "hello"
	data := makePayload(0x11, 0x05, 'h', 'e', 'l', 'l', 'o')
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualString(t, val, "hello")
}

func TestDecodeString16(t *testing.T) {
	// TypeString16 + len(3, big-endian) + "abc"
	data := makePayload(0x12, 0x00, 0x03, 'a', 'b', 'c')
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualString(t, val, "abc")
}

func TestDecodeString32(t *testing.T) {
	// TypeString32 + len(2, big-endian 4 bytes) + "hi"
	data := makePayload(0x13, 0x00, 0x00, 0x00, 0x02, 'h', 'i')
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	assertEqualString(t, val, "hi")
}

// --- String Deduplication ---

func TestDecodeStringDedup(t *testing.T) {
	// Array with 2 entries: key "name" => value "name" (reused via string ID)
	// TypeArray8(1 entry) + key=String8("name") + value=StringID8(0)
	data := makePayload(
		0x14, 0x01, // array of 1 element
		0x11, 0x04, 'n', 'a', 'm', 'e', // key: new string "name" (ID 0)
		0x0E, 0x00, // value: string ID 0 -> "name"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	assertEqualString(t, m["name"], "name")
}

func TestDecodeStringDedupMultiple(t *testing.T) {
	// Array with 3 entries using string dedup:
	// "alpha" => "one", "beta" => "alpha" (via ID), "gamma" => "beta" (via ID)
	data := makePayload(
		0x14, 0x03, // array of 3 elements
		0x11, 0x05, 'a', 'l', 'p', 'h', 'a', // key: "alpha" (ID 0)
		0x11, 0x03, 'o', 'n', 'e', // value: "one" (ID 1)
		0x11, 0x04, 'b', 'e', 't', 'a', // key: "beta" (ID 2)
		0x0E, 0x00, // value: string ID 0 -> "alpha"
		0x11, 0x05, 'g', 'a', 'm', 'm', 'a', // key: "gamma" (ID 3)
		0x0E, 0x02, // value: string ID 2 -> "beta"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	assertEqualString(t, m["alpha"], "one")
	assertEqualString(t, m["beta"], "alpha")
	assertEqualString(t, m["gamma"], "beta")
}

// --- Arrays ---

func TestDecodeEmptyArray(t *testing.T) {
	data := makePayload(0x14, 0x00) // TypeArray8 with 0 elements
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestDecodeArrayWithIntegerKeys(t *testing.T) {
	// PHP array [0 => "a", 1 => "b"]
	data := makePayload(
		0x14, 0x02, // array of 2
		0x06, 0x00, // key: positive int 0
		0x11, 0x01, 'a', // value: "a"
		0x06, 0x01, // key: positive int 1
		0x11, 0x01, 'b', // value: "b"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	assertEqualString(t, m["0"], "a")
	assertEqualString(t, m["1"], "b")
}

func TestDecodeNestedArray(t *testing.T) {
	// {"outer" => {"inner" => 42}}
	data := makePayload(
		0x14, 0x01, // outer array of 1
		0x11, 0x05, 'o', 'u', 't', 'e', 'r', // key: "outer"
		0x14, 0x01, // inner array of 1
		0x11, 0x05, 'i', 'n', 'n', 'e', 'r', // key: "inner"
		0x06, 0x2A, // value: int 42
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	inner := m["outer"].(map[string]any)
	assertEqualInt64(t, inner["inner"], 42)
}

func TestDecodeArrayMixedTypes(t *testing.T) {
	// {"str" => "hello", "num" => 99, "flag" => true, "empty" => nil}
	data := makePayload(
		0x14, 0x04, // array of 4
		0x11, 0x03, 's', 't', 'r', // key: "str"
		0x11, 0x05, 'h', 'e', 'l', 'l', 'o', // value: "hello"
		0x11, 0x03, 'n', 'u', 'm', // key: "num"
		0x06, 0x63, // value: int 99
		0x11, 0x04, 'f', 'l', 'a', 'g', // key: "flag"
		0x05,                                // value: true
		0x11, 0x05, 'e', 'm', 'p', 't', 'y', // key: "empty"
		0x00, // value: nil
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m["str"], "hello")
	assertEqualInt64(t, m["num"], 99)
	assertEqualBool(t, m["flag"], true)
	if m["empty"] != nil {
		t.Errorf("expected nil, got %v", m["empty"])
	}
}

// --- Objects ---

func TestDecodeObject(t *testing.T) {
	// Object of class "User" with property "name" => "Alice"
	data := makePayload(
		0x17, 0x04, 'U', 's', 'e', 'r', // TypeObject8, class name "User"
		0x14, 0x01, // properties: array of 1
		0x11, 0x04, 'n', 'a', 'm', 'e', // key: "name"
		0x11, 0x05, 'A', 'l', 'i', 'c', 'e', // value: "Alice"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "User")
	assertEqualString(t, m["name"], "Alice")
}

func TestDecodeObjectByID(t *testing.T) {
	// First, create a class name in the string table via an object, then reference it.
	// We encode: array of 2 objects, both class "Foo"
	// First object uses TypeObject8, second uses TypeObjectID8
	data := makePayload(
		0x14, 0x02, // array of 2 elements
		0x11, 0x01, 'a', // key: "a" (string ID 0)
		0x17, 0x03, 'F', 'o', 'o', // TypeObject8, class name "Foo" (string ID 1)
		0x14, 0x01, // properties: 1
		0x11, 0x01, 'x', // key: "x" (string ID 2)
		0x06, 0x01, // value: int 1
		0x11, 0x01, 'b', // key: "b" (string ID 3)
		0x1A, 0x01, // TypeObjectID8, class string ID 1 -> "Foo"
		0x14, 0x01, // properties: 1
		0x0E, 0x02, // key: string ID 2 -> "x"
		0x06, 0x02, // value: int 2
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	objA := m["a"].(map[string]any)
	objB := m["b"].(map[string]any)

	assertEqualString(t, objA[igbinary.ClassKey], "Foo")
	assertEqualString(t, objB[igbinary.ClassKey], "Foo")
	assertEqualInt64(t, objA["x"], 1)
	assertEqualInt64(t, objB["x"], 2)
}

func TestDecodeSerializedObject(t *testing.T) {
	// TypeObjectSer8 + class name + serialized data
	data := makePayload(
		0x1D, 0x03, 'B', 'a', 'r', // TypeObjectSer8, class "Bar"
		0x11, 0x05, 'h', 'e', 'l', 'l', 'o', // serialized data: "hello" (5 bytes)
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Bar")
	assertEqualString(t, m[igbinary.SerializedDataKey], "hello")
}

// --- References ---

func TestDecodeArrayRefOutOfRange(t *testing.T) {
	// A bare ArrayRef8 with no prior arrays should return ErrValueRefOutOfRange.
	data := makePayload(0x01, 0x00) // TypeArrayRef8, ID 0
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for array ref with no prior values")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

func TestDecodeObjectRefOutOfRange(t *testing.T) {
	// A bare ObjectRef8 with no prior objects should return ErrValueRefOutOfRange.
	data := makePayload(0x22, 0x00) // TypeObjectRef8, ID 0
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for object ref with no prior values")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

func TestDecodeSimpleRef(t *testing.T) {
	data := makePayload(0x25) // TypeSimpleRef
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	if val != nil {
		t.Errorf("expected nil for simple ref, got %v", val)
	}
}

// --- Strict Mode ---

func TestStrictModeRejectsSimpleRef(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x25) // TypeSimpleRef
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for simple ref")
	}
}

func TestStrictModeRejectsArrayRef(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x01, 0x00) // TypeArrayRef8
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for array ref")
	}
}

// --- Error Cases ---

func TestDecodeEmptyData(t *testing.T) {
	_, err := igbinary.Decode([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
	if !errors.Is(err, igbinary.ErrDataTooShort) {
		t.Errorf("expected ErrDataTooShort, got: %v", err)
	}
}

func TestDecodeNilData(t *testing.T) {
	_, err := igbinary.Decode(nil)
	if err == nil {
		t.Fatal("expected error for nil data")
	}
	if !errors.Is(err, igbinary.ErrDataTooShort) {
		t.Errorf("expected ErrDataTooShort, got: %v", err)
	}
}

func TestDecodeTooShort(t *testing.T) {
	_, err := igbinary.Decode([]byte{0x00, 0x00, 0x00, 0x02}) // header only, no body
	if err == nil {
		t.Fatal("expected error for data with header only")
	}
	if !errors.Is(err, igbinary.ErrDataTooShort) {
		t.Errorf("expected ErrDataTooShort, got: %v", err)
	}
}

func TestDecodeInvalidHeader(t *testing.T) {
	_, err := igbinary.Decode([]byte{0x00, 0x00, 0x00, 0x01, 0x00}) // version 1
	if err == nil {
		t.Fatal("expected error for invalid header")
	}
	if !errors.Is(err, igbinary.ErrInvalidHeader) {
		t.Errorf("expected ErrInvalidHeader, got: %v", err)
	}
}

func TestDecodeInvalidHeaderBytes(t *testing.T) {
	_, err := igbinary.Decode([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x00})
	if err == nil {
		t.Fatal("expected error for garbage header")
	}
	if !errors.Is(err, igbinary.ErrInvalidHeader) {
		t.Errorf("expected ErrInvalidHeader, got: %v", err)
	}
}

func TestDecodeTruncatedInt(t *testing.T) {
	data := makePayload(0x08, 0x01) // TypePosInt16 but only 1 byte of payload
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated int16")
	}
	if !errors.Is(err, igbinary.ErrUnexpectedEnd) {
		t.Errorf("expected ErrUnexpectedEnd, got: %v", err)
	}
}

func TestDecodeTruncatedString(t *testing.T) {
	data := makePayload(0x11, 0x05, 'h', 'e') // String8 with length 5 but only 2 bytes
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated string")
	}
	if !errors.Is(err, igbinary.ErrUnexpectedEnd) {
		t.Errorf("expected ErrUnexpectedEnd, got: %v", err)
	}
}

func TestDecodeInvalidStringID(t *testing.T) {
	data := makePayload(0x0E, 0x05) // StringID8 referencing ID 5, but table is empty
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for invalid string ID")
	}
	if !errors.Is(err, igbinary.ErrStringIDOutOfRange) {
		t.Errorf("expected ErrStringIDOutOfRange, got: %v", err)
	}
}

func TestDecodeUnknownTypeCode(t *testing.T) {
	data := makePayload(0xFF) // Unknown type
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for unknown type code")
	}
	if !errors.Is(err, igbinary.ErrUnknownType) {
		t.Errorf("expected ErrUnknownType, got: %v", err)
	}
}

func TestDecodeTruncatedArray(t *testing.T) {
	data := makePayload(0x14, 0x02, // array of 2
		0x11, 0x01, 'a', // key: "a"
		0x06, 0x01, // value: 1
		// missing second entry
	)
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated array")
	}
}

func TestDecodeInvalidObjectPropertyCode(t *testing.T) {
	data := makePayload(
		0x17, 0x03, 'F', 'o', 'o', // TypeObject8, class "Foo"
		0x06, 0x01, // WRONG: should be array type for properties, but got int
	)
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for invalid object property encoding")
	}
	if !errors.Is(err, igbinary.ErrInvalidObjectProperties) {
		t.Errorf("expected ErrInvalidObjectProperties, got: %v", err)
	}
}

// --- DecodeError ---

func TestDecodeErrorMessage(t *testing.T) {
	_, err := igbinary.Decode([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x00})
	if err == nil {
		t.Fatal("expected error")
	}

	var decErr *igbinary.DecodeError
	if !errors.As(err, &decErr) {
		t.Fatalf("expected DecodeError, got %T: %v", err, err)
	}
	if decErr.Pos != 0 {
		t.Errorf("expected Pos 0, got %d", decErr.Pos)
	}
	if decErr.Detail == "" {
		t.Error("expected non-empty Detail")
	}
}

// --- Decoder reuse ---

func TestDecoderReuse(t *testing.T) {
	dec := igbinary.NewDecoder()

	// Decode an int
	val1, err := dec.Decode(makePayload(0x06, 0x0A))
	assertNoError(t, err)
	assertEqualInt64(t, val1, 10)

	// Decode a string - string table should be fresh
	val2, err := dec.Decode(makePayload(0x11, 0x02, 'h', 'i'))
	assertNoError(t, err)
	assertEqualString(t, val2, "hi")
}

// --- Convenience function vs Decoder ---

func TestConvenienceFunctionMatchesDecoder(t *testing.T) {
	data := makePayload(0x06, 42)
	v1, err1 := igbinary.Decode(data)
	v2, err2 := igbinary.NewDecoder().Decode(data)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if v1 != v2 {
		t.Errorf("Decode() returned %v, NewDecoder().Decode() returned %v", v1, v2)
	}
}

// --- Complex real-world-like payload ---

func TestDecodeComplexPayload(t *testing.T) {
	// Simulate a PHP product-like object:
	// {"id" => 12345, "title" => "Test Product", "price" => 99.99, "active" => true}
	data := makePayload(
		0x14, 0x04, // array of 4
		0x11, 0x02, 'i', 'd', // key: "id" (string ID 0)
		0x08, 0x30, 0x39, // value: int 12345
		0x11, 0x05, 't', 'i', 't', 'l', 'e', // key: "title" (string ID 1)
		0x11, 0x0C, 'T', 'e', 's', 't', ' ', 'P', 'r', 'o', 'd', 'u', 'c', 't', // value: "Test Product"
		0x11, 0x05, 'p', 'r', 'i', 'c', 'e', // key: "price" (string ID 3)
	)
	// Append float64 for 99.99
	bits := math.Float64bits(99.99)
	data = append(data, 0x0C,
		byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
		byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits),
	)
	// Append "active" => true
	data = append(data,
		0x11, 0x06, 'a', 'c', 't', 'i', 'v', 'e', // key: "active"
		0x05, // value: true
	)

	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualInt64(t, m["id"], 12345)
	assertEqualString(t, m["title"], "Test Product")
	assertEqualFloat64(t, m["price"], 99.99)
	assertEqualBool(t, m["active"], true)
}

// --- Array16 and Array32 ---

func TestDecodeArray16(t *testing.T) {
	// Array16 with 1 element
	data := makePayload(
		0x15, 0x00, 0x01, // TypeArray16, count=1
		0x11, 0x01, 'k', // key: "k"
		0x06, 0x01, // value: 1
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualInt64(t, m["k"], 1)
}

func TestDecodeArray32(t *testing.T) {
	// Array32 with 1 element
	data := makePayload(
		0x16, 0x00, 0x00, 0x00, 0x01, // TypeArray32, count=1
		0x11, 0x01, 'k', // key: "k"
		0x06, 0x02, // value: 2
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualInt64(t, m["k"], 2)
}

// --- StringID16 and StringID32 ---

func TestDecodeStringID16(t *testing.T) {
	// Array with string dedup using 16-bit ID
	data := makePayload(
		0x14, 0x01,
		0x11, 0x03, 'k', 'e', 'y', // "key" as string ID 0
		0x0F, 0x00, 0x00, // StringID16 -> ID 0
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m["key"], "key")
}

func TestDecodeStringID32(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x11, 0x03, 'k', 'e', 'y',
		0x10, 0x00, 0x00, 0x00, 0x00, // StringID32 -> ID 0
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m["key"], "key")
}

// --- Object16 and Object32 ---

func TestDecodeObject16(t *testing.T) {
	data := makePayload(
		0x18, 0x00, 0x02, 'O', 'b', // TypeObject16, class name len 2, "Ob"
		0x14, 0x00, // properties: 0
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Ob")
}

func TestDecodeObject32(t *testing.T) {
	data := makePayload(
		0x19, 0x00, 0x00, 0x00, 0x02, 'O', 'b', // TypeObject32, class name len 2
		0x14, 0x00, // properties: 0
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Ob")
}

// --- ObjectID16, ObjectID32 ---

func TestDecodeObjectID16(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x11, 0x01, 'a', // key: "a" (string ID 0)
		// First, register class name via an Object8
		0x17, 0x03, 'C', 'l', 's', // TypeObject8, class "Cls" (string ID 1)
		0x14, 0x00, // 0 properties
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	obj := m["a"].(map[string]any)
	assertEqualString(t, obj[igbinary.ClassKey], "Cls")
}

// --- ObjectSer16, ObjectSer32 ---

func TestDecodeObjectSer16(t *testing.T) {
	data := makePayload(
		0x1E, 0x00, 0x02, 'S', 'r', // TypeObjectSer16, class "Sr"
		0x11, 0x03, 'r', 'a', 'w', // serialized data "raw"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Sr")
	assertEqualString(t, m[igbinary.SerializedDataKey], "raw")
}

func TestDecodeObjectSer32(t *testing.T) {
	data := makePayload(
		0x1F, 0x00, 0x00, 0x00, 0x02, 'S', 'r', // TypeObjectSer32, class "Sr"
		0x11, 0x02, 'o', 'k', // serialized data "ok"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Sr")
	assertEqualString(t, m[igbinary.SerializedDataKey], "ok")
}

// --- Ref16, Ref32 out-of-range ---

func TestDecodeArrayRef16OutOfRange(t *testing.T) {
	data := makePayload(0x02, 0x00, 0x00) // TypeArrayRef16, ID 0
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for array ref16 with no prior values")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

func TestDecodeArrayRef32OutOfRange(t *testing.T) {
	data := makePayload(0x03, 0x00, 0x00, 0x00, 0x00) // TypeArrayRef32, ID 0
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for array ref32 with no prior values")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

func TestDecodeObjectRef16OutOfRange(t *testing.T) {
	data := makePayload(0x23, 0x00, 0x00) // TypeObjectRef16, ID 0
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for object ref16 with no prior values")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

func TestDecodeObjectRef32OutOfRange(t *testing.T) {
	data := makePayload(0x24, 0x00, 0x00, 0x00, 0x00) // TypeObjectRef32, ID 0
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for object ref32 with no prior values")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

// --- Negative integer keys ---

func TestDecodeArrayWithNegativeIntegerKeys(t *testing.T) {
	data := makePayload(
		0x14, 0x01, // array of 1
		0x07, 0x05, // key: negative int -5
		0x11, 0x03, 'v', 'a', 'l', // value: "val"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m["-5"], "val")
}

// --- Array key types (covering decodeArrayKey branches) ---

func TestDecodeArrayKeyString16(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x12, 0x00, 0x02, 'a', 'b', // key: String16 "ab"
		0x06, 0x01, // value: int 1
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualInt64(t, m["ab"], 1)
}

func TestDecodeArrayKeyString32(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x13, 0x00, 0x00, 0x00, 0x02, 'c', 'd', // key: String32 "cd"
		0x06, 0x02, // value: int 2
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualInt64(t, m["cd"], 2)
}

func TestDecodeArrayKeyStringEmpty(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x0D,       // key: StringEmpty
		0x06, 0x07, // value: int 7
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualInt64(t, m[""], 7)
}

func TestDecodeArrayKeyStringID16(t *testing.T) {
	data := makePayload(
		0x14, 0x02,
		0x11, 0x01, 'k', // key: String8 "k" (ID 0)
		0x06, 0x01, // value: 1
		0x0F, 0x00, 0x00, // key: StringID16 -> ID 0 ("k")
		0x06, 0x02, // value: 2
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualInt64(t, m["k"], 2) // second write overwrites
}

func TestDecodeArrayKeyStringID32(t *testing.T) {
	data := makePayload(
		0x14, 0x02,
		0x11, 0x01, 'k', // key: String8 "k" (ID 0)
		0x06, 0x01, // value: 1
		0x10, 0x00, 0x00, 0x00, 0x00, // key: StringID32 -> ID 0 ("k")
		0x06, 0x02, // value: 2
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualInt64(t, m["k"], 2)
}

func TestDecodeArrayKeyPosInt16(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x08, 0x01, 0x00, // key: PosInt16 = 256
		0x11, 0x01, 'v', // value: "v"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m["256"], "v")
}

func TestDecodeArrayKeyNegInt16(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x09, 0x01, 0x00, // key: NegInt16 = -256
		0x11, 0x01, 'v', // value: "v"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m["-256"], "v")
}

func TestDecodeArrayKeyPosInt32(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x0A, 0x00, 0x01, 0x00, 0x00, // key: PosInt32 = 65536
		0x11, 0x01, 'v', // value: "v"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m["65536"], "v")
}

func TestDecodeArrayKeyNegInt32(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x0B, 0x00, 0x01, 0x00, 0x00, // key: NegInt32 = -65536
		0x11, 0x01, 'v', // value: "v"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m["-65536"], "v")
}

func TestDecodeArrayKeyPosInt64(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x20, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // key: PosInt64 = 4294967296
		0x11, 0x01, 'v', // value: "v"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m["4294967296"], "v")
}

func TestDecodeArrayKeyNegInt64(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x21, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // key: NegInt64 = -4294967296
		0x11, 0x01, 'v', // value: "v"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m["-4294967296"], "v")
}

func TestDecodeArrayKeyUnsupported(t *testing.T) {
	data := makePayload(
		0x14, 0x01,
		0x0C, // key: TypeDouble (unsupported as array key)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x06, 0x01,
	)
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for unsupported array key type")
	}
}

// --- ObjectID16 and ObjectID32 ---

func TestDecodeObjectByID16(t *testing.T) {
	data := makePayload(
		0x14, 0x02,
		0x11, 0x01, 'a', // key: "a" (string ID 0)
		0x17, 0x03, 'C', 'l', 's', // TypeObject8, class "Cls" (string ID 1)
		0x14, 0x00, // 0 properties
		0x11, 0x01, 'b', // key: "b" (string ID 2)
		0x1B, 0x00, 0x01, // TypeObjectID16, class string ID 1 -> "Cls"
		0x14, 0x01, // 1 property
		0x11, 0x01, 'x', // key: "x" (string ID 3)
		0x06, 0x05, // value: int 5
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	objB := m["b"].(map[string]any)
	assertEqualString(t, objB[igbinary.ClassKey], "Cls")
	assertEqualInt64(t, objB["x"], 5)
}

func TestDecodeObjectByID32(t *testing.T) {
	data := makePayload(
		0x14, 0x02,
		0x11, 0x01, 'a', // key: "a" (string ID 0)
		0x17, 0x03, 'C', 'l', 's', // TypeObject8, class "Cls" (string ID 1)
		0x14, 0x00, // 0 properties
		0x11, 0x01, 'b', // key: "b" (string ID 2)
		0x1C, 0x00, 0x00, 0x00, 0x01, // TypeObjectID32, class string ID 1 -> "Cls"
		0x14, 0x01, // 1 property
		0x11, 0x01, 'y', // key: "y" (string ID 3)
		0x06, 0x09, // value: int 9
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	objB := m["b"].(map[string]any)
	assertEqualString(t, objB[igbinary.ClassKey], "Cls")
	assertEqualInt64(t, objB["y"], 9)
}

func TestDecodeObjectByIDInvalidClassID(t *testing.T) {
	// ObjectID8 referencing class ID 99 when string table is empty
	data := makePayload(
		0x1A, 0x63, // TypeObjectID8, class string ID 99
		0x14, 0x00,
	)
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for invalid class ID")
	}
}

// --- Object with Array16/Array32 property counts ---

func TestDecodeObjectWithArray16Properties(t *testing.T) {
	data := makePayload(
		0x17, 0x03, 'F', 'o', 'o', // TypeObject8, class "Foo"
		0x15, 0x00, 0x01, // TypeArray16, count=1
		0x11, 0x01, 'x', // key: "x"
		0x06, 0x01, // value: 1
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Foo")
	assertEqualInt64(t, m["x"], 1)
}

func TestDecodeObjectWithArray32Properties(t *testing.T) {
	data := makePayload(
		0x17, 0x03, 'F', 'o', 'o', // TypeObject8, class "Foo"
		0x16, 0x00, 0x00, 0x00, 0x01, // TypeArray32, count=1
		0x11, 0x01, 'x', // key: "x"
		0x06, 0x02, // value: 2
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Foo")
	assertEqualInt64(t, m["x"], 2)
}

// --- Serialized object data length variations ---

func TestDecodeObjectSerializedString16DataLen(t *testing.T) {
	data := makePayload(
		0x1D, 0x03, 'B', 'a', 'r', // TypeObjectSer8, class "Bar"
		0x12, 0x00, 0x03, 'a', 'b', 'c', // TypeString16, len=3, data "abc"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Bar")
	assertEqualString(t, m[igbinary.SerializedDataKey], "abc")
}

func TestDecodeObjectSerializedString32DataLen(t *testing.T) {
	data := makePayload(
		0x1D, 0x03, 'B', 'a', 'r', // TypeObjectSer8, class "Bar"
		0x13, 0x00, 0x00, 0x00, 0x02, 'o', 'k', // TypeString32, len=2, data "ok"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)
	m := val.(map[string]any)
	assertEqualString(t, m[igbinary.ClassKey], "Bar")
	assertEqualString(t, m[igbinary.SerializedDataKey], "ok")
}

func TestDecodeObjectSerializedInvalidDataLenCode(t *testing.T) {
	data := makePayload(
		0x1D, 0x03, 'B', 'a', 'r', // TypeObjectSer8, class "Bar"
		0x06, 0x01, // WRONG: int type instead of string type for data length
	)
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for invalid serialized data length code")
	}
	if !errors.Is(err, igbinary.ErrInvalidSerializedData) {
		t.Errorf("expected ErrInvalidSerializedData, got: %v", err)
	}
}

// --- Strict mode for all ref types ---

func TestStrictModeRejectsObjectRef8(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x22, 0x00) // TypeObjectRef8
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for object ref8")
	}
}

func TestStrictModeRejectsArrayRef16(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x02, 0x00, 0x00) // TypeArrayRef16
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for array ref16")
	}
}

func TestStrictModeRejectsArrayRef32(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x03, 0x00, 0x00, 0x00, 0x00) // TypeArrayRef32
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for array ref32")
	}
}

func TestStrictModeRejectsObjectRef16(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x23, 0x00, 0x00) // TypeObjectRef16
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for object ref16")
	}
}

func TestStrictModeRejectsObjectRef32(t *testing.T) {
	dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
	data := makePayload(0x24, 0x00, 0x00, 0x00, 0x00) // TypeObjectRef32
	_, err := dec.Decode(data)
	if err == nil {
		t.Fatal("expected error in strict mode for object ref32")
	}
}

// --- Truncation errors for wider types ---

func TestDecodeTruncatedUint32(t *testing.T) {
	data := makePayload(0x0A, 0x00, 0x00) // PosInt32 but only 2 payload bytes
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated uint32")
	}
	if !errors.Is(err, igbinary.ErrUnexpectedEnd) {
		t.Errorf("expected ErrUnexpectedEnd, got: %v", err)
	}
}

func TestDecodeTruncatedUint64(t *testing.T) {
	data := makePayload(0x20, 0x00, 0x00, 0x00) // PosInt64 but only 3 payload bytes
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated uint64")
	}
	if !errors.Is(err, igbinary.ErrUnexpectedEnd) {
		t.Errorf("expected ErrUnexpectedEnd, got: %v", err)
	}
}

func TestDecodeTruncatedDouble(t *testing.T) {
	data := makePayload(0x0C, 0x00, 0x00) // Double but only 2 payload bytes
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated double")
	}
	if !errors.Is(err, igbinary.ErrUnexpectedEnd) {
		t.Errorf("expected ErrUnexpectedEnd, got: %v", err)
	}
}

func TestDecodeTruncatedString16Length(t *testing.T) {
	data := makePayload(0x12, 0x00) // String16 but only 1 byte of length
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated string16 length")
	}
}

func TestDecodeTruncatedString32Length(t *testing.T) {
	data := makePayload(0x13, 0x00, 0x00) // String32 but only 2 bytes of length
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated string32 length")
	}
}

func TestDecodeTruncatedNegInt32(t *testing.T) {
	data := makePayload(0x0B, 0x00) // NegInt32 but only 1 payload byte
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated negative int32")
	}
}

func TestDecodeTruncatedNegInt64(t *testing.T) {
	data := makePayload(0x21, 0x00) // NegInt64 but only 1 payload byte
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for truncated negative int64")
	}
}

// --- DecodeError without detail ---

func TestDecodeErrorWithoutDetail(t *testing.T) {
	e := &igbinary.DecodeError{Err: igbinary.ErrUnexpectedEnd, Pos: 5, Detail: ""}
	expected := "igbinary: unexpected end of data at pos 5"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestDecodeErrorWithDetail(t *testing.T) {
	e := &igbinary.DecodeError{Err: igbinary.ErrUnexpectedEnd, Pos: 10, Detail: "need 4 bytes"}
	expected := "igbinary: unexpected end of data at pos 10: need 4 bytes"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestDecodeErrorUnwrap(t *testing.T) {
	e := &igbinary.DecodeError{Err: igbinary.ErrInvalidHeader, Pos: 0}
	if !errors.Is(e, igbinary.ErrInvalidHeader) {
		t.Error("Unwrap should return the underlying error")
	}
}

// --- Reference resolution tests ---

// TestDecodeArrayRefResolvesSharedArray mirrors the PHP promotion_pages scenario:
// Two keys in the same map point to the same array value. igbinary encodes the
// first as a full array and the second as an ArrayRef back-reference.
//
// PHP equivalent:
//
//	$arr = [42, 83];
//	$data = ["landing_pages" => $arr, "promotion_pages" => $arr];
//	// igbinary_serialize($data) uses ArrayRef for the second occurrence.
func TestDecodeArrayRefResolvesSharedArray(t *testing.T) {
	// Outer array (2 entries):
	//   key "landing_pages" => inner array {0 => 42, 1 => 83}   (value ID 1, outer is ID 0)
	//   key "promotion_pages" => ArrayRef8(1)                     (resolves to same inner array)
	data := makePayload(
		0x14, 0x02, // outer array, 2 entries (registered as value ID 0)
		0x11, 0x0D, 'l', 'a', 'n', 'd', 'i', 'n', 'g', '_', 'p', 'a', 'g', 'e', 's', // key: "landing_pages"
		0x14, 0x02, // inner array, 2 entries (registered as value ID 1)
		0x06, 0x00, // key: int 0
		0x06, 0x2A, // value: int 42
		0x06, 0x01, // key: int 1
		0x06, 0x53, // value: int 83
		0x11, 0x0F, 'p', 'r', 'o', 'm', 'o', 't', 'i', 'o', 'n', '_', 'p', 'a', 'g', 'e', 's', // key: "promotion_pages"
		0x01, 0x01, // TypeArrayRef8, ID 1 -> same inner array
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	landing, ok := m["landing_pages"].(map[string]any)
	if !ok {
		t.Fatalf("landing_pages: expected map, got %T", m["landing_pages"])
	}
	promo, ok := m["promotion_pages"].(map[string]any)
	if !ok {
		t.Fatalf("promotion_pages: expected map, got %T (%v)", m["promotion_pages"], m["promotion_pages"])
	}

	// Both should contain the same data
	assertEqualInt64(t, landing["0"], 42)
	assertEqualInt64(t, landing["1"], 83)
	assertEqualInt64(t, promo["0"], 42)
	assertEqualInt64(t, promo["1"], 83)
}

// TestDecodeArrayRefSelfReference tests that an array can reference its own parent.
func TestDecodeArrayRefSelfReference(t *testing.T) {
	// Outer array (1 entry): key "self" => ArrayRef8(0) -> the outer array itself
	data := makePayload(
		0x14, 0x01, // outer array, 1 entry (value ID 0)
		0x11, 0x04, 's', 'e', 'l', 'f', // key: "self"
		0x01, 0x00, // TypeArrayRef8, ID 0 -> self
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	selfRef, ok := m["self"].(map[string]any)
	if !ok {
		t.Fatalf("self: expected map, got %T", m["self"])
	}
	// The self-reference should point back to the same map instance
	if selfRef["self"] == nil {
		t.Fatal("self.self should not be nil (circular reference)")
	}
}

// TestDecodeObjectRefResolution tests that TypeObjectRef resolves to a previously
// decoded object.
func TestDecodeObjectRefResolution(t *testing.T) {
	// Outer array (2 entries):
	//   key "obj1" => Object "Foo" with property "x" => 1   (value ID 1, outer array is 0)
	//   key "obj2" => ObjectRef8(1)                          (resolves to same object)
	data := makePayload(
		0x14, 0x02, // outer array, 2 entries (value ID 0)
		0x11, 0x04, 'o', 'b', 'j', '1', // key: "obj1"
		0x17, 0x03, 'F', 'o', 'o', // TypeObject8, class "Foo" (value ID 1)
		0x14, 0x01, // properties: 1
		0x11, 0x01, 'x', // key: "x"
		0x06, 0x01, // value: int 1
		0x11, 0x04, 'o', 'b', 'j', '2', // key: "obj2"
		0x22, 0x01, // TypeObjectRef8, ID 1 -> same object
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	obj1, ok := m["obj1"].(map[string]any)
	if !ok {
		t.Fatalf("obj1: expected map, got %T", m["obj1"])
	}
	obj2, ok := m["obj2"].(map[string]any)
	if !ok {
		t.Fatalf("obj2: expected map, got %T (%v)", m["obj2"], m["obj2"])
	}

	assertEqualString(t, obj1[igbinary.ClassKey], "Foo")
	assertEqualInt64(t, obj1["x"], 1)
	assertEqualString(t, obj2[igbinary.ClassKey], "Foo")
	assertEqualInt64(t, obj2["x"], 1)
}

// TestDecodeNestedArrayRefResolution tests references to nested arrays.
func TestDecodeNestedArrayRefResolution(t *testing.T) {
	// Outer (value ID 0) -> "inner" => inner array (value ID 1) {0=>10}
	//                     -> "ref"   => ArrayRef8(1)
	data := makePayload(
		0x14, 0x02, // outer array, 2 entries (value ID 0)
		0x11, 0x05, 'i', 'n', 'n', 'e', 'r', // key: "inner"
		0x14, 0x01, // inner array, 1 entry (value ID 1)
		0x06, 0x00, // key: int 0
		0x06, 0x0A, // value: int 10
		0x11, 0x03, 'r', 'e', 'f', // key: "ref"
		0x01, 0x01, // TypeArrayRef8, ID 1 -> inner array
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	inner := m["inner"].(map[string]any)
	ref := m["ref"].(map[string]any)

	assertEqualInt64(t, inner["0"], 10)
	assertEqualInt64(t, ref["0"], 10)
}

// TestDecodeMultipleRefsToSameArray tests multiple references to the same array.
func TestDecodeMultipleRefsToSameArray(t *testing.T) {
	// Outer (ID 0) -> "a" => inner (ID 1) {0=>99}
	//              -> "b" => ArrayRef8(1)
	//              -> "c" => ArrayRef8(1)
	data := makePayload(
		0x14, 0x03, // outer array, 3 entries (value ID 0)
		0x11, 0x01, 'a', // key: "a"
		0x14, 0x01, // inner (value ID 1)
		0x06, 0x00, 0x06, 0x63, // key 0 => 99
		0x11, 0x01, 'b', // key: "b"
		0x01, 0x01, // ArrayRef8(1)
		0x11, 0x01, 'c', // key: "c"
		0x01, 0x01, // ArrayRef8(1)
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	a := m["a"].(map[string]any)
	b := m["b"].(map[string]any)
	c := m["c"].(map[string]any)

	assertEqualInt64(t, a["0"], 99)
	assertEqualInt64(t, b["0"], 99)
	assertEqualInt64(t, c["0"], 99)
}

// TestDecodeEmptyArrayRef tests that a reference to an empty array resolves correctly.
func TestDecodeEmptyArrayRef(t *testing.T) {
	// Outer (ID 0) -> "empty" => empty array (ID 1) {}
	//              -> "ref"   => ArrayRef8(1)
	data := makePayload(
		0x14, 0x02, // outer array, 2 entries (value ID 0)
		0x11, 0x05, 'e', 'm', 'p', 't', 'y', // key: "empty"
		0x14, 0x00, // empty array (value ID 1)
		0x11, 0x03, 'r', 'e', 'f', // key: "ref"
		0x01, 0x01, // ArrayRef8(1)
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	empty := m["empty"].(map[string]any)
	ref := m["ref"].(map[string]any)

	if len(empty) != 0 {
		t.Errorf("expected empty map, got %d entries", len(empty))
	}
	if len(ref) != 0 {
		t.Errorf("expected empty map via ref, got %d entries", len(ref))
	}
}

// TestDecodeValueRefOutOfRangeError tests the error type for invalid ref IDs.
func TestDecodeValueRefOutOfRangeError(t *testing.T) {
	// Outer (ID 0) -> "x" => ArrayRef8(5) -- ID 5 does not exist
	data := makePayload(
		0x14, 0x01,
		0x11, 0x01, 'x',
		0x01, 0x05, // ArrayRef8(5) -- out of range
	)
	_, err := igbinary.Decode(data)
	if err == nil {
		t.Fatal("expected error for out-of-range ref")
	}
	if !errors.Is(err, igbinary.ErrValueRefOutOfRange) {
		t.Errorf("expected ErrValueRefOutOfRange, got: %v", err)
	}
}

// --- TypeStringEmpty dedup table tests ---
//
// PHP igbinary treats TypeStringEmpty (0x0D) as a special marker that does NOT
// occupy a slot in the string deduplication table. If an empty string were
// registered, all subsequent StringID lookups would be off by one.

func TestEmptyStringValueDoesNotShiftStringIDs(t *testing.T) {
	// Scenario: a map where a value is an empty string (TypeStringEmpty), followed
	// by a StringID reference to a previously registered string.
	//
	// Without the fix, TypeStringEmpty would register as string ID 1, and the
	// StringID8(1) on line 9 would incorrectly resolve to "" instead of "hello".
	//
	// Layout:
	//   array(2)
	//     key: String8 "key"  -> registers as string ID 0
	//     val: String8 "hello" -> registers as string ID 1
	//     key: String8 "empty" -> registers as string ID 2
	//     val: StringEmpty      -> does NOT register
	//     key: String8 "ref"   -> registers as string ID 3
	//     val: StringID8(1)    -> should resolve to "hello" (not "")
	data := makePayload(
		0x14, 0x03, // array of 3 elements
		0x11, 0x03, 'k', 'e', 'y', // key: "key" (string ID 0)
		0x11, 0x05, 'h', 'e', 'l', 'l', 'o', // value: "hello" (string ID 1)
		0x11, 0x05, 'e', 'm', 'p', 't', 'y', // key: "empty" (string ID 2)
		0x0D,                      // value: TypeStringEmpty -> NOT registered
		0x11, 0x03, 'r', 'e', 'f', // key: "ref" (string ID 3)
		0x0E, 0x01, // value: StringID8(1) -> should be "hello"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m["key"], "hello")
	assertEqualString(t, m["empty"], "")
	assertEqualString(t, m["ref"], "hello")
}

func TestEmptyStringKeyDoesNotShiftStringIDs(t *testing.T) {
	// Scenario: empty string used as a map key (TypeStringEmpty), followed by
	// a StringID reference to a previously registered key.
	//
	// Layout:
	//   array(2)
	//     key: String8 "first" -> registers as string ID 0
	//     val: int 1
	//     key: StringEmpty      -> does NOT register
	//     val: int 2
	//     key: StringID8(0)    -> should resolve to "first"
	//     val: int 3            -> overwrites first entry
	data := makePayload(
		0x14, 0x03, // array of 3 elements
		0x11, 0x05, 'f', 'i', 'r', 's', 't', // key: "first" (string ID 0)
		0x06, 0x01, // value: int 1
		0x0D,       // key: TypeStringEmpty -> NOT registered
		0x06, 0x02, // value: int 2
		0x0E, 0x00, // key: StringID8(0) -> should be "first"
		0x06, 0x03, // value: int 3
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualInt64(t, m[""], 2)
	assertEqualInt64(t, m["first"], 3) // overwritten by the StringID8(0) entry
}

func TestMultipleEmptyStringsDoNotShiftStringIDs(t *testing.T) {
	// Scenario: multiple empty strings (both as keys and values) appear before
	// a StringID reference. None should occupy a slot.
	//
	// Layout:
	//   array(3)
	//     key: String8 "name" -> string ID 0
	//     val: String8 "Alice" -> string ID 1
	//     key: StringEmpty     -> NOT registered
	//     val: StringEmpty     -> NOT registered
	//     key: String8 "ref"  -> string ID 2
	//     val: StringID8(1)   -> should be "Alice"
	data := makePayload(
		0x14, 0x03, // array of 3 elements
		0x11, 0x04, 'n', 'a', 'm', 'e', // key: "name" (string ID 0)
		0x11, 0x05, 'A', 'l', 'i', 'c', 'e', // value: "Alice" (string ID 1)
		0x0D,                      // key: TypeStringEmpty -> NOT registered
		0x0D,                      // value: TypeStringEmpty -> NOT registered
		0x11, 0x03, 'r', 'e', 'f', // key: "ref" (string ID 2)
		0x0E, 0x01, // value: StringID8(1) -> should be "Alice"
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	assertEqualString(t, m["name"], "Alice")
	assertEqualString(t, m[""], "")
	assertEqualString(t, m["ref"], "Alice")
}

func TestEmptyStringInNestedStructDoesNotShiftStringIDs(t *testing.T) {
	// Scenario: mimics real PHP badge-like data where a nested object has an
	// empty string value, followed by StringID references in sibling objects.
	//
	// Layout (simplified badges):
	//   array(2)
	//     key: String8 "badge1" (ID 0)
	//     val: array(2)
	//       key: String8 "text"  (ID 1)
	//       val: String8 "Sale"  (ID 2)
	//       key: String8 "icon"  (ID 3)
	//       val: StringEmpty      -> NOT registered
	//     key: String8 "badge2" (ID 4)
	//     val: array(2)
	//       key: StringID8(1)    -> "text"
	//       val: String8 "New"   (ID 5)
	//       key: StringID8(3)    -> "icon"
	//       val: String8 "star"  (ID 6)
	data := makePayload(
		0x14, 0x02, // outer array, 2 entries
		0x11, 0x06, 'b', 'a', 'd', 'g', 'e', '1', // key: "badge1" (string ID 0)
		0x14, 0x02, // inner array, 2 entries
		0x11, 0x04, 't', 'e', 'x', 't', // key: "text" (string ID 1)
		0x11, 0x04, 'S', 'a', 'l', 'e', // value: "Sale" (string ID 2)
		0x11, 0x04, 'i', 'c', 'o', 'n', // key: "icon" (string ID 3)
		0x0D,                                     // value: TypeStringEmpty -> NOT registered
		0x11, 0x06, 'b', 'a', 'd', 'g', 'e', '2', // key: "badge2" (string ID 4)
		0x14, 0x02, // inner array, 2 entries
		0x0E, 0x01, // key: StringID8(1) -> "text"
		0x11, 0x03, 'N', 'e', 'w', // value: "New" (string ID 5)
		0x0E, 0x03, // key: StringID8(3) -> "icon"
		0x11, 0x04, 's', 't', 'a', 'r', // value: "star" (string ID 6)
	)
	val, err := igbinary.Decode(data)
	assertNoError(t, err)

	m := val.(map[string]any)
	badge1 := m["badge1"].(map[string]any)
	badge2 := m["badge2"].(map[string]any)

	assertEqualString(t, badge1["text"], "Sale")
	assertEqualString(t, badge1["icon"], "")
	assertEqualString(t, badge2["text"], "New")
	assertEqualString(t, badge2["icon"], "star")
}

// --- Test helpers ---

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqualInt64(t *testing.T, val any, expected int64) {
	t.Helper()
	v, ok := val.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T (%v)", val, val)
	}
	if v != expected {
		t.Errorf("expected %d, got %d", expected, v)
	}
}

func assertEqualFloat64(t *testing.T, val any, expected float64) {
	t.Helper()
	v, ok := val.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T (%v)", val, val)
	}
	if v != expected {
		t.Errorf("expected %f, got %f", expected, v)
	}
}

func assertEqualString(t *testing.T, val any, expected string) {
	t.Helper()
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T (%v)", val, val)
	}
	if v != expected {
		t.Errorf("expected %q, got %q", expected, v)
	}
}

func assertEqualBool(t *testing.T, val any, expected bool) {
	t.Helper()
	v, ok := val.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T (%v)", val, val)
	}
	if v != expected {
		t.Errorf("expected %v, got %v", expected, v)
	}
}
