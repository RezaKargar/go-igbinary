# go-igbinary

[![CI](https://github.com/RezaKargar/go-igbinary/actions/workflows/ci.yml/badge.svg)](https://github.com/RezaKargar/go-igbinary/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/RezaKargar/go-igbinary.svg)](https://pkg.go.dev/github.com/RezaKargar/go-igbinary)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A pure Go decoder for [PHP's igbinary](https://github.com/igbinary/igbinary) serialization format. **Zero external dependencies** for the core decoder.

Use this library when you need to **read PHP igbinary-serialized data from memcached** (or any other source) in a Go application. This is a common need when migrating from PHP to Go, or when Go services need to read cache entries written by PHP.

## Features

- **Pure Go** -- no CGo, no PHP dependency
- **Zero external dependencies** for the core `igbinary` package (stdlib only)
- Decodes all igbinary v2 types: strings, integers, floats, booleans, nil, arrays, objects
- **String deduplication** support (igbinary's compact string table)
- **PHP memcached integration** via the `memcached` sub-package (handles decompression + flag-based dispatch)
- Builder pattern for customizing compressors and serializers
- Comprehensive test suite with 80%+ coverage
- Docker-based integration tests with real PHP memcached data

## Installation

```bash
go get github.com/RezaKargar/go-igbinary
```

## Quick Start

### Decode igbinary data directly

```go
package main

import (
    "fmt"
    "log"

    igbinary "github.com/RezaKargar/go-igbinary"
)

func main() {
    // igbinary-encoded integer 42: header (00 00 00 02) + PosInt8 (06) + value (2a)
    data := []byte{0x00, 0x00, 0x00, 0x02, 0x06, 0x2a}

    val, err := igbinary.Decode(data)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(val) // Output: 42
}
```

### Decode PHP memcached entries

```go
package main

import (
    "fmt"
    "log"

    "github.com/bradfitz/gomemcache/memcache"
    mc "github.com/RezaKargar/go-igbinary/memcached"
)

func main() {
    // Connect to memcached
    client := memcache.New("localhost:11211")

    // Create a codec (pre-configured for PHP memcached defaults)
    codec := mc.NewCodec()

    // Read and decode a cache entry written by PHP
    item, err := client.Get("my-php-cache-key")
    if err != nil {
        log.Fatal(err)
    }

    // The codec handles decompression (FastLZ/Zlib) and deserialization (igbinary/JSON)
    // automatically based on the flags field
    val, err := codec.Decode(item.Value, item.Flags)
    if err != nil {
        log.Fatal(err)
    }

    // PHP arrays become map[string]any in Go
    m := val.(map[string]any)
    fmt.Println(m["title"])
}
```

### Custom codec configuration

```go
codec := mc.NewCodecBuilder().
    WithCompressor(mc.FlagFastlz, &mc.FastlzCompressor{}).
    WithCompressor(mc.FlagZlib, mc.NewZlibCompressor(true)).
    WithSerializer(mc.FlagIgbinary, &mc.IgbinarySerializer{}).
    WithSerializer(mc.FlagJSON, &mc.JSONSerializer{}).
    WithSerializer(mc.FlagString, &mc.StringSerializer{}).
    WithFallbackCompressor(&mc.FastlzCompressor{}).
    WithFallbackSerializer(&mc.IgbinarySerializer{}).
    Build()
```

## Type Mapping

| PHP Type   | Go Type              | Notes                                              |
|------------|----------------------|----------------------------------------------------|
| `array`    | `map[string]any`     | Integer keys are converted to string keys (`"0"`)  |
| `string`   | `string`             | Deduplicated via string ID table                   |
| `integer`  | `int64`              | 8/16/32/64-bit, signed                             |
| `float`    | `float64`            | IEEE 754 double                                    |
| `boolean`  | `bool`               |                                                    |
| `NULL`     | `nil`                |                                                    |
| `object`   | `map[string]any`     | Class name stored under `"__class"` key            |

## Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│  github.com/RezaKargar/go-igbinary         (zero external deps)   │
│                                                                    │
│  Decode(data) -> any                                               │
│  NewDecoder(opts...) -> *Decoder                                   │
│  Type constants, error types                                       │
└──────────────────────────────┬─────────────────────────────────────┘
                               │ imported by
┌──────────────────────────────▼─────────────────────────────────────┐
│  github.com/RezaKargar/go-igbinary/memcached                      │
│                                                                    │
│  Codec          -- decompress + deserialize pipeline               │
│  Compressor     -- interface (FastLZ, Zlib built-in)               │
│  Serializer     -- interface (igbinary, JSON, String built-in)     │
│  Flag constants -- PHP memcached flag parsing                      │
└────────────────────────────────────────────────────────────────────┘
```

The root package has **zero external dependencies** -- it uses only `fmt` and `math` from the standard library. The `memcached` sub-package adds one external dependency ([`go-fastlz`](https://github.com/dgryski/go-fastlz)) for FastLZ decompression.

## How igbinary Works

igbinary is a compact binary serializer for PHP. It replaces PHP's text-based `serialize()` function with a binary format that is both smaller and faster. Understanding the format is useful for debugging and extending the decoder.

### Binary Format Overview

Every igbinary payload starts with a **4-byte header** identifying the format version, followed by a single **root value**:

```
[4-byte header] [root value]
```

**Header** (always `00 00 00 02` for version 2):
```
Offset  Bytes   Meaning
0       00      Reserved
1       00      Reserved  
2       00      Reserved
3       02      Format version (2)
```

### Type Tags

Each value begins with a single **type byte** that identifies what follows:

| Code   | Name           | Payload                                            |
|--------|----------------|----------------------------------------------------|
| `0x00` | nil            | (none)                                             |
| `0x04` | false          | (none)                                             |
| `0x05` | true           | (none)                                             |
| `0x06` | +int8          | 1 byte unsigned                                    |
| `0x07` | -int8          | 1 byte unsigned (negated)                          |
| `0x08` | +int16         | 2 bytes big-endian unsigned                        |
| `0x09` | -int16         | 2 bytes big-endian unsigned (negated)              |
| `0x0A` | +int32         | 4 bytes big-endian unsigned                        |
| `0x0B` | -int32         | 4 bytes big-endian unsigned (negated)              |
| `0x20` | +int64         | 8 bytes big-endian unsigned                        |
| `0x21` | -int64         | 8 bytes big-endian unsigned (negated)              |
| `0x0C` | double         | 8 bytes IEEE 754 big-endian                        |
| `0x0D` | empty string   | (none) -- registered in string table               |
| `0x11` | string8        | 1-byte length + bytes -- registered in string table|
| `0x12` | string16       | 2-byte length + bytes -- registered in string table|
| `0x13` | string32       | 4-byte length + bytes -- registered in string table|
| `0x0E` | string_id8     | 1-byte string table ID (back-reference)            |
| `0x0F` | string_id16    | 2-byte string table ID (back-reference)            |
| `0x10` | string_id32    | 4-byte string table ID (back-reference)            |
| `0x14` | array8         | 1-byte count + key/value pairs                     |
| `0x15` | array16        | 2-byte count + key/value pairs                     |
| `0x16` | array32        | 4-byte count + key/value pairs                     |
| `0x17` | object8        | 1-byte name length + name + properties (as array)  |
| `0x18` | object16       | 2-byte name length + name + properties (as array)  |
| `0x19` | object32       | 4-byte name length + name + properties (as array)  |

All multi-byte integers are **big-endian**.

### Integer Encoding

igbinary uses the smallest encoding that fits the value. Negative values use separate type codes rather than two's complement:

```
PHP:  42          -> 06 2A              (PosInt8)
PHP:  256         -> 08 01 00           (PosInt16, big-endian)
PHP:  -5          -> 07 05              (NegInt8, magnitude 5)
PHP:  100000      -> 0A 00 01 86 A0    (PosInt32, big-endian)
```

### String Deduplication

This is the key feature that makes igbinary compact. Strings are stored in a table indexed by order of first appearance:

```
First occurrence:   0x11 05 "hello"    -> registered as ID 0
Second occurrence:  0x0E 00            -> lookup ID 0 -> "hello"
Third occurrence:   0x0E 00            -> same lookup
```

When serializing PHP arrays where the same keys repeat across entries (e.g., `"id"`, `"name"`, `"email"` in a list of users), each key string is stored in full only once. All subsequent occurrences use a 2-byte reference (`0x0E` + ID) instead of re-encoding the full string. For a typical PHP cache entry with hundreds of repeated keys, this saves significant space.

### Array Encoding

PHP arrays are encoded as a count followed by alternating key-value pairs:

```
PHP:  ["name" => "Alice", "age" => 30]

Binary:
  14              array8 (count in next byte)
  02              2 entries
  11 04 "name"    key: new string "name" (registered as ID 0)
  11 05 "Alice"   value: new string "Alice" (registered as ID 1)
  11 03 "age"     key: new string "age" (registered as ID 2)
  06 1E           value: positive int 30
```

**Array keys** can be either strings or integers. PHP indexed arrays (`[0 => "a", 1 => "b"]`) use integer keys, which this decoder converts to their string representation (`"0"`, `"1"`).

### Object Encoding

PHP objects are encoded with a class name followed by properties (as an array):

```
PHP:  new User(name: "Alice")

Binary:
  17              object8 (class name length in next byte)
  04 "User"       class name "User" (registered in string table)
  14 01           properties: array of 1
  11 04 "name"    property key: "name"
  11 05 "Alice"   property value: "Alice"
```

Objects are decoded as `map[string]any` with the class name stored under the `"__class"` key.

When the same class appears multiple times, subsequent objects use `TypeObjectID` (codes `0x1A`-`0x1C`) to reference the class name by its string table ID instead of re-encoding it.

### PHP Memcached Flags

When PHP's memcached extension stores a value, it sets a 32-bit `flags` field that encodes both the serializer and compression used:

```
Bit layout of flags (uint32):

  31 ... 8   7       6       5       4       3 2 1 0
  [unused]  [zstd]  [fastlz] [zlib] [compr]  [type ]
            bit 7   bit 6    bit 5   bit 4   bits 0-3
```

Common flag values:

| flags | Binary        | Meaning                                               |
|-------|---------------|-------------------------------------------------------|
| `0`   | `0000 0000`   | Raw string, no compression                            |
| `5`   | `0000 0101`   | igbinary serialized, no compression                   |
| `85`  | `0101 0101`   | igbinary serialized + FastLZ compressed (most common)  |
| `53`  | `0011 0101`   | igbinary serialized + Zlib compressed                  |
| `6`   | `0000 0110`   | JSON serialized, no compression                       |

The `memcached` sub-package handles this flag decoding automatically.

## Implementing Custom Extensions

### Custom Compressor

```go
type ZstdCompressor struct{}

func (c *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
    return zstd.Decompress(nil, data)
}

codec := mc.NewCodecBuilder().
    WithCompressor(mc.FlagZstd, &ZstdCompressor{}).
    // ... other config
    Build()
```

### Custom Serializer

```go
type MsgpackSerializer struct{}

func (s *MsgpackSerializer) Deserialize(data []byte) (any, error) {
    var result any
    err := msgpack.Unmarshal(data, &result)
    return result, err
}

codec := mc.NewCodecBuilder().
    WithSerializer(mc.FlagMsgpack, &MsgpackSerializer{}).
    // ... other config
    Build()
```

## Decoder Options

The root package supports options for advanced usage:

```go
// Strict mode: returns errors for unresolved references instead of nil
dec := igbinary.NewDecoder(igbinary.WithStrictMode(true))
val, err := dec.Decode(data)
```

## Integration Testing

The `integration/` directory contains Docker-based tests that verify the decoder against real PHP-serialized memcached data. These tests use Docker Compose to spin up memcached and a PHP container -- **Docker is NOT a dependency of the library**, only of the tests.

```bash
# Run integration tests (requires Docker)
make integration-test
```

See [`integration/README.md`](integration/README.md) for details.

## Development

```bash
# Run unit tests
make test

# Run tests with coverage
make test-cover

# Run linter
make lint

# Run all CI checks
make ci

# Run integration tests (requires Docker)
make integration-test
```

## License

[MIT](LICENSE)
