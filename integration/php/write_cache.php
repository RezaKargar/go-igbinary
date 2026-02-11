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

// --- PHP class objects ---

// 15. Simple stdClass object
$obj = new stdClass();
$obj->name = 'Charlie';
$obj->age = 25;
$obj->active = true;
$mc->set('test:stdclass', $obj);
echo "Set test:stdclass\n";

// 16. Custom class - simple
class Product {
    public int $id;
    public string $title;
    public float $price;
    public bool $inStock;

    public function __construct(int $id, string $title, float $price, bool $inStock) {
        $this->id = $id;
        $this->title = $title;
        $this->price = $price;
        $this->inStock = $inStock;
    }
}

$product = new Product(42, 'Widget', 19.99, true);
$mc->set('test:product', $product);
echo "Set test:product\n";

// 17. Class with nested object
class Address {
    public string $street;
    public string $city;
    public string $country;

    public function __construct(string $street, string $city, string $country) {
        $this->street = $street;
        $this->city = $city;
        $this->country = $country;
    }
}

class User {
    public int $id;
    public string $name;
    public ?string $email;
    public Address $address;
    public array $roles;

    public function __construct(int $id, string $name, ?string $email, Address $address, array $roles) {
        $this->id = $id;
        $this->name = $name;
        $this->email = $email;
        $this->address = $address;
        $this->roles = $roles;
    }
}

$addr = new Address('123 Main St', 'Springfield', 'US');
$user = new User(1, 'Alice', 'alice@example.com', $addr, ['admin', 'editor']);
$mc->set('test:user_obj', $user);
echo "Set test:user_obj\n";

// 18. Class with null property
$userNoEmail = new User(2, 'Bob', null, new Address('456 Oak Ave', 'Portland', 'US'), ['viewer']);
$mc->set('test:user_null_prop', $userNoEmail);
echo "Set test:user_null_prop\n";

// 19. Inheritance
class Animal {
    public string $species;
    public int $legs;

    public function __construct(string $species, int $legs) {
        $this->species = $species;
        $this->legs = $legs;
    }
}

class Dog extends Animal {
    public string $breed;
    public string $name;

    public function __construct(string $breed, string $name) {
        parent::__construct('Canis familiaris', 4);
        $this->breed = $breed;
        $this->name = $name;
    }
}

$dog = new Dog('Labrador', 'Rex');
$mc->set('test:dog', $dog);
echo "Set test:dog\n";

// 20. Array of objects (multiple instances of same class)
$products = [
    new Product(1, 'Alpha', 10.00, true),
    new Product(2, 'Beta', 20.50, false),
    new Product(3, 'Gamma', 30.99, true),
];
$mc->set('test:product_list', $products);
echo "Set test:product_list\n";

// 21. Object with empty array property
class Container {
    public string $label;
    public array $items;

    public function __construct(string $label, array $items) {
        $this->label = $label;
        $this->items = $items;
    }
}

$empty = new Container('empty box', []);
$mc->set('test:empty_container', $empty);
echo "Set test:empty_container\n";

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
