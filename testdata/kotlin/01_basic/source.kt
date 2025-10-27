package com.example.basic

import kotlin.collections.List
import java.time.LocalDateTime

/**
 * Basic Kotlin features demonstration including data classes, sealed classes,
 * null safety, extension functions, and when expressions
 */

/**
 * Data class showcasing Kotlin's concise syntax and automatic methods
 */
data class User(
    val id: Long,
    val name: String,
    val email: String?,
    internal val createdAt: LocalDateTime = LocalDateTime.now()
) {
    /**
     * Public method to get display name
     */
    fun getDisplayName(): String = name.takeIf { it.isNotBlank() } ?: "Anonymous"

    /**
     * Private validation method
     */
    private fun isValidEmail(): Boolean {
        return email?.contains("@") == true
    }

    companion object {
        /**
         * Factory method for creating users
         */
        fun createUser(name: String, email: String? = null): User {
            return User(System.currentTimeMillis(), name, email)
        }
    }
}

/**
 * Sealed class hierarchy for representing different user states
 */
sealed class UserState {
    /**
     * Active user state
     */
    data class Active(val lastLoginAt: LocalDateTime) : UserState()

    /**
     * Suspended user with reason
     */
    data class Suspended(val reason: String, val until: LocalDateTime?) : UserState()

    /**
     * Permanently banned user
     */
    object Banned : UserState()

    /**
     * Pending verification
     */
    object PendingVerification : UserState()
}

/**
 * Extension function on String for email validation
 */
fun String.isValidEmail(): Boolean {
    return contains("@") && contains(".")
}

/**
 * Extension property on List
 */
val <T> List<T>.secondOrNull: T?
    get() = if (size >= 2) this[1] else null

/**
 * Basic service class demonstrating various Kotlin features
 */
class UserService {
    private val users: MutableList<User> = mutableListOf()
    internal var isInitialized: Boolean = false
        private set

    /**
     * Public method to add a user
     */
    fun addUser(user: User): Boolean {
        return when {
            users.any { it.id == user.id } -> false
            user.email?.isValidEmail() != false -> {
                users.add(user)
                true
            }
            else -> false
        }
    }

    /**
     * Find user by ID with null safety
     */
    fun findUserById(id: Long): User? {
        return users.find { it.id == id }
    }

    /**
     * Get user state using when expression with sealed classes
     */
    fun getUserState(userId: Long): String {
        val user = findUserById(userId) ?: return "User not found"

        return when (determineUserState(user)) {
            is UserState.Active -> "User is active"
            is UserState.Suspended -> "User is suspended"
            is UserState.Banned -> "User is banned"
            is UserState.PendingVerification -> "Verification pending"
        }
    }

    /**
     * Private method to determine user state
     */
    private fun determineUserState(user: User): UserState {
        // Simplified logic for demo
        return UserState.Active(LocalDateTime.now())
    }

    /**
     * Protected method for subclasses
     */
    protected fun validateUserData(user: User): Boolean {
        return user.name.isNotBlank() && (user.email == null || user.email.isValidEmail())
    }
}

/**
 * Enum class with methods
 */
enum class UserRole(val displayName: String, val permissions: Set<String>) {
    ADMIN("Administrator", setOf("read", "write", "delete", "manage")),
    MODERATOR("Moderator", setOf("read", "write", "moderate")),
    USER("User", setOf("read", "write_own"));

    /**
     * Check if role has specific permission
     */
    fun hasPermission(permission: String): Boolean = permission in permissions

    companion object {
        /**
         * Get role by display name
         */
        fun fromDisplayName(displayName: String): UserRole? {
            return values().find { it.displayName == displayName }
        }
    }
}

/**
 * Object declaration for constants and utilities
 */
object UserConstants {
    const val MAX_USERNAME_LENGTH = 50
    const val MIN_PASSWORD_LENGTH = 8
    val ALLOWED_EMAIL_DOMAINS = setOf("gmail.com", "yahoo.com", "outlook.com")

    /**
     * Utility function to check if email domain is allowed
     */
    fun isAllowedEmailDomain(email: String): Boolean {
        val domain = email.substringAfter("@")
        return domain in ALLOWED_EMAIL_DOMAINS
    }
}

/**
 * Simple inline function
 */
inline fun <T> T.applyIf(condition: Boolean, block: T.() -> T): T {
    return if (condition) block() else this
}

/**
 * Main function demonstrating usage
 */
fun main() {
    val service = UserService()
    val user = User.createUser("John Doe", "john@example.com")

    service.addUser(user)

    val foundUser = service.findUserById(user.id)
    println("Found user: ${foundUser?.getDisplayName()}")

    // Extension function usage
    val emails = listOf("test@gmail.com", "invalid-email", "user@yahoo.com")
    val validEmails = emails.filter { it.isValidEmail() }

    // Extension property usage
    val secondEmail = validEmails.secondOrNull
    println("Second valid email: $secondEmail")
}