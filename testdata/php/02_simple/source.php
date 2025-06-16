<?php

declare(strict_types=1);

namespace App\Simple;

use DateTimeInterface;
use JsonSerializable;

/**
 * Interface for entities that can be persisted
 */
interface PersistableInterface
{
    /**
     * Get entity ID
     * 
     * @return int|null
     */
    public function getId(): ?int;
    
    /**
     * Set entity ID
     * 
     * @param int $id Entity ID
     */
    public function setId(int $id): void;
    
    /**
     * Get creation timestamp
     * 
     * @return DateTimeInterface|null
     */
    public function getCreatedAt(): ?DateTimeInterface;
}

/**
 * Interface for cacheable entities
 */
interface CacheableInterface
{
    /**
     * Get cache key
     * 
     * @return string
     */
    public function getCacheKey(): string;
    
    /**
     * Get cache TTL in seconds
     * 
     * @return int
     */
    public function getCacheTtl(): int;
}

/**
 * Abstract base entity class
 */
abstract class BaseEntity implements PersistableInterface, JsonSerializable
{
    /**
     * @var int|null Entity ID
     */
    protected ?int $id = null;
    
    /**
     * @var DateTimeInterface|null Creation timestamp
     */
    protected ?DateTimeInterface $createdAt = null;
    
    /**
     * @var DateTimeInterface|null Update timestamp
     */
    protected ?DateTimeInterface $updatedAt = null;

    /**
     * @inheritDoc
     */
    public function getId(): ?int
    {
        return $this->id;
    }

    /**
     * @inheritDoc
     */
    public function setId(int $id): void
    {
        $this->id = $id;
    }

    /**
     * @inheritDoc
     */
    public function getCreatedAt(): ?DateTimeInterface
    {
        return $this->createdAt;
    }

    /**
     * Set creation timestamp
     * 
     * @param DateTimeInterface $createdAt Creation timestamp
     */
    public function setCreatedAt(DateTimeInterface $createdAt): void
    {
        $this->createdAt = $createdAt;
    }

    /**
     * Get update timestamp
     * 
     * @return DateTimeInterface|null
     */
    public function getUpdatedAt(): ?DateTimeInterface
    {
        return $this->updatedAt;
    }

    /**
     * Set update timestamp
     * 
     * @param DateTimeInterface $updatedAt Update timestamp
     */
    public function setUpdatedAt(DateTimeInterface $updatedAt): void
    {
        $this->updatedAt = $updatedAt;
    }

    /**
     * Abstract method to get entity name
     * 
     * @return string
     */
    abstract public function getEntityName(): string;

    /**
     * Abstract method to validate entity
     * 
     * @return bool
     */
    abstract protected function validate(): bool;

    /**
     * @inheritDoc
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'entity_name' => $this->getEntityName(),
            'created_at' => $this->createdAt?->format('c'),
            'updated_at' => $this->updatedAt?->format('c'),
        ];
    }
}

/**
 * Trait for timestampable entities
 */
trait TimestampableTrait
{
    /**
     * Update timestamps
     */
    protected function updateTimestamps(): void
    {
        $now = new \DateTime();
        
        if ($this->createdAt === null) {
            $this->createdAt = $now;
        }
        
        $this->updatedAt = $now;
    }
}

/**
 * Trait for validation functionality
 */
trait ValidatableTrait
{
    /**
     * @var array<string, string> Validation errors
     */
    private array $validationErrors = [];

    /**
     * Add validation error
     * 
     * @param string $field Field name
     * @param string $message Error message
     */
    protected function addValidationError(string $field, string $message): void
    {
        $this->validationErrors[$field] = $message;
    }

    /**
     * Get validation errors
     * 
     * @return array<string, string>
     */
    public function getValidationErrors(): array
    {
        return $this->validationErrors;
    }

    /**
     * Clear validation errors
     */
    protected function clearValidationErrors(): void
    {
        $this->validationErrors = [];
    }

    /**
     * Check if entity has validation errors
     * 
     * @return bool
     */
    public function hasValidationErrors(): bool
    {
        return !empty($this->validationErrors);
    }
}

/**
 * Product entity
 */
class Product extends BaseEntity implements CacheableInterface
{
    use TimestampableTrait;
    use ValidatableTrait;

    /**
     * @var string Product name
     */
    private string $name;
    
    /**
     * @var float Product price
     */
    private float $price;
    
    /**
     * @var string Product description
     */
    private string $description;
    
    /**
     * @var bool Product availability
     */
    private bool $isAvailable = true;
    
    /**
     * @var list<string> Product categories
     */
    private array $categories = [];

    /**
     * Static constant for cache TTL
     */
    public const CACHE_TTL = 3600;

    /**
     * Create new product
     * 
     * @param string $name Product name
     * @param float $price Product price
     * @param string $description Product description
     */
    public function __construct(string $name, float $price, string $description = '')
    {
        $this->name = $name;
        $this->price = $price;
        $this->description = $description;
        $this->updateTimestamps();
    }

    /**
     * Get product name
     * 
     * @return string
     */
    public function getName(): string
    {
        return $this->name;
    }

    /**
     * Set product name
     * 
     * @param string $name Product name
     */
    public function setName(string $name): void
    {
        $this->name = $name;
        $this->updateTimestamps();
    }

    /**
     * Get product price
     * 
     * @return float
     */
    public function getPrice(): float
    {
        return $this->price;
    }

    /**
     * Set product price
     * 
     * @param float $price Product price
     */
    public function setPrice(float $price): void
    {
        $this->price = $price;
        $this->updateTimestamps();
    }

    /**
     * @inheritDoc
     */
    public function getEntityName(): string
    {
        return 'product';
    }

    /**
     * @inheritDoc
     */
    protected function validate(): bool
    {
        $this->clearValidationErrors();
        
        if (empty($this->name)) {
            $this->addValidationError('name', 'Product name is required');
        }
        
        if ($this->price < 0) {
            $this->addValidationError('price', 'Product price must be positive');
        }
        
        return !$this->hasValidationErrors();
    }

    /**
     * @inheritDoc
     */
    public function getCacheKey(): string
    {
        return sprintf('product:%d', $this->id ?? 0);
    }

    /**
     * @inheritDoc
     */
    public function getCacheTtl(): int
    {
        return self::CACHE_TTL;
    }

    /**
     * Add category to product
     * 
     * @param string $category Category name
     */
    public function addCategory(string $category): void
    {
        if (!in_array($category, $this->categories)) {
            $this->categories[] = $category;
        }
    }

    /**
     * Static factory method for creating sale products
     * 
     * @param string $name Product name
     * @param float $originalPrice Original price
     * @param float $discountPercent Discount percentage
     * @return self
     */
    public static function createSaleProduct(string $name, float $originalPrice, float $discountPercent): self
    {
        $salePrice = $originalPrice * (1 - $discountPercent / 100);
        $description = sprintf("Sale item - %d%% off!", $discountPercent);
        
        return new self($name, $salePrice, $description);
    }
}

/**
 * Category entity
 */
class Category extends BaseEntity
{
    /**
     * @var string Category name
     */
    public readonly string $name;
    
    /**
     * @var string Category slug
     */
    public readonly string $slug;
    
    /**
     * @var Category|null Parent category
     */
    private ?Category $parent = null;

    /**
     * Create new category
     * 
     * @param string $name Category name
     * @param string $slug Category slug
     */
    public function __construct(string $name, string $slug)
    {
        $this->name = $name;
        $this->slug = $slug;
    }

    /**
     * @inheritDoc
     */
    public function getEntityName(): string
    {
        return 'category';
    }

    /**
     * @inheritDoc
     */
    protected function validate(): bool
    {
        return !empty($this->name) && !empty($this->slug);
    }

    /**
     * Set parent category
     * 
     * @param Category|null $parent Parent category
     */
    public function setParent(?Category $parent): void
    {
        $this->parent = $parent;
    }

    /**
     * Get parent category
     * 
     * @return Category|null
     */
    public function getParent(): ?Category
    {
        return $this->parent;
    }
}