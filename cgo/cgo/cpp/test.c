#include "bridge.h"
#include <stdio.h>

int main() {
    void* buf = NewBuffer(10);

    char* data = Buffer_Data(buf);
    int size = Buffer_Size(buf);
    printf("data: %s, size: %d \n", data, size);
    DeleteBuffer(buf);
}