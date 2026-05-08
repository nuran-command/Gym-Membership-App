package redis

import (
	"context"
	"errors"
)

type creditCache struct{}

func NewCreditCache() *creditCache {
	return &creditCache{}
}

func (c *creditCache) Get(ctx context.Context, key string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
