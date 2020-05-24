/**
 *C++ 类转换成 C 接口的实现
 */

#include "buffer_cpp.h"

// 在 C++ 源文件包含时需要用 `extern "C"` 语句说明. 另外 Buffer_T 的实现只是从 Buffer
// 继承的类, 这样可以简化包装代码的实现.
// 同时和 CGO 通信时必须通过 Buffer_T 指针.
extern "C" {
    #include "buffer_c.h" // 用于 CGO, 必须采用 C 语言规范的名字修饰规则.
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
