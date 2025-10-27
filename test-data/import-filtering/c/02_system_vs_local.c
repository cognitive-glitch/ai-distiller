// Test Pattern 2: System vs Local Headers
// Tests distinction between <> and "" includes

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include "config.h"
#include "database.h"
#include "network/socket.h"
#include "network/protocol.h"

// Not using: unistd.h, sys/types.h, sys/stat.h, database.h, network/protocol.h

typedef struct {
    char* host;
    int port;
} Config;

Config* load_config() {
    // Using stdlib.h
    Config* cfg = (Config*)malloc(sizeof(Config));
    cfg->host = "localhost";
    cfg->port = 8080;

    // Using stdio.h
    printf("Config loaded: %s:%d\n", cfg->host, cfg->port);

    return cfg;
}

int main() {
    Config* config = load_config();

    // Using stdlib.h
    free(config);

    return 0;
}
