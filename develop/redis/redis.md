## redis数据类型对应的底层数据结构

hash: ziplist 或 dict

list: ziplist 或 quicklist

set: intset 或 dict

sortedset: ziplist 或 skiplist

string: int(可以使用整数表示的字符串) 或 embstr(小于等于44字符串) 或 raw(大于44的字符串)
