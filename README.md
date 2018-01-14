The hredis is one client for Redis Server

you can use it to operate Redis.

hredis is written by Go,I wrote it according to Python redis-py and goredis.

#Features

1. Connection Pool .the pool size is 4 by default.
2. command function is same as Redis command.
3. Transaction


User can use it to operate Redis convienmently.


#Install

      go install github.com/haipersist/hredis


#Usage:

	Host,Port,Password,Db := "127.0.0.1",6379,"hehe",0
    	redis,err := hredis.Redis(Host,Port,Password,Db)

    	defer redis.DisConnect()
    	result,err := redis.Get("haibo"))
    	_,err := redis.Set("hhh","haibo")
    	result,err := redis.Incr("h")
    	result,err := redis.Lrange("sf",0,3)
    	for _,v := range result{
    		fmt.Println(string(v))
    	}

By default,the transation is closed,if you want to use transaction. you should add one like below :

     redis.Transaction = true

 then you can use redis.Multi,redis.Exec and so on.


#Author:

    Haibo Wang