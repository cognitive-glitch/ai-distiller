<?php

namespace App\Data\Repositories;

use App\Data\Models\Product;
use App\Data\Contracts\{Cacheable, Deletable};
use App\Data\Traits\HasSoftDeletes;
use \Serializable; // Implementing a built-in PHP interface

// An attribute for declarative configuration
#[\Attribute(\Attribute::TARGET_CLASS)]
class RepositoryConfig
{
    public function __construct(public string $model) {}
}

interface FindableById
{
    /**
     * @return object|null
     */
    public function find(int $id);
}

#[RepositoryConfig(model: Product::class)]
class ProductRepository extends BaseRepository implements FindableById, Cacheable, Deletable, Serializable
{
    // Trait provides the soft delete functionality
    use HasSoftDeletes;

    private static int $queryCount = 0;
    protected array $searchableFields = ['name', 'sku'];

    public function __construct()
    {
        // Set the model for the parent repository
        parent::__construct(Product::class);
    }

    /**
     * Find a product by its ID.
     *
     * @param int $id
     * @return Product|null The found product or null.
     */
    public function find(int $id): ?Product
    {
        self::$queryCount++;
        // In a real app, this would query a database.
        // We return a mock object for this example.
        if ($id === 1) {
            return new Product(1, 'Laptop', 1500.00);
        }
        return null;
    }

    /**
     * Find products by a specific property.
     * This demonstrates a more complex return type via DocBlock.
     *
     * @param string $field
     * @param mixed $value
     * @return Product[]
     */
    public function findBy(string $field, mixed $value): array
    {
        self::$queryCount++;
        // Imagine DB query logic here...
        return [new Product(1, 'Laptop', 1500.00)];
    }

    // --- Interface Implementations ---

    public function clearCache(): bool { /* ... */ return true; }
    public function serialize(): string { /* ... */ return ''; }
    public function unserialize(string $data): void { /* ... */ }
}

// Note: Other classes like Product, BaseRepository, etc., would be defined elsewhere.
// This construct specifically tests the AI's ability to analyze ProductRepository.