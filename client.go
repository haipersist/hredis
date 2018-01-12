/*
Redis is one client by Go,I wrote it according to Python redis-py and goredis.

User can use it to operate Redis convienmently.

Usage:
	Host,Port,Password,Db := "127.0.0.1",6379,"hehe",0
	redis,err := hredis.Redis(Host,Port,Password,Db)
	fmt.Println(err)
	defer redis.DisConnect()
	redis.ExecCmd()

CopyRight(c)  Haibo Wang 2018
 */
package hredis


import "fmt"


//The function is used to create Connection,it is one interface between conncetion and public.
// every user should use it to get one client first.
func Redis(host string,port int,password string,db int) (*Client,error) {
	//The function is used to create Connection,it is one interface between conncetion and pubpic.
	// every user should use it to get one client first.
	conn := Connection{
		Host:host,
		Port:port,
		Db:db,
		Password:password,
		Pool:&RedisPool{},
	}
	client := &Client{
		Connection:conn,
	}

	return client,nil
}


type Client struct {
	//I create it seperately,just because I want to add some logical code for client.Connection implement basci logic method.
	//as for client,it's method looks like the basic redis-cli.
	//add some field for client api ,which is used by pubic
	Connection
}



func (client *Client) ExecCmd() {
	data,err := client.SendCmd("GET","haibo")
	fmt.Println(data,err)
}

func (client *Client) ParseResponse() {

}

func (client *Client) BgSave() {

}

/*
Below,write all kinds of commands in Redis, according the format of Redis
 */

