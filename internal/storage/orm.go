/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

type StructPointer interface {
	Storable
	SetStorageId(id string) error
	Copy() StructPointer
	Downgrade(any) (StructPointer, error)
}

func LoadFields[T StructPointer](ctx context.Context, obj T) error {
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

func SaveFields[T StructPointer](ctx context.Context, obj T) error {
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

func MapFields[T StructPointer](ctx context.Context, f func(), obj T) error {
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

type String interface {
	~string
	Storable
}

func FetchString[T String](ctx context.Context, obj T) (string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.Get(ctx, key)
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		} else {
			return "", err
		}
	}
	return res.Val(), nil
}

func StoreString[T String](ctx context.Context, obj T, val string) error {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.Set(ctx, key, val, 0)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

type Set interface {
	~string
	Storable
}

func FetchMembers[T Set](ctx context.Context, obj T) ([]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.SMembers(ctx, key)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func AddMembers[T Set](ctx context.Context, obj T, members ...string) error {
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

func RemoveMembers[T Set](ctx context.Context, obj T, members ...string) error {
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

type SortedSet interface {
	~string
	Storable
}

func FetchRangeInterval[T SortedSet](ctx context.Context, obj T, start, end int64) ([]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.ZRange(ctx, key, start, end)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func FetchRangeScoreInterval[T SortedSet](ctx context.Context, obj T, min, max float64) ([]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	minStr := strconv.FormatFloat(min, 'f', -1, 64)
	maxStr := strconv.FormatFloat(max, 'f', -1, 64)
	res := db.ZRangeByScore(ctx, key, &redis.ZRangeBy{Min: minStr, Max: maxStr})
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func AddScoredMember[T SortedSet](ctx context.Context, obj T, score float64, member string) error {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func RemoveMember[T SortedSet](ctx context.Context, obj T, member string) error {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.ZRem(ctx, key, member)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

type List interface {
	Storable
	~string
}

func FetchRange[T List](ctx context.Context, obj T, start int64, end int64) ([]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.LRange(ctx, key, start, end)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func FetchOneBlocking[T List](ctx context.Context, obj T, timeout time.Duration) (string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.BLMove(ctx, key, key, "right", "left", timeout)
	if err := res.Err(); err != nil {
		return "", err
	}
	return res.Val(), nil
}

func PushRange[T List](ctx context.Context, obj T, onLeft bool, members ...string) error {
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

func RemoveElement[T List](ctx context.Context, obj T, count int64, element string) error {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.LRem(ctx, key, count, any(element))
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

type Map interface {
	Storable
	~string
}

func MapGet[T Map](ctx context.Context, obj T, k string) (string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.HGet(ctx, key, k)
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		} else {
			return "", err
		}
	}
	return res.Val(), nil
}

func MapSet[T Map](ctx context.Context, obj T, k string, v string) error {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.HSet(ctx, key, k, v)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}

func MapGetAll[T Map](ctx context.Context, obj T) (map[string]string, error) {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		return nil, err
	}
	return res.Val(), nil
}

func MapRemove[T Map](ctx context.Context, obj T, k string) error {
	db, prefix := GetDb()
	key := prefix + obj.StoragePrefix() + obj.StorageId()
	res := db.HDel(ctx, key, k)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}
