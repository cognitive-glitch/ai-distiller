<?php
/**
 * Basic PHP class for testing the distiller
 * 
 * This file tests core PHP features including:
 * - Classes with visibility modifiers
 * - Methods and properties
 * - Constants
 * - Docblocks
 */

namespace App\Models;

use App\Core\BaseModel;
use App\Traits\Timestampable;
use App\Interfaces\Jsonable as JsonableInterface;

/**
 * User model class
 * 
 * @property string $name User's full name
 * @property string $email User's email address
 */
class User extends BaseModel implements JsonableInterface
{
    use Timestampable;
    
    /** @var string User type constants */
    public const TYPE_ADMIN = 'admin';
    public const TYPE_USER = 'user';
    protected const TYPE_GUEST = 'guest';
    
    /**
     * @var string User's name
     */
    public string $name;
    
    /**
     * @var string User's email (protected)
     */
    protected string $email;
    
    /**
     * @var string User's password (private)
     */
    private string $password;
    
    /**
     * @var array<string, mixed> User attributes
     */
    private array $attributes = [];
    
    /**
     * Create a new user instance
     * 
     * @param string $name User's name
     * @param string $email User's email
     */
    public function __construct(string $name, string $email)
    {
        $this->name = $name;
        $this->email = $email;
    }
    
    /**
     * Get user's full name
     * 
     * @return string
     */
    public function getName(): string
    {
        return $this->name;
    }
    
    /**
     * Set user's name
     * 
     * @param string $name
     * @return void
     */
    public function setName(string $name): void
    {
        $this->name = $name;
    }
    
    /**
     * Get user's email (protected method)
     * 
     * @return string
     */
    protected function getEmail(): string
    {
        return $this->email;
    }
    
    /**
     * Validate password (private method)
     * 
     * @param string $password
     * @return bool
     */
    private function validatePassword(string $password): bool
    {
        return strlen($password) >= 8;
    }
    
    /**
     * Convert to JSON representation
     * 
     * @return string
     */
    public function toJson(): string
    {
        return json_encode([
            'name' => $this->name,
            'email' => $this->email,
            'type' => self::TYPE_USER
        ]);
    }
    
    /**
     * Static factory method
     * 
     * @param array $data
     * @return self
     */
    public static function fromArray(array $data): self
    {
        return new self($data['name'], $data['email']);
    }
    
    /**
     * Abstract method from parent (if any)
     */
    abstract public function validate(): bool;
}

/**
 * Admin user extends regular user
 */
final class AdminUser extends User
{
    private bool $isSuperAdmin = false;
    
    public function validate(): bool
    {
        return !empty($this->name) && filter_var($this->email, FILTER_VALIDATE_EMAIL);
    }
}