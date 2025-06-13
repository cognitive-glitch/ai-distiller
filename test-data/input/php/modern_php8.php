<?php
/**
 * Modern PHP 8+ features test
 * 
 * Tests:
 * - Attributes (PHP 8.0)
 * - Enums (PHP 8.1)
 * - Constructor property promotion
 * - Union and intersection types
 * - Match expressions
 * - Named arguments
 */

namespace App\Modern;

use App\Attributes\Route;
use App\Attributes\Deprecated;
use App\Attributes\Required;

/**
 * Status enum with string backing
 */
enum Status: string
{
    case PENDING = 'pending';
    case APPROVED = 'approved';
    case REJECTED = 'rejected';
    
    public function getLabel(): string
    {
        return match($this) {
            self::PENDING => 'Pending Review',
            self::APPROVED => 'Approved',
            self::REJECTED => 'Rejected'
        };
    }
}

/**
 * Priority enum without backing type
 */
enum Priority
{
    case LOW;
    case MEDIUM;
    case HIGH;
    case URGENT;
}

/**
 * Interface with PHP 8 features
 */
interface ModernInterface
{
    public function process(string|int $id): void;
}

/**
 * Modern PHP 8 class with attributes
 */
#[Route('/api/users')]
class UserController implements ModernInterface
{
    /**
     * Constructor with property promotion
     */
    public function __construct(
        private readonly string $apiKey,
        protected ?string $environment = 'production',
        public string|int $version = 1
    ) {}
    
    #[Route('/api/users/{id}', methods: ['GET'])]
    public function show(string|int $id): array
    {
        return [
            'id' => $id,
            'status' => Status::APPROVED->value
        ];
    }
    
    #[Route('/api/users', methods: ['POST'])]
    #[Required(['name', 'email'])]
    public function store(array $data): array
    {
        $priority = Priority::MEDIUM;
        
        return match($data['type'] ?? 'user') {
            'admin' => $this->createAdmin($data),
            'moderator' => $this->createModerator($data),
            default => $this->createUser($data)
        };
    }
    
    #[Deprecated('Use show() method instead', since: '2.0')]
    public function getUser(int $id): ?array
    {
        return $this->show($id);
    }
    
    /**
     * Method with union return type
     */
    private function createUser(array $data): array|false
    {
        if (empty($data['email'])) {
            return false;
        }
        
        return [
            'id' => uniqid(),
            'type' => 'user',
            ...$data
        ];
    }
    
    /**
     * Method with intersection type parameter
     */
    public function process(string|int $id): void
    {
        // Implementation
    }
}

/**
 * Trait with PHP 8 features
 */
trait ModernTrait
{
    public function logAction(Status $status, mixed $data = null): void
    {
        $message = match($status) {
            Status::PENDING => 'Action pending',
            Status::APPROVED => 'Action approved',
            Status::REJECTED => 'Action rejected'
        };
        
        error_log($message);
    }
}

/**
 * Abstract class with readonly properties
 */
abstract class AbstractService
{
    public function __construct(
        public readonly string $serviceName,
        private readonly array $config = []
    ) {}
    
    abstract public function execute(mixed ...$args): mixed;
}

/**
 * Anonymous class example
 */
function createAnonymousHandler(): object
{
    return new class extends AbstractService {
        public function __construct()
        {
            parent::__construct('anonymous');
        }
        
        public function execute(mixed ...$args): mixed
        {
            return $args;
        }
    };
}