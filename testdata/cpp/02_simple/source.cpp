#include <iostream>
#include <string>
#include <vector>
#include <memory>
#include <algorithm>
#include <iterator>
#include <functional>
#include <map>
#include <set>

/**
 * @brief Forward declaration of Employee class
 */
class Employee;

/**
 * @brief A comprehensive library management system demonstrating STL usage
 */
namespace LibrarySystem {

/**
 * @brief Abstract base class for library items
 */
class LibraryItem {
public:
    /**
     * @brief Constructor
     * @param id Unique identifier
     * @param title Item title
     */
    LibraryItem(const std::string& id, const std::string& title)
        : id_(id), title_(title), isAvailable_(true) {}

    /**
     * @brief Virtual destructor for proper inheritance
     */
    virtual ~LibraryItem() = default;

    /**
     * @brief Get item ID
     * @return The unique identifier
     */
    const std::string& getId() const { return id_; }

    /**
     * @brief Get item title
     * @return The title
     */
    const std::string& getTitle() const { return title_; }

    /**
     * @brief Check if item is available
     * @return true if available, false otherwise
     */
    bool isAvailable() const { return isAvailable_; }

    /**
     * @brief Borrow the item
     * @return true if successfully borrowed
     */
    virtual bool borrow() {
        if (isAvailable_) {
            isAvailable_ = false;
            return true;
        }
        return false;
    }

    /**
     * @brief Return the item
     */
    virtual void returnItem() {
        isAvailable_ = true;
    }

    /**
     * @brief Pure virtual function for item details
     * @return String representation of item details
     */
    virtual std::string getDetails() const = 0;

protected:
    /**
     * @brief Protected method for subclass use
     * @param status New availability status
     */
    void setAvailability(bool status) { isAvailable_ = status; }

private:
    std::string id_;      ///< Unique identifier
    std::string title_;   ///< Item title
    bool isAvailable_;    ///< Availability status
};

/**
 * @brief Book class extending LibraryItem
 */
class Book : public LibraryItem {
public:
    /**
     * @brief Constructor for Book
     * @param id Book ID
     * @param title Book title
     * @param author Book author
     * @param isbn ISBN number
     */
    Book(const std::string& id, const std::string& title,
         const std::string& author, const std::string& isbn)
        : LibraryItem(id, title), author_(author), isbn_(isbn) {}

    /**
     * @brief Get book author
     * @return The author name
     */
    const std::string& getAuthor() const { return author_; }

    /**
     * @brief Get ISBN
     * @return The ISBN number
     */
    const std::string& getISBN() const { return isbn_; }

    /**
     * @brief Implementation of getDetails for Book
     * @return Book details as string
     */
    std::string getDetails() const override {
        return "Book: " + getTitle() + " by " + author_ + " (ISBN: " + isbn_ + ")";
    }

private:
    std::string author_;  ///< Book author
    std::string isbn_;    ///< ISBN number
};

/**
 * @brief Magazine class extending LibraryItem
 */
class Magazine : public LibraryItem {
public:
    /**
     * @brief Constructor for Magazine
     * @param id Magazine ID
     * @param title Magazine title
     * @param issueNumber Issue number
     * @param publisher Publisher name
     */
    Magazine(const std::string& id, const std::string& title,
             int issueNumber, const std::string& publisher)
        : LibraryItem(id, title), issueNumber_(issueNumber), publisher_(publisher) {}

    /**
     * @brief Get issue number
     * @return The issue number
     */
    int getIssueNumber() const { return issueNumber_; }

    /**
     * @brief Get publisher
     * @return The publisher name
     */
    const std::string& getPublisher() const { return publisher_; }

    /**
     * @brief Implementation of getDetails for Magazine
     * @return Magazine details as string
     */
    std::string getDetails() const override {
        return "Magazine: " + getTitle() + " Issue #" +
               std::to_string(issueNumber_) + " (" + publisher_ + ")";
    }

private:
    int issueNumber_;      ///< Issue number
    std::string publisher_; ///< Publisher name
};

/**
 * @brief Library catalog management class
 */
class LibraryCatalog {
public:
    /**
     * @brief Add an item to the catalog
     * @param item Unique pointer to library item
     */
    void addItem(std::unique_ptr<LibraryItem> item) {
        items_[item->getId()] = std::move(item);
    }

    /**
     * @brief Find item by ID
     * @param id Item identifier
     * @return Raw pointer to item or nullptr if not found
     */
    LibraryItem* findItem(const std::string& id) {
        auto it = items_.find(id);
        return (it != items_.end()) ? it->second.get() : nullptr;
    }

    /**
     * @brief Get available items
     * @return Vector of pointers to available items
     */
    std::vector<LibraryItem*> getAvailableItems() const {
        std::vector<LibraryItem*> available;
        for (const auto& pair : items_) {
            if (pair.second->isAvailable()) {
                available.push_back(pair.second.get());
            }
        }
        return available;
    }

    /**
     * @brief Search items by title (case-insensitive)
     * @param searchTerm Search term
     * @return Vector of matching items
     */
    std::vector<LibraryItem*> searchByTitle(const std::string& searchTerm) const {
        std::vector<LibraryItem*> results;

        // Convert search term to lowercase
        std::string lowerSearchTerm = searchTerm;
        std::transform(lowerSearchTerm.begin(), lowerSearchTerm.end(),
                      lowerSearchTerm.begin(), ::tolower);

        for (const auto& pair : items_) {
            std::string lowerTitle = pair.second->getTitle();
            std::transform(lowerTitle.begin(), lowerTitle.end(),
                          lowerTitle.begin(), ::tolower);

            if (lowerTitle.find(lowerSearchTerm) != std::string::npos) {
                results.push_back(pair.second.get());
            }
        }

        return results;
    }

    /**
     * @brief Get total count of items
     * @return Number of items in catalog
     */
    size_t getItemCount() const { return items_.size(); }

private:
    /// Map of item ID to unique pointer
    std::map<std::string, std::unique_ptr<LibraryItem>> items_;
};

/**
 * @brief Function object for sorting items by title
 */
struct TitleComparator {
    /**
     * @brief Comparison operator
     * @param a First item
     * @param b Second item
     * @return true if a's title comes before b's title
     */
    bool operator()(const LibraryItem* a, const LibraryItem* b) const {
        return a->getTitle() < b->getTitle();
    }
};

/**
 * @brief Utility class for library operations
 */
class LibraryUtils {
public:
    /**
     * @brief Sort items by title
     * @param items Vector of item pointers to sort
     */
    static void sortByTitle(std::vector<LibraryItem*>& items) {
        std::sort(items.begin(), items.end(), TitleComparator{});
    }

    /**
     * @brief Filter items using a predicate
     * @tparam Predicate Function object type
     * @param items Input items
     * @param pred Predicate function
     * @return Filtered items
     */
    template<typename Predicate>
    static std::vector<LibraryItem*> filterItems(
        const std::vector<LibraryItem*>& items, Predicate pred) {
        std::vector<LibraryItem*> filtered;
        std::copy_if(items.begin(), items.end(),
                    std::back_inserter(filtered), pred);
        return filtered;
    }

    /**
     * @brief Count items by type
     * @tparam ItemType Specific item type
     * @param items Items to count
     * @return Count of items of specified type
     */
    template<typename ItemType>
    static size_t countItemsByType(const std::vector<LibraryItem*>& items) {
        return std::count_if(items.begin(), items.end(),
                           [](const LibraryItem* item) {
                               return dynamic_cast<const ItemType*>(item) != nullptr;
                           });
    }

private:
    /**
     * @brief Private constructor - utility class
     */
    LibraryUtils() = delete;
};

} // namespace LibrarySystem

/**
 * @brief RAII wrapper for file operations
 */
class FileManager {
public:
    /**
     * @brief Constructor
     * @param filename File to manage
     */
    explicit FileManager(const std::string& filename)
        : filename_(filename), isOpen_(false) {}

    /**
     * @brief Destructor - automatically closes file
     */
    ~FileManager() {
        if (isOpen_) {
            close();
        }
    }

    /**
     * @brief Copy constructor deleted for RAII
     */
    FileManager(const FileManager&) = delete;

    /**
     * @brief Assignment operator deleted for RAII
     */
    FileManager& operator=(const FileManager&) = delete;

    /**
     * @brief Open the file
     * @return true if successfully opened
     */
    bool open() {
        // Simulate file opening
        isOpen_ = true;
        return true;
    }

    /**
     * @brief Close the file
     */
    void close() {
        if (isOpen_) {
            // Simulate file closing
            isOpen_ = false;
        }
    }

    /**
     * @brief Check if file is open
     * @return true if file is open
     */
    bool isOpen() const { return isOpen_; }

private:
    std::string filename_;  ///< Managed filename
    bool isOpen_;          ///< File open status
};

/**
 * @brief Demonstration function
 */
void demonstrateLibrarySystem() {
    using namespace LibrarySystem;

    // Create catalog
    LibraryCatalog catalog;

    // Add items
    catalog.addItem(std::make_unique<Book>("B001", "The C++ Programming Language",
                                          "Bjarne Stroustrup", "978-0321563842"));
    catalog.addItem(std::make_unique<Book>("B002", "Effective C++",
                                          "Scott Meyers", "978-0321334879"));
    catalog.addItem(std::make_unique<Magazine>("M001", "C++ Today", 42, "Tech Publications"));

    // Search and demonstrate
    auto cppItems = catalog.searchByTitle("C++");
    LibraryUtils::sortByTitle(cppItems);

    // Filter available items
    auto availableItems = LibraryUtils::filterItems(
        cppItems, [](const LibraryItem* item) { return item->isAvailable(); });

    // Count books
    size_t bookCount = LibraryUtils::countItemsByType<Book>(cppItems);

    std::cout << "Found " << cppItems.size() << " C++ related items" << std::endl;
    std::cout << "Available: " << availableItems.size() << std::endl;
    std::cout << "Books: " << bookCount << std::endl;
}