<?php
/**
 * Complex namespace and trait usage test
 */

namespace App\Core\Services;

use App\Contracts\ServiceInterface;
use App\Traits\{Loggable, Cacheable};
use App\Events\UserCreated as UserCreatedEvent;
use function App\Helpers\sanitize_input;
use const App\Config\DEFAULT_TIMEOUT;

/**
 * Base service trait
 */
trait ServiceTrait
{
    protected string $serviceName = 'BaseService';
    
    public function getServiceName(): string
    {
        return $this->serviceName;
    }
    
    protected function logServiceAction(string $action): void
    {
        // Log the action
    }
}

/**
 * Advanced trait with conflict resolution
 */
trait AdvancedTrait
{
    use ServiceTrait {
        ServiceTrait::getServiceName as getBaseServiceName;
        ServiceTrait::logServiceAction as protected;
    }
    
    use Loggable {
        Loggable::log insteadof ServiceTrait;
        Loggable::log as writeLog;
    }
    
    public function getServiceName(): string
    {
        return 'Advanced: ' . $this->getBaseServiceName();
    }
}

/**
 * Service implementation
 */
class UserService implements ServiceInterface
{
    use AdvancedTrait, Cacheable {
        Cacheable::get as getCached;
        Cacheable::set as setCached;
    }
    
    private const CACHE_PREFIX = 'user_';
    
    public function createUser(array $data): User
    {
        $sanitized = sanitize_input($data);
        
        $user = new User($sanitized);
        $this->setCached(self::CACHE_PREFIX . $user->id, $user, DEFAULT_TIMEOUT);
        
        event(new UserCreatedEvent($user));
        
        return $user;
    }
}

/**
 * Nested namespace declaration
 */
namespace App\Core\Services\Validators {
    
    use App\Core\Services\ServiceTrait;
    
    class UserValidator
    {
        use ServiceTrait;
        
        public function validate(array $data): bool
        {
            return !empty($data['email']) && filter_var($data['email'], FILTER_VALIDATE_EMAIL);
        }
    }
}

/**
 * Another namespace in same file
 */
namespace App\Core\Services\Repositories {
    
    use App\Core\Services\User;
    use App\Core\Services\ServiceInterface;
    
    interface RepositoryInterface extends ServiceInterface
    {
        public function find(int $id): ?User;
        public function save(User $user): bool;
    }
    
    class UserRepository implements RepositoryInterface
    {
        public function find(int $id): ?User
        {
            // Implementation
            return null;
        }
        
        public function save(User $user): bool
        {
            // Implementation
            return true;
        }
    }
}