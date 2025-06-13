<?php

declare(strict_types=1);

// Using PHP 8 Constructor Property Promotion
class User
{
    public function __construct(
        private readonly int $id,
        private string $name,
        private string $email
    ) {}

    public function getId(): int
    {
        return $this->id;
    }

    public function getDisplayName(): string
    {
        return "User: " . $this->name;
    }

    public function changeEmail(string $newEmail): void
    {
        // Basic validation logic
        if (!filter_var($newEmail, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid email format provided.");
        }
        $this->email = $newEmail;
    }
}

$user = new User(101, 'Alice', 'alice@example.com');
$user->changeEmail('alice.new@example.com');
echo $user->getDisplayName();