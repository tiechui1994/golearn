/*
C 语言接口, CGO 的接口文档
*/

typedef struct Buffer_T Buffer_T;

Buffer_T* NewBuffer(int size);
void DeleteBuffer(Buffer_T* p);

char* Buffer_Data(Buffer_T* p);
int Buffer_Size(Buffer_T* p);
