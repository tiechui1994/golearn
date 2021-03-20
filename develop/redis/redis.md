## redis数据类型对应的底层数据结构

hash: ziplist 或 dict

list: ziplist 或 quicklist

set: intset 或 dict

sortedset: ziplist 或 skiplist

string: int 或 embstr 或 raw
