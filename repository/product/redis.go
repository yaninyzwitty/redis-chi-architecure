package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yaninyzwitty/merch-crud-microservice-go/model"
)

type RedisRepo struct {
	Client *redis.Client
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}
type FindResult struct {
	products []model.Product
	Cursor   uint64
}

var ErrorNotExists = errors.New("product does not exist")

func productIdKey(productId uuid.UUID) string {
	return fmt.Sprintf("product:%s", productId)
}

func (r *RedisRepo) Insert(ctx context.Context, product model.Product) error {
	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to encode product: %v", err)
	}

	key := productIdKey(product.ProductId)

	res := r.Client.SetNX(ctx, key, data, 0)
	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to set product: %v", err)

	}

	return nil
}

func (r *RedisRepo) GetById(ctx context.Context, productId uuid.UUID) (model.Product, error) {

	key := productIdKey(productId)
	val, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Product{}, ErrorNotExists
	} else if err != nil {
		return model.Product{}, fmt.Errorf("get product: %w", err)

	}

	// convert val (json data) to of model type

	var product model.Product
	err = json.Unmarshal([]byte(val), &product)
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to decode product: %w", err)
	}
	return product, nil

}

func (r *RedisRepo) Delete(ctx context.Context, productId uuid.UUID) error {
	key := productIdKey(productId)

	err := r.Client.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		return ErrorNotExists

	} else if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil

}

func (r *RedisRepo) UpdateById(ctx context.Context, product model.Product) error {
	key := productIdKey(product.ProductId)
	data, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to encode product: %v", err)

	}

	err = r.Client.SetXX(ctx, key, string(data), 0).Err() //only if the key already exists
	if errors.Is(err, redis.Nil) {
		return ErrorNotExists
	} else if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil

}

func (r *RedisRepo) GetAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "products", page.Offset, "*", int64(page.Size))
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get product ids: %w", err)

	}

	if len(keys) == 0 {
		return FindResult{}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get products: %w", err)
	}
	products := make([]model.Product, len(xs))
	for i, x := range xs {
		x := x.(string)
		var product model.Product
		err := json.Unmarshal([]byte(x), &product)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode product: %w", err)
		}

		products[i] = product

	}
	return FindResult{products: products, Cursor: cursor}, nil

}
