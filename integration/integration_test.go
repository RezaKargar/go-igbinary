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
}
