// Test Pattern 3: Conditional Includes
// Tests includes that are conditionally compiled based on preprocessor macros

#include <iostream>
#include <vector>
#include <string>

// Platform-specific includes
#ifdef _WIN32
    #include <windows.h>
    #include <direct.h>
    #define GetCurrentDir _getcwd
#else
    #include <unistd.h>
    #include <sys/types.h>
    #include <pwd.h>
    #define GetCurrentDir getcwd
#endif

// Debug/Release specific includes
#ifdef DEBUG
    #include <cassert>
    #include <cstdio>
    #define DEBUG_PRINT(x) std::cout << "DEBUG: " << x << std::endl
#else
    #include <stdexcept>
    #define DEBUG_PRINT(x)
#endif

// Feature-specific includes
#ifdef USE_OPENSSL
    #include <openssl/sha.h>
    #include <openssl/md5.h>
#endif

#ifdef USE_BOOST
    #include <boost/algorithm/string.hpp>
    #include <boost/filesystem.hpp>
#endif

// C++17/20 feature detection
#if __cplusplus >= 201703L
    #include <filesystem>
    #include <optional>
    #include <variant>
    namespace fs = std::filesystem;
#else
    #include <experimental/filesystem>
    namespace fs = std::experimental::filesystem;
#endif

// Threading support detection
#ifdef _REENTRANT
    #include <thread>
    #include <mutex>
    #include <condition_variable>
#endif

// Not using: Many of the conditional includes depending on defines

using namespace std;

class SystemInfo {
public:
    static string getCurrentDirectory() {
        char buff[FILENAME_MAX];
        GetCurrentDir(buff, FILENAME_MAX);
        return string(buff);
    }
    
    static string getUserInfo() {
        #ifdef _WIN32
            // Using Windows.h
            char username[UNLEN + 1];
            DWORD username_len = UNLEN + 1;
            GetUserName(username, &username_len);
            return string(username);
        #else
            // Using pwd.h and unistd.h
            uid_t uid = getuid();
            struct passwd *pw = getpwuid(uid);
            if (pw) {
                return string(pw->pw_name);
            }
            return "unknown";
        #endif
    }
    
    static void debugLog(const string& message) {
        DEBUG_PRINT(message);
        
        #ifdef DEBUG
            // Using cassert in debug mode
            assert(!message.empty());
            // Using cstdio
            printf("Debug log: %s\n", message.c_str());
        #else
            // In release mode, might throw exception
            if (message.empty()) {
                throw runtime_error("Empty log message");
            }
        #endif
    }
};

class FileOperations {
public:
    static vector<string> listDirectory(const string& path) {
        vector<string> files;
        
        #if __cplusplus >= 201703L
            // Using C++17 filesystem
            for (const auto& entry : fs::directory_iterator(path)) {
                files.push_back(entry.path().filename().string());
            }
        #else
            // Using experimental filesystem or boost
            #ifdef USE_BOOST
                namespace bfs = boost::filesystem;
                bfs::directory_iterator end_iter;
                for (bfs::directory_iterator dir_iter(path); dir_iter != end_iter; ++dir_iter) {
                    files.push_back(dir_iter->path().filename().string());
                }
            #else
                // Fallback implementation
                files.push_back("(directory listing not available)");
            #endif
        #endif
        
        return files;
    }
    
    #ifdef USE_OPENSSL
    static string calculateSHA256(const string& data) {
        unsigned char hash[SHA256_DIGEST_LENGTH];
        SHA256_CTX sha256;
        SHA256_Init(&sha256);
        SHA256_Update(&sha256, data.c_str(), data.length());
        SHA256_Final(hash, &sha256);
        
        // Convert to hex string
        stringstream ss;
        for (int i = 0; i < SHA256_DIGEST_LENGTH; i++) {
            ss << hex << setw(2) << setfill('0') << (int)hash[i];
        }
        return ss.str();
    }
    #endif
};

int main() {
    cout << "Current directory: " << SystemInfo::getCurrentDirectory() << endl;
    cout << "Current user: " << SystemInfo::getUserInfo() << endl;
    
    SystemInfo::debugLog("Application started");
    
    auto files = FileOperations::listDirectory(".");
    cout << "Files in current directory:" << endl;
    for (const auto& file : files) {
        cout << "  " << file << endl;
    }
    
    #ifdef USE_OPENSSL
        string data = "Hello, World!";
        cout << "SHA256 of '" << data << "': " << FileOperations::calculateSHA256(data) << endl;
    #endif
    
    return 0;
}