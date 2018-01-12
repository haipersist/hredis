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
		data,err := conn.SendCmd("AUTH",conn.Password)
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
			if data,err := conn.SendCmd("SELECT",strconv.Itoa(conn.Db));data !="OK" {
				return RedisError("Redis Error")
				fmt.Println(err)
			}
		}
	return nil
}

func (conn *Connection) rawSend(c net.Conn,cmd string,args...string) (interface{}, error) {
        str := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(cmd), cmd)
	cmdbuf := bytes.NewBufferString(str)
        //fmt.Println(str,"str")
	for _,item := range args {
                fmt.Println(fmt.Sprintf("$%d\r\n%s\r\n", len(item), item))
		cmdbuf.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(item), item))
	}
        fmt.Println(cmdbuf)
	cmdbytes := cmdbuf.Bytes()
	fmt.Println("cmdbytes",cmdbytes,string(cmdbytes))
        fmt.Println(c)
	_,err := c.Write(cmdbytes)
	if err != nil {
		fmt.Println("error in write",err)
		return nil,err
	}
	reader := bufio.NewReader(c)
	result,err := conn.ReadResponse(reader)
	fmt.Println("get from redis server:",result,err)
	return result,err
}

func (conn *Connection) DisConnect() {
	err := conn.Pool.ClosePoolConn()
	fmt.Println("close err",err)
}

func (conn *Connection) SendCmd(cmd string,args...string) (interface{},error) {
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
	data,err := conn.rawSend(c,cmd,args...)
	return data,err
}


func (conn *Connection) ReadResponse(reader *bufio.Reader) (interface{}, error) {
	var line string
	var err error
	var sum = 0
	//read until the first non-whitespace line
	for {
		line, err = reader.ReadString('\n')
		sum++
		fmt.Println(line,sum)
		
		if len(line) == 0 || err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			break
		}
	}

	if line[0] == '+' {
		return strings.TrimSpace(line[1:]), nil
	}

	if strings.HasPrefix(line, "-ERR ") {
		errmesg := strings.TrimSpace(line[5:])
		return nil, RedisError(errmesg)
	}

	if line[0] == ':' {
		n, err := strconv.ParseInt(strings.TrimSpace(line[1:]), 10, 64)
		if err != nil {
			return nil, RedisError("Int reply is not a number")
		}
		fmt.Println(":,n",n)
		return n, nil
	}

	if line[0] == '*' {
		size, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return nil, RedisError("MultiBulk reply expected a number")
		}
		if size <= 0 {
			return make([][]byte, 0), nil
		}
		res := make([][]byte, size)
		for i := 0; i < size; i++ {
			res[i], err = readBulk(reader, "")
			fmt.Println("*",res[i])
			if err != nil {
				return nil, err
			}
			// dont read end of line as might not have been bulk
		}
		return res, nil
	}
	fmt.Println(line,"line before return")
	return readBulk(reader, line)
}

func readBulk(reader *bufio.Reader, head string) ([]byte, error) {
	var err error
	var data []byte

	if head == "" {
		head, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}
	switch head[0] {
	case ':':
		data = []byte(strings.TrimSpace(head[1:]))

	case '$':
		size, err := strconv.Atoi(strings.TrimSpace(head[1:]))
		if err != nil {
			return nil, err
		}
		if size == -1 {
			return nil, err
		}
		lr := io.LimitReader(reader, int64(size))
		data, err = ioutil.ReadAll(lr)
		fmt.Println("data:",string(data))
		if err == nil {
			// read end of line
			_, err = reader.ReadString('\n')
		}
	default:
		return nil, RedisError("Expecting Prefix '$' or ':'")
	}
	fmt.Println("before return,data",string(data))
	return data,err
}



