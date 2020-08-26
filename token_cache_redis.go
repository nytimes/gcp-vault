package gcpvault

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

var redisPool *redis.Pool

type TokenCacheRedis struct {
	cfg *Config
}

func (t TokenCacheRedis) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.TokenCache != nil {

		redisAddr := t.cfg.TokenCacheStorageRedis
		log.Printf("Redis address is:", redisAddr)
		redisPool = redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", redisAddr)
		}, 1)
		log.Printf("Getting connection")
		conn := redisPool.Get()
		defer conn.Close()
		log.Printf("Reading from redis")
		data, err := redis.String(conn.Do("GET", "token-cache"))
		if err != nil {
			errors.Wrap(err, "Error calling redis GET")
			return nil, nil
		}
		var token Token
		err = json.Unmarshal([]byte(data), &token)
		if err != nil {
			errors.Wrap(err, "Error Unmarshaling data")
			return nil, err
		}
		return &token, nil
	}

	return nil, nil

}

func (t TokenCacheRedis) SaveToken(token Token) error {

	if t.cfg.TokenCache != nil {
		redisAddr := t.cfg.TokenCacheStorageRedis
		redisPool = redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", redisAddr)
		}, 1)

		conn := redisPool.Get()
		defer conn.Close()
		payload, err := json.Marshal(&token)
		if err != nil {
			return err
		}
		_, err = redis.String(conn.Do("SET", "token-cache", payload))
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
