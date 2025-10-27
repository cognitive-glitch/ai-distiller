// Test Pattern 3: Conditional Includes
// Tests preprocessor conditional includes

#include <stdio.h>
#include <stdlib.h>

#ifdef _WIN32
    #include <windows.h>
    #include <winsock2.h>
#else
    #include <pthread.h>
    #include <sys/socket.h>
    #include <netinet/in.h>
#endif

#if defined(__linux__)
    #include <linux/limits.h>
    #include <sys/epoll.h>
#elif defined(__APPLE__)
    #include <sys/event.h>
    #include <mach/mach.h>
#endif

#ifndef NO_OPENSSL
    #include <openssl/ssl.h>
    #include <openssl/err.h>
#endif

// Not using: Most platform-specific headers

void platform_init() {
    // Using stdio.h
    printf("Initializing platform-specific code\n");

#ifdef _WIN32
    printf("Windows platform\n");
#else
    printf("Unix-like platform\n");
#endif
}

int main() {
    platform_init();

    // Using stdlib.h
    void* ptr = malloc(100);
    free(ptr);

    return 0;
}
