package gcpvault

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

type TokenCacheRedis struct {
	cfg *Config
}

func (t TokenCacheRedis) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.TokenCache != nil {

		redisAddr := t.cfg.TokenCacheStorageRedis
		log.Printf("Getting connection to: %v", redisAddr)
		conn, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			errors.Wrap(err, "Error connecting")
			return nil, err
		}
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
		conn, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			errors.Wrap(err, "Error connecting")
			return err
		}
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
