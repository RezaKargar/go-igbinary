package memcached_test

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"

	fastlz "github.com/dgryski/go-fastlz"

	"github.com/RezaKargar/go-igbinary/memcached"
)

func TestZlibCompressorWithLengthPrefix(t *testing.T) {
	original := []byte("hello world, this is a test string for compression")

	// Compress with zlib
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(original); err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Prepend 4-byte LE length prefix (PHP memcached format)
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(original)))
	lenBuf = append(lenBuf, compressed.Bytes()...)

	c := memcached.NewZlibCompressor(true)
	result, err := c.Decompress(lenBuf)
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("got %q, want %q", result, original)
	}
}

func TestZlibCompressorWithoutLengthPrefix(t *testing.T) {
	original := []byte("hello world without prefix")

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(original); err != nil {
		t.Fatal(err)
	}
	w.Close()

	c := memcached.NewZlibCompressor(false)
	result, err := c.Decompress(compressed.Bytes())
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("got %q, want %q", result, original)
	}
}

func TestZlibCompressorTooShort(t *testing.T) {
	c := memcached.NewZlibCompressor(true)
	_, err := c.Decompress([]byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for short data")
	}
}

func TestZlibCompressorInvalidData(t *testing.T) {
	c := memcached.NewZlibCompressor(false)
	_, err := c.Decompress([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	if err == nil {
		t.Fatal("expected error for invalid zlib data")
	}
}

func TestFastlzCompressorTooShort(t *testing.T) {
	c := &memcached.FastlzCompressor{}
	_, err := c.Decompress([]byte{0x01})
	if err == nil {
		t.Fatal("expected error for short data")
	}
}

func TestFastlzCompressorSuccess(t *testing.T) {
	original := []byte("hello world, this is test data for fastlz compression testing")
	encoded, err := fastlz.Encode(nil, original)
	if err != nil {
		t.Fatal(err)
	}

	c := &memcached.FastlzCompressor{}
	result, err := c.Decompress(encoded)
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("got %q, want %q", result, original)
	}
}

func TestFastlzCompressorInvalidData(t *testing.T) {
	// 4+ bytes but not valid fastlz payload
	c := &memcached.FastlzCompressor{}
	_, err := c.Decompress([]byte{0x00, 0x00, 0x00, 0x10, 0xFF, 0xFF})
	if err == nil {
		t.Fatal("expected error for invalid fastlz data")
	}
}

func TestZlibCompressorFallbackToFullData(t *testing.T) {
	// Compress valid data with zlib (no length prefix)
	original := []byte("test data for zlib fallback path")
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(original); err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Use a length-prefixed compressor on non-prefixed data.
	// The first 4 bytes (zlib header 78 9C ...) will be misinterpreted as length,
	// making data[4:] invalid zlib. The fallback tries the full data as zlib.
	c := memcached.NewZlibCompressor(true)
	result, err := c.Decompress(buf.Bytes())
	if err != nil {
		t.Fatalf("Decompress error (fallback should have worked): %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("got %q, want %q", result, original)
	}
}

func TestZlibCompressorNoLengthPrefixInvalidData(t *testing.T) {
	c := memcached.NewZlibCompressor(false)
	_, err := c.Decompress([]byte{0x00, 0x00, 0x00, 0x00, 0xFF})
	if err == nil {
		t.Fatal("expected error for invalid zlib data without prefix")
	}
}
