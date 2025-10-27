package com.example.medium

import kotlinx.coroutines.*
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.collect
import kotlin.reflect.KClass
import kotlin.contracts.ExperimentalContracts
import kotlin.contracts.contract

/**
 * Medium complexity demonstration with generics, higher-order functions,
 * coroutines, inline functions, reified types, and contracts
 */

/**
 * Generic repository interface with covariant type parameter
 */
interface Repository<out T : Any> {
    suspend fun findById(id: String): T?
    suspend fun findAll(): List<T>
    suspend fun count(): Long
}

/**
 * Generic mutable repository with contravariant and invariant type parameters
 */
interface MutableRepository<T : Any> : Repository<T> {
    suspend fun save(entity: T): T
    suspend fun update(entity: T): T
    suspend fun delete(id: String): Boolean
    suspend fun saveAll(entities: List<T>): List<T>
}

/**
 * Generic service class with multiple type parameters and constraints
 */
abstract class BaseService<T : Entity, R : MutableRepository<T>>(
    protected val repository: R
) where T : Comparable<T> {

    /**
     * Generic method with reified type parameter
     */
    inline fun <reified U : T> findByType(): Flow<U> = flow {
        repository.findAll()
            .filterIsInstance<U>()
            .forEach { emit(it) }
    }

    /**
     * Higher-order function with generic parameters
     */
    suspend fun <K> processEntities(
        transformer: suspend (T) -> K,
        filter: (K) -> Boolean = { true }
    ): List<K> {
        return repository.findAll()
            .map { transformer(it) }
            .filter(filter)
    }

    /**
     * Method with generic constraints
     */
    suspend fun <U> findAndMap(
        id: String,
        mapper: suspend (T) -> U
    ): U? where U : Any {
        return repository.findById(id)?.let { mapper(it) }
    }

    /**
     * Abstract method for subclasses
     */
    abstract suspend fun validateEntity(entity: T): ValidationResult
}

/**
 * Entity interface with comparable constraint
 */
interface Entity : Comparable<Entity> {
    val id: String
    val version: Long

    override fun compareTo(other: Entity): Int = id.compareTo(other.id)
}

/**
 * Concrete entity implementation
 */
data class User(
    override val id: String,
    override val version: Long,
    val name: String,
    val email: String,
    val role: UserRole
) : Entity {

    /**
     * Custom comparison based on name
     */
    fun compareByName(other: User): Int = name.compareTo(other.name)
}

/**
 * Enum class for user roles
 */
enum class UserRole {
    ADMIN, MODERATOR, USER, GUEST
}

/**
 * Sealed class for validation results
 */
sealed class ValidationResult {
    object Valid : ValidationResult()
    data class Invalid(val errors: List<String>) : ValidationResult()

    /**
     * Helper methods using when expressions
     */
    fun isValid(): Boolean = this is Valid
    fun getErrors(): List<String> = when (this) {
        is Valid -> emptyList()
        is Invalid -> errors
    }
}

/**
 * User repository implementation with coroutines
 */
class UserRepository : MutableRepository<User> {
    private val users = mutableMapOf<String, User>()
    private val mutex = kotlinx.coroutines.sync.Mutex()

    override suspend fun findById(id: String): User? = withContext(Dispatchers.IO) {
        delay(10) // Simulate database delay
        mutex.withLock { users[id] }
    }

    override suspend fun findAll(): List<User> = withContext(Dispatchers.IO) {
        delay(20)
        mutex.withLock { users.values.toList() }
    }

    override suspend fun count(): Long = withContext(Dispatchers.IO) {
        mutex.withLock { users.size.toLong() }
    }

    override suspend fun save(entity: User): User = withContext(Dispatchers.IO) {
        delay(15)
        mutex.withLock {
            users[entity.id] = entity
            entity
        }
    }

    override suspend fun update(entity: User): User = withContext(Dispatchers.IO) {
        delay(15)
        mutex.withLock {
            users[entity.id] = entity.copy(version = entity.version + 1)
            users[entity.id]!!
        }
    }

    override suspend fun delete(id: String): Boolean = withContext(Dispatchers.IO) {
        delay(10)
        mutex.withLock { users.remove(id) != null }
    }

    override suspend fun saveAll(entities: List<User>): List<User> = withContext(Dispatchers.IO) {
        delay(entities.size * 5L)
        mutex.withLock {
            entities.forEach { users[it.id] = it }
            entities
        }
    }

    /**
     * Repository-specific method with Flow
     */
    fun findByRoleFlow(role: UserRole): Flow<User> = flow {
        findAll()
            .filter { it.role == role }
            .forEach { emit(it) }
    }
}

/**
 * User service extending generic base service
 */
class UserService(
    repository: UserRepository
) : BaseService<User, UserRepository>(repository) {

    private val eventChannel = Channel<UserEvent>(Channel.UNLIMITED)

    override suspend fun validateEntity(entity: User): ValidationResult {
        val errors = mutableListOf<String>()

        if (entity.name.isBlank()) {
            errors.add("Name cannot be blank")
        }

        if (!entity.email.contains("@")) {
            errors.add("Invalid email format")
        }

        return if (errors.isEmpty()) {
            ValidationResult.Valid
        } else {
            ValidationResult.Invalid(errors)
        }
    }

    /**
     * Method using inline function with reified type
     */
    inline fun <reified T : UserEvent> getEventsOfType(): Flow<T> = flow {
        for (event in eventChannel) {
            if (event is T) emit(event)
        }
    }

    /**
     * Method with higher-order function parameter
     */
    suspend fun processUsersInBatches(
        batchSize: Int,
        processor: suspend (List<User>) -> Unit
    ) {
        val allUsers = repository.findAll()
        allUsers.chunked(batchSize).forEach { batch ->
            processor(batch)
        }
    }

    /**
     * Coroutine-based method with structured concurrency
     */
    suspend fun createUsersAsync(userDtos: List<UserDto>): List<User> = coroutineScope {
        userDtos.map { dto ->
            async(Dispatchers.IO) {
                val user = User(
                    id = generateId(),
                    version = 1L,
                    name = dto.name,
                    email = dto.email,
                    role = dto.role
                )

                when (val validation = validateEntity(user)) {
                    is ValidationResult.Valid -> {
                        repository.save(user)
                        eventChannel.trySend(UserEvent.Created(user.id))
                        user
                    }
                    is ValidationResult.Invalid -> {
                        throw IllegalArgumentException("Invalid user: ${validation.errors}")
                    }
                }
            }
        }.awaitAll()
    }

    /**
     * Private method for ID generation
     */
    private fun generateId(): String = "USER_${System.currentTimeMillis()}"
}

/**
 * Sealed class for user events
 */
sealed class UserEvent {
    data class Created(val userId: String) : UserEvent()
    data class Updated(val userId: String) : UserEvent()
    data class Deleted(val userId: String) : UserEvent()
}

/**
 * Data transfer object
 */
data class UserDto(
    val name: String,
    val email: String,
    val role: UserRole
)

/**
 * Generic cache class with type parameters
 */
class Cache<K : Any, V : Any>(
    private val maxSize: Int = 100
) {
    private val data = mutableMapOf<K, CacheEntry<V>>()
    private val accessOrder = mutableListOf<K>()

    /**
     * Generic method with nullable return
     */
    fun get(key: K): V? {
        val entry = data[key] ?: return null

        if (entry.isExpired()) {
            data.remove(key)
            accessOrder.remove(key)
            return null
        }

        // Update access order
        accessOrder.remove(key)
        accessOrder.add(key)

        return entry.value
    }

    /**
     * Put method with eviction logic
     */
    fun put(key: K, value: V, ttlMillis: Long = 3600000) {
        // Evict if necessary
        if (data.size >= maxSize && key !in data) {
            val oldestKey = accessOrder.firstOrNull()
            if (oldestKey != null) {
                data.remove(oldestKey)
                accessOrder.remove(oldestKey)
            }
        }

        data[key] = CacheEntry(value, System.currentTimeMillis() + ttlMillis)
        accessOrder.remove(key)
        accessOrder.add(key)
    }

    /**
     * Clear expired entries
     */
    fun clearExpired() {
        val currentTime = System.currentTimeMillis()
        val expiredKeys = data.filterValues { it.isExpired(currentTime) }.keys
        expiredKeys.forEach { key ->
            data.remove(key)
            accessOrder.remove(key)
        }
    }

    /**
     * Inner data class for cache entries
     */
    private data class CacheEntry<T>(
        val value: T,
        val expirationTime: Long
    ) {
        fun isExpired(currentTime: Long = System.currentTimeMillis()): Boolean {
            return currentTime > expirationTime
        }
    }
}

/**
 * Inline function with contract
 */
@OptIn(ExperimentalContracts::class)
inline fun <T> T?.requireNotNull(message: () -> String): T {
    contract {
        returns() implies (this@requireNotNull != null)
    }
    return this ?: throw IllegalArgumentException(message())
}

/**
 * Higher-order function with multiple generic parameters
 */
inline fun <T, R, S> combineAndTransform(
    first: T,
    second: R,
    crossinline transformer: (T, R) -> S
): S {
    return transformer(first, second)
}

/**
 * Extension function with reified type parameter
 */
inline fun <reified T> Any.safeCast(): T? {
    return this as? T
}

/**
 * Coroutine builder extension
 */
suspend fun <T> List<T>.forEachAsync(
    dispatcher: CoroutineDispatcher = Dispatchers.Default,
    action: suspend (T) -> Unit
) = coroutineScope {
    map { async(dispatcher) { action(it) } }.awaitAll()
}

/**
 * Main function demonstrating the medium complexity features
 */
fun main() = runBlocking {
    val userRepository = UserRepository()
    val userService = UserService(userRepository)
    val cache = Cache<String, User>()

    // Create users asynchronously
    val userDtos = listOf(
        UserDto("Alice", "alice@example.com", UserRole.ADMIN),
        UserDto("Bob", "bob@example.com", UserRole.USER),
        UserDto("Charlie", "charlie@example.com", UserRole.MODERATOR)
    )

    val createdUsers = userService.createUsersAsync(userDtos)
    println("Created ${createdUsers.size} users")

    // Use cache
    createdUsers.forEach { user ->
        cache.put(user.id, user)
    }

    // Process users in batches
    userService.processUsersInBatches(2) { batch ->
        println("Processing batch of ${batch.size} users")
        batch.forEach { user ->
            println("- ${user.name} (${user.role})")
        }
    }

    // Use generic method with reified type
    userService.findByType<User>().collect { user ->
        println("Found user: ${user.name}")
    }

    // Use extension function
    val users = listOf("Alice", "Bob", "Charlie")
    users.forEachAsync {
        println("Processing $it asynchronously")
        delay(100)
    }
}