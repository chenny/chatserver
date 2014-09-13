####redislist: 封装了 redis的 string--> list 操作，采用4个线程访问。



####cid: 处理用户上线，下线，记录保存在内存中



####c2c: 处理用户的C2C离线消息

redisdb：127.0.0.1:6379

expire: 604800秒(7天)



####group/groupmsg: 处理讨论组的离线消息

redisdb：127.0.0.1:6380

expire: 25200秒(7小时)



####group/groupinfo: 处理讨论组和用户之间的关系: 

uuid -> groupid1, groupid2, groupid3, ...

redisdb: 127.0.0.1:6381

expire: -1 (表示不设置expire)


groupid --> group_name, group_owner, uuid1, uuid2, uuid3, ...

redisdb: 127.0.0.1:6382

expire: -1 (表示不设置expire)










