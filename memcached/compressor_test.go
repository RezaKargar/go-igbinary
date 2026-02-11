package memcached_test

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"

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
