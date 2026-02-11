package memcached

import "fmt"

// Codec decodes raw cache values by orchestrating decompression and
// deserialization using pluggable [Compressor] and [Serializer] implementations.
//
// It reads a flags field to determine which compression and serialization
// algorithms to use, then dispatches to the registered implementations.
//
// Create one with [NewCodec] for standard PHP memcached defaults,
// or use [NewCodecBuilder] for custom configurations.
type Codec struct {
	compressors        map[uint32]Compressor
	serializers        map[uint32]Serializer
	fallbackCompressor Compressor
	fallbackSerializer Serializer
}

// NewCodec creates a Codec pre-configured with standard PHP memcached defaults:
// FastLZ + Zlib compressors, and Igbinary + String + JSON serializers.
//
// This is a convenience constructor. Use [NewCodecBuilder] to customize.
func NewCodec() *Codec {
	return NewCodecBuilder().
		WithCompressor(FlagFastlz, &FastlzCompressor{}).
		WithCompressor(FlagZlib, NewZlibCompressor(true)).
		WithSerializer(FlagIgbinary, &IgbinarySerializer{}).
		WithSerializer(FlagString, &StringSerializer{}).
		WithSerializer(FlagJSON, &JSONSerializer{}).
		WithFallbackCompressor(&FastlzCompressor{}).
		WithFallbackSerializer(&IgbinarySerializer{}).
		Build()
}

// Decode decodes a raw cache value using flags to determine the compression
// and serialization format.
//
// Pipeline:
//  1. If [IsCompressed](flags), decompress using the matching [Compressor].
//  2. Deserialize using the matching [Serializer] for [SerializerType](flags).
func (c *Codec) Decode(data []byte, flags uint32) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}

	raw := data

	// Step 1: Decompress if needed
	if IsCompressed(flags) {
		compressor := c.resolveCompressor(flags)
		if compressor == nil {
			return nil, fmt.Errorf("memcached: no compressor registered for flags 0x%08x", flags)
		}
		var err error
		raw, err = compressor.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("memcached: decompression failed: %w", err)
		}
	}

	// Step 2: Deserialize
	serializer := c.resolveSerializer(flags)
	if serializer == nil {
		return nil, fmt.Errorf("memcached: no serializer registered for type %d (flags 0x%08x)",
			SerializerType(flags), flags)
	}
	return serializer.Deserialize(raw)
}

// resolveCompressor finds the appropriate compressor for the given flags.
func (c *Codec) resolveCompressor(flags uint32) Compressor {
	for flag, comp := range c.compressors {
		if flags&flag != 0 {
			return comp
		}
	}
	return c.fallbackCompressor
}

// resolveSerializer finds the appropriate serializer for the given flags.
func (c *Codec) resolveSerializer(flags uint32) Serializer {
	serType := SerializerType(flags)
	if ser, ok := c.serializers[serType]; ok {
		return ser
	}
	return c.fallbackSerializer
}

// --- Builder ---

// CodecBuilder provides a fluent API for constructing a [Codec]
// with custom compressor and serializer implementations.
//
// Example:
//
//	codec := memcached.NewCodecBuilder().
//	    WithCompressor(memcached.FlagFastlz, &memcached.FastlzCompressor{}).
//	    WithCompressor(memcached.FlagZstd, &myZstdCompressor{}).
//	    WithSerializer(memcached.FlagJSON, &memcached.JSONSerializer{}).
//	    WithSerializer(memcached.FlagIgbinary, &memcached.IgbinarySerializer{}).
//	    WithFallbackCompressor(&memcached.FastlzCompressor{}).
//	    Build()
type CodecBuilder struct {
	compressors        map[uint32]Compressor
	serializers        map[uint32]Serializer
	fallbackCompressor Compressor
	fallbackSerializer Serializer
}

// NewCodecBuilder creates a new empty builder.
func NewCodecBuilder() *CodecBuilder {
	return &CodecBuilder{
		compressors: make(map[uint32]Compressor),
		serializers: make(map[uint32]Serializer),
	}
}

// WithCompressor registers a compressor for a specific compression flag bit.
// The flag should be a single bit flag like [FlagFastlz], [FlagZlib], or [FlagZstd].
func (b *CodecBuilder) WithCompressor(flag uint32, c Compressor) *CodecBuilder {
	b.compressors[flag] = c
	return b
}

// WithSerializer registers a serializer for a specific serializer type value.
// The flag should be a type constant like [FlagIgbinary], [FlagJSON], or [FlagString].
func (b *CodecBuilder) WithSerializer(flag uint32, s Serializer) *CodecBuilder {
	b.serializers[flag] = s
	return b
}

// WithFallbackCompressor sets the compressor used when the compressed flag is
// set but no specific compression algorithm flag matches.
func (b *CodecBuilder) WithFallbackCompressor(c Compressor) *CodecBuilder {
	b.fallbackCompressor = c
	return b
}

// WithFallbackSerializer sets the serializer used when the type flag doesn't
// match any registered serializer.
func (b *CodecBuilder) WithFallbackSerializer(s Serializer) *CodecBuilder {
	b.fallbackSerializer = s
	return b
}

// Build creates the [Codec] from the builder configuration.
func (b *CodecBuilder) Build() *Codec {
	return &Codec{
		compressors:        b.compressors,
		serializers:        b.serializers,
		fallbackCompressor: b.fallbackCompressor,
		fallbackSerializer: b.fallbackSerializer,
	}
}
