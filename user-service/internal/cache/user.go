package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awiipp/go-library/user-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	profileKeyPrefix = "profile"
	userTTL          = 10 * time.Minute
)

type ProfileCache struct {
	client *redis.Client
}

func NewProfileCache(client *redis.Client) *ProfileCache {
	return &ProfileCache{client: client}
}

func (c *ProfileCache) Get(ctx context.Context, id string) (*domain.User, error) {
	key := c.profileKey(id)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	profile := &domain.User{}
	if err := json.Unmarshal([]byte(data), profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (c *ProfileCache) Set(ctx context.Context, user *domain.User) error {
	key := c.profileKey(user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, userTTL).Err()
}

func (c *ProfileCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, c.profileKey(id)).Err()
}

func (c *ProfileCache) profileKey(id string) string {
	return fmt.Sprintf("%s:%s", profileKeyPrefix, id)
}
