#include "bridge.h"
#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus_combine
typedef struct TargetDate {
    Date date;
} TargetDate;

TargetDate* NewDate(int year, int month, int day) {
    Date date = Date(year, month, day);
    TargetDate* t = (TargetDate *)malloc(sizeof(TargetDate));
    t->date = date;
    return t;
}

void SetDate(TargetDate* self, int year, int month, int day) {
    self->date.SetDate(year, month, day);
}

int  getYear(TargetDate* self) {
    return self->date.getYear();
}
int  getMonth(TargetDate* self) {
    return self->date.getMonth();
}

int  getDay(TargetDate* self) {
     return self->date.getDay();
}
#endif


#ifdef __cplusplus_inherit
typedef struct TargetDate : Date {
    TargetDate(int year, int month, int day): Date(year, month, day) {}
    ~TargetDate() {}
} TargetDate;

TargetDate* NewDate(int year, int month, int day) {
     TargetDate* p = new TargetDate(year, month, day);
     return p;
}

void SetDate(TargetDate* self, int year, int month, int day) {
    self->SetDate(year, month, day);
}

int  getYear(TargetDate* self) {
    return self->getYear();
}
int  getMonth(TargetDate* self) {
    return self->getMonth();
}

int  getDay(TargetDate* self) {
     return self->getDay();
}
#endif

#ifdef __cplusplus
}
#endif