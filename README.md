# geecache
> geektutu cache learn
* 三种缓存淘汰(失效)算法：FIFO，LFU 和 LRU
  * FIFO(First In First Out):淘汰缓存中最老(最早添加)的记录
  * LFU(Least Frequently Used):淘汰缓存中访问频率最低的记录
  * LRU(Least Recently Used):最近最少使用,LRU 算法的实现非常简单，维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。
