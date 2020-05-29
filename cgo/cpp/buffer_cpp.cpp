#include "buffer_cpp.h"
#include <string>


Buffer::Buffer(int size) {
    this->s_ = new std::string(size, char('\0'));
}

int Buffer::Size() {
    return this->s_->size();
}

char* Buffer::Data() {
    return (char*)this->s_->data();
}