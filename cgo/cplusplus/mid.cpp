#include "mid.h"
#include <stdio.h>
#include "cxx.h"

void c_function(int cmd){
    printf("c_function. \n");
    cxx_function(cmd);
}