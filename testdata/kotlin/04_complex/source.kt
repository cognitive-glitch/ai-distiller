package com.example.complex

import kotlinx.coroutines.*
import kotlinx.coroutines.channels.*
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.selects.select
import kotlin.reflect.KClass
import kotlin.reflect.KProperty
import kotlin.properties.ReadOnlyProperty
import kotlin.properties.ReadWriteProperty
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

/**
 * Complex demonstration with advanced coroutines, channels, flows,
 * DSL builders, delegation patterns, reflection, and advanced generics
 */

/**
 * DSL marker annotation for type-safe builders
 */
@DslMarker
annotation class DatabaseDsl

/**
 * Query DSL builder for creating complex database queries
 */
@DatabaseDsl
class QueryBuilder<T : Any>(private val entityClass: KClass<T>) {
    private val conditions = mutableListOf<Condition<T>>()
    private val orderByClauses = mutableListOf<OrderBy<T>>()
    private var limitValue: Int? = null
    private var offsetValue: Int? = null
    
    /**
     * Add a WHERE condition using DSL
     */
    fun where(condition: ConditionBuilder<T>.() -> Unit) {
        val builder = ConditionBuilder<T>()
        builder.condition()
        conditions.addAll(builder.getConditions())
    }
    
    /**
     * Add ORDER BY clause
     */
    fun orderBy(column: KProperty<*>, direction: SortDirection = SortDirection.ASC) {
        orderByClauses.add(OrderBy(column.name, direction))
    }
    
    /**
     * Set LIMIT
     */
    fun limit(count: Int) {
        limitValue = count
    }
    
    /**
     * Set OFFSET
     */
    fun offset(count: Int) {
        offsetValue = count
    }
    
    /**
     * Build the final query
     */
    internal fun build(): Query<T> {
        return Query(entityClass, conditions, orderByClauses, limitValue, offsetValue)
    }
}

/**
 * Condition builder for WHERE clauses
 */
@DatabaseDsl
class ConditionBuilder<T : Any> {
    private val conditions = mutableListOf<Condition<T>>()
    
    /**
     * Equals condition
     */
    infix fun <V> KProperty<V>.eq(value: V) {
        conditions.add(Condition.Equals(this.name, value))
    }
    
    /**
     * Not equals condition
     */
    infix fun <V> KProperty<V>.ne(value: V) {
        conditions.add(Condition.NotEquals(this.name, value))
    }
    
    /**
     * Greater than condition
     */
    infix fun <V : Comparable<V>> KProperty<V>.gt(value: V) {
        conditions.add(Condition.GreaterThan(this.name, value))
    }
    
    /**
     * Less than condition
     */
    infix fun <V : Comparable<V>> KProperty<V>.lt(value: V) {
        conditions.add(Condition.LessThan(this.name, value))
    }
    
    /**
     * IN condition
     */
    infix fun <V> KProperty<V>.isIn(values: Collection<V>) {
        conditions.add(Condition.In(this.name, values))
    }
    
    /**
     * AND operation
     */
    infix fun ConditionBuilder<T>.and(other: ConditionBuilder<T>.() -> Unit) {
        val otherBuilder = ConditionBuilder<T>()
        otherBuilder.other()
        conditions.add(Condition.And(this.conditions.toList(), otherBuilder.conditions))
    }
    
    /**
     * OR operation
     */
    infix fun ConditionBuilder<T>.or(other: ConditionBuilder<T>.() -> Unit) {
        val otherBuilder = ConditionBuilder<T>()
        otherBuilder.other()
        conditions.add(Condition.Or(this.conditions.toList(), otherBuilder.conditions))
    }
    
    internal fun getConditions(): List<Condition<T>> = conditions
}

/**
 * Sealed class for query conditions
 */
sealed class Condition<T : Any> {
    data class Equals<T : Any>(val column: String, val value: Any?) : Condition<T>()
    data class NotEquals<T : Any>(val column: String, val value: Any?) : Condition<T>()
    data class GreaterThan<T : Any>(val column: String, val value: Comparable<*>) : Condition<T>()
    data class LessThan<T : Any>(val column: String, val value: Comparable<*>) : Condition<T>()
    data class In<T : Any>(val column: String, val values: Collection<*>) : Condition<T>()
    data class And<T : Any>(val left: List<Condition<T>>, val right: List<Condition<T>>) : Condition<T>()
    data class Or<T : Any>(val left: List<Condition<T>>, val right: List<Condition<T>>) : Condition<T>()
}

/**
 * Order by clause
 */
data class OrderBy<T : Any>(
    val column: String,
    val direction: SortDirection
)

/**
 * Sort direction enum
 */
enum class SortDirection { ASC, DESC }

/**
 * Query representation
 */
data class Query<T : Any>(
    val entityClass: KClass<T>,
    val conditions: List<Condition<T>>,
    val orderBy: List<OrderBy<T>>,
    val limit: Int?,
    val offset: Int?
)

/**
 * Repository interface with advanced query support
 */
interface AdvancedRepository<T : Any> {
    suspend fun findByQuery(query: Query<T>): Flow<T>
    suspend fun executeQuery(builder: QueryBuilder<T>.() -> Unit): Flow<T>
}

/**
 * Entity for demonstration
 */
data class Product(
    val id: String,
    val name: String,
    val price: Double,
    val category: String,
    val inStock: Boolean,
    val rating: Double
)

/**
 * Advanced coroutine-based product repository
 */
class ProductRepository : AdvancedRepository<Product> {
    private val products = mutableListOf<Product>()
    private val _productUpdates = MutableSharedFlow<ProductUpdate>()
    
    /**
     * Product updates flow
     */
    val productUpdates: SharedFlow<ProductUpdate> = _productUpdates.asSharedFlow()
    
    init {
        // Initialize with sample data
        runBlocking {
            addSampleProducts()
        }
    }
    
    override suspend fun findByQuery(query: Query<Product>): Flow<Product> = flow {
        products.asSequence()
            .filter { product -> evaluateConditions(product, query.conditions) }
            .sortedWith { a, b -> compareByOrderBy(a, b, query.orderBy) }
            .let { sequence ->
                query.offset?.let { sequence.drop(it) } ?: sequence
            }
            .let { sequence ->
                query.limit?.let { sequence.take(it) } ?: sequence
            }
            .forEach { emit(it) }
    }
    
    override suspend fun executeQuery(builder: QueryBuilder<Product>.() -> Unit): Flow<Product> {
        val queryBuilder = QueryBuilder(Product::class)
        queryBuilder.builder()
        return findByQuery(queryBuilder.build())
    }
    
    /**
     * Add product with real-time updates
     */
    suspend fun addProduct(product: Product) {
        products.add(product)
        _productUpdates.emit(ProductUpdate.Added(product))
    }
    
    /**
     * Update product with optimistic locking simulation
     */
    suspend fun updateProduct(id: String, updater: (Product) -> Product): Product? {
        val index = products.indexOfFirst { it.id == id }
        return if (index >= 0) {
            val oldProduct = products[index]
            val newProduct = updater(oldProduct)
            products[index] = newProduct
            _productUpdates.emit(ProductUpdate.Modified(oldProduct, newProduct))
            newProduct
        } else {
            null
        }
    }
    
    /**
     * Private method to evaluate query conditions
     */
    private fun evaluateConditions(product: Product, conditions: List<Condition<Product>>): Boolean {
        return conditions.all { condition -> evaluateCondition(product, condition) }
    }
    
    /**
     * Private method to evaluate single condition
     */
    private fun evaluateCondition(product: Product, condition: Condition<Product>): Boolean {
        return when (condition) {
            is Condition.Equals -> getPropertyValue(product, condition.column) == condition.value
            is Condition.NotEquals -> getPropertyValue(product, condition.column) != condition.value
            is Condition.GreaterThan -> {
                val value = getPropertyValue(product, condition.column) as? Comparable<Any>
                value?.compareTo(condition.value) ?: 0 > 0
            }
            is Condition.LessThan -> {
                val value = getPropertyValue(product, condition.column) as? Comparable<Any>
                value?.compareTo(condition.value) ?: 0 < 0
            }
            is Condition.In -> condition.values.contains(getPropertyValue(product, condition.column))
            is Condition.And -> evaluateConditions(product, condition.left) && 
                              evaluateConditions(product, condition.right)
            is Condition.Or -> evaluateConditions(product, condition.left) || 
                             evaluateConditions(product, condition.right)
        }
    }
    
    /**
     * Get property value using reflection
     */
    private fun getPropertyValue(product: Product, propertyName: String): Any? {
        return when (propertyName) {
            "id" -> product.id
            "name" -> product.name
            "price" -> product.price
            "category" -> product.category
            "inStock" -> product.inStock
            "rating" -> product.rating
            else -> null
        }
    }
    
    /**
     * Compare products by order by clauses
     */
    private fun compareByOrderBy(a: Product, b: Product, orderBy: List<OrderBy<Product>>): Int {
        for (order in orderBy) {
            val aValue = getPropertyValue(a, order.column) as? Comparable<Any>
            val bValue = getPropertyValue(b, order.column) as? Comparable<Any>
            
            if (aValue != null && bValue != null) {
                val comparison = aValue.compareTo(bValue)
                if (comparison != 0) {
                    return if (order.direction == SortDirection.ASC) comparison else -comparison
                }
            }
        }
        return 0
    }
    
    /**
     * Add sample products for testing
     */
    private suspend fun addSampleProducts() {
        val sampleProducts = listOf(
            Product("1", "Laptop", 999.99, "Electronics", true, 4.5),
            Product("2", "Phone", 599.99, "Electronics", true, 4.2),
            Product("3", "Book", 19.99, "Books", true, 4.8),
            Product("4", "Chair", 149.99, "Furniture", false, 4.0),
            Product("5", "Desk", 299.99, "Furniture", true, 4.3)
        )
        
        sampleProducts.forEach { addProduct(it) }
    }
}

/**
 * Sealed class for product updates
 */
sealed class ProductUpdate {
    data class Added(val product: Product) : ProductUpdate()
    data class Modified(val oldProduct: Product, val newProduct: Product) : ProductUpdate()
    data class Removed(val product: Product) : ProductUpdate()
}

/**
 * Advanced service with channels and select statements
 */
class ProductNotificationService(
    private val repository: ProductRepository
) {
    private val notificationChannel = Channel<Notification>(Channel.UNLIMITED)
    private val subscriptions = mutableMapOf<String, Channel<Notification>>()
    
    /**
     * Start the notification service
     */
    fun start(): Job = GlobalScope.launch {
        // Collect product updates
        launch {
            repository.productUpdates.collect { update ->
                val notification = when (update) {
                    is ProductUpdate.Added -> Notification.ProductAdded(update.product.name)
                    is ProductUpdate.Modified -> Notification.ProductUpdated(
                        update.newProduct.name,
                        update.oldProduct.price,
                        update.newProduct.price
                    )
                    is ProductUpdate.Removed -> Notification.ProductRemoved(update.product.name)
                }
                notificationChannel.send(notification)
            }
        }
        
        // Distribute notifications to subscribers
        launch {
            for (notification in notificationChannel) {
                subscriptions.values.forEach { channel ->
                    channel.trySend(notification)
                }
            }
        }
    }
    
    /**
     * Subscribe to notifications
     */
    suspend fun subscribe(subscriberId: String): ReceiveChannel<Notification> {
        val channel = Channel<Notification>(Channel.BUFFERED)
        subscriptions[subscriberId] = channel
        return channel
    }
    
    /**
     * Unsubscribe from notifications
     */
    suspend fun unsubscribe(subscriberId: String) {
        subscriptions.remove(subscriberId)?.close()
    }
    
    /**
     * Send custom notification
     */
    suspend fun sendNotification(notification: Notification) {
        notificationChannel.send(notification)
    }
}

/**
 * Sealed class for notifications
 */
sealed class Notification {
    data class ProductAdded(val productName: String) : Notification()
    data class ProductUpdated(val productName: String, val oldPrice: Double, val newPrice: Double) : Notification()
    data class ProductRemoved(val productName: String) : Notification()
    data class CustomMessage(val message: String) : Notification()
}

/**
 * Advanced property delegation with caching and validation
 */
class CachedProperty<T>(
    private val initialValue: T,
    private val validator: (T) -> Boolean = { true },
    private val ttl: Duration = 60.seconds
) : ReadWriteProperty<Any?, T> {
    
    private var value: T = initialValue
    private var lastUpdated: Long = System.currentTimeMillis()
    
    override fun getValue(thisRef: Any?, property: KProperty<*>): T {
        val now = System.currentTimeMillis()
        if (now - lastUpdated > ttl.inWholeMilliseconds) {
            // Value expired, return initial value
            value = initialValue
            lastUpdated = now
        }
        return value
    }
    
    override fun setValue(thisRef: Any?, property: KProperty<*>, value: T) {
        if (validator(value)) {
            this.value = value
            this.lastUpdated = System.currentTimeMillis()
        } else {
            throw IllegalArgumentException("Invalid value for property ${property.name}: $value")
        }
    }
}

/**
 * Lazy async property delegate
 */
class AsyncLazy<T>(
    private val initializer: suspend () -> T
) : ReadOnlyProperty<Any?, Deferred<T>> {
    
    private val deferred: Deferred<T> by lazy {
        GlobalScope.async { initializer() }
    }
    
    override fun getValue(thisRef: Any?, property: KProperty<*>): Deferred<T> = deferred
}

/**
 * Configuration class using advanced property delegation
 */
class ServiceConfiguration {
    /**
     * Cached property with validation
     */
    var maxConcurrentRequests: Int by CachedProperty(
        initialValue = 100,
        validator = { it > 0 && it <= 1000 },
        ttl = 30.seconds
    )
    
    /**
     * Async lazy property
     */
    val databaseConnection: Deferred<String> by AsyncLazy {
        delay(1000) // Simulate connection establishment
        "Connected to database"
    }
    
    /**
     * Observable property with custom setter
     */
    var isEnabled: Boolean = true
        set(value) {
            field = value
            println("Service ${if (value) "enabled" else "disabled"}")
        }
}

/**
 * Complex coroutine orchestration with select statements
 */
class DataProcessingOrchestrator {
    private val inputChannel = Channel<DataItem>(Channel.UNLIMITED)
    private val processingResults = Channel<ProcessingResult>(Channel.UNLIMITED)
    private val errorChannel = Channel<ProcessingError>(Channel.UNLIMITED)
    
    /**
     * Start processing with multiple workers and error handling
     */
    fun startProcessing(workerCount: Int = 4): Job = GlobalScope.launch {
        // Start worker coroutines
        repeat(workerCount) { workerId ->
            launch {
                processData(workerId)
            }
        }
        
        // Error handling coroutine
        launch {
            for (error in errorChannel) {
                println("Error in worker ${error.workerId}: ${error.message}")
                // Could implement retry logic here
            }
        }
        
        // Results aggregation
        launch {
            val results = mutableListOf<ProcessingResult>()
            for (result in processingResults) {
                results.add(result)
                if (results.size >= 10) {
                    println("Processed batch of ${results.size} items")
                    results.clear()
                }
            }
        }
    }
    
    /**
     * Process data with select statement for timeouts
     */
    private suspend fun processData(workerId: Int) {
        while (true) {
            select<Unit> {
                inputChannel.onReceive { item ->
                    try {
                        val result = processItem(item)
                        processingResults.send(ProcessingResult(workerId, result))
                    } catch (e: Exception) {
                        errorChannel.send(ProcessingError(workerId, e.message ?: "Unknown error"))
                    }
                }
                
                // Timeout after 5 seconds of inactivity
                onTimeout(5000) {
                    println("Worker $workerId timed out, taking a break")
                    delay(1000)
                }
            }
        }
    }
    
    /**
     * Add item for processing
     */
    suspend fun addItem(item: DataItem) {
        inputChannel.send(item)
    }
    
    /**
     * Simulate item processing
     */
    private suspend fun processItem(item: DataItem): String {
        delay(kotlin.random.Random.nextLong(100, 500)) // Simulate processing time
        return "Processed: ${item.data}"
    }
    
    /**
     * Data classes for processing
     */
    data class DataItem(val id: String, val data: String)
    data class ProcessingResult(val workerId: Int, val result: String)
    data class ProcessingError(val workerId: Int, val message: String)
}

/**
 * Main function demonstrating complex features
 */
fun main() = runBlocking {
    // DSL Query example
    val repository = ProductRepository()
    
    println("=== DSL Query Example ===")
    repository.executeQuery {
        where {
            Product::price gt 100.0
            Product::category eq "Electronics"
            Product::inStock eq true
        }
        orderBy(Product::price, SortDirection.DESC)
        limit(3)
    }.collect { product ->
        println("Found: ${product.name} - $${product.price}")
    }
    
    // Notification service example
    println("\n=== Notification Service Example ===")
    val notificationService = ProductNotificationService(repository)
    val serviceJob = notificationService.start()
    
    val subscription = notificationService.subscribe("user1")
    
    // Listen to notifications in background
    launch {
        for (notification in subscription) {
            when (notification) {
                is Notification.ProductAdded -> println("🆕 New product: ${notification.productName}")
                is Notification.ProductUpdated -> println("📝 Price changed for ${notification.productName}: $${notification.oldPrice} → $${notification.newPrice}")
                is Notification.ProductRemoved -> println("🗑️ Removed: ${notification.productName}")
                is Notification.CustomMessage -> println("📢 ${notification.message}")
            }
        }
    }
    
    // Add a new product to trigger notification
    delay(1000)
    repository.addProduct(Product("6", "Tablet", 399.99, "Electronics", true, 4.1))
    
    // Update a product to trigger notification
    delay(1000)
    repository.updateProduct("1") { it.copy(price = 899.99) }
    
    delay(2000)
    notificationService.unsubscribe("user1")
    
    // Configuration with advanced delegation
    println("\n=== Advanced Property Delegation ===")
    val config = ServiceConfiguration()
    println("Max requests: ${config.maxConcurrentRequests}")
    
    config.maxConcurrentRequests = 200
    println("Updated max requests: ${config.maxConcurrentRequests}")
    
    // Access async property
    val connection = config.databaseConnection.await()
    println("Database: $connection")
    
    // Data processing orchestrator
    println("\n=== Data Processing Orchestrator ===")
    val orchestrator = DataProcessingOrchestrator()
    val processingJob = orchestrator.startProcessing(2)
    
    // Add some items for processing
    repeat(20) { i ->
        orchestrator.addItem(DataProcessingOrchestrator.DataItem("item_$i", "data_$i"))
        delay(50)
    }
    
    delay(3000)
    
    // Cleanup
    serviceJob.cancelAndJoin()
    processingJob.cancelAndJoin()
}