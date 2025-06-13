// Advanced Kotlin features test
package com.example.advanced

import kotlinx.coroutines.*
import java.util.concurrent.atomic.AtomicInteger

// Annotations
@Target(AnnotationTarget.CLASS, AnnotationTarget.FUNCTION)
@Retention(AnnotationRetention.RUNTIME)
annotation class Experimental(val reason: String)

@Target(AnnotationTarget.PROPERTY)
annotation class Inject

// Interface with default implementation
interface Repository<T> {
    fun findById(id: Long): T?
    fun findAll(): List<T>
    fun save(entity: T): T
    fun delete(entity: T)
    
    // Default implementation
    fun count(): Int = findAll().size
}

// Sealed class hierarchy
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: Exception) : Result<Nothing>()
    object Loading : Result<Nothing>()
    
    inline fun <R> map(transform: (T) -> R): Result<R> = when (this) {
        is Success -> Success(transform(data))
        is Error -> this
        is Loading -> this
    }
}

// Abstract class with generics
abstract class BaseViewModel<S : Any> {
    protected abstract val initialState: S
    private var state: S? = null
    
    fun getState(): S = state ?: initialState
    
    protected abstract fun reduce(currentState: S, action: Any): S
}

// Data class with secondary constructor
data class User(
    val id: Long,
    val name: String,
    val email: String,
    val isActive: Boolean = true
) {
    var lastLogin: Long? = null
        private set
    
    constructor(name: String, email: String) : this(0, name, email)
    
    fun updateLastLogin(timestamp: Long) {
        lastLogin = timestamp
    }
}

// Class with companion object
@Experimental("Testing new repository pattern")
class UserRepository : Repository<User> {
    @Inject
    private lateinit var database: Database
    
    private val users = mutableListOf<User>()
    private val idCounter = AtomicInteger(0)
    
    companion object {
        const val TABLE_NAME = "users"
        private const val MAX_BATCH_SIZE = 100
        
        @JvmStatic
        fun validateEmail(email: String): Boolean {
            return email.contains("@") && email.contains(".")
        }
    }
    
    override fun findById(id: Long): User? = users.find { it.id == id }
    
    override fun findAll(): List<User> = users.toList()
    
    override fun save(entity: User): User {
        val savedUser = if (entity.id == 0L) {
            entity.copy(id = idCounter.incrementAndGet().toLong())
        } else {
            entity
        }
        users.add(savedUser)
        return savedUser
    }
    
    override fun delete(entity: User) {
        users.removeIf { it.id == entity.id }
    }
    
    // Extension function inside class
    private fun User.isValid(): Boolean {
        return name.isNotBlank() && validateEmail(email)
    }
}

// Interface for database (just for example)
interface Database {
    fun execute(query: String): Any
}

// Inline class (value class)
@JvmInline
value class Password(private val value: String) {
    init {
        require(value.length >= 8) { "Password must be at least 8 characters" }
    }
    
    fun hash(): String = value.hashCode().toString()
}

// Object with operator overloading
object Vector3D {
    data class Vec3(val x: Float, val y: Float, val z: Float) {
        operator fun plus(other: Vec3) = Vec3(x + other.x, y + other.y, z + other.z)
        operator fun minus(other: Vec3) = Vec3(x - other.x, y - other.y, z - other.z)
        operator fun times(scalar: Float) = Vec3(x * scalar, y * scalar, z * scalar)
        
        infix fun dot(other: Vec3) = x * other.x + y * other.y + z * other.z
        infix fun cross(other: Vec3) = Vec3(
            y * other.z - z * other.y,
            z * other.x - x * other.z,
            x * other.y - y * other.x
        )
    }
}

// Higher-order functions and lambdas
inline fun <T> measureTime(block: () -> T): Pair<T, Long> {
    val start = System.currentTimeMillis()
    val result = block()
    val time = System.currentTimeMillis() - start
    return result to time
}

// Suspend functions (coroutines)
class AsyncService {
    suspend fun fetchUser(id: Long): Result<User> = coroutineScope {
        delay(1000) // Simulate network delay
        Result.Success(User(id, "John Doe", "john@example.com"))
    }
    
    suspend fun fetchUsers(): List<User> = withContext(Dispatchers.IO) {
        delay(2000)
        listOf(
            User(1, "John", "john@example.com"),
            User(2, "Jane", "jane@example.com")
        )
    }
}

// Type aliases
typealias UserId = Long
typealias UserPredicate = (User) -> Boolean

// Extension properties
val String.lastChar: Char?
    get() = if (isEmpty()) null else this[length - 1]

// Delegation
class LoggingList<T>(private val delegate: MutableList<T> = mutableListOf()) : 
    MutableList<T> by delegate {
    
    override fun add(element: T): Boolean {
        println("Adding element: $element")
        return delegate.add(element)
    }
}

// DSL example
class HtmlBuilder {
    private val elements = mutableListOf<String>()
    
    fun tag(name: String, block: HtmlBuilder.() -> Unit) {
        elements.add("<$name>")
        block()
        elements.add("</$name>")
    }
    
    fun text(content: String) {
        elements.add(content)
    }
    
    fun build(): String = elements.joinToString("")
}

fun html(block: HtmlBuilder.() -> Unit): String {
    return HtmlBuilder().apply(block).build()
}

// Generic constraints
fun <T> List<T>.secondOrNull(): T? where T : Comparable<T> {
    return if (size >= 2) this[1] else null
}

// Tail recursive function
tailrec fun fibonacci(n: Int, a: Long = 0, b: Long = 1): Long {
    return if (n == 0) a else fibonacci(n - 1, b, a + b)
}