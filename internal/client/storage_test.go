/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package client

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-test/deep"

	"clickonetwo.io/whisper/internal/internaltest"
	"clickonetwo.io/whisper/internal/storage"
)

func TestClientStorableInterfaces(t *testing.T) {
	var c *Data = nil
	if c.StoragePrefix() != "cli:" {
		t.Errorf("Clients have a non-'cli:' prefix: %s", c.StoragePrefix())
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

func TestClientJsonMarshaling(t *testing.T) {
	c1 := Data{Id: internaltest.KnownClientId}
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
	c1 := Data{Id: internaltest.KnownClientId}
	if err := storage.LoadFields(context.Background(), &c1); err != nil {
		t.Fatal(err)
	}
	if c1.UserName != internaltest.KnownClientUserName {
		t.Errorf("c1.UserName (%s) != internaltest.KnownClientUserName (%s)", c1.UserName, internaltest.KnownClientUserName)
	}
	c2 := c1
	c2.Id = internaltest.NewTestId()
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
