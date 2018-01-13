package hredis

import (
	"strings"
)

type RedisError string

func (r RedisError) Error() string {
	return "Redis Error: " + string(r)
}

type AuthenticationError string

func (r AuthenticationError) Error() string {
	return RedisError.Error("Authentication Error")
}


type ConnectionError string

func (cerr ConnectionError) Error() string {
	return ConnectionError.Error("Connection Error")
}

type SocketCloseError  string

func (sr SocketCloseError) Error() string {
	return "SocketError:"
}

type ReplyError map[string]error

func(r ReplyError) ParseError(line string) error {
	prefix,errstring := strings.Split(line," ")[0],strings.Join(strings.Split(line," ")[1:]," ")
	if key,ok:= r[prefix];ok {
		if errstring == "" {
			return key
		}
		return RedisError(errstring)
	}
	return nil
}



