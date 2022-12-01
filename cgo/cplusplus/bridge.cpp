#include "bridge.h"
#include "func.h"
#include <stdio.h>

void c_function(int cmd){
    printf("c_function. \n");
    cxx_function(cmd);
}