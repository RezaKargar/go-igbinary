//go:build integration

// Package integration provides end-to-end tests that verify the Go igbinary
// decoder against real PHP-serialized memcached data.
//
// These tests require Docker (memcached + PHP containers) and are gated behind
// the "integration" build tag. Run via: make integration-test
package integration

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"gopkg.in/yaml.v3"

	igbinary "github.com/RezaKargar/go-igbinary"
	"github.com/RezaKargar/go-igbinary/memcached"
)

// config holds integration test settings.
type config struct {
	Memcached struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"memcached"`
}

// loadConfig reads config.yml, then .env, then environment variables.
// Priority: env vars > .env > config.yml
func loadConfig() (*config, error) {
	cfg := &config{}
	cfg.Memcached.Host = "localhost"
	cfg.Memcached.Port = 11211

	// 1. Read config.yml defaults
	yamlData, err := os.ReadFile("config.yml")
	if err == nil {
		_ = yaml.Unmarshal(yamlData, cfg)
	}

	// 2. Read .env file (overrides yaml)
	loadDotEnv(".env")

	// 3. Environment variables (highest priority)
	if h := os.Getenv("MEMCACHED_HOST"); h != "" {
		cfg.Memcached.Host = h
	}
	if p := os.Getenv("MEMCACHED_PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err == nil {
			cfg.Memcached.Port = port
		}
	}

	return cfg, nil
}

// loadDotEnv reads a .env file and sets environment variables (does not override existing).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func TestIntegration(t *testing.T) {
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Memcached.Host, cfg.Memcached.Port)
	t.Logf("Connecting to memcached at %s", addr)

	mc := memcache.New(addr)
	if err := mc.Ping(); err != nil {
		t.Fatalf("cannot connect to memcached at %s: %v\n"+
			"Make sure Docker containers are running (make integration-test)", addr, err)
	}

	codec := memcached.NewCodec()

	// --- Test: Simple string ---
	t.Run("string", func(t *testing.T) {
		item, err := mc.Get("test:string")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		str, ok := val.(string)
		if !ok {
			t.Fatalf("expected string, got %T", val)
		}
		if str != "hello world" {
			t.Errorf("expected %q, got %q", "hello world", str)
		}
	})

	// --- Test: Integer ---
	t.Run("int", func(t *testing.T) {
		item, err := mc.Get("test:int")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
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
	})

	// --- Test: Float ---
	t.Run("float", func(t *testing.T) {
		item, err := mc.Get("test:float")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		v, ok := val.(float64)
		if !ok {
			t.Fatalf("expected float64, got %T (%v)", val, val)
		}
		if v != 3.14 {
			t.Errorf("expected 3.14, got %f", v)
		}
	})

	// --- Test: Booleans ---
	t.Run("bool_true", func(t *testing.T) {
		item, err := mc.Get("test:bool_true")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		v, ok := val.(bool)
		if !ok {
			t.Fatalf("expected bool, got %T (%v)", val, val)
		}
		if !v {
			t.Error("expected true")
		}
	})

	t.Run("bool_false", func(t *testing.T) {
		item, err := mc.Get("test:bool_false")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		v, ok := val.(bool)
		if !ok {
			t.Fatalf("expected bool, got %T (%v)", val, val)
		}
		if v {
			t.Error("expected false")
		}
	})

	// --- Test: Null ---
	t.Run("null", func(t *testing.T) {
		item, err := mc.Get("test:null")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})

	// --- Test: Associative array ---
	t.Run("assoc", func(t *testing.T) {
		item, err := mc.Get("test:assoc")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m["name"] != "Alice" {
			t.Errorf("name: expected Alice, got %v", m["name"])
		}
		if m["age"] != int64(30) {
			t.Errorf("age: expected 30, got %v", m["age"])
		}
		if m["email"] != "alice@example.com" {
			t.Errorf("email: expected alice@example.com, got %v", m["email"])
		}
	})

	// --- Test: Nested array ---
	t.Run("nested", func(t *testing.T) {
		item, err := mc.Get("test:nested")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		user, ok := m["user"].(map[string]any)
		if !ok {
			t.Fatalf("expected user to be map, got %T", m["user"])
		}
		if user["id"] != int64(123) {
			t.Errorf("user.id: expected 123, got %v", user["id"])
		}
		if user["name"] != "Bob" {
			t.Errorf("user.name: expected Bob, got %v", user["name"])
		}
		if m["active"] != true {
			t.Errorf("active: expected true, got %v", m["active"])
		}
	})

	// --- Test: Indexed array (list) ---
	t.Run("list", func(t *testing.T) {
		item, err := mc.Get("test:list")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		// PHP indexed arrays become map with string keys "0", "1", "2"
		if m["0"] != "apple" {
			t.Errorf("0: expected apple, got %v", m["0"])
		}
		if m["1"] != "banana" {
			t.Errorf("1: expected banana, got %v", m["1"])
		}
		if m["2"] != "cherry" {
			t.Errorf("2: expected cherry, got %v", m["2"])
		}
	})

	// --- Test: Mixed types ---
	t.Run("mixed", func(t *testing.T) {
		item, err := mc.Get("test:mixed")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m["string_val"] != "test" {
			t.Errorf("string_val: expected test, got %v", m["string_val"])
		}
		if m["int_val"] != int64(999) {
			t.Errorf("int_val: expected 999, got %v", m["int_val"])
		}
		if m["bool_val"] != false {
			t.Errorf("bool_val: expected false, got %v", m["bool_val"])
		}
		if m["null_val"] != nil {
			t.Errorf("null_val: expected nil, got %v", m["null_val"])
		}
	})

	// --- Test: Large integer ---
	t.Run("large_int", func(t *testing.T) {
		item, err := mc.Get("test:large_int")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		v, ok := val.(int64)
		if !ok {
			t.Fatalf("expected int64, got %T (%v)", val, val)
		}
		if v != 9999999999 {
			t.Errorf("expected 9999999999, got %d", v)
		}
	})

	// --- Test: Negative integer ---
	t.Run("negative_int", func(t *testing.T) {
		item, err := mc.Get("test:negative_int")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		v, ok := val.(int64)
		if !ok {
			t.Fatalf("expected int64, got %T (%v)", val, val)
		}
		if v != -42 {
			t.Errorf("expected -42, got %d", v)
		}
	})

	// --- Test: Empty array ---
	t.Run("empty_array", func(t *testing.T) {
		item, err := mc.Get("test:empty_array")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if len(m) != 0 {
			t.Errorf("expected empty map, got %d entries", len(m))
		}
	})

	// --- Test: Compressed data ---
	t.Run("compressed", func(t *testing.T) {
		item, err := mc.Get("test:compressed")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		t.Logf("Compressed item flags: %d (%s)", item.Flags, memcached.ExplainFlags(item.Flags))

		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m["title"] != "Compressed Data Test" {
			t.Errorf("title: expected 'Compressed Data Test', got %v", m["title"])
		}
		if m["count"] != int64(42) {
			t.Errorf("count: expected 42, got %v", m["count"])
		}
	})

	// =====================
	// PHP class object tests
	// =====================

	// --- Test: stdClass ---
	t.Run("stdclass", func(t *testing.T) {
		item, err := mc.Get("test:stdclass")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m[igbinary.ClassKey] != "stdClass" {
			t.Errorf("__class: expected stdClass, got %v", m[igbinary.ClassKey])
		}
		if m["name"] != "Charlie" {
			t.Errorf("name: expected Charlie, got %v", m["name"])
		}
		if m["age"] != int64(25) {
			t.Errorf("age: expected 25, got %v", m["age"])
		}
		if m["active"] != true {
			t.Errorf("active: expected true, got %v", m["active"])
		}
	})

	// --- Test: Custom class (Product) ---
	t.Run("product", func(t *testing.T) {
		item, err := mc.Get("test:product")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m[igbinary.ClassKey] != "Product" {
			t.Errorf("__class: expected Product, got %v", m[igbinary.ClassKey])
		}
		if m["id"] != int64(42) {
			t.Errorf("id: expected 42, got %v", m["id"])
		}
		if m["title"] != "Widget" {
			t.Errorf("title: expected Widget, got %v", m["title"])
		}
		if m["price"] != 19.99 {
			t.Errorf("price: expected 19.99, got %v", m["price"])
		}
		if m["inStock"] != true {
			t.Errorf("inStock: expected true, got %v", m["inStock"])
		}
	})

	// --- Test: Nested objects (User with Address) ---
	t.Run("user_obj", func(t *testing.T) {
		item, err := mc.Get("test:user_obj")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m[igbinary.ClassKey] != "User" {
			t.Errorf("__class: expected User, got %v", m[igbinary.ClassKey])
		}
		if m["id"] != int64(1) {
			t.Errorf("id: expected 1, got %v", m["id"])
		}
		if m["name"] != "Alice" {
			t.Errorf("name: expected Alice, got %v", m["name"])
		}
		if m["email"] != "alice@example.com" {
			t.Errorf("email: expected alice@example.com, got %v", m["email"])
		}

		// Nested Address object
		addr, ok := m["address"].(map[string]any)
		if !ok {
			t.Fatalf("address: expected map, got %T", m["address"])
		}
		if addr[igbinary.ClassKey] != "Address" {
			t.Errorf("address.__class: expected Address, got %v", addr[igbinary.ClassKey])
		}
		if addr["street"] != "123 Main St" {
			t.Errorf("address.street: expected '123 Main St', got %v", addr["street"])
		}
		if addr["city"] != "Springfield" {
			t.Errorf("address.city: expected Springfield, got %v", addr["city"])
		}
		if addr["country"] != "US" {
			t.Errorf("address.country: expected US, got %v", addr["country"])
		}

		// Roles array (indexed array inside object)
		roles, ok := m["roles"].(map[string]any)
		if !ok {
			t.Fatalf("roles: expected map, got %T", m["roles"])
		}
		if roles["0"] != "admin" {
			t.Errorf("roles[0]: expected admin, got %v", roles["0"])
		}
		if roles["1"] != "editor" {
			t.Errorf("roles[1]: expected editor, got %v", roles["1"])
		}
	})

	// --- Test: Object with null property ---
	t.Run("user_null_prop", func(t *testing.T) {
		item, err := mc.Get("test:user_null_prop")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m[igbinary.ClassKey] != "User" {
			t.Errorf("__class: expected User, got %v", m[igbinary.ClassKey])
		}
		if m["id"] != int64(2) {
			t.Errorf("id: expected 2, got %v", m["id"])
		}
		if m["name"] != "Bob" {
			t.Errorf("name: expected Bob, got %v", m["name"])
		}
		if m["email"] != nil {
			t.Errorf("email: expected nil, got %v", m["email"])
		}
	})

	// --- Test: Inherited class (Dog extends Animal) ---
	t.Run("dog_inheritance", func(t *testing.T) {
		item, err := mc.Get("test:dog")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		// igbinary serializes the concrete class, not the parent
		if m[igbinary.ClassKey] != "Dog" {
			t.Errorf("__class: expected Dog, got %v", m[igbinary.ClassKey])
		}
		// Inherited properties from Animal
		if m["species"] != "Canis familiaris" {
			t.Errorf("species: expected 'Canis familiaris', got %v", m["species"])
		}
		if m["legs"] != int64(4) {
			t.Errorf("legs: expected 4, got %v", m["legs"])
		}
		// Own properties from Dog
		if m["breed"] != "Labrador" {
			t.Errorf("breed: expected Labrador, got %v", m["breed"])
		}
		if m["name"] != "Rex" {
			t.Errorf("name: expected Rex, got %v", m["name"])
		}
	})

	// --- Test: Array of objects ---
	t.Run("product_list", func(t *testing.T) {
		item, err := mc.Get("test:product_list")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}

		// PHP indexed array of objects: keys "0", "1", "2"
		for i, expected := range []struct {
			id    int64
			title string
		}{
			{1, "Alpha"},
			{2, "Beta"},
			{3, "Gamma"},
		} {
			key := fmt.Sprintf("%d", i)
			obj, ok := m[key].(map[string]any)
			if !ok {
				t.Fatalf("product[%d]: expected map, got %T", i, m[key])
			}
			if obj[igbinary.ClassKey] != "Product" {
				t.Errorf("product[%d].__class: expected Product, got %v", i, obj[igbinary.ClassKey])
			}
			if obj["id"] != expected.id {
				t.Errorf("product[%d].id: expected %d, got %v", i, expected.id, obj["id"])
			}
			if obj["title"] != expected.title {
				t.Errorf("product[%d].title: expected %s, got %v", i, expected.title, obj["title"])
			}
		}
	})

	// --- Test: Object with empty array property ---
	t.Run("empty_container", func(t *testing.T) {
		item, err := mc.Get("test:empty_container")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		val, err := codec.Decode(item.Value, item.Flags)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		m, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if m[igbinary.ClassKey] != "Container" {
			t.Errorf("__class: expected Container, got %v", m[igbinary.ClassKey])
		}
		if m["label"] != "empty box" {
			t.Errorf("label: expected 'empty box', got %v", m["label"])
		}
		items, ok := m["items"].(map[string]any)
		if !ok {
			t.Fatalf("items: expected map, got %T", m["items"])
		}
		if len(items) != 0 {
			t.Errorf("items: expected empty map, got %d entries", len(items))
		}
	})
}
