package memcached_test

import (
	"testing"

	"github.com/RezaKargar/go-igbinary/memcached"
)

func TestIgbinarySerializer(t *testing.T) {
	s := &memcached.IgbinarySerializer{}

	// igbinary header + int(42)
	data := []byte{0x00, 0x00, 0x00, 0x02, 0x06, 0x2A}
	val, err := s.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v, ok := val.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", val)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestIgbinarySerializerInvalid(t *testing.T) {
	s := &memcached.IgbinarySerializer{}
	_, err := s.Deserialize([]byte{0x01})
	if err == nil {
		t.Fatal("expected error for invalid igbinary data")
	}
}

func TestStringSerializer(t *testing.T) {
	s := &memcached.StringSerializer{}
	val, err := s.Deserialize([]byte("hello world"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if str != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", str)
	}
}

func TestStringSerializerEmpty(t *testing.T) {
	s := &memcached.StringSerializer{}
	val, err := s.Deserialize([]byte{})
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if str != "" {
		t.Errorf("expected empty string, got %q", str)
	}
}

// --- LongSerializer ---

func TestLongSerializer(t *testing.T) {
	s := &memcached.LongSerializer{}
	val, err := s.Deserialize([]byte("42"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v, ok := val.(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", val)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestLongSerializerNegative(t *testing.T) {
	s := &memcached.LongSerializer{}
	val, err := s.Deserialize([]byte("-42"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v := val.(int64)
	if v != -42 {
		t.Errorf("expected -42, got %d", v)
	}
}

func TestLongSerializerLarge(t *testing.T) {
	s := &memcached.LongSerializer{}
	val, err := s.Deserialize([]byte("9999999999"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v := val.(int64)
	if v != 9999999999 {
		t.Errorf("expected 9999999999, got %d", v)
	}
}

func TestLongSerializerEmpty(t *testing.T) {
	s := &memcached.LongSerializer{}
	val, err := s.Deserialize([]byte(""))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v := val.(int64)
	if v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
}

func TestLongSerializerInvalid(t *testing.T) {
	s := &memcached.LongSerializer{}
	_, err := s.Deserialize([]byte("not_a_number"))
	if err == nil {
		t.Fatal("expected error for invalid long")
	}
}

// --- DoubleSerializer ---

func TestDoubleSerializer(t *testing.T) {
	s := &memcached.DoubleSerializer{}
	val, err := s.Deserialize([]byte("3.14"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v, ok := val.(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", val)
	}
	if v != 3.14 {
		t.Errorf("expected 3.14, got %f", v)
	}
}

func TestDoubleSerializerNegative(t *testing.T) {
	s := &memcached.DoubleSerializer{}
	val, err := s.Deserialize([]byte("-1.5"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v := val.(float64)
	if v != -1.5 {
		t.Errorf("expected -1.5, got %f", v)
	}
}

func TestDoubleSerializerEmpty(t *testing.T) {
	s := &memcached.DoubleSerializer{}
	val, err := s.Deserialize([]byte(""))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v := val.(float64)
	if v != 0 {
		t.Errorf("expected 0, got %f", v)
	}
}

func TestDoubleSerializerInvalid(t *testing.T) {
	s := &memcached.DoubleSerializer{}
	_, err := s.Deserialize([]byte("not_a_float"))
	if err == nil {
		t.Fatal("expected error for invalid double")
	}
}

// --- BoolSerializer ---

func TestBoolSerializerTrue(t *testing.T) {
	s := &memcached.BoolSerializer{}
	val, err := s.Deserialize([]byte("1"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v, ok := val.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", val)
	}
	if !v {
		t.Error("expected true")
	}
}

func TestBoolSerializerFalse(t *testing.T) {
	s := &memcached.BoolSerializer{}
	val, err := s.Deserialize([]byte(""))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v, ok := val.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", val)
	}
	if v {
		t.Error("expected false")
	}
}

func TestBoolSerializerZero(t *testing.T) {
	s := &memcached.BoolSerializer{}
	val, err := s.Deserialize([]byte("0"))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	v := val.(bool)
	if v {
		t.Error("expected false for '0'")
	}
}

// --- JSONSerializer ---

func TestJSONSerializer(t *testing.T) {
	s := &memcached.JSONSerializer{}
	val, err := s.Deserialize([]byte(`{"name":"Alice","age":30}`))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	if m["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", m["name"])
	}
	if m["age"] != float64(30) {
		t.Errorf("expected 30, got %v", m["age"])
	}
}

func TestJSONSerializerInvalid(t *testing.T) {
	s := &memcached.JSONSerializer{}
	_, err := s.Deserialize([]byte("{invalid"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestJSONSerializerArray(t *testing.T) {
	s := &memcached.JSONSerializer{}
	val, err := s.Deserialize([]byte(`[1,2,3]`))
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	arr, ok := val.([]any)
	if !ok {
		t.Fatalf("expected slice, got %T", val)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
}
