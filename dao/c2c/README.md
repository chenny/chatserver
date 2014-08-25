####用户的C2C离线消息处理：offlinemsg.go

规则：

保证每条离线消息的存活期是7天 == 604800秒

策略：

用户关联离线消息(uuid, msg1, msg2, ...)按每天计算保存在7个dbindex中，
redisdb：127.0.0.1:6380，例如：
第一天保存在 index: 1 中，expire为7天 == 604800秒；
第二天保存在 index: 2 中， expire为6天 == 518400秒；
......
也就是说同一天的离线消息的expire都是一样的，expire一到就把这个List删除。


####根据测试， 采用4个以上的线程访问redis效率最佳

###注意tcp的conn（redis的conn）不能多线程使用





