package database

import "context"

type Databse interface {
	Put(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Delete(ctx context.Context, key string) (bool, error)
	All(ctx context.Context) (map[string][]byte, error)
}
