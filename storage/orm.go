package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storable interface {
	prefix() string
	id() string
}

func DeleteStorage[T Storable](ctx context.Context, obj *T) error {
	if obj == nil {
		return fmt.Errorf("Storable pointer is nil")
	}
	db, prefix := GetDb()
	key := prefix + (*obj).prefix() + (*obj).id()
	res := db.Del(ctx, key)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

type StorableStruct interface {
	Storable
}

func LoadFields[T StorableStruct](ctx context.Context, obj *T) error {
	if obj == nil {
		return fmt.Errorf("StorableStruct pointer is nil")
	}
	db, prefix := GetDb()
	key := prefix + (*obj).prefix() + (*obj).id()
	res := db.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		return err
	}
	if len(res.Val()) == 0 {
		return fmt.Errorf("stored object %s has no fields", key)
	}
	if err := res.Scan(obj); err != nil {
		return err
	}
	return nil
}

func SaveFields[T StorableStruct](ctx context.Context, obj *T) error {
	if obj == nil {
		return fmt.Errorf("StorableStruct pointer is nil")
	}
	db, prefix := GetDb()
	key := prefix + (*obj).prefix() + (*obj).id()
	res := db.HSet(ctx, key, obj)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func MapFields[T StorableStruct](ctx context.Context, f func(), obj *T) error {
	if obj == nil {
		return fmt.Errorf("StorableStruct pointer is nil")
	}
	db, prefix := GetDb()
	iter := db.Scan(ctx, 0, prefix+(*obj).prefix()+"*", 20).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		res := db.HGetAll(ctx, key)
		if err := res.Err(); err != nil {
			return err
		}
		if err := res.Scan(obj); err != nil {
			return err
		}
		f()
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}

type StorableSet interface {
	~string
	Storable
}

func FetchMembers[T StorableSet](ctx context.Context, obj T) ([]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	res := db.SMembers(ctx, key)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func AddMembers[T StorableSet](ctx context.Context, obj T, members ...string) error {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	args := make([]interface{}, len(members))
	for i, member := range members {
		args[i] = any(member)
	}
	res := db.SAdd(ctx, key, args...)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func RemoveMembers[T StorableSet](ctx context.Context, obj T, members ...string) error {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	args := make([]interface{}, len(members))
	for i, member := range members {
		args[i] = any(member)
	}
	res := db.SRem(ctx, key, args...)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

type StorableList interface {
	Storable
	~string
}

func FetchRange[T StorableList](ctx context.Context, obj T, start int64, end int64) ([]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	res := db.LRange(ctx, key, start, end)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func FetchOneBlocking[T StorableList](ctx context.Context, obj T, timeout time.Duration) (string, error) {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	res := db.BLMove(ctx, key, key, "right", "left", timeout)
	if err := res.Err(); err != nil {
		return "", err
	}
	return res.Val(), nil
}

func PushRange[T StorableList](ctx context.Context, obj T, onLeft bool, members ...string) error {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	args := make([]interface{}, len(members))
	for i, member := range members {
		args[i] = any(member)
	}
	var res *redis.IntCmd
	if onLeft {
		res = db.LPush(ctx, key, args...)
	} else {
		res = db.RPush(ctx, key, args...)
	}
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func RemoveElement[T StorableList](ctx context.Context, obj T, count int64, element string) error {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	res := db.LRem(ctx, key, count, any(element))
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}
