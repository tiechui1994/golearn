#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define SIZE 5

// C 当中 malloc, realloc, memset, free 案例
void main() {
    long int* arr, d;
    int i;
    int length = SIZE;

    arr = (long int*)malloc(length*sizeof(long int));
    memset(arr, 0, length*sizeof(long int));

    printf("input numbers util you input zero: \n");

    for(;;) {
        printf("> ");
        scanf("%ld", &d);
        if (d==0) {
            arr[i++]=0;
            break;
        } else {
            arr[i++]=d;
        }

        if (i >= length) {
            arr = (long int*)realloc(arr, 2*length*sizeof(long int));
            memset(arr+length, 0, length*sizeof(long int));
            length *= 2;
        }
    }

    printf("\n");
    printf("output all numbers: \n");
    for (int idx=0; idx<i; idx++) {
        if (idx && idx%5==0) {
            printf("\n");
        }

        printf("%ld\t", *(arr+idx));
    }

    printf("\n");

    free(arr);
}