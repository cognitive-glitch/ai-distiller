<?php

declare(strict_types=1);

namespace App\Medium;

use Generator;
use Closure;
use ReflectionClass;
use ReflectionException;
use InvalidArgumentException;

/**
 * Interface for dependency injection container
 */
interface ContainerInterface
{
    /**
     * Register a service in the container
     * 
     * @param string $id Service identifier
     * @param callable|object $service Service factory or instance
     */
    public function register(string $id, callable|object $service): void;
    
    /**
     * Resolve a service from the container
     * 
     * @param string $id Service identifier
     * @return mixed
     */
    public function resolve(string $id): mixed;
    
    /**
     * Check if service is registered
     * 
     * @param string $id Service identifier
     * @return bool
     */
    public function has(string $id): bool;
}

/**
 * Simple dependency injection container
 */
class Container implements ContainerInterface
{
    /**
     * @var array<string, callable|object> Service definitions
     */
    private array $services = [];
    
    /**
     * @var array<string, object> Resolved instances
     */
    private array $instances = [];

    /**
     * @inheritDoc
     */
    public function register(string $id, callable|object $service): void
    {
        $this->services[$id] = $service;
        unset($this->instances[$id]); // Clear cached instance
    }

    /**
     * @inheritDoc
     */
    public function resolve(string $id): mixed
    {
        if (isset($this->instances[$id])) {
            return $this->instances[$id];
        }

        if (!isset($this->services[$id])) {
            throw new InvalidArgumentException("Service '{$id}' not found");
        }

        $service = $this->services[$id];
        
        $instance = is_callable($service) ? $service($this) : $service;
        $this->instances[$id] = $instance;
        
        return $instance;
    }

    /**
     * @inheritDoc
     */
    public function has(string $id): bool
    {
        return isset($this->services[$id]);
    }

    /**
     * Auto-wire a class using reflection
     * 
     * @param string $className Class name
     * @return object
     * @throws ReflectionException
     */
    public function autowire(string $className): object
    {
        $reflection = new ReflectionClass($className);
        
        if (!$reflection->isInstantiable()) {
            throw new InvalidArgumentException("Class '{$className}' is not instantiable");
        }

        $constructor = $reflection->getConstructor();
        if (!$constructor) {
            return $reflection->newInstance();
        }

        $dependencies = [];
        foreach ($constructor->getParameters() as $parameter) {
            $type = $parameter->getType();
            if ($type && !$type->isBuiltin()) {
                $dependencies[] = $this->resolve($type->getName());
            } else {
                $dependencies[] = $parameter->getDefaultValue();
            }
        }

        return $reflection->newInstanceArgs($dependencies);
    }
}

/**
 * Generic repository interface
 * 
 * @template T
 */
interface RepositoryInterface
{
    /**
     * Find entity by ID
     * 
     * @param int $id Entity ID
     * @return T|null
     */
    public function find(int $id): ?object;
    
    /**
     * Find all entities
     * 
     * @return list<T>
     */
    public function findAll(): array;
    
    /**
     * Save entity
     * 
     * @param T $entity Entity to save
     * @return T
     */
    public function save(object $entity): object;
    
    /**
     * Delete entity
     * 
     * @param T $entity Entity to delete
     */
    public function delete(object $entity): void;
}

/**
 * Abstract base repository
 * 
 * @template T
 * @implements RepositoryInterface<T>
 */
abstract class AbstractRepository implements RepositoryInterface
{
    /**
     * @var array<int, T> In-memory storage
     */
    protected array $entities = [];
    
    /**
     * @var int Next ID counter
     */
    protected int $nextId = 1;

    /**
     * @inheritDoc
     */
    public function find(int $id): ?object
    {
        return $this->entities[$id] ?? null;
    }

    /**
     * @inheritDoc
     */
    public function findAll(): array
    {
        return array_values($this->entities);
    }

    /**
     * @inheritDoc
     */
    public function save(object $entity): object
    {
        $reflection = new ReflectionClass($entity);
        
        if ($reflection->hasMethod('getId') && $reflection->hasMethod('setId')) {
            $id = $entity->getId();
            if ($id === null) {
                $entity->setId($this->nextId++);
                $id = $entity->getId();
            }
            $this->entities[$id] = $entity;
        }
        
        return $entity;
    }

    /**
     * @inheritDoc
     */
    public function delete(object $entity): void
    {
        $reflection = new ReflectionClass($entity);
        
        if ($reflection->hasMethod('getId')) {
            $id = $entity->getId();
            if ($id !== null) {
                unset($this->entities[$id]);
            }
        }
    }

    /**
     * Find entities by criteria using closures
     * 
     * @param Closure $criteria Criteria closure
     * @return Generator<T>
     */
    public function findBy(Closure $criteria): Generator
    {
        foreach ($this->entities as $entity) {
            if ($criteria($entity)) {
                yield $entity;
            }
        }
    }

    /**
     * Create a lazy-loaded collection
     * 
     * @param Closure $loader Loader closure
     * @return Closure
     */
    protected function createLazyCollection(Closure $loader): Closure
    {
        return function() use ($loader) {
            yield from $loader();
        };
    }
}

/**
 * Event interface
 */
interface EventInterface
{
    /**
     * Get event name
     * 
     * @return string
     */
    public function getName(): string;
    
    /**
     * Get event data
     * 
     * @return array
     */
    public function getData(): array;
}

/**
 * Event listener interface
 */
interface EventListenerInterface
{
    /**
     * Handle the event
     * 
     * @param EventInterface $event Event to handle
     */
    public function handle(EventInterface $event): void;
}

/**
 * Simple event dispatcher
 */
class EventDispatcher
{
    /**
     * @var array<string, list<EventListenerInterface>> Event listeners
     */
    private array $listeners = [];

    /**
     * Subscribe to an event
     * 
     * @param string $eventName Event name
     * @param EventListenerInterface $listener Event listener
     */
    public function subscribe(string $eventName, EventListenerInterface $listener): void
    {
        $this->listeners[$eventName][] = $listener;
    }

    /**
     * Dispatch an event
     * 
     * @param EventInterface $event Event to dispatch
     */
    public function dispatch(EventInterface $event): void
    {
        $eventName = $event->getName();
        
        if (!isset($this->listeners[$eventName])) {
            return;
        }

        foreach ($this->listeners[$eventName] as $listener) {
            $listener->handle($event);
        }
    }

    /**
     * Create anonymous event listener
     * 
     * @param string $eventName Event name
     * @param Closure $handler Event handler
     */
    public function listen(string $eventName, Closure $handler): void
    {
        $this->subscribe($eventName, new class($handler) implements EventListenerInterface {
            public function __construct(private Closure $handler) {}
            
            public function handle(EventInterface $event): void
            {
                ($this->handler)($event);
            }
        });
    }
}

/**
 * Simple event implementation
 */
class Event implements EventInterface
{
    /**
     * Create new event
     * 
     * @param string $name Event name
     * @param array<string, mixed> $data Event data
     */
    public function __construct(
        private string $name,
        private array $data = []
    ) {}

    /**
     * @inheritDoc
     */
    public function getName(): string
    {
        return $this->name;
    }

    /**
     * @inheritDoc
     */
    public function getData(): array
    {
        return $this->data;
    }
}

/**
 * User entity with advanced features
 */
class User
{
    /**
     * Create new user
     * 
     * @param int|null $id User ID
     * @param string $name User name
     * @param string $email User email
     */
    public function __construct(
        private ?int $id,
        private string $name,
        private string $email
    ) {}

    public function getId(): ?int { return $this->id; }
    public function setId(int $id): void { $this->id = $id; }
    public function getName(): string { return $this->name; }
    public function getEmail(): string { return $this->email; }
}

/**
 * User repository
 * 
 * @extends AbstractRepository<User>
 */
class UserRepository extends AbstractRepository
{
    /**
     * Find users by email domain
     * 
     * @param string $domain Email domain
     * @return Generator<User>
     */
    public function findByEmailDomain(string $domain): Generator
    {
        return $this->findBy(fn(User $user) => str_ends_with($user->getEmail(), "@{$domain}"));
    }

    /**
     * Get user statistics
     * 
     * @return array
     */
    public function getStatistics(): array
    {
        $total = count($this->entities);
        $domains = [];
        
        foreach ($this->entities as $user) {
            $domain = substr($user->getEmail(), strpos($user->getEmail(), '@') + 1);
            $domains[$domain] = ($domains[$domain] ?? 0) + 1;
        }
        
        return [
            'total' => $total,
            'domains' => $domains,
        ];
    }
}

/**
 * Service with dependency injection
 */
class UserService
{
    /**
     * Create user service
     * 
     * @param UserRepository $repository User repository
     * @param EventDispatcher $dispatcher Event dispatcher
     */
    public function __construct(
        private UserRepository $repository,
        private EventDispatcher $dispatcher
    ) {}

    /**
     * Create new user
     * 
     * @param string $name User name
     * @param string $email User email
     * @return User
     */
    public function createUser(string $name, string $email): User
    {
        $user = new User(null, $name, $email);
        $savedUser = $this->repository->save($user);
        
        $this->dispatcher->dispatch(new Event('user.created', [
            'user_id' => $savedUser->getId(),
            'name' => $name,
            'email' => $email,
        ]));
        
        return $savedUser;
    }

    /**
     * Get user statistics using generator
     * 
     * @return Generator<string, mixed>
     */
    public function getUserStatistics(): Generator
    {
        yield 'total' => count($this->repository->findAll());
        yield 'by_domain' => $this->repository->getStatistics()['domains'];
        yield 'recent' => iterator_to_array($this->repository->findBy(
            fn(User $user) => $user->getId() > 0
        ));
    }
}