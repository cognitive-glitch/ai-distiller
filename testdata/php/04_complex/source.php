<?php

declare(strict_types=1);

namespace App\Complex;

use Attribute;
use ReflectionClass;
use ReflectionMethod;
use ReflectionProperty;
use ReflectionAttribute;
use Generator;
use Closure;
use WeakMap;
use SplObjectStorage;

/**
 * Custom attribute for route definition
 */
#[Attribute(Attribute::TARGET_METHOD)]
class Route
{
    /**
     * Create route attribute
     *
     * @param string $path Route path
     * @param string $method HTTP method
     * @param list<string> $middleware Route middleware
     */
    public function __construct(
        public readonly string $path,
        public readonly string $method = 'GET',
        public readonly array $middleware = []
    ) {}
}

/**
 * Custom attribute for validation
 */
#[Attribute(Attribute::TARGET_PROPERTY | Attribute::IS_REPEATABLE)]
class Validate
{
    /**
     * Create validation attribute
     *
     * @param string $rule Validation rule
     * @param string $message Error message
     */
    public function __construct(
        public readonly string $rule,
        public readonly string $message = ''
    ) {}
}

/**
 * Custom attribute for dependency injection
 */
#[Attribute(Attribute::TARGET_PARAMETER)]
class Inject
{
    /**
     * Create injection attribute
     *
     * @param string|null $service Service identifier
     */
    public function __construct(
        public readonly ?string $service = null
    ) {}
}

/**
 * Advanced reflection-based attribute processor
 */
class AttributeProcessor
{
    /**
     * @var WeakMap Cache for processed classes
     */
    private WeakMap $classCache;

    public function __construct()
    {
        $this->classCache = new WeakMap();
    }

    /**
     * Process attributes for a class
     *
     * @param object $instance Object instance
     * @return array{class: list<array{name: string, arguments: array<mixed>, instance: object}>, methods: array<string, list<array{name: string, arguments: array<mixed>, instance: object}>>, properties: array<string, list<array{name: string, arguments: array<mixed>, instance: object}>>} Processed attributes
     */
    public function processAttributes(object $instance): array
    {
        $class = $instance::class;
        $reflection = new ReflectionClass($class);

        if (isset($this->classCache[$instance])) {
            return $this->classCache[$instance];
        }

        $result = [
            'class' => $this->processClassAttributes($reflection),
            'methods' => $this->processMethodAttributes($reflection),
            'properties' => $this->processPropertyAttributes($reflection),
        ];

        $this->classCache[$instance] = $result;
        return $result;
    }

    /**
     * Process class-level attributes
     *
     * @param ReflectionClass $reflection Class reflection
     * @return array
     */
    private function processClassAttributes(ReflectionClass $reflection): array
    {
        $attributes = [];

        foreach ($reflection->getAttributes() as $attribute) {
            $attributes[] = [
                'name' => $attribute->getName(),
                'arguments' => $attribute->getArguments(),
                'instance' => $attribute->newInstance(),
            ];
        }

        return $attributes;
    }

    /**
     * Process method attributes
     *
     * @param ReflectionClass $reflection Class reflection
     * @return array
     */
    private function processMethodAttributes(ReflectionClass $reflection): array
    {
        $methods = [];

        foreach ($reflection->getMethods() as $method) {
            $attributes = [];

            foreach ($method->getAttributes() as $attribute) {
                $attributes[] = [
                    'name' => $attribute->getName(),
                    'arguments' => $attribute->getArguments(),
                    'instance' => $attribute->newInstance(),
                ];
            }

            if (!empty($attributes)) {
                $methods[$method->getName()] = $attributes;
            }
        }

        return $methods;
    }

    /**
     * Process property attributes
     *
     * @param ReflectionClass $reflection Class reflection
     * @return array
     */
    private function processPropertyAttributes(ReflectionClass $reflection): array
    {
        $properties = [];

        foreach ($reflection->getProperties() as $property) {
            $attributes = [];

            foreach ($property->getAttributes() as $attribute) {
                $attributes[] = [
                    'name' => $attribute->getName(),
                    'arguments' => $attribute->getArguments(),
                    'instance' => $attribute->newInstance(),
                ];
            }

            if (!empty($attributes)) {
                $properties[$property->getName()] = $attributes;
            }
        }

        return $properties;
    }

    /**
     * Find methods with specific attribute
     *
     * @param ReflectionClass $reflection Class reflection
     * @param string $attributeClass Attribute class name
     * @return Generator<ReflectionMethod, ReflectionAttribute>
     */
    public function findMethodsWithAttribute(ReflectionClass $reflection, string $attributeClass): Generator
    {
        foreach ($reflection->getMethods() as $method) {
            $attributes = $method->getAttributes($attributeClass);
            foreach ($attributes as $attribute) {
                yield $method => $attribute;
            }
        }
    }
}

/**
 * Advanced route controller with attributes
 */
class ApiController
{
    /**
     * @var AttributeProcessor Attribute processor
     */
    private AttributeProcessor $attributeProcessor;

    public function __construct()
    {
        $this->attributeProcessor = new AttributeProcessor();
    }

    /**
     * Get user profile
     */
    #[Route('/api/users/{id}', 'GET', ['auth', 'throttle'])]
    public function getUserProfile(int $id): array
    {
        return ['id' => $id, 'name' => 'John Doe'];
    }

    /**
     * Update user profile
     *
     * @param int $id User ID
     * @param array<string, mixed> $data User data
     * @return array{id: int, updated: bool, data: array<string, mixed>}
     */
    #[Route('/api/users/{id}', 'PUT', ['auth', 'validate'])]
    public function updateUserProfile(int $id, array $data): array
    {
        return ['id' => $id, 'updated' => true, 'data' => $data];
    }

    /**
     * Delete user
     */
    #[Route('/api/users/{id}', 'DELETE', ['auth', 'admin'])]
    public function deleteUser(int $id): array
    {
        return ['id' => $id, 'deleted' => true];
    }

    /**
     * Get all routes defined in this controller
     *
     * @return array
     */
    public function getRoutes(): array
    {
        $reflection = new ReflectionClass($this);
        $routes = [];

        foreach ($this->attributeProcessor->findMethodsWithAttribute($reflection, Route::class) as $method => $attribute) {
            $route = $attribute->newInstance();
            $routes[] = [
                'method' => $method->getName(),
                'path' => $route->path,
                'http_method' => $route->method,
                'middleware' => $route->middleware,
            ];
        }

        return $routes;
    }
}

/**
 * Data Transfer Object with validation attributes
 */
class UserCreateDto
{
    /**
     * @var string User name
     */
    #[Validate('required', 'Name is required')]
    #[Validate('min:2', 'Name must be at least 2 characters')]
    public string $name;

    /**
     * @var string User email
     */
    #[Validate('required', 'Email is required')]
    #[Validate('email', 'Invalid email format')]
    public string $email;

    /**
     * @var int User age
     */
    #[Validate('required', 'Age is required')]
    #[Validate('min:18', 'Must be at least 18 years old')]
    #[Validate('max:120', 'Age cannot exceed 120')]
    public int $age;

    /**
     * @var array<string, mixed> User preferences
     */
    public array $preferences = [];

    /**
     * Create DTO from array
     *
     * @param array{name?: string, email?: string, age?: int, preferences?: array<string, mixed>} $data Input data
     * @return self
     */
    public static function fromArray(array $data): self
    {
        $dto = new self();

        foreach ($data as $key => $value) {
            if (property_exists($dto, $key)) {
                $dto->$key = $value;
            }
        }

        return $dto;
    }

    /**
     * Validate the DTO using reflection and attributes
     *
     * @return array Validation errors
     */
    public function validate(): array
    {
        $errors = [];
        $reflection = new ReflectionClass($this);

        foreach ($reflection->getProperties() as $property) {
            $propertyName = $property->getName();
            $value = $property->getValue($this);

            foreach ($property->getAttributes(Validate::class) as $attribute) {
                $validator = $attribute->newInstance();
                $error = $this->validateRule($value, $validator->rule, $validator->message);

                if ($error) {
                    $errors[$propertyName][] = $error;
                }
            }
        }

        return $errors;
    }

    /**
     * Validate a single rule
     *
     * @param mixed $value Value to validate
     * @param string $rule Validation rule
     * @param string $message Error message
     * @return string|null Error message or null if valid
     */
    private function validateRule(mixed $value, string $rule, string $message): ?string
    {
        switch (true) {
            case $rule === 'required':
                return empty($value) ? ($message ?: 'Field is required') : null;

            case str_starts_with($rule, 'min:'):
                $min = (int) substr($rule, 4);
                if (is_string($value) && strlen($value) < $min) {
                    return $message ?: "Must be at least {$min} characters";
                }
                if (is_numeric($value) && $value < $min) {
                    return $message ?: "Must be at least {$min}";
                }
                return null;

            case str_starts_with($rule, 'max:'):
                $max = (int) substr($rule, 4);
                if (is_string($value) && strlen($value) > $max) {
                    return $message ?: "Must be at most {$max} characters";
                }
                if (is_numeric($value) && $value > $max) {
                    return $message ?: "Must be at most {$max}";
                }
                return null;

            case $rule === 'email':
                return filter_var($value, FILTER_VALIDATE_EMAIL) ? null : ($message ?: 'Invalid email format');

            default:
                return null;
        }
    }
}

/**
 * Advanced service locator with lazy loading
 */
class ServiceLocator
{
    /**
     * @var array<string, callable> Service factories
     */
    private array $factories = [];

    /**
     * @var SplObjectStorage Singleton instances
     */
    private SplObjectStorage $singletons;

    /**
     * @var array<string, array{singleton: bool}> Service metadata
     */
    private array $metadata = [];

    public function __construct()
    {
        $this->singletons = new SplObjectStorage();
    }

    /**
     * Register a service factory
     *
     * @param string $id Service identifier
     * @param Closure $factory Service factory
     * @param bool $singleton Whether to create as singleton
     * @param array<string, mixed> $metadata Service metadata
     */
    public function register(string $id, Closure $factory, bool $singleton = false, array $metadata = []): void
    {
        $this->factories[$id] = $factory;
        $this->metadata[$id] = $metadata + ['singleton' => $singleton];
    }

    /**
     * Resolve a service
     *
     * @param string $id Service identifier
     * @return mixed
     */
    public function resolve(string $id): mixed
    {
        if (!isset($this->factories[$id])) {
            throw new \InvalidArgumentException("Service '{$id}' not registered");
        }

        $metadata = $this->metadata[$id];

        if ($metadata['singleton']) {
            // Check if singleton already exists
            foreach ($this->singletons as $service) {
                if ($this->singletons[$service] === $id) {
                    return $service;
                }
            }

            // Create new singleton
            $instance = $this->factories[$id]($this);
            $this->singletons[$instance] = $id;
            return $instance;
        }

        return $this->factories[$id]($this);
    }

    /**
     * Create service with dependency injection using attributes
     *
     * @param string $className Class name
     * @return object
     */
    public function createWithDependencies(string $className): object
    {
        $reflection = new ReflectionClass($className);
        $constructor = $reflection->getConstructor();

        if (!$constructor) {
            return new $className();
        }

        $dependencies = [];

        foreach ($constructor->getParameters() as $parameter) {
            $injectAttributes = $parameter->getAttributes(Inject::class);

            if (!empty($injectAttributes)) {
                $inject = $injectAttributes[0]->newInstance();
                $serviceId = $inject->service ?? $parameter->getType()?->getName();

                if ($serviceId && $this->has($serviceId)) {
                    $dependencies[] = $this->resolve($serviceId);
                } else {
                    $dependencies[] = $parameter->getDefaultValue();
                }
            } else {
                $type = $parameter->getType();
                if ($type && !$type->isBuiltin() && $this->has($type->getName())) {
                    $dependencies[] = $this->resolve($type->getName());
                } else {
                    $dependencies[] = $parameter->getDefaultValue();
                }
            }
        }

        return $reflection->newInstanceArgs($dependencies);
    }

    /**
     * Check if service is registered
     *
     * @param string $id Service identifier
     * @return bool
     */
    public function has(string $id): bool
    {
        return isset($this->factories[$id]);
    }

    /**
     * Get service metadata
     *
     * @param string $id Service identifier
     * @return array
     */
    public function getMetadata(string $id): array
    {
        return $this->metadata[$id] ?? [];
    }

    /**
     * Get all registered service IDs
     *
     * @return array
     */
    public function getServiceIds(): array
    {
        return array_keys($this->factories);
    }
}

/**
 * Complex service with multiple dependencies
 */
class ComplexUserService
{
    /**
     * Create service with dependency injection
     *
     * @param UserRepository $repository User repository
     * @param EventDispatcher $dispatcher Event dispatcher
     * @param AttributeProcessor $processor Attribute processor
     */
    public function __construct(
        #[Inject] private UserRepository $repository,
        #[Inject] private EventDispatcher $dispatcher,
        #[Inject('attribute_processor')] private AttributeProcessor $processor
    ) {}

    /**
     * Create user with validation
     *
     * @param UserCreateDto $dto User data
     * @return array
     */
    public function createUser(UserCreateDto $dto): array
    {
        $errors = $dto->validate();

        if (!empty($errors)) {
            return ['success' => false, 'errors' => $errors];
        }

        // Create user logic here
        return ['success' => true, 'user_id' => 123];
    }

    /**
     * Process user data with generators
     *
     * @param array $users User data
     * @return Generator<array>
     */
    public function processUsers(array $users): Generator
    {
        foreach ($users as $userData) {
            $dto = UserCreateDto::fromArray($userData);
            $result = $this->createUser($dto);

            yield [
                'original' => $userData,
                'processed' => $result,
                'timestamp' => time(),
            ];
        }
    }
}