package igbinary

// FormatVersion is the igbinary format version supported by this decoder.
const FormatVersion = 2

// igbinary type codes (format version 2).
//
// These constants define the single-byte tags used in the igbinary binary
// format to identify the type and encoding of each value. All multi-byte
// numeric payloads are big-endian.
const (
	// TypeNil represents a PHP NULL value. No payload follows.
	TypeNil byte = 0x00

	// TypeArrayRef8 is an 8-bit back-reference to a previously seen array.
	TypeArrayRef8 byte = 0x01
	// TypeArrayRef16 is a 16-bit back-reference to a previously seen array.
	TypeArrayRef16 byte = 0x02
	// TypeArrayRef32 is a 32-bit back-reference to a previously seen array.
	TypeArrayRef32 byte = 0x03

	// TypeBoolFalse represents PHP boolean false. No payload follows.
	TypeBoolFalse byte = 0x04
	// TypeBoolTrue represents PHP boolean true. No payload follows.
	TypeBoolTrue byte = 0x05

	// TypePosInt8 is a positive integer stored in 1 unsigned byte.
	TypePosInt8 byte = 0x06
	// TypeNegInt8 is a negative integer; 1 unsigned byte is negated.
	TypeNegInt8 byte = 0x07
	// TypePosInt16 is a positive integer stored in 2 big-endian bytes.
	TypePosInt16 byte = 0x08
	// TypeNegInt16 is a negative integer; 2 big-endian bytes are negated.
	TypeNegInt16 byte = 0x09
	// TypePosInt32 is a positive integer stored in 4 big-endian bytes.
	TypePosInt32 byte = 0x0A
	// TypeNegInt32 is a negative integer; 4 big-endian bytes are negated.
	TypeNegInt32 byte = 0x0B

	// TypeDouble is an IEEE 754 float64 stored in 8 big-endian bytes.
	TypeDouble byte = 0x0C

	// TypeStringEmpty is an empty string. Registered in the string table.
	TypeStringEmpty byte = 0x0D

	// TypeStringID8 references a previously seen string by 8-bit table index.
	TypeStringID8 byte = 0x0E
	// TypeStringID16 references a previously seen string by 16-bit table index.
	TypeStringID16 byte = 0x0F
	// TypeStringID32 references a previously seen string by 32-bit table index.
	TypeStringID32 byte = 0x10

	// TypeString8 is a new string with 8-bit length prefix.
	TypeString8 byte = 0x11
	// TypeString16 is a new string with 16-bit length prefix.
	TypeString16 byte = 0x12
	// TypeString32 is a new string with 32-bit length prefix.
	TypeString32 byte = 0x13

	// TypeArray8 is an array with 8-bit element count.
	TypeArray8 byte = 0x14
	// TypeArray16 is an array with 16-bit element count.
	TypeArray16 byte = 0x15
	// TypeArray32 is an array with 32-bit element count.
	TypeArray32 byte = 0x16

	// TypeObject8 is an object with 8-bit class name length.
	TypeObject8 byte = 0x17
	// TypeObject16 is an object with 16-bit class name length.
	TypeObject16 byte = 0x18
	// TypeObject32 is an object with 32-bit class name length.
	TypeObject32 byte = 0x19

	// TypeObjectID8 is an object referencing a class by 8-bit string table index.
	TypeObjectID8 byte = 0x1A
	// TypeObjectID16 is an object referencing a class by 16-bit string table index.
	TypeObjectID16 byte = 0x1B
	// TypeObjectID32 is an object referencing a class by 32-bit string table index.
	TypeObjectID32 byte = 0x1C

	// TypeObjectSer8 is a serialized object with 8-bit class name length.
	TypeObjectSer8 byte = 0x1D
	// TypeObjectSer16 is a serialized object with 16-bit class name length.
	TypeObjectSer16 byte = 0x1E
	// TypeObjectSer32 is a serialized object with 32-bit class name length.
	TypeObjectSer32 byte = 0x1F

	// TypePosInt64 is a positive integer stored in 8 big-endian bytes.
	TypePosInt64 byte = 0x20
	// TypeNegInt64 is a negative integer; 8 big-endian bytes are negated.
	TypeNegInt64 byte = 0x21

	// TypeObjectRef8 is an 8-bit back-reference to a previously seen object.
	TypeObjectRef8 byte = 0x22
	// TypeObjectRef16 is a 16-bit back-reference to a previously seen object.
	TypeObjectRef16 byte = 0x23
	// TypeObjectRef32 is a 32-bit back-reference to a previously seen object.
	TypeObjectRef32 byte = 0x24

	// TypeSimpleRef is a simple PHP reference (&$var). Currently decoded as nil.
	TypeSimpleRef byte = 0x25
)

// ClassKey is the map key used to store the PHP class name when decoding objects.
const ClassKey = "__class"

// SerializedDataKey is the map key used to store raw serialized data
// when decoding objects that implement PHP's Serializable interface.
const SerializedDataKey = "__serialized_raw"
