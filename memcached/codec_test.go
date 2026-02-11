package memcached_test

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"

	"github.com/RezaKargar/go-igbinary/memcached"
)

func TestCodecDecodeIgbinaryUncompressed(t *testing.T) {
	codec := memcached.NewCodec()

	// igbinary header + string "hello"
	data := []byte{0x00, 0x00, 0x00, 0x02, 0x11, 0x05, 'h', 'e', 'l', 'l', 'o'}
	flags := memcached.FlagIgbinary // 5 = igbinary, no compression

	val, err := codec.Decode(data, flags)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if str != "hello" {
		t.Errorf("expected %q, got %q", "hello", str)
	}
}

func TestCodecDecodeJSONUncompressed(t *testing.T) {
	codec := memcached.NewCodec()

	data := []byte(`{"key":"value"}`)
	flags := memcached.FlagJSON // 6 = JSON, no compression

	val, err := codec.Decode(data, flags)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

func TestCodecDecodeRawString(t *testing.T) {
	codec := memcached.NewCodec()

	data := []byte("plain text")
	flags := memcached.FlagString // 0 = string, no compression

	val, err := codec.Decode(data, flags)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if str != "plain text" {
		t.Errorf("expected %q, got %q", "plain text", str)
	}
}

func TestCodecDecodeEmptyBoolFalse(t *testing.T) {
	// PHP stores false as empty bytes with FlagBool
	codec := memcached.NewCodec()
	val, err := codec.Decode([]byte{}, memcached.FlagBool)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := val.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T (%v)", val, val)
	}
	if v {
		t.Error("expected false")
	}
}

func TestCodecDecodeNilData(t *testing.T) {
	codec := memcached.NewCodec()
	val, err := codec.Decode(nil, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestCodecNoSerializerRegistered(t *testing.T) {
	// Build a codec with no serializers at all
	codec := memcached.NewCodecBuilder().Build()
	_, err := codec.Decode([]byte("data"), memcached.FlagIgbinary)
	if err == nil {
		t.Fatal("expected error for unregistered serializer")
	}
}

func TestCodecNoCompressorRegistered(t *testing.T) {
	// Build a codec with no compressors
	codec := memcached.NewCodecBuilder().
		WithSerializer(memcached.FlagIgbinary, &memcached.IgbinarySerializer{}).
		Build()

	// Set compressed flag
	flags := memcached.FlagIgbinary | memcached.FlagCompressed | memcached.FlagFastlz
	_, err := codec.Decode([]byte("data"), flags)
	if err == nil {
		t.Fatal("expected error for unregistered compressor")
	}
}

func TestCodecBuilderFluent(t *testing.T) {
	// Verify the builder pattern works end-to-end
	codec := memcached.NewCodecBuilder().
		WithCompressor(memcached.FlagFastlz, &memcached.FastlzCompressor{}).
		WithCompressor(memcached.FlagZlib, memcached.NewZlibCompressor(true)).
		WithSerializer(memcached.FlagIgbinary, &memcached.IgbinarySerializer{}).
		WithSerializer(memcached.FlagJSON, &memcached.JSONSerializer{}).
		WithSerializer(memcached.FlagString, &memcached.StringSerializer{}).
		WithFallbackCompressor(&memcached.FastlzCompressor{}).
		WithFallbackSerializer(&memcached.IgbinarySerializer{}).
		Build()

	// Decode a simple uncompressed igbinary int
	data := []byte{0x00, 0x00, 0x00, 0x02, 0x06, 0x0A} // int 10
	val, err := codec.Decode(data, memcached.FlagIgbinary)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	v, ok := val.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", val)
	}
	if v != 10 {
		t.Errorf("expected 10, got %d", v)
	}
}

func TestCodecFallbackSerializer(t *testing.T) {
	// Use a codec with a fallback serializer for an unknown type
	codec := memcached.NewCodecBuilder().
		WithFallbackSerializer(&memcached.StringSerializer{}).
		Build()

	data := []byte("raw data")
	flags := uint32(15) // Unknown serializer type
	val, err := codec.Decode(data, flags)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if str != "raw data" {
		t.Errorf("expected %q, got %q", "raw data", str)
	}
}

func TestCodecDecodeCompressedZlib(t *testing.T) {
	// Create igbinary data: header + int 42
	igData := []byte{0x00, 0x00, 0x00, 0x02, 0x06, 0x2A}

	// Compress with zlib
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(igData); err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Prepend 4-byte LE length prefix (PHP format)
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(igData)))
	lenBuf = append(lenBuf, buf.Bytes()...)
	data := lenBuf

	codec := memcached.NewCodec()
	flags := memcached.FlagIgbinary | memcached.FlagCompressed | memcached.FlagZlib

	val, err := codec.Decode(data, flags)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	v, ok := val.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T (%v)", val, val)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestCodecDecompressError(t *testing.T) {
	codec := memcached.NewCodec()
	// Compressed flag set but data is invalid
	flags := memcached.FlagIgbinary | memcached.FlagCompressed | memcached.FlagFastlz
	_, err := codec.Decode([]byte{0x00, 0x00, 0x00, 0x10, 0xFF, 0xFF}, flags)
	if err == nil {
		t.Fatal("expected error for decompression failure")
	}
}

func TestCodecFallbackCompressor(t *testing.T) {
	// Compress "hello" with zlib (no length prefix)
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Build codec with only a fallback compressor (no specific flag compressors)
	codec := memcached.NewCodecBuilder().
		WithFallbackCompressor(memcached.NewZlibCompressor(false)).
		WithSerializer(memcached.FlagString, &memcached.StringSerializer{}).
		Build()

	// Use FlagCompressed without a specific compression flag bit
	flags := memcached.FlagString | memcached.FlagCompressed
	val, err := codec.Decode(buf.Bytes(), flags)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if str != "hello" {
		t.Errorf("expected %q, got %q", "hello", str)
	}
}
