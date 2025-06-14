package com.example.simple

import kotlin.properties.Delegates
import kotlin.random.Random

/**
 * Advanced OOP demonstration with interfaces, inheritance, delegation,
 * property delegates, and companion objects
 */

/**
 * Base interface for all entities
 */
interface Entity {
    val id: String
    val createdAt: Long
    
    /**
     * Default implementation of a method
     */
    fun getAge(): Long = System.currentTimeMillis() - createdAt
    
    /**
     * Abstract method to be implemented
     */
    fun getEntityType(): String
}

/**
 * Auditable interface for tracking changes
 */
interface Auditable {
    var lastModified: Long
    var modifiedBy: String?
    
    /**
     * Method to mark entity as modified
     */
    fun markModified(by: String) {
        lastModified = System.currentTimeMillis()
        modifiedBy = by
    }
}

/**
 * Abstract base class for all persistent entities
 */
abstract class BaseEntity(
    override val id: String,
    override val createdAt: Long = System.currentTimeMillis()
) : Entity {
    
    /**
     * Protected property for subclasses
     */
    protected var isActive: Boolean = true
        private set
    
    /**
     * Public method to activate/deactivate
     */
    fun setActive(active: Boolean) {
        isActive = active
    }
    
    /**
     * Abstract method for validation
     */
    abstract fun validate(): Boolean
    
    /**
     * Open method that can be overridden
     */
    open fun getStatusMessage(): String {
        return if (isActive) "Active" else "Inactive"
    }
}

/**
 * Product class extending BaseEntity
 */
class Product(
    id: String,
    val name: String,
    val category: String,
    private var _price: Double
) : BaseEntity(id), Auditable {
    
    override var lastModified: Long = System.currentTimeMillis()
    override var modifiedBy: String? = null
    
    /**
     * Property with custom getter and setter
     */
    var price: Double
        get() = _price
        set(value) {
            if (value >= 0) {
                _price = value
                markModified("system")
            }
        }
    
    /**
     * Lazy property
     */
    val description: String by lazy {
        "Product: $name in category $category"
    }
    
    /**
     * Observable property
     */
    var stock: Int by Delegates.observable(0) { _, oldValue, newValue ->
        println("Stock changed from $oldValue to $newValue for product $name")
    }
    
    /**
     * Vetoable property
     */
    var discount: Double by Delegates.vetoable(0.0) { _, _, newValue ->
        newValue in 0.0..1.0
    }
    
    override fun getEntityType(): String = "Product"
    
    override fun validate(): Boolean {
        return name.isNotBlank() && 
               category.isNotBlank() && 
               price >= 0 &&
               stock >= 0
    }
    
    override fun getStatusMessage(): String {
        return super.getStatusMessage() + " - Stock: $stock"
    }
    
    /**
     * Product-specific method
     */
    fun applyDiscount(discountPercent: Double): Double {
        return if (discountPercent in 0.0..1.0) {
            price * (1 - discountPercent)
        } else {
            price
        }
    }
    
    companion object {
        private var nextId = 1
        
        /**
         * Factory method for creating products
         */
        fun create(name: String, category: String, price: Double): Product {
            return Product("PROD_${nextId++}", name, category, price)
        }
        
        /**
         * Validation constants
         */
        const val MAX_NAME_LENGTH = 100
        const val MIN_PRICE = 0.01
    }
}

/**
 * Category class with delegation
 */
class Category(
    id: String,
    val name: String,
    private val auditDelegate: Auditable
) : BaseEntity(id), Auditable by auditDelegate {
    
    private val _products = mutableListOf<Product>()
    
    /**
     * Read-only property exposing products
     */
    val products: List<Product> get() = _products.toList()
    
    /**
     * Property with custom getter
     */
    val productCount: Int get() = _products.size
    
    override fun getEntityType(): String = "Category"
    
    override fun validate(): Boolean {
        return name.isNotBlank() && name.length <= 50
    }
    
    /**
     * Add product to category
     */
    fun addProduct(product: Product): Boolean {
        return if (product.category == name && product !in _products) {
            _products.add(product)
            markModified("system")
            true
        } else {
            false
        }
    }
    
    /**
     * Remove product from category
     */
    fun removeProduct(product: Product): Boolean {
        return if (_products.remove(product)) {
            markModified("system")
            true
        } else {
            false
        }
    }
    
    /**
     * Get products by price range
     */
    fun getProductsByPriceRange(min: Double, max: Double): List<Product> {
        return _products.filter { it.price in min..max }
    }
}

/**
 * Audit implementation for delegation
 */
class AuditImpl : Auditable {
    override var lastModified: Long = System.currentTimeMillis()
    override var modifiedBy: String? = null
}

/**
 * Service class demonstrating composition over inheritance
 */
class ProductService(
    private val auditService: AuditService
) {
    private val products = mutableMapOf<String, Product>()
    private val categories = mutableMapOf<String, Category>()
    
    /**
     * Public method to add product
     */
    fun addProduct(product: Product): Result<Unit> {
        return try {
            if (product.validate()) {
                products[product.id] = product
                auditService.logAction("ADD_PRODUCT", product.id)
                Result.success(Unit)
            } else {
                Result.failure(IllegalArgumentException("Invalid product data"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    /**
     * Find product by ID
     */
    fun findProduct(id: String): Product? = products[id]
    
    /**
     * Get all products in category
     */
    fun getProductsInCategory(categoryName: String): List<Product> {
        return products.values.filter { it.category == categoryName }
    }
    
    /**
     * Private helper method
     */
    private fun validateProductData(product: Product): Boolean {
        return product.name.length <= Product.MAX_NAME_LENGTH &&
               product.price >= Product.MIN_PRICE
    }
}

/**
 * Audit service for logging actions
 */
class AuditService {
    private val auditLog = mutableListOf<AuditEntry>()
    
    /**
     * Log an action
     */
    fun logAction(action: String, entityId: String) {
        auditLog.add(AuditEntry(action, entityId, System.currentTimeMillis()))
    }
    
    /**
     * Get audit log
     */
    fun getAuditLog(): List<AuditEntry> = auditLog.toList()
    
    /**
     * Internal audit entry class
     */
    data class AuditEntry(
        val action: String,
        val entityId: String,
        val timestamp: Long
    )
}

/**
 * Singleton object for managing configuration
 */
object ConfigManager {
    private val config = mutableMapOf<String, String>()
    
    /**
     * Set configuration value
     */
    fun setConfig(key: String, value: String) {
        config[key] = value
    }
    
    /**
     * Get configuration value
     */
    fun getConfig(key: String): String? = config[key]
    
    /**
     * Get configuration with default
     */
    fun getConfigOrDefault(key: String, default: String): String {
        return config[key] ?: default
    }
}

/**
 * Demonstration of class with multiple constructors
 */
class ShoppingCart {
    private val items = mutableListOf<CartItem>()
    private var _customerId: String? = null
    
    /**
     * Primary constructor
     */
    constructor()
    
    /**
     * Secondary constructor with customer ID
     */
    constructor(customerId: String) : this() {
        _customerId = customerId
    }
    
    /**
     * Property for customer ID
     */
    val customerId: String? get() = _customerId
    
    /**
     * Add item to cart
     */
    fun addItem(product: Product, quantity: Int = 1) {
        val existingItem = items.find { it.product.id == product.id }
        if (existingItem != null) {
            existingItem.quantity += quantity
        } else {
            items.add(CartItem(product, quantity))
        }
    }
    
    /**
     * Calculate total price
     */
    fun calculateTotal(): Double {
        return items.sumOf { it.product.price * it.quantity }
    }
    
    /**
     * Nested data class for cart items
     */
    data class CartItem(
        val product: Product,
        var quantity: Int
    )
}

/**
 * Main function demonstrating the functionality
 */
fun main() {
    val auditService = AuditService()
    val productService = ProductService(auditService)
    
    // Create products
    val laptop = Product.create("Gaming Laptop", "Electronics", 1299.99)
    laptop.stock = 50
    laptop.discount = 0.1
    
    val phone = Product.create("Smartphone", "Electronics", 699.99)
    phone.stock = 100
    
    // Add products to service
    productService.addProduct(laptop)
    productService.addProduct(phone)
    
    // Create category with delegation
    val electronics = Category("CAT_1", "Electronics", AuditImpl())
    electronics.addProduct(laptop)
    electronics.addProduct(phone)
    
    // Use shopping cart
    val cart = ShoppingCart("CUSTOMER_123")
    cart.addItem(laptop, 1)
    cart.addItem(phone, 2)
    
    println("Cart total: ${cart.calculateTotal()}")
    println("Electronics category has ${electronics.productCount} products")
}