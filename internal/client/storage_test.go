/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package client

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
	knownClientId       = "561E5E8E-EA35-405A-A256-69C74713FAFD"
	knownClientUserName = "Dan Brotsky"
)

func TestClientJsonMarshaling(t *testing.T) {
	c1 := Data{Id: knownClientId}
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

func TestTransferClientData(t *testing.T) {
	c1 := Data{Id: knownClientId}
	if err := storage.LoadFields(context.Background(), &c1); err != nil {
		t.Fatal(err)
	}
	if c1.UserName != knownClientUserName {
		t.Errorf("c1.UserName (%s) != knownClientUserName (%s)", c1.UserName, knownClientUserName)
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
		t.Fatalf("Failed to delete transfered client")
	}
}
