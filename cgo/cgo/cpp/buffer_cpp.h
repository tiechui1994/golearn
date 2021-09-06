#include <string>

/*
C++ 类
*/

class Buffer {
    private:
        std::string* s_;

    public:
        Buffer(int size);
        ~Buffer(){}

        int Size();
        char* Data();
};
