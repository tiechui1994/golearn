# C++ 和 C 的桥接

## C call C++ 

C++ class

```
class Date
{
private:
    int m_year;
    int m_month;
    int m_day;

public:
    Date(int year, int month, int day);

    void SetDate(int year, int month, int day);

    int getYear() { return m_year; }
    int getMonth() { return m_month; }
    int getDay()  { return m_day; }
};
```

组合方式:

```
typedef struct TargetDate {
    Date date;
} TargetDate;
```

继承:

```
typedef struct TargetDate : Date {
    TargetDate(int year, int month, int day): Date(year, month, day) {}
    ~TargetDate() {}
} TargetDate;
```

