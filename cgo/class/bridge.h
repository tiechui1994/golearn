#include "date.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct TargetDate TargetDate;

extern TargetDate* NewDate(int year, int month, int day);

extern void SetDate(TargetDate* self, int year, int month, int day);
extern int  getYear(TargetDate* self);
extern int  getMonth(TargetDate* self);
extern int  getDay(TargetDate* self);

#ifdef __cplusplus
}
#endif