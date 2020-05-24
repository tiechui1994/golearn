#include <string>

/*
C++ 类
*/

struct Buffer {
    std::string* s_;

    Buffer(int size) {
        this->s_ = new std::string(size, char('\0'));
    }

    ~Buffer() {
        delete this->s_;
    }

    int Size() {
        return this->s_->size();
    }

    char* Data() {
        return (char*)this->s_->data();
    }
};
