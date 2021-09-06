#include "math.h"
#include <stdio.h>

int add(int a, int b) {
    printf("add args: %d, %d \n", a, b);
    return a+b;
}

int sub(int a, int b) {
    printf("sub args: %d, %d \n", a, b);
    return a-b;
}