package storage

import (
	"context"
	"github.com/go-test/deep"
	"github.com/google/uuid"
	"testing"
	"time"
)

type OrmTester struct {
	IdField           string    `redis:"id"`
	CreateDate        time.Time `redis:"createDate"`
	CreateDateMillis  int64     `redis:"createDateMillis"`
	CreateDateSeconds float64   `redis:"createDateSeconds"`
	Secret            string    `redis:"secret"`
}

func (data OrmTester) prefix() string {
	return "ormTestPrefix:"
}

func (data OrmTester) id() string {
	return data.IdField
}

func TestNilOrmTester(t *testing.T) {
	var data *OrmTester = nil
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
	data := &OrmTester{IdField: uuid.New().String()}
	if err := LoadFields(context.Background(), data); err == nil {
		t.Errorf("Found stored data for new client %q", data.IdField)
	}
}

func TestSaveLoadDeleteOrmTester(t *testing.T) {
	id := uuid.New().String()
	now := time.Now()
	millis := now.UnixMilli()
	var seconds = float64(now.UnixMicro()) / 1_000_000
	saved := OrmTester{IdField: id, CreateDate: now, CreateDateMillis: millis, CreateDateSeconds: seconds, Secret: "shh!"}
	if err := SaveFields(context.Background(), &saved); err != nil {
		t.Errorf("Failed to save stored data for %q: %v", id, err)
	}
	loaded := OrmTester{IdField: id}
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
	saved := OrmTester{IdField: id, CreateDate: now, CreateDateMillis: millis, CreateDateSeconds: seconds, Secret: id}
	if err := SaveFields(ctx, &saved); err != nil {
		t.Errorf("Failed to save stored data for %q: %v", id, err)
	}
	count := 0
	found := false
	loaded := OrmTester{}
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
