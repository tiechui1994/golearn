/**
 *C++ 类转换成 C 接口的实现
 */

#include "./buffer_cpp.h"

extern "C" {
    #include "./buffer_c.h"
}

struct Buffer_T: Buffer {
    Buffer_T(int size): Buffer(size) {}
    ~Buffer_T() {}
};

Buffer_T* NewBuffer(int size) {
    Buffer_T* p = new Buffer_T(size);
    return p;
}

void DeleteBuffer(Buffer_T* p) {
    delete p;
}

char* Buffer_Data(Buffer_T* p) {
    return p->Data();
}

int Buffer_Size(Buffer_T* p) {
    return p->Size();
}
