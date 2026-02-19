package memcached_test

import (
	"testing"

	igbinary "github.com/RezaKargar/go-igbinary"
	"github.com/RezaKargar/go-igbinary/memcached"
)

func TestSerializerType(t *testing.T) {
	tests := []struct {
		flags    uint32
		expected uint32
	}{
		{0, memcached.FlagString},
		{1, memcached.FlagLong},
		{2, memcached.FlagDouble},
		{3, memcached.FlagBool},
		{4, memcached.FlagSerialized},
		{5, memcached.FlagIgbinary},
		{6, memcached.FlagJSON},
		{7, memcached.FlagMsgpack},
		{85, memcached.FlagIgbinary},   // 0x55 = igbinary + fastlz + compressed
		{53, memcached.FlagIgbinary},   // 0x35 = igbinary + zlib + compressed
		{0x16, memcached.FlagJSON},     // JSON + compressed
		{0xF5, memcached.FlagIgbinary}, // igbinary with many compression flags
	}
	for _, tc := range tests {
		got := memcached.SerializerType(tc.flags)
		if got != tc.expected {
			t.Errorf("SerializerType(0x%x) = %d, want %d", tc.flags, got, tc.expected)
		}
	}
}

func TestIsCompressed(t *testing.T) {
	tests := []struct {
		flags    uint32
		expected bool
	}{
		{0, false},
		{5, false}, // igbinary, no compression
		{85, true}, // 0x55 = igbinary + fastlz + compressed
		{53, true}, // 0x35 = igbinary + zlib + compressed
		{memcached.FlagCompressed, true},
		{memcached.FlagCompressed | memcached.FlagFastlz, true},
	}
	for _, tc := range tests {
		got := memcached.IsCompressed(tc.flags)
		if got != tc.expected {
			t.Errorf("IsCompressed(0x%x) = %v, want %v", tc.flags, got, tc.expected)
		}
	}
}

func TestIsIgbinaryFlags(t *testing.T) {
	tests := []struct {
		flags    uint32
		expected bool
	}{
		{memcached.FlagIgbinary, true},
		{memcached.FlagIgbinary | memcached.FlagCompressed | memcached.FlagZlib, true},
		{memcached.FlagIgbinary | memcached.FlagCompressed | memcached.FlagFastlz, true},
		{memcached.FlagJSON, false},
		{0, false},
	}

	for _, tc := range tests {
		got := memcached.IsIgbinaryFlags(tc.flags)
		if got != tc.expected {
			t.Errorf("IsIgbinaryFlags(0x%x) = %v, want %v", tc.flags, got, tc.expected)
		}
	}
}

func TestHasIgbinaryHeader(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "valid header",
			data:     []byte{0x00, 0x00, 0x00, igbinary.FormatVersion, 0x06, 0x2a},
			expected: true,
		},
		{
			name:     "too short",
			data:     []byte{0x00, 0x00, 0x00},
			expected: false,
		},
		{
			name:     "invalid version",
			data:     []byte{0x00, 0x00, 0x00, 0x01, 0x06, 0x2a},
			expected: false,
		},
		{
			name:     "invalid prefix",
			data:     []byte{0x01, 0x00, 0x00, igbinary.FormatVersion, 0x06, 0x2a},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := memcached.HasIgbinaryHeader(tc.data)
			if got != tc.expected {
				t.Errorf("HasIgbinaryHeader(%v) = %v, want %v", tc.data, got, tc.expected)
			}
		})
	}
}

func TestShouldTryIgbinary(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint32
		data     []byte
		expected bool
	}{
		{
			name:     "igbinary flags",
			flags:    memcached.FlagIgbinary,
			data:     []byte("not-igbinary"),
			expected: true,
		},
		{
			name:     "igbinary header",
			flags:    0,
			data:     []byte{0x00, 0x00, 0x00, igbinary.FormatVersion, 0x06, 0x2a},
			expected: true,
		},
		{
			name:     "neither",
			flags:    memcached.FlagJSON,
			data:     []byte("json-ish"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := memcached.ShouldTryIgbinary(tc.flags, tc.data)
			if got != tc.expected {
				t.Errorf("ShouldTryIgbinary(0x%x, %v) = %v, want %v", tc.flags, tc.data, got, tc.expected)
			}
		})
	}
}

func TestSerializerName(t *testing.T) {
	tests := []struct {
		flags    uint32
		expected string
	}{
		{0, "string"},
		{1, "long"},
		{2, "double"},
		{3, "bool"},
		{4, "php_serialize"},
		{5, "igbinary"},
		{6, "json"},
		{7, "msgpack"},
		{85, "igbinary"}, // works with compression flags too
		{15, "unknown(15)"},
	}
	for _, tc := range tests {
		got := memcached.SerializerName(tc.flags)
		if got != tc.expected {
			t.Errorf("SerializerName(0x%x) = %q, want %q", tc.flags, got, tc.expected)
		}
	}
}

func TestCompressionName(t *testing.T) {
	tests := []struct {
		flags    uint32
		expected string
	}{
		{0, "none"},
		{5, "none"},
		{memcached.FlagCompressed | memcached.FlagFastlz, "fastlz"},
		{memcached.FlagCompressed | memcached.FlagZlib, "zlib"},
		{memcached.FlagCompressed | memcached.FlagZstd, "zstd"},
		{memcached.FlagCompressed, "unknown"}, // compressed but no algo flag
	}
	for _, tc := range tests {
		got := memcached.CompressionName(tc.flags)
		if got != tc.expected {
			t.Errorf("CompressionName(0x%x) = %q, want %q", tc.flags, got, tc.expected)
		}
	}
}

func TestExplainFlags(t *testing.T) {
	result := memcached.ExplainFlags(85)
	if result == "" {
		t.Error("ExplainFlags returned empty string")
	}
	// Should contain key information
	for _, substr := range []string{"igbinary", "fastlz", "true", "0x00000055"} {
		if !contains(result, substr) {
			t.Errorf("ExplainFlags(85) = %q, want to contain %q", result, substr)
		}
	}
}

func TestExplainFlagsUncompressed(t *testing.T) {
	result := memcached.ExplainFlags(5) // igbinary, no compression
	for _, substr := range []string{"igbinary", "none", "false"} {
		if !contains(result, substr) {
			t.Errorf("ExplainFlags(5) = %q, want to contain %q", result, substr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
