package cursor

import (
	"github.com/garyburd/redigo/redis"
)

const (
	FlowchainSyncCursor = "flowchain:cursors"
)

func Get(conn redis.Conn, name string) (uint64, error) {
	index, err := redis.Uint64(conn.Do("HGET", FlowchainSyncCursor, name))
	if err == redis.ErrNil {
		err = nil
	}
	return index, err
}

func Incr(conn redis.Conn, name string) error {
	_, err := conn.Do("HINCRBY", FlowchainSyncCursor, name, 1)
	return err
}

func Set(conn redis.Conn, name string, index uint64) error {
	_, err := conn.Do("HSET", FlowchainSyncCursor, name, index)
	return err
}
