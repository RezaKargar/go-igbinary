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
