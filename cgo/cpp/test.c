#include "buffer_c.h"

int main() {
    Buffer_T* buf = NewBuffer(1024);

    char* data = Buffer_Data(buf);
    int size = Buffer_Size(buf);

    DeleteBuffer(buf);
}