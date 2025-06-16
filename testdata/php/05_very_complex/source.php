<?php

declare(strict_types=1);

namespace App\VeryComplex;

use Attribute;
use ReflectionClass;
use ReflectionMethod;
use ReflectionProperty;
use ReflectionFunction;
use ReflectionAttribute;
use Generator;
use Closure;
use WeakMap;
use WeakReference;
use SplObjectStorage;
use Fiber;
use FiberError;

/**
 * Advanced meta-programming attribute for dynamic proxy creation
 */
#[Attribute(Attribute::TARGET_CLASS)]
class ProxyTarget
{
    /**
     * Create proxy target attribute
     * 
     * @param list<string> $interceptMethods Methods to intercept
     * @param string $proxyClass Proxy class name
     */
    public function __construct(
        public readonly array $interceptMethods = [],
        public readonly string $proxyClass = ''
    ) {}
}

/**
 * Method interceptor attribute
 */
#[Attribute(Attribute::TARGET_METHOD)]
class Intercept
{
    /**
     * Create interceptor attribute
     * 
     * @param string $before Before method hook
     * @param string $after After method hook
     * @param bool $cache Whether to cache results
     */
    public function __construct(
        public readonly string $before = '',
        public readonly string $after = '',
        public readonly bool $cache = false
    ) {}
}

/**
 * Memoization attribute for expensive operations
 */
#[Attribute(Attribute::TARGET_METHOD)]
class Memoize
{
    /**
     * Create memoization attribute
     * 
     * @param int $ttl Time to live in seconds
     * @param string $keyGenerator Key generator method
     */
    public function __construct(
        public readonly int $ttl = 3600,
        public readonly string $keyGenerator = ''
    ) {}
}

/**
 * Advanced dynamic proxy factory using reflection and code generation
 */
class DynamicProxyFactory
{
    /**
     * @var WeakMap Cache for generated proxy classes
     */
    private WeakMap $proxyCache;
    
    /**
     * @var array<string, WeakReference<callable>> Method interceptors
     */
    private array $interceptors = [];
    
    /**
     * @var SplObjectStorage Memoization cache
     */
    private SplObjectStorage $memoCache;

    public function __construct()
    {
        $this->proxyCache = new WeakMap();
        $this->memoCache = new SplObjectStorage();
    }

    /**
     * Create dynamic proxy for an object
     * 
     * @param object $target Target object
     * @return object Proxy object
     */
    public function createProxy(object $target): object
    {
        $reflection = new ReflectionClass($target);
        $proxyAttributes = $reflection->getAttributes(ProxyTarget::class);
        
        if (empty($proxyAttributes)) {
            throw new \InvalidArgumentException('Target class must have ProxyTarget attribute');
        }

        $proxyConfig = $proxyAttributes[0]->newInstance();
        
        // Check cache first
        if (isset($this->proxyCache[$target])) {
            return $this->proxyCache[$target];
        }

        $proxy = $this->generateProxy($target, $reflection, $proxyConfig);
        $this->proxyCache[$target] = $proxy;
        
        return $proxy;
    }

    /**
     * Generate proxy object dynamically
     * 
     * @param object $target Target object
     * @param ReflectionClass $reflection Target reflection
     * @param ProxyTarget $config Proxy configuration
     * @return object
     */
    private function generateProxy(object $target, ReflectionClass $reflection, ProxyTarget $config): object
    {
        $proxyClass = $this->generateProxyClass($reflection, $config);
        return new $proxyClass($target, $this);
    }

    /**
     * Generate proxy class code dynamically
     * 
     * @param ReflectionClass $reflection Target reflection
     * @param ProxyTarget $config Proxy configuration
     * @return string Generated proxy class name
     */
    private function generateProxyClass(ReflectionClass $reflection, ProxyTarget $config): string
    {
        $targetClassName = $reflection->getName();
        $proxyClassName = $config->proxyClass ?: $targetClassName . 'Proxy' . uniqid();
        
        if (class_exists($proxyClassName)) {
            return $proxyClassName;
        }

        $classCode = $this->buildProxyClassCode($reflection, $proxyClassName, $config);
        eval($classCode);
        
        return $proxyClassName;
    }

    /**
     * Build proxy class code
     * 
     * @param ReflectionClass $reflection Target reflection
     * @param string $proxyClassName Proxy class name
     * @param ProxyTarget $config Proxy configuration
     * @return string Generated class code
     */
    private function buildProxyClassCode(ReflectionClass $reflection, string $proxyClassName, ProxyTarget $config): string
    {
        $targetClassName = $reflection->getName();
        $methods = [];

        foreach ($reflection->getMethods(ReflectionMethod::IS_PUBLIC) as $method) {
            if ($method->isConstructor() || $method->isDestructor() || $method->isStatic()) {
                continue;
            }

            $methodName = $method->getName();
            $interceptMethods = $config->interceptMethods;
            
            if (empty($interceptMethods) || in_array($methodName, $interceptMethods)) {
                $methods[] = $this->generateProxyMethod($method);
            }
        }

        $methodsCode = implode("\n\n", $methods);

        return "
class {$proxyClassName} implements ProxyInterface
{
    private object \$target;
    private DynamicProxyFactory \$factory;

    public function __construct(object \$target, DynamicProxyFactory \$factory)
    {
        \$this->target = \$target;
        \$this->factory = \$factory;
    }

    public function __getTarget(): object
    {
        return \$this->target;
    }

    {$methodsCode}
}";
    }

    /**
     * Generate proxy method code
     * 
     * @param ReflectionMethod $method Method reflection
     * @return string Generated method code
     */
    private function generateProxyMethod(ReflectionMethod $method): string
    {
        $methodName = $method->getName();
        $parameters = $this->buildParameterList($method);
        $parameterNames = $this->buildParameterNames($method);
        $returnType = $method->getReturnType() ? ': ' . $method->getReturnType() : '';

        return "
    public function {$methodName}({$parameters}){$returnType}
    {
        return \$this->factory->interceptMethod(
            \$this->target,
            '{$methodName}',
            [{$parameterNames}]
        );
    }";
    }

    /**
     * Build parameter list for method
     * 
     * @param ReflectionMethod $method Method reflection
     * @return string Parameter list
     */
    private function buildParameterList(ReflectionMethod $method): string
    {
        $params = [];
        
        foreach ($method->getParameters() as $param) {
            $paramStr = '';
            
            if ($param->getType()) {
                $paramStr .= $param->getType() . ' ';
            }
            
            $paramStr .= '$' . $param->getName();
            
            if ($param->isDefaultValueAvailable()) {
                $paramStr .= ' = ' . var_export($param->getDefaultValue(), true);
            }
            
            $params[] = $paramStr;
        }
        
        return implode(', ', $params);
    }

    /**
     * Build parameter names for method call
     * 
     * @param ReflectionMethod $method Method reflection
     * @return string Parameter names
     */
    private function buildParameterNames(ReflectionMethod $method): string
    {
        $names = [];
        
        foreach ($method->getParameters() as $param) {
            $names[] = '$' . $param->getName();
        }
        
        return implode(', ', $names);
    }

    /**
     * Intercept method call
     * 
     * @param object $target Target object
     * @param string $methodName Method name
     * @param array<int, mixed> $arguments Method arguments
     * @return mixed
     */
    public function interceptMethod(object $target, string $methodName, array $arguments): mixed
    {
        $reflection = new ReflectionClass($target);
        $method = $reflection->getMethod($methodName);
        
        // Check for memoization
        $memoizeAttributes = $method->getAttributes(Memoize::class);
        if (!empty($memoizeAttributes)) {
            $memoConfig = $memoizeAttributes[0]->newInstance();
            $cacheKey = $this->generateCacheKey($target, $methodName, $arguments, $memoConfig);
            
            if ($this->memoCache->contains($target) && isset($this->memoCache[$target][$cacheKey])) {
                return $this->memoCache[$target][$cacheKey]['value'];
            }
        }

        // Check for interceptors
        $interceptAttributes = $method->getAttributes(Intercept::class);
        $interceptConfig = !empty($interceptAttributes) ? $interceptAttributes[0]->newInstance() : null;

        // Before interceptor
        if ($interceptConfig && $interceptConfig->before) {
            $this->callInterceptor($target, $interceptConfig->before, $arguments);
        }

        // Call original method
        $result = $method->invokeArgs($target, $arguments);

        // After interceptor
        if ($interceptConfig && $interceptConfig->after) {
            $this->callInterceptor($target, $interceptConfig->after, [$result]);
        }

        // Store in memo cache
        if (!empty($memoizeAttributes)) {
            if (!$this->memoCache->contains($target)) {
                $this->memoCache[$target] = [];
            }
            $this->memoCache[$target][$cacheKey] = [
                'value' => $result,
                'timestamp' => time(),
                'ttl' => $memoConfig->ttl
            ];
        }

        return $result;
    }

    /**
     * Generate cache key for memoization
     * 
     * @param object $target Target object
     * @param string $methodName Method name
     * @param array $arguments Method arguments
     * @param Memoize $config Memoization configuration
     * @return string Cache key
     */
    private function generateCacheKey(object $target, string $methodName, array $arguments, Memoize $config): string
    {
        if ($config->keyGenerator && method_exists($target, $config->keyGenerator)) {
            return $target->{$config->keyGenerator}($methodName, $arguments);
        }
        
        return md5($methodName . serialize($arguments));
    }

    /**
     * Call interceptor method
     * 
     * @param object $target Target object
     * @param string $interceptorMethod Interceptor method name
     * @param array $arguments Arguments
     */
    private function callInterceptor(object $target, string $interceptorMethod, array $arguments): void
    {
        if (method_exists($target, $interceptorMethod)) {
            $target->$interceptorMethod(...$arguments);
        }
    }
}

/**
 * Proxy interface
 */
interface ProxyInterface
{
    /**
     * Get the target object
     * 
     * @return object
     */
    public function __getTarget(): object;
}

/**
 * Advanced async operation manager using Fibers
 */
class AsyncOperationManager
{
    /**
     * @var array<string, Fiber<mixed, mixed, mixed, mixed>> Active fibers
     */
    private array $fibers = [];
    
    /**
     * @var array<string, mixed> Fiber results
     */
    private array $results = [];

    /**
     * Execute async operation
     * 
     * @param string $id Operation ID
     * @param Closure $operation Operation to execute
     */
    public function execute(string $id, Closure $operation): void
    {
        $this->fibers[$id] = new Fiber($operation);
        $this->fibers[$id]->start();
    }

    /**
     * Resume fiber execution
     * 
     * @param string $id Fiber ID
     * @param mixed $value Value to resume with
     */
    public function resume(string $id, mixed $value = null): void
    {
        if (isset($this->fibers[$id]) && $this->fibers[$id]->isSuspended()) {
            try {
                $result = $this->fibers[$id]->resume($value);
                if ($this->fibers[$id]->isTerminated()) {
                    $this->results[$id] = $result;
                    unset($this->fibers[$id]);
                }
            } catch (FiberError $e) {
                $this->results[$id] = ['error' => $e->getMessage()];
                unset($this->fibers[$id]);
            }
        }
    }

    /**
     * Get operation result
     * 
     * @param string $id Operation ID
     * @return mixed
     */
    public function getResult(string $id): mixed
    {
        return $this->results[$id] ?? null;
    }

    /**
     * Check if operation is complete
     * 
     * @param string $id Operation ID
     * @return bool
     */
    public function isComplete(string $id): bool
    {
        return isset($this->results[$id]);
    }

    /**
     * Wait for all operations to complete
     * 
     * @return Generator<string, mixed>
     */
    public function waitAll(): Generator
    {
        while (!empty($this->fibers)) {
            foreach (array_keys($this->fibers) as $id) {
                $this->resume($id);
                if ($this->isComplete($id)) {
                    yield $id => $this->getResult($id);
                }
            }
            
            // Yield control to prevent blocking
            Fiber::suspend();
        }
    }
}

/**
 * Complex business entity with advanced features
 */
#[ProxyTarget(['calculatePrice', 'processOrder'], 'OrderServiceProxy')]
class OrderService
{
    /**
     * @var AsyncOperationManager Async manager
     */
    private AsyncOperationManager $asyncManager;

    public function __construct()
    {
        $this->asyncManager = new AsyncOperationManager();
    }

    /**
     * Calculate order price with memoization
     * 
     * @param list<array{id: int, quantity: int, price: float}> $items Order items
     * @param string $currency Currency code
     * @return float
     */
    #[Memoize(ttl: 1800, keyGenerator: 'generatePriceKey')]
    #[Intercept(before: 'logPriceCalculation', after: 'validatePriceResult')]
    public function calculatePrice(array $items, string $currency = 'USD'): float
    {
        $total = 0.0;
        
        foreach ($items as $item) {
            $total += $item['price'] * $item['quantity'];
        }
        
        // Simulate complex calculation
        usleep(100000); // 100ms
        
        return $total * $this->getCurrencyMultiplier($currency);
    }

    /**
     * Process order asynchronously
     * 
     * @param array{customer_id: int, items: list<array{id: int, quantity: int}>, payment_method: string} $orderData Order data
     * @return string Order ID
     */
    #[Intercept(before: 'validateOrder', after: 'notifyOrderProcessed')]
    public function processOrder(array $orderData): string
    {
        $orderId = uniqid('order_');
        
        $this->asyncManager->execute($orderId, function() use ($orderData, $orderId) {
            // Simulate async processing
            Fiber::suspend();
            
            // Process payment
            $this->processPayment($orderData['payment']);
            Fiber::suspend();
            
            // Update inventory
            $this->updateInventory($orderData['items']);
            Fiber::suspend();
            
            // Send confirmation
            $this->sendConfirmation($orderData['customer']);
            
            return ['order_id' => $orderId, 'status' => 'completed'];
        });
        
        return $orderId;
    }

    /**
     * Generate cache key for price calculation
     * 
     * @param string $methodName Method name
     * @param array $arguments Method arguments
     * @return string Cache key
     */
    public function generatePriceKey(string $methodName, array $arguments): string
    {
        [$items, $currency] = $arguments;
        return md5($methodName . serialize($items) . $currency);
    }

    /**
     * Log price calculation
     * 
     * @param array $items Order items
     * @param string $currency Currency
     */
    protected function logPriceCalculation(array $items, string $currency): void
    {
        error_log("Calculating price for " . count($items) . " items in {$currency}");
    }

    /**
     * Validate price result
     * 
     * @param float $result Calculated price
     */
    protected function validatePriceResult(float $result): void
    {
        if ($result < 0) {
            throw new \InvalidArgumentException('Price cannot be negative');
        }
    }

    /**
     * Validate order before processing
     * 
     * @param array $orderData Order data
     */
    protected function validateOrder(array $orderData): void
    {
        if (empty($orderData['items'])) {
            throw new \InvalidArgumentException('Order must contain items');
        }
    }

    /**
     * Notify that order was processed
     * 
     * @param string $orderId Order ID
     */
    protected function notifyOrderProcessed(string $orderId): void
    {
        error_log("Order {$orderId} processing initiated");
    }

    /**
     * Get currency multiplier
     * 
     * @param string $currency Currency code
     * @return float Multiplier
     */
    private function getCurrencyMultiplier(string $currency): float
    {
        return match($currency) {
            'EUR' => 0.85,
            'GBP' => 0.73,
            'JPY' => 110.0,
            default => 1.0,
        };
    }

    /**
     * Process payment (simulated)
     * 
     * @param array $paymentData Payment data
     */
    private function processPayment(array $paymentData): void
    {
        // Simulate payment processing
        usleep(500000); // 500ms
    }

    /**
     * Update inventory (simulated)
     * 
     * @param array $items Order items
     */
    private function updateInventory(array $items): void
    {
        // Simulate inventory update
        usleep(200000); // 200ms
    }

    /**
     * Send confirmation (simulated)
     * 
     * @param array $customerData Customer data
     */
    private function sendConfirmation(array $customerData): void
    {
        // Simulate sending confirmation
        usleep(100000); // 100ms
    }

    /**
     * Get async operation status
     * 
     * @param string $orderId Order ID
     * @return array Status information
     */
    public function getOrderStatus(string $orderId): array
    {
        if ($this->asyncManager->isComplete($orderId)) {
            return $this->asyncManager->getResult($orderId);
        }
        
        return ['order_id' => $orderId, 'status' => 'processing'];
    }
}

/**
 * Advanced metaprogramming demonstration
 */
class MetaProgrammingDemo
{
    /**
     * Create dynamic class at runtime
     * 
     * @param string $className Class name
     * @param array $properties Class properties
     * @param array $methods Class methods
     * @return string Generated class name
     */
    public function createDynamicClass(string $className, array $properties, array $methods): string
    {
        $classCode = "class {$className} {\n";
        
        // Add properties
        foreach ($properties as $name => $type) {
            $classCode .= "    public {$type} \${$name};\n";
        }
        
        // Add methods
        foreach ($methods as $methodName => $methodCode) {
            $classCode .= "\n    public function {$methodName}() {\n";
            $classCode .= "        {$methodCode}\n";
            $classCode .= "    }\n";
        }
        
        $classCode .= "}\n";
        
        eval($classCode);
        
        return $className;
    }

    /**
     * Create method dynamically using Closure
     * 
     * @param object $object Target object
     * @param string $methodName Method name
     * @param Closure $implementation Method implementation
     */
    public function addMethod(object $object, string $methodName, Closure $implementation): void
    {
        $boundClosure = $implementation->bindTo($object, $object);
        $object->$methodName = $boundClosure;
    }

    /**
     * Analyze object structure using advanced reflection
     * 
     * @param object $object Object to analyze
     * @return array Analysis result
     */
    public function analyzeObject(object $object): array
    {
        $reflection = new ReflectionClass($object);
        
        return [
            'class' => $reflection->getName(),
            'interfaces' => $reflection->getInterfaceNames(),
            'traits' => $reflection->getTraitNames(),
            'properties' => $this->analyzeProperties($reflection),
            'methods' => $this->analyzeMethods($reflection),
            'constants' => $reflection->getConstants(),
            'attributes' => $this->analyzeAttributes($reflection),
        ];
    }

    /**
     * Analyze class properties
     * 
     * @param ReflectionClass $reflection Class reflection
     * @return array Properties analysis
     */
    private function analyzeProperties(ReflectionClass $reflection): array
    {
        $properties = [];
        
        foreach ($reflection->getProperties() as $property) {
            $properties[] = [
                'name' => $property->getName(),
                'type' => $property->getType()?->getName(),
                'visibility' => $this->getVisibility($property),
                'static' => $property->isStatic(),
                'readonly' => $property->isReadOnly(),
                'attributes' => array_map(
                    fn($attr) => $attr->getName(),
                    $property->getAttributes()
                ),
            ];
        }
        
        return $properties;
    }

    /**
     * Analyze class methods
     * 
     * @param ReflectionClass $reflection Class reflection
     * @return array Methods analysis
     */
    private function analyzeMethods(ReflectionClass $reflection): array
    {
        $methods = [];
        
        foreach ($reflection->getMethods() as $method) {
            $methods[] = [
                'name' => $method->getName(),
                'visibility' => $this->getVisibility($method),
                'static' => $method->isStatic(),
                'abstract' => $method->isAbstract(),
                'final' => $method->isFinal(),
                'parameters' => array_map(
                    fn($param) => [
                        'name' => $param->getName(),
                        'type' => $param->getType()?->getName(),
                        'optional' => $param->isOptional(),
                    ],
                    $method->getParameters()
                ),
                'return_type' => $method->getReturnType()?->getName(),
                'attributes' => array_map(
                    fn($attr) => $attr->getName(),
                    $method->getAttributes()
                ),
            ];
        }
        
        return $methods;
    }

    /**
     * Analyze class attributes
     * 
     * @param ReflectionClass $reflection Class reflection
     * @return array Attributes analysis
     */
    private function analyzeAttributes(ReflectionClass $reflection): array
    {
        $attributes = [];
        
        foreach ($reflection->getAttributes() as $attribute) {
            $attributes[] = [
                'name' => $attribute->getName(),
                'arguments' => $attribute->getArguments(),
            ];
        }
        
        return $attributes;
    }

    /**
     * Get visibility string
     * 
     * @param ReflectionProperty|ReflectionMethod $member Class member
     * @return string Visibility
     */
    private function getVisibility(ReflectionProperty|ReflectionMethod $member): string
    {
        return match(true) {
            $member->isPrivate() => 'private',
            $member->isProtected() => 'protected',
            default => 'public',
        };
    }
}