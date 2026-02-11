# Integration Tests

This directory contains integration tests that verify the Go igbinary decoder against
real PHP-serialized memcached data.

**Docker and Docker Compose are required only for running these tests.** They are NOT
dependencies of the `go-igbinary` library itself.

## How It Works

1. Docker Compose starts a **memcached** container and a **PHP** container.
2. The PHP script (`php/write_cache.php`) writes various data types to memcached
   using PHP's igbinary serializer.
3. The Go integration test connects to memcached, reads the entries, and verifies
   that the decoded values match the expected PHP values.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- Go 1.21+

## Running

The easiest way is via the Makefile from the project root:

```bash
make integration-test
```

Or manually:

```bash
# Start memcached and populate test data
cd integration
docker compose up -d --build
docker compose run --rm php-writer

# Run the Go integration test
cd ..
MEMCACHED_HOST=localhost MEMCACHED_PORT=11211 go test -v -tags=integration ./integration/

# Clean up
cd integration
docker compose down -v
```

## Configuration

Default settings are in `config.yml`. Override with a `.env` file or environment variables:

```bash
cp .env.example .env
# Edit .env as needed
```

Priority: environment variables > `.env` file > `config.yml`
