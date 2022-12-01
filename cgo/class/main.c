#include "bridge.h"
#include <stdio.h>

int main(int argc, char **argv)
{
  TargetDate* d = NewDate(2022,11,22);
  printf("before Year:%d, Mon:%d, Day:%d\n", getYear(d), getMonth(d), getDay(d));

  SetDate(d, 2022, 11, 30);
  printf("after Year:%d, Mon:%d, Day:%d\n", getYear(d), getMonth(d), getDay(d));
  
  return 0;
}