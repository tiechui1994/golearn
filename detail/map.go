package main

/**

map性能总结:

1000个桶保存了10000个键值对, 它的性能是保存1000个键值对的 1/10;
它的性能比一个链表(10000)的性能好1000倍.


type hmap struct {
	count     int    // 记录当前哈希表元素数量
	flags     uint8
	B         uint8  // 当前哈希表持有的buckets数量. 但是因为哈希表的扩容是以2的倍数进行的, 所以这里使用对数来存储 len(buckets) == 2^B
	noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details
	hash0     uint32 // 哈希的种子,这个值再调用哈希函数的时候作为参数传进去. 为哈希函数的结果引入一定的随机性.

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // 哈希在扩容时用于保存之前buckets的字段, 它的大小是当前buckets的一半
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	extra *mapextra // optional fields
}

**/
