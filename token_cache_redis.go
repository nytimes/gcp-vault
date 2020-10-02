package gcpvault

import (
	"context"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"time"
)

type TokenCacheRedis struct {
	cfg *Config
}

func (t TokenCacheRedis) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.TokenCache != nil {

		redisAddr := t.cfg.TokenCacheStorageRedis
		tokenKey := t.cfg.TokenCacheKeyName

		conn, err := redis.Dial("tcp", redisAddr, redis.DialConnectTimeout(time.Second*time.Duration(t.cfg.TokenCacheCtxTimeout)))
		if err != nil {
			return nil, errors.Wrap(err, "error connecting")
		}
		defer conn.Close()

		data, err := redis.String(conn.Do("GET", tokenKey))
		if err != nil {
			// swallowing the error here since we may not have cached a token yet
			return nil, nil
		}
		var token Token
		err = json.Unmarshal([]byte(data), &token)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling data")
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
			return errors.Wrap(err, "Error connecting")
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
