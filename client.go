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

import (
	"strconv"
	"fmt"
)


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

func (client *Client) Exist(key string) (bool,error) {
	reply,err := client.send_cmd("EXISTS",key)
	if err != nil {
		return false,err
	}
	if reply == 1{
		return true,nil
	}
	return false,nil
}


func (client *Client) Del(key string) (bool,error) {
	reply,err := client.send_cmd("DEL",key)
	if err != nil {
		return false,err
	}
	if reply == 1{
		return true,nil
	}
	return false,nil
}

func (client *Client) Get(key string) (string,error) {
	data,err := client.send_cmd("GET",key)
	return string(data.([]byte)),err
}

func (client *Client) Set(key string,value string) (string,error) {
	reply,err := client.send_cmd("SET",key,value)
	return reply.(string),err
}

func (client *Client) Incr(key string) (int64,error) {
	result,err := client.send_cmd("INCR",key)
	if err != nil {
		return 0,err
	}
	return result.(int64),nil
}

func (client *Client) Lpush(key string,args...string)(int64,error) {
	para := make([]string,len(args)+1)
	para = append(para,key)
	para = append(para,args...)
	fmt.Println("lpush para",para)
	reply,err := client.send_cmd("LPUSH",para...)
	if err != nil {
		return 0,err
	}
	return reply.(int64),nil
}

func (client *Client) Lrange(key string,start int,end int) ([][]byte,error) {
	start_index,end_index := strconv.Itoa(start),strconv.Itoa(end)
	result,err := client.send_cmd("LRANGE",key,start_index,end_index)
	if err == nil {
		return result.([][]byte),nil
	}
	return nil,err
}


func (client *Client) BgSave() {

}

/*
Below,write all kinds of commands in Redis, according the format of Redis
 */

