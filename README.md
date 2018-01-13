The hredis is one client for Redis Server

you can use it to operate Redis.

hredis is one client by Go,I wrote it according to Python redis-py and goredis.

User can use it to operate Redis convienmently.

Usage:


	    Host,Port,Password,Db := "127.0.0.1",6379,"hehe",0
    	redis,err := hredis.Redis(Host,Port,Password,Db)
    	fmt.Println(err)
    	defer redis.DisConnect()
    	fmt.Println(redis.Get("haibo"))
    	fmt.Println(redis.Set("hhh","haibo"))

    	fmt.Println(redis.Incr("h"))
    	result,err := redis.Lrange("sf",0,3)
    	for _,v := range result{
    		fmt.Println(string(v))
    	}
    	fmt.Println(redis.Exist("haibo"))
    	fmt.Println(redis.Del("adf"))


Author:

    Haibo Wang

