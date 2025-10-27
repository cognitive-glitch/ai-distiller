// Test Pattern 4: POSIX and System Headers
// Tests various system-level includes

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <sys/time.h>
#include <signal.h>
#include <dirent.h>

// Not using: errno.h, fcntl.h, sys/mman.h, sys/time.h, signal.h, dirent.h

typedef struct {
    char* data;
    size_t size;
} Buffer;

Buffer* create_buffer(size_t size) {
    // Using stdlib.h
    Buffer* buf = (Buffer*)malloc(sizeof(Buffer));
    buf->data = (char*)malloc(size);
    buf->size = size;

    // Using string.h
    memset(buf->data, 0, size);

    return buf;
}

void destroy_buffer(Buffer* buf) {
    if (buf) {
        // Using stdlib.h
        free(buf->data);
        free(buf);
    }
}

int main() {
    // Using stdio.h
    printf("Creating buffer\n");

    Buffer* buffer = create_buffer(1024);

    // Using string.h
    strcpy(buffer->data, "Hello, World!");
    printf("Buffer content: %s\n", buffer->data);

    destroy_buffer(buffer);

    return 0;
}
