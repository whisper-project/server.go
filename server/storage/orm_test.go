/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"context"
	"fmt"
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

func (data *OrmTestStruct) StoragePrefix() string {
	return "ormTestPrefix:"
}

func (data *OrmTestStruct) StorageId() string {
	if data == nil {
		return ""
	}
	return data.IdField
}

func (data *OrmTestStruct) SetStorageId(id string) error {
	if data == nil {
		return fmt.Errorf("can't set storage id of nil struct")
	}
	data.IdField = id
	return nil
}

func (data *OrmTestStruct) Copy() StructPointer {
	if data == nil {
		return nil
	}
	n := new(OrmTestStruct)
	*n = *data
	return n
}

func (data *OrmTestStruct) Downgrade(in any) (StructPointer, error) {
	if o, ok := in.(OrmTestStruct); ok {
		return &o, nil
	}
	if o, ok := in.(*OrmTestStruct); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not an OrmTestStruct: %#v", in)
}

func TestNilOrmTester(t *testing.T) {
	var data *OrmTestStruct = nil
	if err := LoadFields(context.Background(), data); err == nil {
		t.Errorf("LoadFields on nil pointer didn't fail!")
	}
	if err := SaveFields(context.Background(), data); err == nil {
		t.Errorf("SaveFields on nil pointer didn't fail!")
	}
	if err := MapFields(context.Background(), func() {}, data); err == nil {
		t.Errorf("MapFields on nil pointer didn't fail!")
	}
	if err := DeleteStorage(context.Background(), data); err == nil {
		t.Errorf("DeleteStorage on nil pointer didn't fail!")
	}
}

func TestLoadMissingOrmTester(t *testing.T) {
	data := &OrmTestStruct{IdField: uuid.New().String()}
	if err := LoadFields(context.Background(), data); err == nil {
		t.Errorf("Found stored data for new test object %q", data.IdField)
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
	var loaded OrmTestStruct
	_ = loaded.SetStorageId(id)
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
		t.Fatalf("Failed to map stored data in pass 1: %v", err)
	}
	if !found {
		t.Errorf("Mapped over %#v objects; never found one with secret %q", count, id)
	}
	count = 0
	if err := MapFields(ctx, mapper, &loaded); err != nil {
		t.Errorf("Failed to map stored data in pass 2: %v", err)
	}
	if count != 0 {
		t.Fatalf("Mapped over %#v objects; wanted %#v", count, 0)
	}
}

func TestCopyDowngradeOrmTester(t *testing.T) {
	var orig *OrmTestStruct = nil
	//goland:noinspection GoDfaNilDereference
	copiedOrig := orig.Copy()
	if copiedOrig != nil {
		t.Errorf("copy of nil OrmTestStruct pointer wasn't nil!")
	}
	orig = &OrmTestStruct{IdField: uuid.New().String(), Secret: "shh!"}
	copiedOrig = orig.Copy()
	c, ok := copiedOrig.(*OrmTestStruct)
	if !ok {
		t.Errorf("copy of OrmTestStruct was not a OrmTestStruct: %#v", copiedOrig)
	}
	if diff := deep.Equal(c, copiedOrig); diff != nil {
		t.Errorf("copy of OrmTestStruct differs: %v", diff)
	}
	if c == orig {
		t.Errorf("copy of OrmTestStruct has same address as original: %p, %p", c, orig)
	}
	var template OrmTestStruct
	downgradedCopy, err := template.Downgrade(*c)
	if err != nil {
		t.Errorf("Failed to downgrade OrmTestStruct: %v", err)
	}
	if _, ok := downgradedCopy.(*OrmTestStruct); !ok {
		t.Errorf("Downgraded copy has wrong type: %T", downgradedCopy)
	}
	if diff := deep.Equal(downgradedCopy, orig); diff != nil {
		t.Errorf("Downgraded copy differs from original: %v", diff)
	}
}

type OrmTestString string

func (s OrmTestString) StoragePrefix() string {
	return "ormTestString:"
}

func (s OrmTestString) StorageId() string {
	return string(s)
}

func TestFetchSetFetchString(t *testing.T) {
	ctx := context.Background()
	id := OrmTestString(uuid.New().String())
	if val, err := FetchString(ctx, id); err != nil || val != "" {
		t.Errorf("FetchString of missing string failed (%v), expected success with empty value (%s)", err, val)
	}
	if err := StoreString(ctx, id, string(id)); err != nil {
		t.Error(err)
	}
	if val, err := FetchString(ctx, id); err != nil || val != string(id) {
		t.Errorf("FetchString of failed (%v), expected %q got %q", err, string(id), val)
	}
	if err := DeleteStorage(ctx, id); err != nil {
		t.Error(err)
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

type OrmTestSortedSet string

func (s OrmTestSortedSet) StoragePrefix() string {
	return "ormTestSortedSet:"
}

func (s OrmTestSortedSet) StorageId() string {
	return string(s)
}

func TestSortedFetchAddFetchRemoveMembers(t *testing.T) {
	ctx := context.Background()
	id := OrmTestSortedSet(uuid.New().String())
	sorted := []string{"a", "b", "c"}
	if members, err := FetchRangeInterval(ctx, id, 0, -1); err != nil || len(members) != 0 {
		t.Errorf("FetchRangeInterval of empty failed (%v) or has members: %v", err, members)
	}
	if err := AddScoredMember(ctx, id, 3, "c"); err != nil {
		t.Error(err)
	}
	if err := AddScoredMember(ctx, id, 2, "b"); err != nil {
		t.Error(err)
	}
	if err := AddScoredMember(ctx, id, 1, "a"); err != nil {
		t.Error(err)
	}
	if found, err := FetchRangeInterval(ctx, id, 0, -1); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(sorted, found); diff != nil {
		t.Error(diff)
	}
	if found, err := FetchRangeScoreInterval(ctx, id, 2, 3); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(sorted[1:3], found); diff != nil {
		t.Error(diff)
	}
	if err := RemoveMember(ctx, id, "a"); err != nil {
		t.Error(err)
	}
	if found, err := FetchRangeInterval(ctx, id, 0, -1); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(sorted[1:3], found); diff != nil {
		t.Error(diff)
	}
	if err := DeleteStorage(ctx, &id); err != nil {
		t.Error(err)
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
