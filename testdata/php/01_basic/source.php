<?php

declare(strict_types=1);

namespace App\Basic;

use DateTime;
use InvalidArgumentException;

/**
 * Basic user model demonstrating fundamental PHP features
 */
class User
{
    /**
     * @var int User ID
     */
    public int $id;
    
    /**
     * @var string User name
     */
    public string $name;
    
    /**
     * @var string Email address
     */
    private string $email;
    
    /**
     * @var DateTime|null Creation timestamp
     */
    protected ?DateTime $createdAt;
    
    /**
     * @var array User preferences
     */
    private array $preferences = [];

    /**
     * Create a new user instance
     * 
     * @param int $id User ID
     * @param string $name User name
     * @param string $email Email address
     */
    public function __construct(int $id, string $name, string $email)
    {
        $this->id = $id;
        $this->name = $name;
        $this->setEmail($email);
        $this->createdAt = new DateTime();
    }

    /**
     * Get user email
     * 
     * @return string
     */
    public function getEmail(): string
    {
        return $this->email;
    }

    /**
     * Set user email with validation
     * 
     * @param string $email Email address
     * @throws InvalidArgumentException
     */
    public function setEmail(string $email): void
    {
        if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid email format");
        }
        $this->email = $email;
    }

    /**
     * Get user preference
     * 
     * @param string $key Preference key
     * @param mixed $default Default value
     * @return mixed
     */
    public function getPreference(string $key, mixed $default = null): mixed
    {
        return $this->preferences[$key] ?? $default;
    }

    /**
     * Set user preference
     * 
     * @param string $key Preference key
     * @param mixed $value Preference value
     */
    public function setPreference(string $key, mixed $value): void
    {
        $this->preferences[$key] = $value;
    }

    /**
     * Check if user is active (created within last 30 days)
     * 
     * @return bool
     */
    public function isActive(): bool
    {
        if ($this->createdAt === null) {
            return false;
        }
        
        $thirtyDaysAgo = new DateTime('-30 days');
        return $this->createdAt > $thirtyDaysAgo;
    }

    /**
     * Get formatted user info
     * 
     * @return string
     */
    protected function formatUserInfo(): string
    {
        return sprintf("User #%d: %s <%s>", $this->id, $this->name, $this->email);
    }

    /**
     * Convert to array representation
     * 
     * @return array
     */
    public function toArray(): array
    {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'email' => $this->email,
            'created_at' => $this->createdAt?->format('Y-m-d H:i:s'),
            'is_active' => $this->isActive(),
        ];
    }
}

/**
 * Simple user manager
 */
class UserManager
{
    /**
     * @var array<int, User> Users storage
     */
    private array $users = [];

    /**
     * Add user to manager
     * 
     * @param User $user User instance
     */
    public function addUser(User $user): void
    {
        $this->users[$user->id] = $user;
    }

    /**
     * Find user by ID
     * 
     * @param int $id User ID
     * @return User|null
     */
    public function findUser(int $id): ?User
    {
        return $this->users[$id] ?? null;
    }

    /**
     * Get all active users
     * 
     * @return array<User>
     */
    public function getActiveUsers(): array
    {
        return array_filter($this->users, fn(User $user) => $user->isActive());
    }

    /**
     * Get user count
     * 
     * @return int
     */
    public function getUserCount(): int
    {
        return count($this->users);
    }
}

/**
 * Utility functions for user operations
 */
function validateEmail(string $email): bool
{
    return filter_var($email, FILTER_VALIDATE_EMAIL) !== false;
}

/**
 * Create user from array data
 * 
 * @param array $data User data
 * @return User
 * @throws InvalidArgumentException
 */
function createUserFromArray(array $data): User
{
    if (!isset($data['id'], $data['name'], $data['email'])) {
        throw new InvalidArgumentException("Missing required user data");
    }
    
    return new User($data['id'], $data['name'], $data['email']);
}

// Constants for user status
const USER_STATUS_ACTIVE = 'active';
const USER_STATUS_INACTIVE = 'inactive';
const MAX_USERS_PER_PAGE = 25;