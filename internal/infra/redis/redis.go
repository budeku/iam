package redis

import (
	"budeku/iam/internal/conf"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const pingTimeout = 5 * time.Second

var ErrNil = errors.New("redis 键不存在")

var (
	mu  sync.RWMutex
	rdb *Redis
)

type Redis struct {
	client *goredis.Client
	prefix string
}

func Init(cfg conf.Redis) error {
	opt := &goredis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	}

	c := goredis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		return fmt.Errorf("redis 连接失败: %w", err)
	}

	mu.Lock()
	rdb = &Redis{client: c, prefix: cfg.Prefix}
	mu.Unlock()

	return nil
}

func DB() *Redis {
	mu.RLock()
	defer mu.RUnlock()
	if rdb == nil {
		log.Fatal("redis 客户端未初始化，请先调用 Init()")
	}
	return rdb
}

func (r *Redis) key(k string) string {
	if r.prefix == "" {
		return k
	}
	return r.prefix + ":" + k
}

func (r *Redis) Key(k string) string {
	return r.key(k)
}

func (r *Redis) Client() *goredis.Client {
	return r.client
}

func Close() error {
	mu.Lock()
	r := rdb
	rdb = nil
	mu.Unlock()

	if r == nil || r.client == nil {
		return nil
	}
	if err := r.client.Close(); err != nil {
		log.Printf("关闭 redis 失败: %v\n", err)
		return err
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	s, err := r.client.Get(ctx, r.key(key)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", ErrNil
	}
	if err != nil {
		return "", err
	}
	return s, nil
}

func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, r.key(key), value, ttl).Err()
}

func (r *Redis) SetForever(ctx context.Context, key string, value any) error {
	return r.client.Set(ctx, r.key(key), value, 0).Err()
}

func (r *Redis) SetNX(ctx context.Context, key string, value any, exp time.Duration) (bool, error) {
	return r.client.SetNX(ctx, r.key(key), value, exp).Result()
}

func (r *Redis) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.key(key)).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, r.key(key)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Redis) SAdd(ctx context.Context, key string, members ...any) error {
	return r.client.SAdd(ctx, r.key(key), members...).Err()
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, r.key(key)).Result()
}

func (r *Redis) SRem(ctx context.Context, key string, members ...any) error {
	return r.client.SRem(ctx, r.key(key), members...).Err()
}

func (r *Redis) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化键 %s 的值失败: %w", key, err)
	}

	return r.Set(ctx, key, b, ttl)
}

func (r *Redis) GetJSON(ctx context.Context, key string, dst any) error {
	s, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), dst); err != nil {
		return fmt.Errorf("反序列化键 %s 的值失败: %w", key, err)
	}
	return nil
}

func (r *Redis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, r.key(key), ttl).Err()
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, r.key(key)).Result()
}

func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, r.key(key), value).Result()
}

func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, r.key(key)).Result()
}

func (r *Redis) BloomAdd(ctx context.Context, key, item string) error {
	return r.client.Do(ctx, "BF.ADD", r.key(key), item).Err()
}

func (r *Redis) BloomExists(ctx context.Context, key, item string) (bool, error) {
	res, err := r.client.Do(ctx, "BF.EXISTS", r.key(key), item).Result()
	if err != nil {
		return false, err
	}
	n, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("布隆过滤器键 %s 的返回结果类型不匹配", key)
	}
	return n == 1, nil
}
