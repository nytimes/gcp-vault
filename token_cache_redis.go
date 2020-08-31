package gcpvault

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

type TokenCacheRedis struct {
	cfg *Config
}

func (t TokenCacheRedis) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.TokenCache != nil {

		redisAddr := t.cfg.TokenCacheStorageRedis
		tokenKey := t.cfg.TokenCacheKeyName
		log.Printf("Getting connection to: %v", redisAddr)
		conn, err := redis.Dial("tcp", redisAddr, redis.DialConnectTimeout(time.Second*time.Duration(t.cfg.TokenCacheCtxTimeout)))
		if err != nil {
			log.Printf("Error connecting %v", err)
			return nil, err
		}
		defer conn.Close()
		log.Printf("Reading from redis")
		data, err := redis.String(conn.Do("GET", tokenKey))
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

func (t TokenCacheRedis) SaveToken(ctx context.Context, token Token) error {

	if t.cfg.TokenCache != nil {
		redisAddr := t.cfg.TokenCacheStorageRedis
		tokenKey := t.cfg.TokenCacheKeyName
		conn, err := redis.Dial("tcp", redisAddr, redis.DialConnectTimeout(time.Second*time.Duration(t.cfg.TokenCacheCtxTimeout)))
		if err != nil {
			errors.Wrap(err, "Error connecting")
			return err
		}
		defer conn.Close()
		payload, err := json.Marshal(&token)
		if err != nil {
			return err
		}
		_, err = redis.String(conn.Do("SET", tokenKey, payload))
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
