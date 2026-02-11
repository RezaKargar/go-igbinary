<?php
/**
 * Writes test data to memcached using igbinary serialization.
 *
 * This script is used by the integration tests to populate memcached with
 * igbinary-serialized data that the Go decoder will then read and verify.
 *
 * Environment variables:
 *   MEMCACHED_HOST - memcached hostname (default: memcached)
 *   MEMCACHED_PORT - memcached port (default: 11211)
 */

$host = getenv('MEMCACHED_HOST') ?: 'memcached';
$port = (int)(getenv('MEMCACHED_PORT') ?: 11211);

echo "Connecting to memcached at {$host}:{$port}\n";

$mc = new Memcached();
$mc->addServer($host, $port);

// Use igbinary serializer
$mc->setOption(Memcached::OPT_SERIALIZER, Memcached::SERIALIZER_IGBINARY);

// Disable compression to test pure igbinary decoding
$mc->setOption(Memcached::OPT_COMPRESSION, false);

// --- Test data ---

// 1. Simple string
$mc->set('test:string', 'hello world');
echo "Set test:string\n";

// 2. Integer
$mc->set('test:int', 42);
echo "Set test:int\n";

// 3. Float
$mc->set('test:float', 3.14);
echo "Set test:float\n";

// 4. Boolean
$mc->set('test:bool_true', true);
$mc->set('test:bool_false', false);
echo "Set test:bool_true and test:bool_false\n";

// 5. Null
$mc->set('test:null', null);
echo "Set test:null\n";

// 6. Simple associative array
$mc->set('test:assoc', [
    'name'  => 'Alice',
    'age'   => 30,
    'email' => 'alice@example.com',
]);
echo "Set test:assoc\n";

// 7. Nested array
$mc->set('test:nested', [
    'user' => [
        'id'   => 123,
        'name' => 'Bob',
        'tags' => ['admin', 'editor'],
    ],
    'active' => true,
]);
echo "Set test:nested\n";

// 8. Indexed array (PHP list)
$mc->set('test:list', ['apple', 'banana', 'cherry']);
echo "Set test:list\n";

// 9. Mixed-type array
$mc->set('test:mixed', [
    'string_val' => 'test',
    'int_val'    => 999,
    'float_val'  => 2.718,
    'bool_val'   => false,
    'null_val'   => null,
]);
echo "Set test:mixed\n";

// 10. Large integer
$mc->set('test:large_int', 9999999999);
echo "Set test:large_int\n";

// 11. Negative integer
$mc->set('test:negative_int', -42);
echo "Set test:negative_int\n";

// 12. Empty array
$mc->set('test:empty_array', []);
echo "Set test:empty_array\n";

// 13. Empty string
$mc->set('test:empty_string', '');
echo "Set test:empty_string\n";

// Now also write with compression enabled for the compressed test
$mc->setOption(Memcached::OPT_COMPRESSION, true);

// 14. Compressed igbinary array (will use FastLZ by default)
$mc->set('test:compressed', [
    'title'       => 'Compressed Data Test',
    'description' => str_repeat('This is a long string for compression testing. ', 10),
    'count'       => 42,
]);
echo "Set test:compressed (with compression)\n";

echo "\nAll test data written successfully.\n";
echo "Result code: " . $mc->getResultCode() . "\n";
