<?php

declare(strict_types=1);

namespace App\EdgeCases;

use Closure;
use Generator;
use Countable;
use Iterator;

/**
 * Edge case demonstrations for PHP docblock annotations
 * Testing advanced type annotations used in real-world PHP projects
 */

/**
 * Class demonstrating magic properties and complex array shapes
 *
 * @property-read int $id Auto-generated ID
 * @property string $name Mutable name property
 * @property-write array<string, mixed> $metadata Write-only metadata
 */
class MagicModel
{
    /**
     * @var array<string, mixed>
     */
    private array $data = [];

    /**
     * Magic getter
     *
     * @param string $key Property name
     * @return mixed
     */
    public function __get(string $key): mixed
    {
        return $this->data[$key] ?? null;
    }

    /**
     * Magic setter
     *
     * @param string $key Property name
     * @param mixed $value Property value
     */
    public function __set(string $key, mixed $value): void
    {
        $this->data[$key] = $value;
    }
}

/**
 * Service demonstrating class-string and template parameters
 */
class ContainerService
{
    /**
     * Create instance of a class
     *
     * @template T of object
     * @param class-string<T> $className Fully qualified class name
     * @param array<string, mixed> $parameters Constructor parameters
     * @return T
     */
    public function make(string $className, array $parameters = []): object
    {
        return new $className(...$parameters);
    }

    /**
     * Register a factory
     *
     * @param literal-string $id Service identifier (must be literal)
     * @param callable(): object $factory Factory function
     */
    public function register(string $id, callable $factory): void
    {
        // Registration logic
    }
}

/**
 * Data transfer object with complex array shapes
 */
class ComplexDto
{
    /**
     * Process complex payload
     *
     * @param array{
     *   id: int,
     *   name?: non-empty-string,
     *   tags: non-empty-list<string>,
     *   status: key-of<self::STATUS_MAP>,
     *   meta: array{
     *     created_at: string,
     *     updated_at?: string,
     *     extra?: array<string, mixed>
     *   },
     *   0?: string,
     *   1?: int
     * } $payload Complex nested structure with optional keys
     * @return array{success: bool, data?: array<string, mixed>, errors?: list<string>}
     */
    public function process(array $payload): array
    {
        // Processing logic
        return ['success' => true];
    }

    /**
     * @var array<string, int>
     */
    public const STATUS_MAP = [
        'draft' => 0,
        'published' => 1,
        'archived' => 2
    ];

    /**
     * Get random status
     *
     * @return key-of<self::STATUS_MAP>
     */
    public function getRandomStatus(): string
    {
        $keys = array_keys(self::STATUS_MAP);
        return $keys[array_rand($keys)];
    }

    /**
     * Get status value
     *
     * @param key-of<self::STATUS_MAP> $status
     * @return value-of<self::STATUS_MAP>
     */
    public function getStatusValue(string $status): int
    {
        return self::STATUS_MAP[$status];
    }
}

/**
 * Collection class with method-level templates
 */
class TypedCollection
{
    /**
     * @var list<mixed>
     */
    private array $items = [];

    /**
     * Get first item matching predicate
     *
     * @template T
     * @param callable(mixed): bool $predicate
     * @param callable(mixed): T $mapper Transform function
     * @return T|null
     */
    public function firstWhere(callable $predicate, callable $mapper): mixed
    {
        foreach ($this->items as $item) {
            if ($predicate($item)) {
                return $mapper($item);
            }
        }
        return null;
    }

    /**
     * Map items to new type
     *
     * @template TKey of array-key
     * @template TValue
     * @param callable(mixed): array{0: TKey, 1: TValue} $mapper
     * @return array<TKey, TValue>
     */
    public function mapToAssoc(callable $mapper): array
    {
        $result = [];
        foreach ($this->items as $item) {
            [$key, $value] = $mapper($item);
            $result[$key] = $value;
        }
        return $result;
    }
}

/**
 * Callable type demonstrations
 */
class CallableTypes
{
    /**
     * @var callable(int, string): bool
     */
    private $validator;

    /**
     * @var Closure(array<string, mixed>): Generator<int, string, mixed, void>
     */
    private Closure $generator;

    /**
     * Set validator
     *
     * @param callable(int, string=): bool $validator Validator function
     */
    public function setValidator(callable $validator): void
    {
        $this->validator = $validator;
    }

    /**
     * Execute with callback
     *
     * @param callable-string $callback Callback function name
     * @param array<int, mixed> $args Arguments
     * @return mixed
     */
    public function execute(string $callback, array $args = []): mixed
    {
        return $callback(...$args);
    }

    /**
     * Create processor
     *
     * @return Closure(non-empty-list<int>): array{min: int, max: int, avg: float}
     */
    public function createProcessor(): Closure
    {
        return function (array $numbers): array {
            return [
                'min' => min($numbers),
                'max' => max($numbers),
                'avg' => array_sum($numbers) / count($numbers)
            ];
        };
    }
}

/**
 * Intersection and union type demonstrations
 */
interface Timestampable
{
    public function getTimestamp(): int;
}

interface Identifiable
{
    public function getId(): int;
}

/**
 * @template T of (Timestampable&Identifiable)
 */
class IntersectionHandler
{
    /**
     * Process items with intersection types
     *
     * @param array<int, (Timestampable&Identifiable)> $items
     * @return list<array{id: int, timestamp: int}>
     */
    public function process(array $items): array
    {
        $result = [];
        foreach ($items as $item) {
            $result[] = [
                'id' => $item->getId(),
                'timestamp' => $item->getTimestamp()
            ];
        }
        return $result;
    }

    /**
     * Filter by type
     *
     * @param array<int, (Countable|Iterator)> $items Mixed types
     * @return array{countables: list<Countable>, iterators: list<Iterator>}
     */
    public function categorize(array $items): array
    {
        $countables = [];
        $iterators = [];

        foreach ($items as $item) {
            if ($item instanceof Countable) {
                $countables[] = $item;
            }
            if ($item instanceof Iterator) {
                $iterators[] = $item;
            }
        }

        return [
            'countables' => $countables,
            'iterators' => $iterators
        ];
    }
}

/**
 * Enum and int-mask demonstrations
 */
enum Permission: int
{
    case READ = 1;
    case WRITE = 2;
    case DELETE = 4;
    case ADMIN = 8;
}

class PermissionChecker
{
    /**
     * Check permissions
     *
     * @param int-mask-of<Permission::*> $userPermissions
     * @param int-mask<1, 2, 4, 8> $requiredPermissions
     * @return bool
     */
    public function hasPermission(int $userPermissions, int $requiredPermissions): bool
    {
        return ($userPermissions & $requiredPermissions) === $requiredPermissions;
    }

    /**
     * Get all permissions
     *
     * @return list<value-of<Permission>>
     */
    public function getAllPermissionValues(): array
    {
        return array_map(fn($case) => $case->value, Permission::cases());
    }
}

/**
 * Conditional return types and assertions
 */
class ValidationService
{
    /**
     * Validate non-empty string
     *
     * @psalm-assert-if-true non-empty-string $value
     * @psalm-assert-if-false null|'' $value
     */
    public function isNonEmpty(?string $value): bool
    {
        return $value !== null && $value !== '';
    }

    /**
     * Parse or default
     *
     * @template T
     * @param string $json
     * @param T $default
     * @psalm-return (T is null ? array<string, mixed>|null : array<string, mixed>|T)
     * @return mixed
     */
    public function parseJsonOrDefault(string $json, mixed $default = null): mixed
    {
        $result = json_decode($json, true);
        return $result !== null ? $result : $default;
    }

    /**
     * Ensure array
     *
     * @template T
     * @param T $value
     * @psalm-return (T is array ? T : array{0: T})
     * @return array
     */
    public function ensureArray(mixed $value): array
    {
        return is_array($value) ? $value : [$value];
    }
}

/**
 * Advanced Generator types
 */
class GeneratorTypes
{
    /**
     * Generate pairs
     *
     * @template TKey
     * @template TValue
     * @param array<TKey, TValue> $data
     * @return Generator<int, array{key: TKey, value: TValue}, mixed, void>
     */
    public function generatePairs(array $data): Generator
    {
        foreach ($data as $key => $value) {
            yield ['key' => $key, 'value' => $value];
        }
    }

    /**
     * Bidirectional generator
     *
     * @return Generator<int, string, bool, int>
     */
    public function bidirectionalGenerator(): Generator
    {
        $count = 0;
        $continue = true;

        while ($continue) {
            $continue = yield "Item $count";
            $count++;
        }

        return $count;
    }
}

/**
 * Type aliases for reusability
 *
 * @psalm-type UserId = positive-int
 * @psalm-type UserData = array{
 *   id: UserId,
 *   email: non-empty-string,
 *   roles: non-empty-list<non-empty-string>,
 *   metadata?: array<string, scalar>
 * }
 * @phpstan-type ErrorResponse array{error: true, message: string, code?: int}
 */
class TypeAliasDemo
{
    /**
     * Get user
     *
     * @param UserId $id
     * @return UserData|ErrorResponse
     */
    public function getUser(int $id): array
    {
        if ($id <= 0) {
            return ['error' => true, 'message' => 'Invalid ID'];
        }

        return [
            'id' => $id,
            'email' => 'user@example.com',
            'roles' => ['user']
        ];
    }

    /**
     * Batch operation
     *
     * @param list<UserId> $ids
     * @return array<UserId, UserData|ErrorResponse>
     */
    public function batchGet(array $ids): array
    {
        $result = [];
        foreach ($ids as $id) {
            $result[$id] = $this->getUser($id);
        }
        return $result;
    }
}

/**
 * Numeric literal keys and special array constructs
 */
class NumericKeyArrays
{
    /**
     * Process tuple
     *
     * @param array{0: string, 1: int, 2?: bool} $tuple
     * @return array{name: string, age: int, active: bool}
     */
    public function processTuple(array $tuple): array
    {
        return [
            'name' => $tuple[0],
            'age' => $tuple[1],
            'active' => $tuple[2] ?? true
        ];
    }

    /**
     * Matrix operations
     *
     * @param non-empty-array<int<0, 10>, non-empty-array<int<0, 10>, float>> $matrix
     * @return array{rows: int<1, 11>, cols: int<1, 11>, sum: float}
     */
    public function analyzeMatrix(array $matrix): array
    {
        $sum = 0.0;
        $rows = count($matrix);
        $cols = count(reset($matrix));

        foreach ($matrix as $row) {
            $sum += array_sum($row);
        }

        return [
            'rows' => $rows,
            'cols' => $cols,
            'sum' => $sum
        ];
    }
}

/**
 * Global function with type imports
 *
 * @param \App\EdgeCases\TypeAliasDemo::UserId $userId
 * @return \App\EdgeCases\TypeAliasDemo::UserData
 */
function getUserGlobal(int $userId): array
{
    return [
        'id' => $userId,
        'email' => 'global@example.com',
        'roles' => ['admin']
    ];
}