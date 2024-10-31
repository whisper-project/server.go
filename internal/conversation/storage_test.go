/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package conversation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/google/uuid"

	"clickonetwo.io/whisper/internal/storage"
)

//goland:noinspection SpellCheckingInspection
var (
	knownConversationId   = "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA"
	knownConversationName = "Anyone"
	knownStateId          = "d7dfb2b5-f25a-4de7-8c4a-52af08f1e7f3"
)

func TestConversationStorableInterfaces(t *testing.T) {
	var c *Data = nil
	if c.StoragePrefix() != "con:" {
		t.Errorf("Conversations have a non-'con:' prefix: %s", c.StoragePrefix())
	}
	if c.StorageId() != "" {
		t.Errorf("nil Data.StorageId() should return empty string")
	}
	if err := c.SetStorageId("test"); err == nil {
		t.Errorf("nil Data.SetStorageId() should error out")
	}
	if dup := c.Copy(); dup != nil {
		t.Errorf("nil Data.Copy() should return nil")
	}

	c = &Data{Id: "before"}
	if c.StorageId() != "before" {
		t.Errorf("StorageId is wrong: %s != %s", c.StorageId(), "before")
	}
	if err := c.SetStorageId("after"); err != nil {
		t.Errorf("Failed to set storage id: %v", err)
	}
	if c.StorageId() != "after" {
		t.Errorf("StorageId is wrong: %s != %s", c.StorageId(), "after")
	}
	dup := c.Copy()
	if diff := deep.Equal(dup, c); diff != nil {
		t.Error(diff)
	}
	if dg, err := c.Downgrade(any(c)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, c); diff != nil {
		t.Error(diff)
	}
	if dg, err := c.Downgrade(any(*c)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, c); diff != nil {
		t.Error(diff)
	}
	if _, err := (*c).Downgrade(any(nil)); err == nil {
		t.Errorf("Data.Downgrade(nil) should error out")
	}
}

func TestConversationJsonMarshaling(t *testing.T) {
	c1 := Data{Id: knownConversationId}
	if err := storage.LoadFields(context.Background(), &c1); err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(c1)
	if err != nil {
		t.Fatal(err)
	}
	var c2 Data
	if err := json.Unmarshal(bytes, &c2); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(c1, c2); diff != nil {
		t.Error(diff)
	}
}

func TestTransferConversationData(t *testing.T) {
	c1 := Data{Id: knownConversationId}
	if err := storage.LoadFields(context.Background(), &c1); err != nil {
		t.Fatal(err)
	}
	if c1.Name != knownConversationName {
		t.Errorf("c1.Name (%s) != knownConversationName (%s)", c1.Name, knownConversationName)
	}
	c2 := c1
	if id, err := uuid.NewRandom(); err != nil {
		t.Fatal(err)
	} else {
		c2.Id = strings.ToUpper(id.String())
	}
	if err := storage.SaveFields(context.Background(), &c2); err != nil {
		t.Fatal(err)
	}
	c3 := Data{Id: c2.Id}
	if err := storage.LoadFields(context.Background(), &c3); err != nil {
		t.Fatal(err)
	}
	c3.Id = c1.Id
	if diff := deep.Equal(c1, c3); diff != nil {
		t.Error(diff)
	}
	if err := storage.DeleteStorage(context.Background(), &c2); err != nil {
		t.Fatalf("Failed to delete transfered conversation")
	}
}

func TestStateStorableInterfaces(t *testing.T) {
	var s *State = nil
	if s.StoragePrefix() != "tra:" {
		t.Errorf("States have a non-'tra:' prefix: %s", s.StoragePrefix())
	}
	if s.StorageId() != "" {
		t.Errorf("nil State.StorageId() should return empty string")
	}
	if err := s.SetStorageId("test"); err == nil {
		t.Errorf("nil State.SetStorageId() should error out")
	}
	if dup := s.Copy(); dup != nil {
		t.Errorf("nil State.Copy() should return nil")
	}

	s = &State{Id: "before"}
	if s.StorageId() != "before" {
		t.Errorf("StorageId is wrong: %s != %s", s.StorageId(), "before")
	}
	if err := s.SetStorageId("after"); err != nil {
		t.Errorf("Failed to set storage id: %v", err)
	}
	if s.StorageId() != "after" {
		t.Errorf("StorageId is wrong: %s != %s", s.StorageId(), "after")
	}
	dup := s.Copy()
	if diff := deep.Equal(dup, s); diff != nil {
		t.Error(diff)
	}
	if dg, err := s.Downgrade(any(s)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, s); diff != nil {
		t.Error(diff)
	}
	if dg, err := s.Downgrade(any(*s)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, s); diff != nil {
		t.Error(diff)
	}
	if _, err := (*s).Downgrade(any(nil)); err == nil {
		t.Errorf("State.Downgrade(nil) should error out")
	}
}

func TestStateJsonMarshaling(t *testing.T) {
	s1 := State{Id: knownStateId}
	if err := storage.LoadFields(context.Background(), &s1); err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(s1)
	if err != nil {
		t.Fatal(err)
	}
	var s2 State
	if err := json.Unmarshal(bytes, &s2); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(s1, s2); diff != nil {
		t.Error(diff)
	}
}

func TestTransferStateData(t *testing.T) {
	s1 := State{Id: knownStateId}
	if err := storage.LoadFields(context.Background(), &s1); err != nil {
		t.Fatal(err)
	}
	if s1.ConversationId != knownConversationId {
		t.Errorf("s1.ConversationId (%s) != knownConversationId (%s)", s1.ConversationId, knownConversationId)
	}
	s2 := s1
	if id, err := uuid.NewRandom(); err != nil {
		t.Fatal(err)
	} else {
		s2.Id = strings.ToUpper(id.String())
	}
	if err := storage.SaveFields(context.Background(), &s2); err != nil {
		t.Fatal(err)
	}
	s3 := State{Id: s2.Id}
	if err := storage.LoadFields(context.Background(), &s3); err != nil {
		t.Fatal(err)
	}
	s3.Id = s1.Id
	if diff := deep.Equal(s1, s3); diff != nil {
		t.Error(diff)
	}
	if err := storage.DeleteStorage(context.Background(), &s2); err != nil {
		t.Fatalf("Failed to delete transfered state")
	}
}
