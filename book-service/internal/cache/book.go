package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awiipp/go-library/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	bookKeyPrefix = "book"
	bookTTL       = 10 * time.Minute
)

type BookCache struct {
	client *redis.Client
}

func NewBookCache(client *redis.Client) *BookCache {
	return &BookCache{client: client}
}

func (c *BookCache) Get(ctx context.Context, id string) (*domain.Book, error) {
	key := c.key(id)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	book := &domain.Book{}
	if err := json.Unmarshal([]byte(data), book); err != nil {
		return nil, err
	}

	return book, nil
}

func (c *BookCache) Set(ctx context.Context, book *domain.Book) error {
	key := c.key(book.ID)

	data, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, bookTTL).Err()
}

func (c *BookCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, c.key(id)).Err()
}

func (c *BookCache) key(id string) string {
	return fmt.Sprintf("%s:%s", bookKeyPrefix, id)
}
