<?php

namespace App\Services\Notification;

use App\Utils\Timestampable;
use Psr\Log\LoggerInterface; // A common external interface

class EmailPayload
{
    // Using a trait
    use Timestampable;

    public function __construct(
        public readonly string $recipient,
        public readonly string $subject,
        public readonly string $body
    ) {
        $this->touch(); // Method from the Timestampable trait
    }
}

class Notifier
{
    public function __construct(private LoggerInterface $logger) {}

    // Union Type hint
    public function send(EmailPayload|string $payload): void
    {
        if (is_string($payload)) {
            $this->logger->info("Raw string notification: {$payload}");
            return;
        }

        $this->logger->info(
            "Sending email to {$payload->recipient} with subject '{$payload->subject}'"
        );
        // Actual sending logic would be here
    }
}

// --- In another file: App/Utils/Timestampable.php ---
namespace App\Utils;

trait Timestampable
{
    private ?\DateTimeImmutable $createdAt = null;

    public function touch(): void
    {
        if ($this->createdAt === null) {
            $this->createdAt = new \DateTimeImmutable();
        }
    }

    public function getCreationDate(): ?\DateTimeImmutable
    {
        return $this->createdAt;
    }
}