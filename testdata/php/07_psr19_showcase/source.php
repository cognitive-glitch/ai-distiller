<?php

declare(strict_types=1);

namespace App\PSR19Showcase;

use InvalidArgumentException;
use RuntimeException;
use DateTimeInterface;

/**
 * User entity with comprehensive PSR-19 annotations
 * 
 * @author John Doe <john@example.com>
 * @copyright 2024 Example Corp
 * @version 2.1.0
 * @since 1.0.0
 * 
 * @package App\Entities
 * @api
 * 
 * @property-read int $id User ID (auto-generated)
 * @property string $name User's full name
 * @property string $email User's email address
 * @property-write array<string, mixed> $preferences User preferences (write-only for security)
 * @property-read DateTimeInterface $createdAt Account creation date
 * @property-read DateTimeInterface $lastLogin Last login timestamp
 * 
 * @method static User|null find(int $id) Find user by ID
 * @method static list<User> findByEmail(string $email) Find users by email
 * @method bool hasRole(string $role) Check if user has specific role
 * @method void sendNotification(string $message, array $options = []) Send notification to user
 * 
 * @see https://example.com/docs/user-entity
 * @link https://github.com/example/app/blob/main/docs/USER.md Documentation
 * 
 * @todo Implement two-factor authentication support
 * @todo Add user avatar support
 */
class User
{
    /**
     * @var int User identifier
     * @internal Do not set directly, use factory methods
     */
    private int $id;
    
    /**
     * @var string User's display name
     * @api
     */
    protected string $displayName;
    
    /**
     * @var array<string, mixed> Internal data storage
     * @internal
     */
    private array $data = [];
    
    /**
     * Create new user instance
     * 
     * @param string $email User's email address
     * @param string $name User's full name
     * 
     * @throws InvalidArgumentException When email format is invalid
     * @throws RuntimeException When user creation fails
     * 
     * @see User::validateEmail() For email validation rules
     * @since 1.0.0
     * @api
     */
    public function __construct(string $email, string $name)
    {
        $this->validateEmail($email);
        $this->data['email'] = $email;
        $this->data['name'] = $name;
        $this->displayName = $name;
    }
    
    /**
     * Get user's email address
     * 
     * @return string The user's email
     * 
     * @api
     * @since 1.0.0
     */
    public function getEmail(): string
    {
        return $this->data['email'];
    }
    
    /**
     * Update user's email address
     * 
     * @param string $email New email address
     * @return void
     * 
     * @throws InvalidArgumentException When email format is invalid
     * @throws RuntimeException When email is already taken
     * 
     * @uses EmailValidator::validate() For email validation
     * @usedby UserController::updateEmail()
     * 
     * @api
     * @since 1.2.0
     */
    public function setEmail(string $email): void
    {
        $this->validateEmail($email);
        $this->checkEmailUniqueness($email);
        $this->data['email'] = $email;
    }
    
    /**
     * Magic getter for virtual properties
     * 
     * @param string $name Property name
     * @return mixed Property value
     * 
     * @throws RuntimeException When property doesn't exist
     * 
     * @internal
     */
    public function __get(string $name): mixed
    {
        return match($name) {
            'id' => $this->id,
            'name' => $this->data['name'],
            'email' => $this->data['email'],
            'createdAt' => $this->data['created_at'] ?? null,
            'lastLogin' => $this->data['last_login'] ?? null,
            default => throw new RuntimeException("Property {$name} does not exist")
        };
    }
    
    /**
     * Magic setter for virtual properties
     * 
     * @param string $name Property name
     * @param mixed $value Property value
     * @return void
     * 
     * @throws RuntimeException When property is read-only
     * 
     * @internal
     */
    public function __set(string $name, mixed $value): void
    {
        match($name) {
            'name' => $this->data['name'] = $value,
            'email' => $this->setEmail($value),
            'preferences' => $this->data['preferences'] = $value,
            default => throw new RuntimeException("Property {$name} is read-only or does not exist")
        };
    }
    
    /**
     * Validate email format
     * 
     * @param string $email Email to validate
     * @return void
     * 
     * @throws InvalidArgumentException When email format is invalid
     * 
     * @internal
     * @since 1.0.0
     */
    private function validateEmail(string $email): void
    {
        if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid email format");
        }
    }
    
    /**
     * Check if email is unique in the system
     * 
     * @param string $email Email to check
     * @return void
     * 
     * @throws RuntimeException When email is already taken
     * 
     * @internal
     * @since 1.2.0
     * @todo Implement actual database check
     */
    private function checkEmailUniqueness(string $email): void
    {
        // Placeholder for uniqueness check
        if ($email === 'taken@example.com') {
            throw new RuntimeException("Email already taken");
        }
    }
}

/**
 * User repository for data persistence
 * 
 * @author Jane Smith <jane@example.com>
 * @version 1.5.0
 * @since 1.0.0
 * 
 * @method User|null findOneBy(array $criteria) Find single user by criteria
 * @method list<User> findBy(array $criteria, array $orderBy = []) Find users by criteria
 * @method int count(array $criteria = []) Count users matching criteria
 * 
 * @deprecated 2.0.0 Use UserService instead
 * @see UserService For the new implementation
 */
class UserRepository
{
    /**
     * @var array<int, User> In-memory user storage
     * @internal
     */
    private array $users = [];
    
    /**
     * Find user by ID
     * 
     * @param int $id User identifier
     * @return User|null Found user or null
     * 
     * @api
     * @since 1.0.0
     * @deprecated 2.0.0 Use UserService::getUser() instead
     */
    public function find(int $id): ?User
    {
        return $this->users[$id] ?? null;
    }
    
    /**
     * Save user to repository
     * 
     * @param User $user User to save
     * @return void
     * 
     * @throws RuntimeException When save operation fails
     * 
     * @api
     * @since 1.0.0
     * @deprecated 2.0.0 Use UserService::saveUser() instead
     */
    public function save(User $user): void
    {
        // Implementation
    }
}

/**
 * Modern user service replacing UserRepository
 * 
 * @author Development Team <dev@example.com>
 * @version 1.0.0
 * @since 2.0.0
 * 
 * @api
 * @generated Partially generated by code generator v3.2
 * 
 * @uses CacheInterface For result caching
 * @uses LoggerInterface For operation logging
 * @usedby UserController
 * @usedby AuthenticationService
 */
class UserService
{
    /**
     * Get user by ID with caching
     * 
     * @param int $id User identifier
     * @return User|null Found user or null
     * 
     * @throws RuntimeException When database connection fails
     * 
     * @api
     * @since 2.0.0
     * @see User::find() Legacy method this replaces
     */
    public function getUser(int $id): ?User
    {
        // Implementation with caching
        return null;
    }
    
    /**
     * Create new user
     * 
     * @param array{email: string, name: string, roles?: list<string>} $data User data
     * @return User Created user instance
     * 
     * @throws InvalidArgumentException When required data is missing
     * @throws RuntimeException When user creation fails
     * 
     * @api
     * @since 2.0.0
     * 
     * @example
     * ```php
     * $user = $userService->createUser([
     *     'email' => 'user@example.com',
     *     'name' => 'John Doe',
     *     'roles' => ['user', 'editor']
     * ]);
     * ```
     */
    public function createUser(array $data): User
    {
        // Implementation
        return new User($data['email'], $data['name']);
    }
}