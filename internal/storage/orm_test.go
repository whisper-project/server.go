/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/google/uuid"
)

type OrmTestStruct struct {
	IdField           string    `redis:"id"`
	CreateDate        time.Time `redis:"createDate"`
	CreateDateMillis  int64     `redis:"createDateMillis"`
	CreateDateSeconds float64   `redis:"createDateSeconds"`
	Secret            string    `redis:"secret"`
}

func (data OrmTestStruct) StoragePrefix() string {
	return "ormTestPrefix:"
}

func (data OrmTestStruct) StorageId() string {
	return data.IdField
}

func TestNilOrmTester(t *testing.T) {
	var data *OrmTestStruct = nil
	if err := LoadFields(context.Background(), data); err == nil {
		t.Errorf("LoadFields on nil pointer didn't fail!")
	}
	if err := SaveFields(context.Background(), data); err == nil {
		t.Errorf("SaveFields on nil pointer didn't fail!")
	}
	if err := DeleteStorage(context.Background(), data); err == nil {
		t.Errorf("DeleteStorage on nil pointer didn't fail!")
	}
	if err := MapFields(context.Background(), func() {}, data); err == nil {
		t.Errorf("MapFields on nil pointer didn't fail!")
	}
}

func TestLoadMissingOrmTester(t *testing.T) {
	data := &OrmTestStruct{IdField: uuid.New().String()}
	if err := LoadFields(context.Background(), data); err == nil {
		t.Errorf("Found stored data for new client %q", data.IdField)
	}
}

func TestSaveLoadDeleteOrmTester(t *testing.T) {
	id := uuid.New().String()
	now := time.Now()
	millis := now.UnixMilli()
	seconds := float64(now.UnixMicro()) / 1_000_000
	saved := OrmTestStruct{IdField: id, CreateDate: now, CreateDateMillis: millis, CreateDateSeconds: seconds, Secret: "shh!"}
	if err := SaveFields(context.Background(), &saved); err != nil {
		t.Errorf("Failed to save stored data for %q: %v", id, err)
	}
	loaded := OrmTestStruct{IdField: id}
	if err := LoadFields(context.Background(), &loaded); err != nil {
		t.Errorf("Failed to load stored data for %q: %v", id, err)
	}
	if diff := deep.Equal(saved, loaded); diff != nil {
		t.Errorf("LoadSave data differs: %v", diff)
	}
	if err := DeleteStorage(context.Background(), &loaded); err != nil {
		t.Errorf("Failed to delete stored data for %q: %v", id, err)
	}
	if err := LoadFields(context.Background(), &loaded); err == nil {
		t.Errorf("Succeeded in loading deleted data for %q: %v", id, err)
	}
	if diff := deep.Equal(saved, loaded); diff != nil {
		t.Errorf("Failed load altered fields: %v", diff)
	}
}

func TestSaveMapDeleteOrmTester(t *testing.T) {
	ctx := context.Background()
	id := uuid.New().String()
	now := time.Now()
	millis := now.UnixMilli()
	seconds := float64(now.UnixMicro()) / 1_000_000
	saved := OrmTestStruct{IdField: id, CreateDate: now, CreateDateMillis: millis, CreateDateSeconds: seconds, Secret: id}
	if err := SaveFields(ctx, &saved); err != nil {
		t.Errorf("Failed to save stored data for %q: %v", id, err)
	}
	count := 0
	found := false
	loaded := OrmTestStruct{}
	mapper := func() {
		count++
		if loaded.Secret == id {
			found = true
		}
		if err := DeleteStorage(ctx, &loaded); err != nil {
			t.Errorf("Failed to delete stored data for %q: %v", id, err)
		}
	}
	if err := MapFields(ctx, mapper, &loaded); err != nil {
		t.Errorf("Failed to map stored data in pass 1: %v", err)
	}
	if !found {
		t.Errorf("Mapped over %#v objects; never found one with secret %q", count, id)
	}
	count = 0
	if err := MapFields(ctx, mapper, &loaded); err != nil {
		t.Errorf("Failed to map stored data in pass 2: %v", err)
	}
	if count != 0 {
		t.Errorf("Mapped over %#v objects; wanted %#v", count, 0)
	}
}

type OrmTestSet string

func (s OrmTestSet) StoragePrefix() string {
	return "ormTestSet:"
}

func (s OrmTestSet) StorageId() string {
	return string(s)
}

func TestFetchNoMembers(t *testing.T) {
	ctx := context.Background()
	id := OrmTestSet(uuid.New().String())
	if members, err := FetchMembers(ctx, id); err != nil || len(members) != 0 {
		t.Errorf("FetchMembers of empty set failed, expected success with no members")
	}
}

func TestAddFetchRemoveMembers(t *testing.T) {
	ctx := context.Background()
	id := OrmTestSet(uuid.New().String())
	saved := []string{"a", "b", "c", "b", "a"}
	if err := AddMembers(ctx, id, saved...); err != nil {
		t.Errorf("Failed to add saved: %v", err)
	}
	if err := AddMembers(ctx, id); err != nil {
		t.Errorf("Failed to add empty: %v", err)
	}
	if found, err := FetchMembers(ctx, id); err != nil {
		t.Errorf("FetchMembers failed: %v", err)
	} else if len(found) != 3 {
		t.Errorf("FetchMembers returned %d results, expected 3: %#v", len(found), found)
	}
	if err := RemoveMembers(ctx, id, "b", "c"); err != nil {
		t.Errorf("Failed to remove members: %v", err)
	}
	if err := RemoveMembers(ctx, id); err != nil {
		t.Errorf("Failed to remove empty: %v", err)
	}
	if found, err := FetchMembers(ctx, id); err != nil {
		t.Errorf("FetchMembers failed: %v", err)
	} else if len(found) != 1 {
		t.Errorf("FetchMembers returned %d results, expected 1: %#v", len(found), found)
	}
	if err := DeleteStorage(ctx, &id); err != nil {
		t.Errorf("Failed to delete stored data for %q: %v", id, err)
	}
}

type OrmTestList string

func (s OrmTestList) StoragePrefix() string {
	return "ormTestList:"
}

func (s OrmTestList) StorageId() string {
	return string(s)
}

func TestFetchEmptyRange(t *testing.T) {
	ctx := context.Background()
	id := OrmTestList(uuid.New().String())
	if elements, err := FetchRange(ctx, id, 0, -1); err != nil || len(elements) != 0 {
		t.Errorf("FetchRange of empty list failed, expected success with no elements")
	}
}

func TestAddFetchRemoveRange(t *testing.T) {
	ctx := context.Background()
	id := OrmTestList(uuid.New().String())
	if err := PushRange(ctx, id, true, "|"); err != nil {
		t.Errorf("Failed to push center: %v", err)
	}
	if err := PushRange(ctx, id, true, "a", "b", "c"); err != nil {
		t.Errorf("Failed to push left: %v", err)
	}
	if err := PushRange(ctx, id, false, "a", "b", "c"); err != nil {
		t.Errorf("Failed to push right: %v", err)
	}
	if before, err := FetchRange(ctx, id, 0, -1); err != nil {
		t.Errorf("FetchRange of before list failed, expected success")
	} else if diff := deep.Equal(before, []string{"c", "b", "a", "|", "a", "b", "c"}); diff != nil {
		t.Errorf("FetchRange of before list is:\n%v\nwith differences:\n%v", before, diff)
	}
	if err := RemoveElement(ctx, id, 0, "b"); err != nil {
		t.Errorf("Failed to remove 'b': %v", err)
	}
	if after, err := FetchRange(ctx, id, 0, -1); err != nil {
		t.Errorf("FetchRange of after list failed, expected success")
	} else if diff := deep.Equal(after, []string{"c", "a", "|", "a", "c"}); diff != nil {
		t.Errorf("FetchRange of after list is:\n%v\nwith differences:\n%v", after, diff)
	}
	if err := DeleteStorage(ctx, &id); err != nil {
		t.Errorf("Failed to delete stored data for %q: %v", id, err)
	}
}

func TestFetchOneBlocking(t *testing.T) {
	ctx := context.Background()
	id := OrmTestList(uuid.New().String())
	defer func() {
		if err := DeleteStorage(ctx, &id); err != nil {
			t.Errorf("Failed to delete stored data for %q: %v", id, err)
		}
	}()
	c := make(chan string)
	go func() {
		if element, err := FetchOneBlocking(ctx, id, 2*time.Second); err != nil {
			t.Errorf("FetchOneBlocking failed: %v", err)
			c <- "failed"
		} else {
			c <- element
		}
	}()
	time.Sleep(500 * time.Millisecond)
	if err := PushRange(ctx, id, false, "a", "b", "c"); err != nil {
		t.Fatalf("Failed to push right: %v", err)
	}
	received := <-c
	if received != "c" {
		t.Errorf("FetchOneBlocking got %q", received)
	}
	if remaining, err := FetchRange(ctx, id, 0, -1); err != nil {
		t.Errorf("FetchRange of remaining list failed, expected success")
	} else if diff := deep.Equal(remaining, []string{"c", "a", "b"}); diff != nil {
		t.Errorf("FetchRange of remaining list is:\n%v\ndifferences are:\n%v", remaining, diff)
	}
}
