package hredis

import (
	"strings"
	//"reflect"
	"strconv"
	"net"
	"fmt"
	"bytes"
	"bufio"
	"io/ioutil"
	"io"
	//"os"
)

type Connection struct {
	Host           string
	Port           int
	Db             int
	Password       string
	Pool           ConnectionPool
}

type ConnectionPool interface {
	GetConn()               (net.Conn,error)
	PutConn(conn net.Conn)  error
	UsingConn()             []Connection
	ClosePoolConn()         []error
}

type RedisPool struct {
	pool         chan net.Conn
}

//set the size of the connecion pool is 4
const MAXCONNUM = 4

var exception ReplyError


func (r *RedisPool) GetConn() (net.Conn,error) {
	// If the pool channel has created con,it return one,or else return nil.
	//the pool should keep full value,avoid it is blocked.
	//when new conn request comes before the first conn push back,and the chan has no value,it will be blocked until other conn
	//push back conn to the Pool
	if r.pool == nil {
		r.pool = make(chan net.Conn,MAXCONNUM)
		for i:=0;i<MAXCONNUM;i++ {
			r.pool <- nil
		}
	}
	c := <- r.pool
	if c == nil {
		return nil,RedisError("No created connection")
	}
	fmt.Println("I'm in Pool")
	return c,nil
}

func(r *RedisPool) ClosePoolConn() []error {
	e := make([]error,len(r.pool))
	sum := 1
	for c := range r.pool {
		fmt.Println("c in pool",c)
		if c != nil {
			err:= c.Close()
			if err != nil {
				fmt.Println(err)
				e = append(e,err)
			} else {
			 fmt.Println("the con is closed")
			}
		}
		if sum==len(r.pool) {
		   close(r.pool)
		   break
		
		} else {
		sum++
		}
	}
	fmt.Println("after for in chan",e)
	return e
}

func (r *RedisPool) PutConn(conn net.Conn) error{
	if len(r.pool) < MAXCONNUM {
		r.pool <- conn
		return nil
	}
	return RedisError("the Pool is Full,sorry")
}

func (r *RedisPool) UsingConn() []Connection {
	return []Connection{}
}

/*
Connection methods
 */

func(conn *Connection) GetAddr() string {
	addr := make([]string,2)
	port := strconv.Itoa(conn.Port)
        fmt.Println(port,"port")
	addr = append(addr,conn.Host,port)
	tcpaddr := strings.Join([]string{conn.Host,port},":")
	fmt.Println("tcpaddr:",tcpaddr)
	return tcpaddr
}


func (conn *Connection) Connect() (net.Conn,error) {
	tcpaddr := conn.GetAddr()
	c,err := net.Dial("tcp",tcpaddr)
	fmt.Println(c,err)
	if err != nil {
		e := ConnectionError("Connection")
		err = e
		c.Close()
		return nil,err
	} else {
		fmt.Println("TCP Connection Created Successfully!")
		//err = conn.OnConnect()
		//conn.Pool.PutConn(c)
		return c,err
	}

}

func (conn *Connection) OnConnect() error {
	if conn.Password != "" {
		fmt.Println(conn.Password)
		data,err := conn.send_cmd("AUTH",conn.Password)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(data)
		if data != "OK" {
			fmt.Println(data,"auth")
			ar := AuthenticationError("authen fail")
			fmt.Println(ar,err)
			return ar
		}
	}else {
		fmt.Println("password is not set. ")
	}
	if conn.Db != 0 {
			fmt.Println(conn.Db)
			if data,err := conn.send_cmd("SELECT",strconv.Itoa(conn.Db));data !="OK" {
				return RedisError("Redis Error")
				fmt.Println(err)
			}
		}
	return nil
}


// pack the cmd and args into the format which is according to Redis protocal
func (conn *Connection) pack_send(c net.Conn,cmd string,args...string) (interface{}, error) {
	//the cmd_str includes:1.the num of cmd and args,args is one array, 2.the length of cmd string,3.cmd
	cmd_str := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(cmd), cmd)
	cmd_buffer := bytes.NewBufferString(cmd_str)
        //fmt.Println(str,"str")
	for _,arg := range args {
		//as for each of args,should set its length and string itself
		fmt.Println(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
		cmd_buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	fmt.Println(cmd_buffer)
	//should convert string into bytes before sending to socket.
	cmdbytes := cmd_buffer.Bytes()
	//write bytes into connect socket
	_,err := c.Write(cmdbytes)
	if err != nil {
		//I think,as long as we add defer conn.Disconnect() in main func,defer will continue before exit.
		panic(err)
		//return nil,err
	}
	//read response from redis server
	reader := bufio.NewReader(c)
	result,err := conn.read_response(reader)
	fmt.Println("get from redis server:",result,err)
	return result,err
}

func (conn *Connection) DisConnect() {
	err := conn.Pool.ClosePoolConn()
	fmt.Println("close err",err)
}

func (conn *Connection) send_cmd(cmd string,args...string) (interface{},error) {
	fmt.Println("first: connect")
	c,err := conn.Pool.GetConn()
	if err != nil {
		fmt.Println(err)
		c,err = conn.Connect()
		fmt.Println("new",c)
		fmt.Println("create one new connection")
	} else {
		fmt.Println("use existed connection")
	}
	fmt.Println("connect success before send cmd")
	err = conn.Pool.PutConn(c)
	if err == nil {
		fmt.Println("put new con into chan successfully")
	}
	data,err := conn.pack_send(c,cmd,args...)
	return data,err
}


func (conn *Connection) read_response(reader *bufio.Reader) (interface{}, error) {
	var line string
	var err error
	//read until the first non-whitespace line,should use '',it is byte type
	line,err = reader.ReadString('\n')
	if len(line) == 0 || err != nil {
		panic(err)
	}
	line = strings.TrimSpace(line)
	switch head:=line[0];head {
	//it is byte format
	case '+' :
		return line[1:],nil
	case '-':
		exception = ReplyError{"ERR":RedisError("ERRError"),
			"EXECABORT": RedisError("ExecAbortError"),
			"LOADING": RedisError("BusyLoadingError"),
			"NOSCRIPT": RedisError("NoScriptError"),
			"READONLY": RedisError("READONLY"),
		}
		return nil,exception.ParseError(line)
	case ':':
		//add for :,the func return int type
		num,err := strconv.ParseInt(line[1:],10,64)
		if err != nil {
			return nil,RedisError("the reply is not inerger as our expection")
		}
		return num,nil
	case '$':
		//return byte type and nil
		reply,err := conn.read_bulk(reader,line)
		if err != nil {
			return nil,err
		}
		return reply,nil
	case '*':
		num,err := strconv.Atoi(line[1:])
		if err != nil {
			return nil,RedisError("the reply is not inerger as our expection")
		}
		if num == 0 {
			return nil,RedisError("the key you want to query does not exists.")
		}
		if num == -1 {
			//different from [] and error
			return nil,RedisError("may me timeoutfor blop.")
		}
		reply := make([][]byte,num)
		for i:=0;i<num;i++ {
			item,err := conn.read_bulk(reader,"")
			if err != nil {
				return nil,err
			}
			reply[i] = item
		}
		return reply,nil
	default:
		fmt.Println("reply head:",head)
		panic("the redis response is not invalid! it not in -,+,:,$,*")
	}
	return nil,err
}

func(conn *Connection) read_bulk(reader *bufio.Reader, head string) ([]byte, error) {
	var err error
	var result []byte

	if head == "" {
		head, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}
	if head[0] == '$' {
		//In head,the next should be the length of reply string
		size, err := strconv.Atoi(strings.TrimSpace(head[1:]))
		if err != nil {
			return nil, err
		}
		if size == -1 {
			return nil, RedisError("the key you get dose not exists")
		}

		lr := io.LimitReader(reader, int64(size))
		result, err = ioutil.ReadAll(lr)
		if err == nil {
			// read end of line
			_, err = reader.ReadString('\n')
		}
		return result,err
	} else if head[0] == ':'{
		return []byte(strings.TrimSpace(head[1:])),nil
	}
	return nil,RedisError("the muliple reply is invalid")

}



