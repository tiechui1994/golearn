#include <stdio.h>
#include <unistd.h>
#include <pthread.h>
#include <asm/prctl.h>
#include <sys/prctl.h>

__thread int g = 0;

void print_fs_base() {
    unsigned long addr;
    int ret = arch_prctl(ARCH_GET_FS, &addr);
    if (ret < 0) {
        perror("error");
        return;
    }

    printf("fs base addr: %p\n", (void*)addr);
    return;
}

void* start(void* arg) {
    print_fs_base();
    printf("start, g[%p]: %d\n", &g, g);

    g++;

    return NULL;
}

int main(int argc, char* argv[]) {
    pthread_t tid;
    g = 100;
    pthread_create(&tid, NULL, start, NULL);
    pthread_join(tid, NULL);

    print_fs_base();
    printf("main, g[%p]: %d\n", &g, g);

    return 0;
}