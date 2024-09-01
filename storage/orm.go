package storage

import (
	"context"
	"fmt"
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
	iter := db.Scan(ctx, 0, prefix+(*obj).prefix()+"*", 0).Iterator()
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
	Storable
	~string
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

func AddMembers[T StorableSet](ctx context.Context, obj T, members ...interface{}) error {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	res := db.SAdd(ctx, key, members...)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func RemoveMembers[T StorableSet](ctx context.Context, obj T, members ...interface{}) error {
	db, prefix := GetDb()
	key := prefix + obj.prefix() + obj.id()
	res := db.SRem(ctx, key, members...)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}
