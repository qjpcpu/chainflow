package db

import (
	"github.com/garyburd/redigo/redis"
	"gopkg.in/redsync.v1"
	"time"
)

var g_unblock_redis_pool *redis.Pool
var g_locker *redsync.Redsync

func RedisPool() *redis.Pool {
	return g_unblock_redis_pool
}

func DistributionLocker() *redsync.Redsync {
	return g_locker
}

func InitRedis(conn string, redis_db, passwd string) {
	g_unblock_redis_pool = &redis.Pool{
		MaxIdle:     200,
		MaxActive:   200,
		IdleTimeout: 2 * time.Second,
		Dial: func() (redis.Conn, error) {
			connect_timeout := 2 * time.Second
			read_timeout := 2 * time.Second
			write_timeout := 2 * time.Second
			c, err := redis.DialTimeout("tcp", conn, connect_timeout,
				read_timeout, write_timeout)
			if err != nil {
				return nil, err
			}

			if passwd != "" {
				if _, err := c.Do("AUTH", passwd); err != nil {
					c.Close()
					return nil, err
				}
			}

			if redis_db != "" {
				if _, err = c.Do("SELECT", redis_db); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	g_locker = redsync.New([]redsync.Pool{g_unblock_redis_pool})
}
