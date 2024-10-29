/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storable interface {
	StoragePrefix() string
	StorageId() string
}

func DeleteStorage[T Storable](ctx context.Context, obj T) error {
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.Del(ctx, key)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

type StorableStruct interface {
	Storable
	SetStorageId(id string) error
	Copy() any
}

type StorableStructDowngrader interface {
	Downgrade(any) (StorableStruct, error)
}

func LoadFields[T StorableStruct](ctx context.Context, obj T) error {
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to fetch fields of stored object %s: %v", key, err)
	}
	if len(res.Val()) == 0 {
		return fmt.Errorf("stored object %s has no fields", key)
	}
	if err := res.Scan(obj); err != nil {
		return fmt.Errorf("stored object %s cannot be read: %v", key, err)
	}
	return nil
}

func SaveFields[T StorableStruct](ctx context.Context, obj T) error {
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.HSet(ctx, key, obj)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func MapFields[T StorableStruct](ctx context.Context, f func(), obj T) error {
	if err := obj.SetStorageId(""); err != nil {
		return fmt.Errorf("storable ID cannot be set")
	}
	db, prefix := GetDb()
	iter := db.Scan(ctx, 0, prefix+obj.StoragePrefix()+"*", 20).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		res := db.HGetAll(ctx, key)
		if err := res.Err(); err != nil {
			return fmt.Errorf("failed to fetch fields of stored object %s: %v", key, err)
		}
		if err := res.Scan(obj); err != nil {
			return fmt.Errorf("stored object %s cannot be read: %v", key, err)
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
	if obj.StorageId() == "" {
		return nil, fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.SMembers(ctx, key)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func AddMembers[T StorableSet](ctx context.Context, obj T, members ...string) error {
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	if len(members) == 0 {
		// nothing to add
		return nil
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
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
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	if len(members) == 0 {
		// nothing to delete
		return nil
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
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
	if obj.StorageId() == "" {
		return nil, fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.LRange(ctx, key, start, end)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func FetchOneBlocking[T StorableList](ctx context.Context, obj T, timeout time.Duration) (string, error) {
	if obj.StorageId() == "" {
		return "", fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.BLMove(ctx, key, key, "right", "left", timeout)
	if err := res.Err(); err != nil {
		return "", err
	}
	return res.Val(), nil
}

func PushRange[T StorableList](ctx context.Context, obj T, onLeft bool, members ...string) error {
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
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
	if obj.StorageId() == "" {
		return fmt.Errorf("storable has no ID")
	}
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.LRem(ctx, key, count, any(element))
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}
