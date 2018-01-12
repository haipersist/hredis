package hredis



type RedisError string

func (r RedisError) Error() string {
	return "Redis Error" + string(r)
}

type AuthenticationError string

func (r AuthenticationError) Error() string {
	return RedisError.Error("Authentication Error")
}


type ConnectionError string

func (cerr ConnectionError) Error() string {
	return ConnectionError.Error("Connection Error")
}




