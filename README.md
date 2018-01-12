The hredis is one client for Redis Server

you can use it to operate Redis.

hredis is one client by Go,I wrote it according to Python redis-py and goredis.

User can use it to operate Redis convienmently.

Usage:


	Host,Port,Password,Db := "127.0.0.1",6379,"hehe",0
	redis,err := hredis.Redis(Host,Port,Password,Db)
	fmt.Println(err)
	defer redis.DisConnect()
	redis.ExecCmd()


Author:

    Haibo Wang

