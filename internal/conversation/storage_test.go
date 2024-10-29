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

func TestConversationNilStorable(t *testing.T) {
	var c *Data
	var s *State
	if c.StorageId() != "" {
		t.Errorf("nil Data.StorageId() should return empty string")
	}
	if s.StorageId() != "" {
		t.Errorf("nil State.StorageId() should return empty string")
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
