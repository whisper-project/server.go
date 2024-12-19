/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-test/deep"

	"github.com/whisper-project/server.go/internal/internaltest"
	"github.com/whisper-project/server.go/internal/storage"
)

func TestUserProfileStorableInterfaces(t *testing.T) {
	var p *UserProfile = nil
	if p.StoragePrefix() != "pro:" {
		t.Errorf("UserProfiles have a non-'pro:' prefix: %s", p.StoragePrefix())
	}
	if p.StorageId() != "" {
		t.Errorf("nil UserProfile.StorageId() should return empty string")
	}
	if err := p.SetStorageId("test"); err == nil {
		t.Errorf("nil UserProfile.SetStorageId() should error out")
	}
	if dup := p.Copy(); dup != nil {
		t.Errorf("nil UserProfile.Copy() should return nil")
	}

	p = &UserProfile{Id: "before"}
	if p.StorageId() != "before" {
		t.Errorf("StorageId is wrong: %s != %s", p.StorageId(), "before")
	}
	if err := p.SetStorageId("after"); err != nil {
		t.Errorf("Failed to set storage id: %v", err)
	}
	if p.StorageId() != "after" {
		t.Errorf("StorageId is wrong: %s != %s", p.StorageId(), "after")
	}
	dup := p.Copy()
	if diff := deep.Equal(dup, p); diff != nil {
		t.Error(diff)
	}
	if dg, err := p.Downgrade(any(p)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, p); diff != nil {
		t.Error(diff)
	}
	if dg, err := p.Downgrade(any(*p)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, p); diff != nil {
		t.Error(diff)
	}
	if _, err := (*p).Downgrade(any(nil)); err == nil {
		t.Errorf("UserProfile.Downgrade(nil) should error out")
	}
}

func TestWhisperProfileJsonMarshaling(t *testing.T) {
	p1 := UserProfile{Id: internaltest.KnownUserId}
	if err := storage.LoadFields(context.Background(), &p1); err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(p1.WhisperProfile)
	if err != nil {
		t.Fatal(err)
	}
	var p2 WhisperProfile
	if err := json.Unmarshal(bytes, &p2); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(p1.WhisperProfile, p2); diff != nil {
		t.Error(diff)
	}
}

func TestUserProfileJsonMarshaling(t *testing.T) {
	p1 := UserProfile{Id: internaltest.KnownUserId}
	if err := storage.LoadFields(context.Background(), &p1); err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(p1)
	if err != nil {
		t.Fatal(err)
	}
	var p2 UserProfile
	if err := json.Unmarshal(bytes, &p2); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(p1, p2); diff != nil {
		t.Error(diff)
	}
}

func TestTransferProfileData(t *testing.T) {
	p1 := UserProfile{Id: internaltest.KnownUserId}
	if err := storage.LoadFields(context.Background(), &p1); err != nil {
		t.Fatal(err)
	}
	if p1.Name != internaltest.KnownUserName {
		t.Errorf("p1.Name (%s) != internaltest.KnownUserName (%s)", p1.Name, internaltest.KnownUserName)
	}
	p2 := p1
	p2.Id = internaltest.NewTestId()
	if err := storage.SaveFields(context.Background(), &p2); err != nil {
		t.Fatal(err)
	}
	p3 := UserProfile{Id: p2.Id}
	if err := storage.LoadFields(context.Background(), &p3); err != nil {
		t.Fatal(err)
	}
	p3.Id = p1.Id
	if diff := deep.Equal(p1, p3); diff != nil {
		t.Error(diff)
	}
	if err := storage.DeleteStorage(context.Background(), &p2); err != nil {
		t.Fatalf("Failed to delete transfered profile")
	}
}
