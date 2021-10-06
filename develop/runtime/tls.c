#include <stdio.h>
#include <unistd.h>
#include <pthread.h>

__thread int g = 0;

void* start(void* arg) {
    printf("start, g[%p]: %d\n", &g, g);

    g++;

    return NULL;
}

int main(int argc, char* argv[]) {
    pthread_t tid;
    g = 100;
    pthread_create(&tid, NULL, start, NULL);
    pthread_join(tid, NULL);

    printf("main, g[%p]: %d\n", &g, g);

    return 0;
}