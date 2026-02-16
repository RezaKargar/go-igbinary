package igbinary

import (
	"strconv"
	"testing"
)

func TestNormalizeArrays_SequentialMapToSlice(t *testing.T) {
	input := map[string]any{"0": "a", "1": "b", "2": "c"}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s) != 3 {
		t.Fatalf("expected length 3, got %d", len(s))
	}
	if s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Errorf("unexpected values: %v", s)
	}
}

func TestNormalizeArrays_NonSequentialMapStaysMap(t *testing.T) {
	input := map[string]any{"name": "alice", "age": 30}
	result := NormalizeArrays(input)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if m["name"] != "alice" {
		t.Errorf("expected name=alice, got %v", m["name"])
	}
}

func TestNormalizeArrays_EmptyMapToEmptySlice(t *testing.T) {
	input := map[string]any{}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s) != 0 {
		t.Errorf("expected empty slice, got %v", s)
	}
}

func TestNormalizeArrays_NestedSequential(t *testing.T) {
	input := map[string]any{
		"0": map[string]any{"0": "x", "1": "y"},
		"1": "z",
	}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s) != 2 {
		t.Fatalf("expected length 2, got %d", len(s))
	}
	inner, ok := s[0].([]any)
	if !ok {
		t.Fatalf("expected inner []any, got %T", s[0])
	}
	if inner[0] != "x" || inner[1] != "y" {
		t.Errorf("unexpected inner values: %v", inner)
	}
	if s[1] != "z" {
		t.Errorf("expected s[1]=z, got %v", s[1])
	}
}

func TestNormalizeArrays_SliceRecursive(t *testing.T) {
	input := []any{
		map[string]any{"0": "a", "1": "b"},
		"plain",
	}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	inner, ok := s[0].([]any)
	if !ok {
		t.Fatalf("expected inner []any, got %T", s[0])
	}
	if inner[0] != "a" || inner[1] != "b" {
		t.Errorf("unexpected inner values: %v", inner)
	}
	if s[1] != "plain" {
		t.Errorf("expected s[1]=plain, got %v", s[1])
	}
}

func TestNormalizeArrays_NonMapNonSlice(t *testing.T) {
	if NormalizeArrays(42) != 42 {
		t.Error("expected 42")
	}
	if NormalizeArrays("hello") != "hello" {
		t.Error("expected hello")
	}
	if NormalizeArrays(nil) != nil {
		t.Error("expected nil")
	}
	if NormalizeArrays(true) != true {
		t.Error("expected true")
	}
}

func TestNormalizeArrays_MixedKeysStaysMap(t *testing.T) {
	input := map[string]any{"0": "a", "1": "b", "name": "c"}
	result := NormalizeArrays(input)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if m["0"] != "a" {
		t.Errorf("expected m[0]=a, got %v", m["0"])
	}
	if m["name"] != "c" {
		t.Errorf("expected m[name]=c, got %v", m["name"])
	}
}

func TestNormalizeArrays_GapInSequenceStaysMap(t *testing.T) {
	// Keys 0, 2 -- missing 1, not sequential
	input := map[string]any{"0": "a", "2": "c"}
	result := NormalizeArrays(input)
	if _, ok := result.(map[string]any); !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
}

func TestNormalizeArrays_SingleElementSequential(t *testing.T) {
	input := map[string]any{"0": "only"}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s) != 1 || s[0] != "only" {
		t.Errorf("unexpected values: %v", s)
	}
}

func TestNormalizeArrays_DeeplyNested(t *testing.T) {
	// 3 levels deep: sequential -> assoc -> sequential
	input := map[string]any{
		"0": map[string]any{
			"items": map[string]any{
				"0": "deep",
				"1": "nested",
			},
		},
	}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	inner, ok := s[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", s[0])
	}
	items, ok := inner["items"].([]any)
	if !ok {
		t.Fatalf("expected []any for items, got %T", inner["items"])
	}
	if items[0] != "deep" || items[1] != "nested" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestNormalizeArrays_LargeSequential(t *testing.T) {
	input := make(map[string]any, 100)
	for i := 0; i < 100; i++ {
		input[strconv.Itoa(i)] = i
	}
	result := NormalizeArrays(input)
	s, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(s) != 100 {
		t.Fatalf("expected length 100, got %d", len(s))
	}
	for i := 0; i < 100; i++ {
		if s[i] != i {
			t.Errorf("expected s[%d]=%d, got %v", i, i, s[i])
		}
	}
}

func TestNormalizeArrays_WithPhpStyleData(t *testing.T) {
	// Simulate a typical PHP igbinary decoded structure:
	// A product with photos_raw as sequential array and product as assoc map.
	input := map[string]any{
		"product": map[string]any{
			"id":       int64(42),
			"title_fa": "محصول",
			"platforms": map[string]any{
				"0": "desktop",
				"1": "mobile",
			},
		},
		"photos_raw": map[string]any{
			"0": map[string]any{
				"url":      "img1.jpg",
				"webp_url": "img1.webp",
			},
			"1": map[string]any{
				"url":      "img2.jpg",
				"webp_url": "img2.webp",
			},
		},
	}

	result := NormalizeArrays(input)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}

	// photos_raw should now be []any
	photos, ok := m["photos_raw"].([]any)
	if !ok {
		t.Fatalf("expected photos_raw as []any, got %T", m["photos_raw"])
	}
	if len(photos) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(photos))
	}

	// product should still be map[string]any
	product, ok := m["product"].(map[string]any)
	if !ok {
		t.Fatalf("expected product as map[string]any, got %T", m["product"])
	}
	if product["id"] != int64(42) {
		t.Errorf("expected id=42, got %v", product["id"])
	}

	// product.platforms should be []any
	platforms, ok := product["platforms"].([]any)
	if !ok {
		t.Fatalf("expected platforms as []any, got %T", product["platforms"])
	}
	if platforms[0] != "desktop" || platforms[1] != "mobile" {
		t.Errorf("unexpected platforms: %v", platforms)
	}
}

func TestWithNormalizeArrays_DecoderOption(t *testing.T) {
	// Encode a simple PHP indexed array: [10, 20]
	// igbinary: header (00 00 00 02) + array8(2) + key(posint8 0) + val(posint8 10) + key(posint8 1) + val(posint8 20)
	data := []byte{
		0x00, 0x00, 0x00, 0x02, // header
		0x14, 0x02, // array8, 2 entries
		0x06, 0x00, // key: posint8(0)
		0x06, 0x0A, // value: posint8(10)
		0x06, 0x01, // key: posint8(1)
		0x06, 0x14, // value: posint8(20)
	}

	// Without normalization
	dec := NewDecoder()
	val, err := dec.Decode(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := val.(map[string]any); !ok {
		t.Fatalf("expected map without normalization, got %T", val)
	}

	// With normalization
	decNorm := NewDecoder(WithNormalizeArrays())
	val, err = decNorm.Decode(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	s, ok := val.([]any)
	if !ok {
		t.Fatalf("expected []any with normalization, got %T", val)
	}
	if len(s) != 2 {
		t.Fatalf("expected length 2, got %d", len(s))
	}
	if s[0] != int64(10) || s[1] != int64(20) {
		t.Errorf("unexpected values: %v", s)
	}
}

func TestWithNormalizeArrays_AssocArrayUnchanged(t *testing.T) {
	// Encode {"name": "test"} -- assoc array should stay as map
	data := []byte{
		0x00, 0x00, 0x00, 0x02, // header
		0x14, 0x01, // array8, 1 entry
		0x11, 0x04, 'n', 'a', 'm', 'e', // key: string "name"
		0x11, 0x04, 't', 'e', 's', 't', // value: string "test"
	}

	dec := NewDecoder(WithNormalizeArrays())
	val, err := dec.Decode(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", val)
	}
	if m["name"] != "test" {
		t.Errorf("expected name=test, got %v", m["name"])
	}
}
